// SPDX-FileCopyrightText: the original author or authors
// SPDX-License-Identifier: Apache-2.0

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
