// SPDX-FileCopyrightText: the original author or authors
// SPDX-License-Identifier: Apache-2.0

package encoder

import (
	"encoding/binary"
	"fmt"
	"io"

	"google.golang.org/protobuf/proto"

	"m4o.io/pbf/v2/internal/pb"
)

// writeBlob marshals a Protobuf Message, msg, into a PBF blob and writes its
// blob header and blob data to w.
func writeBlob(w io.Writer, msg proto.Message, c BlobCompression) (err error) {
	bb, err := Pack(msg, c)
	if err != nil {
		return fmt.Errorf("could not marshal blob data: %w", err)
	}

	hdr := &pb.BlobHeader{
		Type:     proto.String("OSMHeader"),
		Datasize: proto.Int32(int32(len(bb))), //nolint:gosec // The encoder does not check the 32 MiB blob limit. See maguro/pbf#11.
	}

	hb, err := proto.Marshal(hdr)
	if err != nil {
		return fmt.Errorf("could not marshal blob header: %w", err)
	}

	//nolint:gosec // A BlobHeader is much smaller than the 64 KiB limit.
	if err = binary.Write(w, binary.BigEndian, uint32(len(hb))); err != nil {
		return fmt.Errorf("could not write header size: %w", err)
	}

	if _, err = w.Write(hb); err != nil {
		return fmt.Errorf("could not write blob header: %w", err)
	}

	if _, err = w.Write(bb); err != nil {
		return fmt.Errorf("could not write blob data: %w", err)
	}

	return nil
}
