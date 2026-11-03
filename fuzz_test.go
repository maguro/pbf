// SPDX-FileCopyrightText: the original author or authors
// SPDX-License-Identifier: Apache-2.0

package pbf

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"m4o.io/pbf/v2/internal/encoder"
	"m4o.io/pbf/v2/model"
)

// The range of seconds that time.Time can show as a calendar date from year 1
// to year 9999.
const (
	minTimestamp = -62135596800
	maxTimestamp = 253402300799
)

// FuzzDecode gives arbitrary bytes to the decoder. The decoder must return an
// error or entities. It must not panic or hang.
func FuzzDecode(f *testing.F) {
	sample, err := os.ReadFile("testdata/sample.osm.pbf")
	require.NoError(f, err)

	f.Add(sample)
	f.Add(sample[:len(sample)/2])
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, data []byte) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		reader := bytes.NewReader(data)

		dec, err := NewDecoder(ctx, reader, WithNCpus(1))
		if err != nil {
			return
		}
		defer dec.Close()

		for {
			_, err := dec.Decode()
			if err != nil {
				return
			}
		}
	})
}

// FuzzRoundTrip encodes a node, a way, and a relation that it makes from the
// fuzz values, then decodes the result. The decoded entities must be equal to
// the encoded entities, after the precision loss that the PBF format causes.
func FuzzRoundTrip(f *testing.F) {
	f.Add(int64(1001), 42.1234567, -71.9876543, "name", "a", "alice", int64(1_700_000_000),
		int32(3), int32(101), int64(55), int64(7), true)
	f.Add(int64(-1), -90.0, 180.0, "", "", "", int64(0), int32(0), int32(0), int64(0), int64(0), false)
	f.Add(int64(1<<62), 90.0, -180.0, "k", "", "u", int64(maxTimestamp),
		int32(1<<30), int32(-1), int64(-1), int64(-1<<62), true)

	f.Fuzz(func(t *testing.T, id int64, lat, lon float64, key, value, user string, timestamp int64,
		version, uid int32, changeset, refDelta int64, visible bool,
	) {
		if lat < float64(model.MinLat) || lat > float64(model.MaxLat) ||
			lon < float64(model.MinLon) || lon > float64(model.MaxLon) {
			t.Skip("coordinate is not on the earth")
		}

		if timestamp < minTimestamp || timestamp > maxTimestamp {
			t.Skip("timestamp is not a calendar date")
		}

		info := model.Info{
			Version:   version,
			UID:       model.UID(uid),
			Timestamp: time.Unix(timestamp, 0).UTC(),
			Changeset: changeset,
			User:      user,
			Visible:   visible,
		}
		tags := map[string]string{key: value}
		nodeID := model.ID(id)

		want := []model.Entity{
			&model.Node{ID: nodeID, Tags: tags, Info: &info, Lat: model.Degrees(lat), Lon: model.Degrees(lon)},
			&model.Way{ID: nodeID, Tags: tags, Info: &info, NodeIDs: []model.ID{nodeID, nodeID + model.ID(refDelta)}},
			&model.Relation{
				ID: nodeID, Tags: tags, Info: &info,
				Members: []model.Member{{ID: nodeID, Type: model.NODE, Role: value}},
			},
		}

		got := roundTrip(t, want)

		// The decoder does not keep the order of entities of different types.
		require.Len(t, got, len(want))
		assert.ElementsMatch(t, normalizeEntities(want), normalizeEntities(got))
	})
}

// roundTrip encodes the entities and returns the entities that the decoder
// reads back, in the order that it reads them.
func roundTrip(t *testing.T, entities []model.Entity) []model.Entity {
	t.Helper()

	var encoded bytes.Buffer

	enc, err := NewEncoder(&encoded)
	require.NoError(t, err)

	err = enc.EncodeBatch(entities)
	require.NoError(t, err)

	enc.Close()

	reader := bytes.NewReader(encoded.Bytes())

	dec, err := NewDecoder(context.Background(), reader, WithNCpus(1))
	require.NoError(t, err)

	defer dec.Close()

	var decoded []model.Entity

	for {
		batch, err := dec.Decode()
		if errors.Is(err, io.EOF) {
			return decoded
		}

		require.NoError(t, err)

		decoded = append(decoded, batch...)
	}
}

func normalizeEntities(entities []model.Entity) []model.Entity {
	normalized := make([]model.Entity, 0, len(entities))

	for _, entity := range entities {
		normalized = append(normalized, normalizeEntity(entity))
	}

	return normalized
}

// normalizeEntity returns a copy of the entity with the precision of the PBF
// format: coordinates in units of the encoder granularity, timestamps in
// whole seconds, and nil for empty tag maps.
func normalizeEntity(entity model.Entity) model.Entity {
	switch e := entity.(type) {
	case *model.Node:
		n := *e
		n.Tags = normalizeTags(n.Tags)
		n.Info = normalizeInfo(n.Info)
		n.Lat = normalizeDegrees(n.Lat)
		n.Lon = normalizeDegrees(n.Lon)

		return &n
	case *model.Way:
		w := *e
		w.Tags = normalizeTags(w.Tags)
		w.Info = normalizeInfo(w.Info)

		return &w
	case *model.Relation:
		r := *e
		r.Tags = normalizeTags(r.Tags)
		r.Info = normalizeInfo(r.Info)

		return &r
	default:
		return entity
	}
}

func normalizeTags(tags map[string]string) map[string]string {
	if len(tags) == 0 {
		return nil
	}

	return tags
}

func normalizeInfo(info *model.Info) *model.Info {
	if info == nil {
		return nil
	}

	i := *info
	seconds := i.Timestamp.Unix()
	i.Timestamp = time.Unix(seconds, 0).UTC()

	return &i
}

func normalizeDegrees(degrees model.Degrees) model.Degrees {
	coordinate := model.ToCoordinate(encoder.LatOffset, encoder.Granularity, degrees)

	return model.ToDegrees(encoder.LatOffset, encoder.Granularity, coordinate)
}
