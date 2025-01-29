package encoder

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

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
