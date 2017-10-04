// Copyright 2017 the original author or authors.
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
	"log"
	"os"
	"testing"
	"time"

	"github.com/maguro/pbf"
	"github.com/stretchr/testify/assert"
)

func TestRunInfo(t *testing.T) {
	f, err := os.Open("../../../testdata/greater-london.osm.pbf")
	if err != nil {
		log.Fatal(err)
	}
	info := runInfo(f, 2, false)

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
	assert.Equal(t, info.NodeCount, int64(0))
	assert.Equal(t, info.WayCount, int64(0))
	assert.Equal(t, info.RelationCount, int64(0))
}

func TestRunInfoExtended(t *testing.T) {
	f, err := os.Open("../../../testdata/greater-london.osm.pbf")
	if err != nil {
		log.Fatal(err)
	}
	info := runInfo(f, 2, true)

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
	assert.Equal(t, info.NodeCount, int64(2729006))
	assert.Equal(t, info.WayCount, int64(459055))
	assert.Equal(t, info.RelationCount, int64(12833))
}
