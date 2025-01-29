// Copyright 2025 the original author or authors.
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

package encoder

import (
	"fmt"
	"io"

	"google.golang.org/protobuf/proto"

	"m4o.io/pbf/v2/internal/pb"
	"m4o.io/pbf/v2/model"
)

func SaveHeader(wrtr io.Writer, hdr model.Header, compression BlobCompression) error {
	bbox := hdr.BoundingBox
	hb := &pb.HeaderBlock{
		Bbox: &pb.HeaderBBox{
			Top:    proto.Int64(bbox.Top.Coordinate()),
			Left:   proto.Int64(bbox.Left.Coordinate()),
			Bottom: proto.Int64(bbox.Bottom.Coordinate()),
			Right:  proto.Int64(bbox.Right.Coordinate()),
		},
		RequiredFeatures:                 hdr.RequiredFeatures,
		OptionalFeatures:                 hdr.OptionalFeatures,
		Writingprogram:                   proto.String(hdr.WritingProgram),
		Source:                           proto.String(hdr.Source),
		OsmosisReplicationTimestamp:      proto.Int64(fromTimestamp(DateGranularityMs, hdr.OsmosisReplicationTimestamp)),
		OsmosisReplicationSequenceNumber: proto.Int64(hdr.OsmosisReplicationSequenceNumber),
		OsmosisReplicationBaseUrl:        proto.String(hdr.OsmosisReplicationBaseURL),
	}

	if err := writeBlob(wrtr, hb, compression); err != nil {
		return fmt.Errorf("could not write header: %w", err)
	}

	return nil
}
