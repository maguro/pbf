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
	"encoding/binary"
	"fmt"
	"io"

	"google.golang.org/protobuf/proto"

	"m4o.io/pbf/v2/internal/pb"
)

// writeBlob marshals a Protobuf Message, msg, into a PBF blob and writes its
// blob header and blob data to the wrtr.
func writeBlob(wrtr io.Writer, msg proto.Message, c BlobCompression) (err error) {
	bb, err := Pack(msg, c)
	if err != nil {
		return fmt.Errorf("could not marshal blob data: %w", err)
	}

	hdr := &pb.BlobHeader{
		Type:     proto.String("OSMHeader"),
		Datasize: proto.Int32(int32(len(bb))),
	}

	hb, err := proto.Marshal(hdr)
	if err != nil {
		return fmt.Errorf("could not marshal blob header: %w", err)
	}

	if err = binary.Write(wrtr, binary.BigEndian, uint32(len(hb))); err != nil {
		return fmt.Errorf("could not write header size: %w", err)
	}

	if _, err = wrtr.Write(hb); err != nil {
		return fmt.Errorf("could not write blob header: %w", err)
	}

	if _, err = wrtr.Write(bb); err != nil {
		return fmt.Errorf("could not write blob data: %w", err)
	}

	return nil
}
