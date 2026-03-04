//go:build integration
// +build integration

package pbf

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
	"sort"
	"testing"
	"time"

	"m4o.io/pbf/v2/model"
)

type entityDigest struct {
	total     int
	nodes     int
	ways      int
	relations int
	sum       [sha256.Size]byte
	nodeSum   [sha256.Size]byte
	waySum    [sha256.Size]byte
	relSum    [sha256.Size]byte
}

type multisetHasher struct {
	sum [4]uint64
	xor [4]uint64
}

func TestRoundTripOSMPBFDatasets(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		path              string
		compareNodeDigest bool
	}{
		{name: "sample", path: "testdata/sample.osm.pbf", compareNodeDigest: true},
		{name: "bremen", path: "testdata/bremen.osm.pbf", compareNodeDigest: true},
		{name: "london", path: "testdata/greater-london.osm.pbf", compareNodeDigest: true},
		{name: "san-francisco", path: "testdata/san-francisco.osm.pbf", compareNodeDigest: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input, err := os.Open(tc.path)
			if err != nil {
				t.Fatalf("open %s: %v", tc.path, err)
			}
			defer input.Close()

			decoded, err := NewDecoder(context.Background(), input)
			if err != nil {
				t.Fatalf("create source decoder: %v", err)
			}
			defer decoded.Close()

			var encoded bytes.Buffer
			reencoded, err := NewEncoder(&encoded)
			if err != nil {
				t.Fatalf("create encoder: %v", err)
			}

			sourceDigest, err := digestAndEncode(decoded, reencoded)
			if err != nil {
				t.Fatalf("digest and encode source %s: %v", tc.path, err)
			}

			reencoded.Close()

			roundTrip, err := NewDecoder(context.Background(), bytes.NewReader(encoded.Bytes()))
			if err != nil {
				t.Fatalf("create round-trip decoder: %v", err)
			}
			defer roundTrip.Close()

			roundTripDigest, err := digestDecoder(roundTrip)
			if err != nil {
				t.Fatalf("digest round-trip %s: %v", tc.path, err)
			}

			if sourceDigest.total != roundTripDigest.total ||
				sourceDigest.nodes != roundTripDigest.nodes ||
				sourceDigest.ways != roundTripDigest.ways ||
				sourceDigest.relations != roundTripDigest.relations ||
				sourceDigest.waySum != roundTripDigest.waySum ||
				sourceDigest.relSum != roundTripDigest.relSum {
				t.Fatalf("source and round-trip digests differ\nsource: %+v\nround-trip: %+v", sourceDigest, roundTripDigest)
			}
			if tc.compareNodeDigest && sourceDigest.nodeSum != roundTripDigest.nodeSum {
				t.Fatalf("source and round-trip node digests differ\nsource: %+v\nround-trip: %+v", sourceDigest, roundTripDigest)
			}
		})
	}
}

func digestAndEncode(dec *Decoder, enc *Encoder) (entityDigest, error) {
	hasher := multisetHasher{}
	nodeHasher := multisetHasher{}
	wayHasher := multisetHasher{}
	relHasher := multisetHasher{}
	stats := entityDigest{}

	for {
		entities, err := dec.Decode()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return entityDigest{}, fmt.Errorf("decode source: %w", err)
		}

		if err := digestEntities(&hasher, &nodeHasher, &wayHasher, &relHasher, &stats, entities); err != nil {
			return entityDigest{}, err
		}

		if err := enc.EncodeBatch(entities); err != nil {
			return entityDigest{}, fmt.Errorf("encode batch: %w", err)
		}
	}

	stats.sum = hasher.Digest()
	stats.nodeSum = nodeHasher.Digest()
	stats.waySum = wayHasher.Digest()
	stats.relSum = relHasher.Digest()

	return stats, nil
}

func digestDecoder(dec *Decoder) (entityDigest, error) {
	hasher := multisetHasher{}
	nodeHasher := multisetHasher{}
	wayHasher := multisetHasher{}
	relHasher := multisetHasher{}
	stats := entityDigest{}

	for {
		entities, err := dec.Decode()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return entityDigest{}, fmt.Errorf("decode: %w", err)
		}

		if err := digestEntities(&hasher, &nodeHasher, &wayHasher, &relHasher, &stats, entities); err != nil {
			return entityDigest{}, err
		}
	}

	stats.sum = hasher.Digest()
	stats.nodeSum = nodeHasher.Digest()
	stats.waySum = wayHasher.Digest()
	stats.relSum = relHasher.Digest()

	return stats, nil
}

