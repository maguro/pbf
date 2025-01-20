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

package decoder

import (
	"fmt"
	"io"
	"time"

	"google.golang.org/protobuf/proto"

	"m4o.io/pbf/v2/internal/core"
	"m4o.io/pbf/v2/internal/pb"
	"m4o.io/pbf/v2/model"
)

func LoadHeader(reader io.Reader) (model.Header, error) {
	buf := core.NewPooledBuffer()
	defer buf.Close()

	blob, err := readBlob(reader)
	if err != nil {
		return model.Header{}, fmt.Errorf("error reading blob for header: %w", err)
	}

	unpacked, err := unpack(buf, blob)
	if err != nil {
		return model.Header{}, fmt.Errorf("error unpacking blob: %w", err)
	}

	var hb pb.HeaderBlock
	if err = proto.Unmarshal(unpacked, &hb); err != nil {
		return model.Header{}, fmt.Errorf("error unmarshalling header: %w", err)
	}

	hdr := model.Header{
		RequiredFeatures:                 hb.GetRequiredFeatures(),
		OptionalFeatures:                 hb.GetOptionalFeatures(),
		WritingProgram:                   hb.GetWritingprogram(),
		Source:                           hb.GetSource(),
		OsmosisReplicationBaseURL:        hb.GetOsmosisReplicationBaseUrl(),
		OsmosisReplicationSequenceNumber: hb.GetOsmosisReplicationSequenceNumber(),
	}

	if hb.Bbox != nil {
		hdr.BoundingBox = &model.BoundingBox{
			Left:   model.ToDegrees(0, 1, hb.Bbox.GetLeft()),
			Right:  model.ToDegrees(0, 1, hb.Bbox.GetRight()),
			Top:    model.ToDegrees(0, 1, hb.Bbox.GetTop()),
			Bottom: model.ToDegrees(0, 1, hb.Bbox.GetBottom()),
		}
	}

	if hb.OsmosisReplicationTimestamp != nil {
		hdr.OsmosisReplicationTimestamp = time.Unix(*hb.OsmosisReplicationTimestamp, 0)
	}

	return hdr, nil
}
