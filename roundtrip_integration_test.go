// SPDX-FileCopyrightText: the original author or authors
// SPDX-License-Identifier: Apache-2.0

//go:build integration
// +build integration

package pbf

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"testing"
	"time"

	"m4o.io/pbf/v2/model"
)

// pbfPath names one PBF file to round-trip in place of the files in
// testdata. The test writes the encoded copy and the encoder store to the
// directory that TMPDIR names. Because of this, that directory must have
// space for two copies of the file.
var pbfPath = flag.String("pbf", "", "round-trip this PBF file in place of the files in testdata")

type entityDigest struct {
	total     int
	nodes     int
	ways      int
	relations int
	nodeSum   [sha256.Size]byte
	waySum    [sha256.Size]byte
	relSum    [sha256.Size]byte
	// extent is the bounding box of the node coordinates. The encoder writes
	// this extent to the header, not the bounding box of the source header.
	extent *model.BoundingBox
}

type multisetHasher struct {
	sum [4]uint64
	xor [4]uint64
}

func TestRoundTripOSMPBFDatasets(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name              string
		path              string
		compareNodeDigest bool
	}

	testCases := []testCase{
		{name: "sample", path: "testdata/sample.osm.pbf", compareNodeDigest: true},
		{name: "bremen", path: "testdata/bremen.osm.pbf", compareNodeDigest: true},
		{name: "london", path: "testdata/greater-london.osm.pbf", compareNodeDigest: true},
		{name: "san-francisco", path: "testdata/san-francisco.osm.pbf", compareNodeDigest: true},
		{name: "edge-cases", path: "testdata/edge-cases.osm.pbf", compareNodeDigest: true},
	}

	if *pbfPath != "" {
		name := filepath.Base(*pbfPath)
		testCases = []testCase{{name: name, path: *pbfPath, compareNodeDigest: true}}
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

			encodedPath := filepath.Join(t.TempDir(), "roundtrip.osm.pbf")
			encoded, err := os.Create(encodedPath)
			if err != nil {
				t.Fatalf("create %s: %v", encodedPath, err)
			}
			defer encoded.Close()

			buffered := bufio.NewWriter(encoded)
			reencoded, err := NewEncoder(buffered, headerOptions(decoded.Header)...)
			if err != nil {
				t.Fatalf("create encoder: %v", err)
			}

			sourceDigest, err := digestAndEncode(decoded, reencoded)
			if err != nil {
				t.Fatalf("digest and encode source %s: %v", tc.path, err)
			}

			reencoded.Close()

			if err := buffered.Flush(); err != nil {
				t.Fatalf("flush %s: %v", encodedPath, err)
			}
			if _, err := encoded.Seek(0, io.SeekStart); err != nil {
				t.Fatalf("rewind %s: %v", encodedPath, err)
			}

			roundTrip, err := NewDecoder(context.Background(), encoded)
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

			assertHeader(t, decoded.Header, sourceDigest.extent, roundTrip.Header)
		})
	}
}

// headerOptions returns the encoder options that copy the source header.
// headerOptions does not copy the writing program or the bounding box. The
// encoder writes its own values for these fields.
func headerOptions(hdr model.Header) []EncoderOption {
	return []EncoderOption{
		WithRequiredFeatures(hdr.RequiredFeatures...),
		WithOptionalFeatures(hdr.OptionalFeatures...),
		WithSource(hdr.Source),
		WithOsmosisReplicationTimestamp(hdr.OsmosisReplicationTimestamp),
		WithOsmosisReplicationSequenceNumber(hdr.OsmosisReplicationSequenceNumber),
		WithOsmosisReplicationBaseURL(hdr.OsmosisReplicationBaseURL),
	}
}

