// Copyright 2026 the original author or authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package decoder

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protowire"

	"m4o.io/pbf/v2/model"
)

// TestParsePrimitiveBlockDecodesNonDenseNode encodes the block by hand, with
// the field numbers of the canonical osmformat.proto, so that the test does
// not depend on the generated code.
func TestParsePrimitiveBlockDecodesNonDenseNode(t *testing.T) {
	var info []byte
	info = protowire.AppendTag(info, 1, protowire.VarintType) // version
	info = protowire.AppendVarint(info, 3)

	var node []byte
	node = protowire.AppendTag(node, 1, protowire.VarintType) // id
	node = protowire.AppendVarint(node, protowire.EncodeZigZag(42))
	node = protowire.AppendTag(node, 2, protowire.BytesType) // keys
	node = protowire.AppendBytes(node, protowire.AppendVarint(nil, 1))
	node = protowire.AppendTag(node, 3, protowire.BytesType) // vals
	node = protowire.AppendBytes(node, protowire.AppendVarint(nil, 2))
	node = protowire.AppendTag(node, 4, protowire.BytesType) // info
	node = protowire.AppendBytes(node, info)
	node = protowire.AppendTag(node, 8, protowire.VarintType) // lat
	node = protowire.AppendVarint(node, protowire.EncodeZigZag(515_000_000))
	node = protowire.AppendTag(node, 9, protowire.VarintType) // lon
	node = protowire.AppendVarint(node, protowire.EncodeZigZag(-1_000_000))

	var group []byte
	group = protowire.AppendTag(group, 1, protowire.BytesType) // nodes
	group = protowire.AppendBytes(group, node)

	var table []byte
	for _, s := range []string{"", "name", "value"} {
		table = protowire.AppendTag(table, 1, protowire.BytesType) // s
		table = protowire.AppendString(table, s)
	}

	var block []byte
	block = protowire.AppendTag(block, 1, protowire.BytesType) // stringtable
	block = protowire.AppendBytes(block, table)
	block = protowire.AppendTag(block, 2, protowire.BytesType) // primitivegroup
	block = protowire.AppendBytes(block, group)

	entities, err := parsePrimitiveBlock(block)
	require.NoError(t, err)
	require.Len(t, entities, 1)

	n, ok := entities[0].(*model.Node)
	require.True(t, ok)

	assert.Equal(t, model.ID(42), n.ID)
	assert.Equal(t, map[string]string{"name": "value"}, n.Tags)
	assert.Equal(t, int32(3), n.Info.Version)
	assert.InDelta(t, 51.5, float64(n.Lat), 1e-9)
	assert.InDelta(t, -0.1, float64(n.Lon), 1e-9)
}

// TestParsePrimitiveBlockDecodesTimestampsAfter2038 uses a timestamp that is
// larger than math.MaxInt32 seconds. The canonical osmformat.proto stores it
// in an int64 field for Info and in a sint64 field for DenseInfo.
func TestParsePrimitiveBlockDecodesTimestampsAfter2038(t *testing.T) {
	const seconds = 3_000_000_000

	want := time.Unix(seconds, 0).UTC()

	tests := []struct {
		name  string
		group []byte
	}{
		{
			name: "info of a node",
			group: field(1, message( // nodes
				varint(1, protowire.EncodeZigZag(42)), // id
				field(4, varint(2, seconds)),          // info.timestamp
				varint(8, 0),                          // lat
				varint(9, 0),                          // lon
			)),
		},
		{
			name: "dense info of a dense node",
			group: field(2, message( // dense
				field(1, packedZigZag(42)), // id
				field(5, message( // denseinfo
					field(1, packedVarint(1)),       // version
					field(2, packedZigZag(seconds)), // timestamp
					field(3, packedZigZag(0)),       // changeset
					field(4, packedZigZag(0)),       // uid
					field(5, packedZigZag(0)),       // user_sid
				)),
				field(8, packedZigZag(0)), // lat
				field(9, packedZigZag(0)), // lon
			)),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			block := message(
				field(1, field(1, nil)), // stringtable with the empty string
				field(2, tc.group),      // primitivegroup
			)

			entities, err := parsePrimitiveBlock(block)
			require.NoError(t, err)
			require.Len(t, entities, 1)

			assert.Equal(t, want, entities[0].GetInfo().Timestamp)
		})
	}
}

// TestParsePrimitiveBlockSkipsChangeSets decodes a group that has a ChangeSet
// next to a node. The decoder does not return change sets.
func TestParsePrimitiveBlockSkipsChangeSets(t *testing.T) {
	group := message(
		field(1, message( // nodes
			varint(1, protowire.EncodeZigZag(42)), // id
			varint(8, 0),                          // lat
			varint(9, 0),                          // lon
		)),
		field(5, varint(1, 7)), // changesets.id
	)
	block := message(
		field(1, field(1, nil)), // stringtable with the empty string
		field(2, group),         // primitivegroup
	)

	entities, err := parsePrimitiveBlock(block)
	require.NoError(t, err)
	require.Len(t, entities, 1)

	assert.Equal(t, model.ID(42), entities[0].GetID())
}

func message(fields ...[]byte) []byte {
	var b []byte
	for _, f := range fields {
		b = append(b, f...)
	}

	return b
}

func field(num protowire.Number, value []byte) []byte {
	b := protowire.AppendTag(nil, num, protowire.BytesType)

	return protowire.AppendBytes(b, value)
}

func varint(num protowire.Number, value uint64) []byte {
	b := protowire.AppendTag(nil, num, protowire.VarintType)

	return protowire.AppendVarint(b, value)
}

func packedVarint(values ...uint64) []byte {
	var b []byte
	for _, v := range values {
		b = protowire.AppendVarint(b, v)
	}

	return b
}

func packedZigZag(values ...int64) []byte {
	var b []byte
	for _, v := range values {
		b = protowire.AppendVarint(b, protowire.EncodeZigZag(v))
	}

	return b
}
