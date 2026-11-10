// SPDX-FileCopyrightText: the original author or authors
// SPDX-License-Identifier: Apache-2.0

package encoder

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"m4o.io/pbf/v2/model"
)

func TestCalcDeltasInt64(t *testing.T) {
	nodes := []model.ID{1, 1, 2, 3, 5, 7, 12}
	deltas := []model.ID{1, 0, 1, 1, 2, 2, 5}

	assert.Equal(t, deltas, calcDeltas(nodes))
}

func TestCalcDeltasFloat(t *testing.T) {
	nodes := []float32{1, 1, 2, 3, 5, 7, 12}
	deltas := []float32{1, 0, 1, 1, 2, 2, 5}

	assert.Equal(t, deltas, calcDeltas(nodes))
}

func TestCalcIDs(t *testing.T) {
	tags := map[string]string{"a": "b", "c": "d", "e": "f"}
	expectedKeyIDs := []uint32{1, 3, 5}
	expectedTagIDs := []uint32{2, 4, 6}

	strings := NewStrings()
	strings.Add("a")
	strings.Add("b")
	strings.Add("c")
	strings.Add("d")
	strings.Add("e")
	strings.Add("f")

	keyIDs, tagIDs := calcTagIDs(tags, strings.CalcTable())

	assert.Equal(t, expectedKeyIDs, keyIDs)
	assert.Equal(t, expectedTagIDs, tagIDs)
}

func TestFromTimestamp(t *testing.T) {
	ts, _ := time.Parse(time.RFC3339, "2022-02-13T20:40:22Z")

	assert.Equal(t, int64(1644784822), fromTimestamp(DateGranularityMs, ts))
	assert.Equal(t, int64(1644784822), fromTimestamp(DateGranularityMs, ts.Local()))
}

// TestToInfoPbKeepsTimestampsAfter2038 uses a timestamp that is larger than
// math.MaxInt32 seconds. The canonical osmformat.proto stores it in an int64
// field.
func TestToInfoPbKeepsTimestampsAfter2038(t *testing.T) {
	const seconds = 3_000_000_000

	strings := NewStrings()
	strings.Add("")
	table := strings.CalcTable()

	info := &model.Info{Timestamp: time.Unix(seconds, 0)}

	pbInfo := toInfoPb(info, table)

	assert.Equal(t, int64(seconds), pbInfo.GetTimestamp())
}

// TestToInfoPbOmitsVisible checks that the encoder writes no visible flag. A
// current data file does not declare HistoricalInformation, so the spec does
// not let the writer set the flag.
func TestToInfoPbOmitsVisible(t *testing.T) {
	strings := NewStrings()
	strings.Add("")
	table := strings.CalcTable()

	pbInfo := toInfoPb(&model.Info{Visible: false}, table)

	assert.Nil(t, pbInfo.Visible)
}

// TestExtractDenseNodesOmitsVisible checks the dense node path for the same
// rule as TestToInfoPbOmitsVisible.
func TestExtractDenseNodesOmitsVisible(t *testing.T) {
	entities := []model.Entity{
		&model.Node{ID: 1, Info: &model.Info{Visible: false}},
		&model.Node{ID: 2, Info: &model.Info{Visible: true}},
	}

	dense := newBlockContext(entities).extractDenseNodes(entities)

	assert.Empty(t, dense.GetDenseinfo().GetVisible())
}

func TestExtractPrimitiveBlockGroupsRunsInOrder(t *testing.T) {
	t.Parallel()

	info := func() *model.Info { return &model.Info{Visible: true} }

	entities := []model.Entity{
		&model.Node{ID: 1, Info: info()},
		&model.Node{ID: 2, Info: info()},
		&model.Way{ID: 10, Info: info()},
		&model.Relation{ID: 20, Info: info()},
		&model.Node{ID: 5, Info: info()},
	}

	groups := newBlockContext(entities).extractPrimitiveBlock().GetPrimitivegroup()

	require.Len(t, groups, 4)

	// The dense IDs are deltas. Each group starts its deltas from zero.
	assert.Equal(t, []int64{1, 1}, groups[0].GetDense().GetId())
	assert.Empty(t, groups[0].GetWays())

	require.Len(t, groups[1].GetWays(), 1)
	assert.Equal(t, int64(10), groups[1].GetWays()[0].GetId())
	assert.Nil(t, groups[1].GetDense())

	require.Len(t, groups[2].GetRelations(), 1)
	assert.Equal(t, int64(20), groups[2].GetRelations()[0].GetId())

	assert.Equal(t, []int64{5}, groups[3].GetDense().GetId())
}

func TestNewBlockContextNodeWithNilInfo(t *testing.T) {
	t.Parallel()

	entities := []model.Entity{
		&model.Node{ID: 1, Info: nil},
	}
	require.NotPanics(t, func() {
		bc := newBlockContext(entities)
		require.NotNil(t, bc)
	})
}

func TestExtractDenseNodesWithNilInfo(t *testing.T) {
	t.Parallel()

	entities := []model.Entity{
		&model.Node{ID: 1, Lat: 10.0, Lon: 20.0, Info: nil},
		&model.Node{ID: 2, Lat: 10.5, Lon: 20.5, Info: nil},
	}

	bc := newBlockContext(entities)
	dense := bc.extractDenseNodes(entities)
	require.NotNil(t, dense)
	assert.Nil(t, dense.Denseinfo)
	assert.Equal(t, []int64{1, 1}, dense.GetId())
}

func TestExtractDenseNodesWithMixedInfo(t *testing.T) {
	t.Parallel()

	entities := []model.Entity{
		&model.Node{ID: 1, Lat: 10.0, Lon: 20.0, Info: nil},
		&model.Node{ID: 2, Lat: 10.5, Lon: 20.5, Info: &model.Info{Version: 3, User: "bob"}},
	}

	bc := newBlockContext(entities)
	dense := bc.extractDenseNodes(entities)
	require.NotNil(t, dense)
	require.NotNil(t, dense.Denseinfo)
	assert.Equal(t, []int32{0, 3}, dense.GetDenseinfo().GetVersion())
}

func TestToInfoPbNil(t *testing.T) {
	t.Parallel()

	strings := NewStrings()
	table := strings.CalcTable()
	assert.Nil(t, toInfoPb(nil, table))
}

func TestExtractWaysAndRelationsNilInfo(t *testing.T) {
	t.Parallel()

	entities := []model.Entity{
		&model.Way{ID: 100, NodeIDs: []model.ID{1, 2}, Info: nil},
		&model.Relation{ID: 200, Info: nil},
	}

	bc := newBlockContext(entities)
	block := bc.extractPrimitiveBlock()
	require.NotNil(t, block)
	groups := block.GetPrimitivegroup()
	require.Len(t, groups, 2)
	assert.Nil(t, groups[0].GetWays()[0].GetInfo())
	assert.Nil(t, groups[1].GetRelations()[0].GetInfo())
}