func digestEntities(hasher, nodeHasher, wayHasher, relHasher *multisetHasher, stats *entityDigest, entities []model.Entity) error {
	for _, entity := range entities {
		stats.total++
		switch typed := entity.(type) {
		case model.Node:
			stats.nodes++
			hasher.AddNode(typed)
			nodeHasher.AddNode(typed)
		case *model.Node:
			stats.nodes++
			hasher.AddNode(*typed)
			nodeHasher.AddNode(*typed)
		case model.Way:
			stats.ways++
			hasher.AddWay(typed)
			wayHasher.AddWay(typed)
		case *model.Way:
			stats.ways++
			hasher.AddWay(*typed)
			wayHasher.AddWay(*typed)
		case model.Relation:
			stats.relations++
			hasher.AddRelation(typed)
			relHasher.AddRelation(typed)
		case *model.Relation:
			stats.relations++
			hasher.AddRelation(*typed)
			relHasher.AddRelation(*typed)
		default:
			return fmt.Errorf("unknown entity type %T", entity)
		}
	}

	return nil
}

func (h *multisetHasher) AddNode(n model.Node) {
	entityHash := sha256.New()
	writeByte(entityHash, 1)
	writeNode(entityHash, n)
	h.addEntityHash(entityHash.Sum(nil))
}

func (h *multisetHasher) AddWay(w model.Way) {
	entityHash := sha256.New()
	writeByte(entityHash, 2)
	writeWay(entityHash, w)
	h.addEntityHash(entityHash.Sum(nil))
}

func (h *multisetHasher) AddRelation(r model.Relation) {
	entityHash := sha256.New()
	writeByte(entityHash, 3)
	writeRelation(entityHash, r)
	h.addEntityHash(entityHash.Sum(nil))
}

func (h *multisetHasher) addEntityHash(sum []byte) {
	for i := 0; i < 4; i++ {
		start := i * 8
		v := binary.LittleEndian.Uint64(sum[start : start+8])
		h.sum[i] += v
		h.xor[i] ^= v
	}
}

func (h *multisetHasher) Digest() [sha256.Size]byte {
	var mixed [64]byte
	for i := 0; i < 4; i++ {
		binary.LittleEndian.PutUint64(mixed[i*8:(i+1)*8], h.sum[i])
	}
	for i := 0; i < 4; i++ {
		binary.LittleEndian.PutUint64(mixed[32+i*8:32+(i+1)*8], h.xor[i])
	}
	return sha256.Sum256(mixed[:])
}

func writeNode(h hash.Hash, n model.Node) {
	writeInt64(h, int64(n.ID))
	writeTags(h, n.Tags)
	writeInt64(h, model.ToCoordinate(0, 100, n.Lat))
	writeInt64(h, model.ToCoordinate(0, 100, n.Lon))
}

func writeWay(h hash.Hash, w model.Way) {
	writeInt64(h, int64(w.ID))
	writeInfo(h, w.Info)
	writeTags(h, w.Tags)
	writeInt(h, len(w.NodeIDs))
	for _, nodeID := range w.NodeIDs {
		writeInt64(h, int64(nodeID))
	}
}

func writeRelation(h hash.Hash, r model.Relation) {
	writeInt64(h, int64(r.ID))
	writeInfo(h, r.Info)
	writeTags(h, r.Tags)
	writeInt(h, len(r.Members))
	for _, member := range r.Members {
		writeInt64(h, int64(member.ID))
		writeInt32(h, int32(member.Type))
		writeString(h, member.Role)
	}
}

func writeInfo(h hash.Hash, info *model.Info) {
	if info == nil {
		writeByte(h, 0)
		return
	}

	writeByte(h, 1)
	writeInt32(h, info.Version)
	writeInt32(h, int32(info.UID))
	writeInt64(h, normalizeTimestamp(info.Timestamp).UnixNano())
	writeInt64(h, info.Changeset)
	writeString(h, info.User)
	if info.Visible {
		writeByte(h, 1)
	} else {
		writeByte(h, 0)
	}
}

func writeTags(h hash.Hash, tags map[string]string) {
	writeInt(h, len(tags))
	if len(tags) == 0 {
		return
	}

	keys := make([]string, 0, len(tags))
	for key := range tags {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		writeString(h, key)
		writeString(h, tags[key])
	}
}

func writeByte(h hash.Hash, value byte) {
	_, _ = h.Write([]byte{value})
}

func writeString(h hash.Hash, value string) {
	writeInt(h, len(value))
	_, _ = h.Write([]byte(value))
}

func writeInt(h hash.Hash, value int) {
	writeInt64(h, int64(value))
}

func writeInt32(h hash.Hash, value int32) {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], uint32(value))
	_, _ = h.Write(buf[:])
}

func writeInt64(h hash.Hash, value int64) {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], uint64(value))
	_, _ = h.Write(buf[:])
}

func writeUint64(h hash.Hash, value uint64) {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], value)
	_, _ = h.Write(buf[:])
}

func normalizeTimestamp(value time.Time) time.Time {
	return value.UTC().Truncate(time.Second)
}
