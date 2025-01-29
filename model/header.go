// Copyright 2017-25 the original author or authors.
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

package model

import (
	"time"
)

// Header is the contents of the OpenStreetMap PBF data file.
type Header struct {
	BoundingBox                      *BoundingBox `json:"bounding_box,omitempty"`
	RequiredFeatures                 []string     `json:"required_features,omitempty"`
	OptionalFeatures                 []string     `json:"optional_features,omitempty"`
	WritingProgram                   string       `json:"writing_program,omitempty"`
	Source                           string       `json:"source,omitempty"`
	OsmosisReplicationTimestamp      time.Time    `json:"osmosis_replication_timestamp,omitempty"`
	OsmosisReplicationSequenceNumber int64        `json:"osmosis_replication_sequence_number,omitempty"`
	OsmosisReplicationBaseURL        string       `json:"osmosis_replication_base_url,omitempty"`
}
