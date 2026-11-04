// SPDX-FileCopyrightText: the original author or authors
// SPDX-License-Identifier: Apache-2.0

package pbf

import (
	"context"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"m4o.io/pbf/v2/internal/edgecases"
	"m4o.io/pbf/v2/model"
)

// TestEdgeCasesFixture decodes testdata/edge-cases.osm.pbf and compares each
// entity with its case in package internal/edgecases. The comparison allows
// for the precision of the PBF format (see normalizeEntity). A failure names
// the case ID. The package doc of internal/edgecases tells what each case
// covers.
//
// The encoder keeps the order of the entities. Thus the decoder returns the
// entities in the order of the cases.
func TestEdgeCasesFixture(t *testing.T) {
	t.Parallel()

	f, err := os.Open("testdata/edge-cases.osm.pbf")
	require.NoError(t, err)

	defer f.Close()

	dec, err := NewDecoder(context.Background(), f)
	require.NoError(t, err)

	defer dec.Close()

	var got []model.Entity

	for {
		batch, err := dec.Decode()
		if errors.Is(err, io.EOF) {
			break
		}

		require.NoError(t, err)

		got = append(got, batch...)
	}

	cases := edgecases.Cases()

	require.Len(t, got, len(cases), "number of entities")

	for i, c := range cases {
		assert.Equal(t, normalizeEdgeCase(c.Entity), normalizeEdgeCase(got[i]), "case %s: %s", c.ID, c.Covers)
	}

	assertEdgeCasesHeader(t, &dec.Header)
}

func assertEdgeCasesHeader(t *testing.T, decoded *model.Header) {
	t.Helper()

	got := *decoded

	want := edgecases.Header()
	want.RequiredFeatures = defaultEncoderConfig.requiredFeatures
	want.BoundingBox = model.InitialBoundingBox()

	for _, c := range edgecases.Cases() {
		if n, ok := c.Entity.(*model.Node); ok {
			want.BoundingBox.ExpandWithLatLng(n.Lat, n.Lon)
		}
	}

	assert.True(t, got.BoundingBox.EqualWithin(want.BoundingBox, model.E7),
		"case H2: bounding box = %s, want %s", got.BoundingBox, want.BoundingBox)

	got.BoundingBox = want.BoundingBox
	got.OsmosisReplicationTimestamp = got.OsmosisReplicationTimestamp.UTC()
	assert.Equal(t, want, got, "case H1: header")
}

// normalizeEdgeCase is normalizeEntity, and it also sets an empty list of
// refs or members to nil. The decoder returns an empty list where a case
// has nil.
func normalizeEdgeCase(entity model.Entity) model.Entity {
	normalized := normalizeEntity(entity)

	switch e := normalized.(type) {
	case *model.Way:
		if len(e.NodeIDs) == 0 {
			e.NodeIDs = nil
		}
	case *model.Relation:
		if len(e.Members) == 0 {
			e.Members = nil
		}
	}

	return normalized
}
