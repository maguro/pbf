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

package decoder

import (
	"bytes"
	"compress/zlib"
	"errors"
	"fmt"
	"io"

	"github.com/klauspost/compress/zstd"
	"github.com/pierrec/lz4"
	"github.com/ulikunitz/xz/lzma"

	"m4o.io/pbf/v2/internal/core"
	"m4o.io/pbf/v2/internal/pb"
)

var ErrUnknownCompressionType = errors.New("unknown blob compression type")

// unpack uncompresses the blob.
//
// This method is not "buried" within the readBlob function so that decompression
// of blobs can be performed concurrently.
func unpack(buf *core.PooledBuffer, blob *pb.Blob) ([]byte, error) {
	var factory func(blob *pb.Blob) (io.Reader, error)

	switch blob.Data.(type) {
	case *pb.Blob_Raw:
		return blob.GetRaw(), nil
	case *pb.Blob_ZlibData:
		factory = func(b *pb.Blob) (io.Reader, error) {
			return zlib.NewReader(bytes.NewReader(b.GetZlibData()))
		}
	case *pb.Blob_LzmaData:
		factory = func(b *pb.Blob) (io.Reader, error) {
			return lzma.NewReader(bytes.NewReader(b.GetLzmaData()))
		}
	case *pb.Blob_Lz4Data:
		factory = func(b *pb.Blob) (io.Reader, error) {
			return lz4.NewReader(bytes.NewReader(b.GetLz4Data())), nil
		}
	case *pb.Blob_ZstdData:
		factory = func(b *pb.Blob) (io.Reader, error) {
			return zstd.NewReader(bytes.NewReader(b.GetZstdData()))
		}
	default:
		return nil, ErrUnknownCompressionType
	}

	rawBufferSize := int(blob.GetRawSize() + bytes.MinRead)
	if rawBufferSize > buf.Cap() {
		buf.Grow(rawBufferSize)
	}

	rdr, err := factory(blob)
	if err != nil {
		return nil, fmt.Errorf("unpacker factory error: %w", err)
	}

	if n, err := buf.ReadFrom(rdr); err != nil {
		return nil, fmt.Errorf("unpacker read error: %w", err)
	} else if n != int64(blob.GetRawSize()) {
		return nil, fmt.Errorf("raw blob data size %d but expected %d", buf.Len(), blob.GetRawSize())
	}

	return buf.Bytes(), nil
}