// assertHeader compares the round-trip header with the source header. The
// round-trip bounding box must be equal to the extent of the source nodes.
func assertHeader(t *testing.T, source model.Header, extent *model.BoundingBox, got model.Header) {
	t.Helper()

	if !slices.Equal(source.RequiredFeatures, got.RequiredFeatures) {
		t.Errorf("required features = %q, want %q", got.RequiredFeatures, source.RequiredFeatures)
	}
	if !slices.Equal(source.OptionalFeatures, got.OptionalFeatures) {
		t.Errorf("optional features = %q, want %q", got.OptionalFeatures, source.OptionalFeatures)
	}
	if got.Source != source.Source {
		t.Errorf("source = %q, want %q", got.Source, source.Source)
	}
	if !got.OsmosisReplicationTimestamp.Equal(source.OsmosisReplicationTimestamp) {
		t.Errorf("replication timestamp = %v, want %v", got.OsmosisReplicationTimestamp, source.OsmosisReplicationTimestamp)
	}
	if got.OsmosisReplicationSequenceNumber != source.OsmosisReplicationSequenceNumber {
		t.Errorf("replication sequence number = %d, want %d", got.OsmosisReplicationSequenceNumber, source.OsmosisReplicationSequenceNumber)
	}
	if got.OsmosisReplicationBaseURL != source.OsmosisReplicationBaseURL {
		t.Errorf("replication base URL = %q, want %q", got.OsmosisReplicationBaseURL, source.OsmosisReplicationBaseURL)
	}
	if !got.BoundingBox.EqualWithin(extent, model.E7) {
		t.Errorf("bounding box = %s, want %s", got.BoundingBox, extent)
	}
}

func digestAndEncode(dec *Decoder, enc *Encoder) (entityDigest, error) {
	nodeHasher := multisetHasher{}
	wayHasher := multisetHasher{}
	relHasher := multisetHasher{}
	stats := entityDigest{extent: model.InitialBoundingBox()}

	for {
		entities, err := dec.Decode()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return entityDigest{}, fmt.Errorf("decode source: %w", err)
		}

		if err := digestEntities(&nodeHasher, &wayHasher, &relHasher, &stats, entities); err != nil {
			return entityDigest{}, err
		}

		if err := enc.EncodeBatch(entities); err != nil {
			return entityDigest{}, fmt.Errorf("encode batch: %w", err)
		}
	}

	stats.nodeSum = nodeHasher.Digest()
	stats.waySum = wayHasher.Digest()
	stats.relSum = relHasher.Digest()

	return stats, nil
}

func digestDecoder(dec *Decoder) (entityDigest, error) {
	nodeHasher := multisetHasher{}
	wayHasher := multisetHasher{}
	relHasher := multisetHasher{}
	stats := entityDigest{extent: model.InitialBoundingBox()}

	for {
		entities, err := dec.Decode()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return entityDigest{}, fmt.Errorf("decode: %w", err)
		}

		if err := digestEntities(&nodeHasher, &wayHasher, &relHasher, &stats, entities); err != nil {
			return entityDigest{}, err
		}
	}

	stats.nodeSum = nodeHasher.Digest()
	stats.waySum = wayHasher.Digest()
	stats.relSum = relHasher.Digest()

	return stats, nil
}

func digestEntities(nodeHasher, wayHasher, relHasher *multisetHasher, stats *entityDigest, entities []model.Entity) error {
	for _, entity := range entities {
		stats.total++
		switch typed := entity.(type) {
		case model.Node:
			stats.nodes++
			nodeHasher.AddNode(typed)
		case *model.Node:
			stats.nodes++
			nodeHasher.AddNode(*typed)
			stats.extent.ExpandWithLatLng(typed.Lat, typed.Lon)
		case model.Way:
			stats.ways++
			wayHasher.AddWay(typed)
		case *model.Way:
			stats.ways++
			wayHasher.AddWay(*typed)
		case model.Relation:
			stats.relations++
			relHasher.AddRelation(typed)
		case *model.Relation:
			stats.relations++
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
	writeInfo(h, n.Info)
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
