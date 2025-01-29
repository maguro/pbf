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

	"m4o.io/pbf/v2/internal/encoder/packers"
	"m4o.io/pbf/v2/internal/pb"
)

// Packer is the interface that groups methods for packing the contents of a
// PBF blob and saving the packed data in the correct place.
type Packer interface {
	// WriteCloser is used to write the contents of the blob to be packed.
	// Be sure to call the Close method to ensure that all the contents are
	// packed.
	io.WriteCloser

	// SaveTo will save the packed contents to the blob using the correct
	// Protobuf data class.
	SaveTo(blob *pb.Blob)
}

// Pack marshals and compresses the blob.
func Pack(msg proto.Message, c BlobCompression) (bb []byte, err error) {
	p := newPacker(c)

	b, err := proto.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("could not marshal message: %zw", err)
	}

	if _, err = p.Write(b); err != nil {
		return nil, fmt.Errorf("could not compress message: %w", err)
	}

	if err = p.Close(); err != nil {
		return nil, fmt.Errorf("could not close writer: %w", err)
	}

	blob := &pb.Blob{
		RawSize: proto.Int32(int32(len(b))),
	}

	p.SaveTo(blob)

	bb, err = proto.Marshal(blob)
	if err != nil {
		return nil, fmt.Errorf("could not marshal blob data: %w", err)
	}

	return bb, nil
}

// newPacker creates the appropriate Packer for the compression.
func newPacker(c BlobCompression) Packer {
	switch c {
	case RAW:
		return packers.NewRawPacker()
	case ZLIB:
		return packers.NewZlibPacker()
	case LZMA:
		return packers.NewLzmaPacker()
	case LZ4:
		return packers.NewLz4Packer()
	case ZSTD:
		return packers.NewZstdPacker()
	default:
		panic(fmt.Errorf("unknown compression type: %v", c))
	}
}
