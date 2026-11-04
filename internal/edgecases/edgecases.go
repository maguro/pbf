// SPDX-FileCopyrightText: the original author or authors
// SPDX-License-Identifier: Apache-2.0

// Package edgecases holds the content of testdata/edge-cases.osm.pbf. Each
// case is one entity, with an ID and the encoder or decoder code path that it
// covers. The command in internal/tools/edgecases writes the file from these
// cases. TestEdgeCasesFixture decodes the file and compares each entity with
// its case.
//
// The cases cover what the OSM extracts in testdata do not have: extreme IDs
// and Info values, deltas that wrap, empty lists, and unusual strings.
//
// Each entity has a visible Info. The encoder panics on a nil Info (#13), and
// it writes no historical information (#24).
package edgecases

import (
	"fmt"
	"math"
	"time"

	"m4o.io/pbf/v2/model"
)

// Case is one entity in the fixture.
type Case struct {
	// ID identifies the case in test failures and in testdata/README.md.
	ID string
	// Covers names the code path that the case tests.
	Covers string
	// Entity is the value that the generator encodes.
	Entity model.Entity
}

const (
	utf8Text  = "Zoë 地図"
	roleOuter = "outer"
	keyCopy   = "copy"

	maxID = model.ID(math.MaxInt64)
	minID = model.ID(math.MinInt64)
)

// Header returns the header values that the generator writes. This is case
// H1: each header field that a caller can set has a value that is not the
// default, and the source is UTF-8. The encoder computes the bounding box
// from the nodes, so Header has no bounding box. The writing program and the
// required features are the encoder defaults. The optional features are not
// standard features, because the file has none of the properties that the
// standard features state.
func Header() model.Header {
	return model.Header{
		OptionalFeatures:                 []string{"edge-cases", "test-only"},
		Source:                           "edge cases, " + utf8Text,
		OsmosisReplicationTimestamp:      time.Date(2026, time.November, 3, 12, 34, 56, 0, time.UTC),
		OsmosisReplicationSequenceNumber: 4_321_987,
		OsmosisReplicationBaseURL:        "https://planet.example.org/replication/minute/",
	}
}

