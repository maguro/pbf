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
