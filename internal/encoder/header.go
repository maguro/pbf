// SPDX-FileCopyrightText: the original author or authors
// SPDX-License-Identifier: Apache-2.0

package encoder

import (
	"fmt"
	"io"

	"google.golang.org/protobuf/proto"

	"m4o.io/pbf/v2/internal/pb"
	"m4o.io/pbf/v2/model"
)

func SaveHeader(w io.Writer, hdr *model.Header, compression BlobCompression) error {
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
		OsmosisReplicationTimestamp:      proto.Int64(hdr.OsmosisReplicationTimestamp.Unix()),
		OsmosisReplicationSequenceNumber: proto.Int64(hdr.OsmosisReplicationSequenceNumber),
		OsmosisReplicationBaseUrl:        proto.String(hdr.OsmosisReplicationBaseURL),
	}

	if err := writeBlob(w, hb, compression); err != nil {
		return fmt.Errorf("could not write header: %w", err)
	}

	return nil
}
