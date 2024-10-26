// Copyright 2017-24 the original author or authors.
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

package info

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"m4o.io/pbf"
)

func TestRunInfo(t *testing.T) {
	testRunInfoWith(t, false, 0, 0, 0)
}

func testRunInfoWith(t *testing.T, extended bool, node int64, way int64, relation int64) {
	f, err := os.Open("../../../testdata/greater-london.osm.pbf")
	if err != nil {
		t.Fatalf("Unable to read data file %v", err)
	}

	info := runInfo(f, 2, extended)
	bbox := pbf.BoundingBox{Left: -0.511482, Right: 0.335437, Top: 51.69344, Bottom: 51.28554}
	ts, _ := time.Parse(time.RFC3339, "2014-03-24T21:55:02Z")

	assert.True(t, info.BoundingBox.EqualWithin(bbox, pbf.E6))
	assert.Equal(t, info.RequiredFeatures, []string{"OsmSchema-V0.6", "DenseNodes"})
	assert.Equal(t, info.OptionalFeatures, []string(nil))
	assert.Equal(t, info.WritingProgram, "Osmium (http://wiki.openstreetmap.org/wiki/Osmium)")
	assert.Equal(t, info.Source, "")
	assert.Equal(t, info.OsmosisReplicationTimestamp.UTC(), ts)
	assert.Equal(t, info.OsmosisReplicationSequenceNumber, int64(0))
	assert.Equal(t, info.OsmosisReplicationBaseURL, "")
	assert.Equal(t, info.NodeCount, node)
	assert.Equal(t, info.WayCount, way)
	assert.Equal(t, info.RelationCount, relation)
}
func TestRenderJSON(t *testing.T) {
	bbox := pbf.BoundingBox{Left: -0.511482, Right: 0.335437, Top: 51.69344, Bottom: 51.28554}
	ts, _ := time.Parse(time.RFC3339, "2014-03-24T21:55:02Z")
	h := pbf.Header{
		BoundingBox:                      bbox,
		RequiredFeatures:                 []string{"OsmSchema-V0.6", "DenseNodes"},
		OptionalFeatures:                 []string(nil),
		WritingProgram:                   "Osmium (http://wiki.openstreetmap.org/wiki/Osmium)",
		Source:                           "",
		OsmosisReplicationTimestamp:      ts,
		OsmosisReplicationSequenceNumber: int64(0),
		OsmosisReplicationBaseURL:        "",
	}
	eh := &extendedHeader{
		Header:        h,
		NodeCount:     int64(2729006),
		WayCount:      int64(459055),
		RelationCount: int64(12833),
	}

	// mock out to collect JSON output
	buf := bytes.NewBuffer(make([]byte, 8192))
	buf.Reset()

	saved := out

	defer func() { out = saved }()

	out = buf

	renderJSON(eh, true)

	info := &extendedHeader{}
	if err := json.Unmarshal(buf.Bytes(), info); err != nil {
		t.Fatalf("Unable to unmarshal json %v", err)
	}

	assert.True(t, info.BoundingBox.EqualWithin(bbox, pbf.E6))
	assert.Equal(t, info.RequiredFeatures, []string{"OsmSchema-V0.6", "DenseNodes"})
	assert.Equal(t, info.OptionalFeatures, []string(nil))
	assert.Equal(t, info.WritingProgram, "Osmium (http://wiki.openstreetmap.org/wiki/Osmium)")
	assert.Equal(t, info.Source, "")
	assert.Equal(t, info.OsmosisReplicationTimestamp.UTC(), ts)
	assert.Equal(t, info.OsmosisReplicationSequenceNumber, int64(0))
	assert.Equal(t, info.OsmosisReplicationBaseURL, "")
	assert.Equal(t, info.NodeCount, int64(2729006))
	assert.Equal(t, info.WayCount, int64(459055))
	assert.Equal(t, info.RelationCount, int64(12833))
}

func TestRenderText(t *testing.T) {
	bbox := pbf.BoundingBox{Left: -0.511482, Right: 0.335437, Top: 51.69344, Bottom: 51.28554}
	ts, _ := time.Parse(time.RFC3339, "2014-03-24T21:55:02Z")
	h := pbf.Header{
		BoundingBox:                      bbox,
		RequiredFeatures:                 []string{"OsmSchema-V0.6", "DenseNodes"},
		OptionalFeatures:                 []string{"Pbf"},
		WritingProgram:                   "Osmium (http://wiki.openstreetmap.org/wiki/Osmium)",
		Source:                           "pbf",
		OsmosisReplicationTimestamp:      ts,
		OsmosisReplicationSequenceNumber: int64(0),
		OsmosisReplicationBaseURL:        "https://github.com/maguro/pbf",
	}
	eh := &extendedHeader{
		Header:        h,
		NodeCount:     int64(2729006),
		WayCount:      int64(459055),
		RelationCount: int64(12833),
	}

	// mock out to collect text output
	buf := bytes.NewBuffer(make([]byte, 8192))
	buf.Reset()

	saved := out

	defer func() { out = saved }()

	out = buf

	renderTxt(eh, true)

	assert.Equal(t, `BoundingBox: [-0.511482, 51.28554, 0.335437, 51.69344]
RequiredFeatures: OsmSchema-V0.6, DenseNodes
OptionalFeatures: Pbf
WritingProgram: Osmium (http://wiki.openstreetmap.org/wiki/Osmium)
Source: pbf
OsmosisReplicationTimestamp: 2014-03-24T21:55:02Z
OsmosisReplicationSequenceNumber: 0
OsmosisReplicationBaseURL: https://github.com/maguro/pbf
NodeCount: 2,729,006
WayCount: 459,055
RelationCount: 12,833
`, buf.String())
}
