package model_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"m4o.io/pbf/v2/model"
)

func TestHeader_JSON(t *testing.T) {
	ts, _ := time.Parse(time.RFC3339, "2024-10-28T14:21:30-07:00")
	h := model.Header{
		BoundingBox: &model.BoundingBox{
			Top:    51.69344,
			Left:   -0.511482,
			Bottom: 51.28554,
			Right:  0.335437,
		},
		RequiredFeatures:                 []string{"OsmSchema-V0.6", "DenseNodes"},
		OptionalFeatures:                 []string{"Sort.Type_then_ID"},
		WritingProgram:                   "osmium/1.14.0",
		OsmosisReplicationTimestamp:      ts,
		OsmosisReplicationSequenceNumber: 4221,
		OsmosisReplicationBaseURL:        "http://download.geofabrik.de/europe/united-kingdom/england/greater-london-updates",
	}

	b, err := json.Marshal(h)
	assert.NoError(t, err)
	assert.Equal(t, `{"bounding_box":{"top":51.69344,"left":-0.511482,"bottom":51.28554,"right":0.335437},"required_features":["OsmSchema-V0.6","DenseNodes"],"optional_features":["Sort.Type_then_ID"],"writing_program":"osmium/1.14.0","osmosis_replication_timestamp":"2024-10-28T14:21:30-07:00","osmosis_replication_sequence_number":4221,"osmosis_replication_base_url":"http://download.geofabrik.de/europe/united-kingdom/england/greater-london-updates"}`, string(b))
}
