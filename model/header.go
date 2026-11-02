// SPDX-FileCopyrightText: the original author or authors
// SPDX-License-Identifier: Apache-2.0

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