// Cases returns the cases in the order that the generator encodes them. The
// decoder returns the entities in this order.
//
//nolint:funlen // The function is a table of cases.
func Cases() []Case {
	epoch := time.Unix(0, 0).UTC()
	recent := time.Date(2026, time.November, 3, 12, 34, 56, 0, time.UTC)
	before1970 := time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC)

	zero := func() *model.Info {
		return &model.Info{Timestamp: epoch, Visible: true}
	}
	full := func() *model.Info {
		return &model.Info{
			Version:   1 << 30,
			UID:       math.MaxInt32,
			Timestamp: recent,
			Changeset: 1 << 40,
			User:      utf8Text,
			Visible:   true,
		}
	}
	negative := func() *model.Info {
		return &model.Info{Version: -1, UID: -1, Timestamp: before1970, Changeset: -1, User: "neg", Visible: true}
	}

	return []Case{
		{
			ID:     "N1",
			Covers: "Zero Info values. The empty user is string-table index 1, because index 0 is a dummy entry.",
			Entity: &model.Node{ID: 1, Info: zero()},
		},
		{
			ID:     "N2",
			Covers: "Many tags in the dense keys_vals list, with UTF-8, spaces, = signs, and an empty value.",
			Entity: &model.Node{ID: 2, Info: full(), Tags: manyTags(), Lat: 51.5074123, Lon: -0.1278456},
		},
		{
			ID: "N3",
			Covers: "A tag with an empty key. Index 0 ends the tags of a dense node, so the empty key must use " +
				"index 1. The next case has tags, so a wrong end of tags moves tags to the wrong node.",
			Entity: &model.Node{ID: 3, Info: full(), Tags: map[string]string{"": "x", "k": "v"}},
		},
		{
			ID:     "N4",
			Covers: "The maximum latitude and longitude.",
			Entity: &model.Node{ID: 4, Info: zero(), Tags: map[string]string{"corner": "max"}, Lat: model.MaxLat, Lon: model.MaxLon},
		},
		{
			ID:     "N5",
			Covers: "The minimum latitude and longitude.",
			Entity: &model.Node{ID: 5, Info: zero(), Lat: model.MinLat, Lon: model.MinLon},
		},
		{
			ID:     "N6a",
			Covers: "The minimum ID. The dense ID delta from the node before wraps.",
			Entity: &model.Node{ID: minID, Info: full(), Lat: -33.8688197, Lon: 151.2092955},
		},
		{
			ID:     "N6b",
			Covers: "The maximum ID after the minimum ID. The dense ID delta of 2^64-1 wraps.",
			Entity: &model.Node{ID: maxID, Info: full(), Lat: 35.6761919, Lon: 139.6503106},
		},
		{
			ID:     "N7a",
			Covers: "The maximum UID.",
			Entity: &model.Node{ID: 7, Info: &model.Info{UID: math.MaxInt32, Timestamp: recent, Visible: true}},
		},
		{
			ID:     "N7b",
			Covers: "The minimum UID after the maximum UID. The int32 dense UID delta wraps.",
			Entity: &model.Node{ID: 8, Info: &model.Info{UID: math.MinInt32, Timestamp: recent, Visible: true}},
		},
		{
			ID:     "N8",
			Covers: "Negative version, UID, and changeset, and a timestamp before 1970, in the dense Info columns.",
			Entity: &model.Node{ID: 9, Info: negative()},
		},
		{
			ID:     "N9",
			Covers: "A long UTF-8 user name in the string table.",
			Entity: &model.Node{ID: 10, Info: &model.Info{User: longUser(), Timestamp: recent, Visible: true}},
		},

		{
			ID:     "W1",
			Covers: "A way without refs and without tags.",
			Entity: &model.Way{ID: 1, Info: zero()},
		},
		{
			ID:     "W2",
			Covers: "A way with one ref and many tags.",
			Entity: &model.Way{ID: 2, Info: full(), Tags: manyTags(), NodeIDs: []model.ID{1}},
		},
		{
			ID:     "W3",
			Covers: "A closed way with a repeated ref. The refs keep their order.",
			Entity: &model.Way{ID: 3, Info: full(), NodeIDs: []model.ID{1, 2, 3, 4, 1}},
		},
		{
			ID:     "W4",
			Covers: "Refs of the minimum and maximum ID. The ref deltas wrap.",
			Entity: &model.Way{ID: 4, Info: full(), NodeIDs: []model.ID{minID, maxID, minID}},
		},
		{
			ID:     "W5",
			Covers: "The minimum way ID, and refs to nodes that are not in the file.",
			Entity: &model.Way{ID: minID, Info: zero(), NodeIDs: []model.ID{100, 101}},
		},
		{
			ID:     "W6",
			Covers: "Negative version, UID, and changeset, and a timestamp before 1970, in the Info of a way.",
			Entity: &model.Way{ID: 6, Info: negative(), NodeIDs: []model.ID{1, 2}},
		},

		{
			ID:     "R1",
			Covers: "A relation without members.",
			Entity: &model.Relation{ID: 1, Info: zero()},
		},
		{
			ID:     "R2",
			Covers: "Members of each type, an empty role, and a member that is the relation itself.",
			Entity: &model.Relation{
				ID: 2, Info: full(), Tags: manyTags(),
				Members: []model.Member{
					{ID: 1, Type: model.NODE, Role: "stop"},
					{ID: 3, Type: model.WAY, Role: roleOuter},
					{ID: 1, Type: model.RELATION, Role: ""},
					{ID: 2, Type: model.RELATION, Role: "self"},
				},
			},
		},
		{
			ID:     "R3",
			Covers: "Member IDs of the minimum and maximum ID. The member ID deltas wrap. The role is UTF-8.",
			Entity: &model.Relation{
				ID: 3, Info: negative(),
				Members: []model.Member{
					{ID: minID, Type: model.NODE, Role: utf8Text},
					{ID: maxID, Type: model.WAY, Role: roleOuter},
					{ID: minID, Type: model.RELATION, Role: "inner"},
				},
			},
		},
		{
			ID:     "R4",
			Covers: "The maximum relation ID, and a member that is not in the file.",
			Entity: &model.Relation{
				ID: maxID, Info: full(),
				Members: []model.Member{{ID: 999, Type: model.RELATION, Role: "parent"}},
			},
		},
		{
			ID:     "R5a",
			Covers: "The first of two relations with the same ID. Nothing checks for duplicate IDs.",
			Entity: &model.Relation{ID: 5, Info: zero(), Tags: map[string]string{keyCopy: "1"}},
		},
		{
			ID:     "R5b",
			Covers: "The second of two relations with the same ID.",
			Entity: &model.Relation{ID: 5, Info: zero(), Tags: map[string]string{keyCopy: "2"}},
		},

		{
			ID: "M1",
			Covers: "A node after the relations. The block holds a second dense group, and the dense ID " +
				"deltas of that group start from zero again.",
			Entity: &model.Node{ID: 11, Info: full(), Tags: map[string]string{"after": "relations"}},
		},
	}
}

func manyTags() map[string]string {
	tags := map[string]string{
		"name":    utf8Text,
		"name:en": "Map",
		"note":    "a value with spaces, commas, and = signs",
		"empty":   "",
	}

	for i := range 36 {
		key := fmt.Sprintf("key%02d", i)
		tags[key] = fmt.Sprintf("value%02d", i)
	}

	return tags
}

func longUser() string {
	user := ""
	for range 64 {
		user += "地図"
	}

	return user
}
