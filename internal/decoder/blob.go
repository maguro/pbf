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
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"google.golang.org/protobuf/proto"

	"m4o.io/pbf/v2/internal/core"
	"m4o.io/pbf/v2/internal/pb"
)

// GenerateBlobReader creates an iterator that returns primitive blobs read
// off of the reader.
func GenerateBlobReader(ctx context.Context, reader io.Reader) func(yield func(enc *pb.Blob, err error) bool) {
	return func(yield func(enc *pb.Blob, err error) bool) {
		buffer := core.NewPooledBuffer()
		defer buffer.Close()

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			blob, err := readBlob(reader)
			if err != nil {
				if !errors.Is(err, io.EOF) {
					slog.Error("unable to read blob", "error", err)
					yield(nil, err)
				}

				return
			}

			if !yield(blob, nil) {
				return
			}

			buffer.Reset()
		}
	}
}

// readBlob reads a PBF blob from the rdr.
func readBlob(rdr io.Reader) (*pb.Blob, error) {
	h, err := readBlobHeader(rdr)
	if err != nil {
		return nil, fmt.Errorf("error reading blob header: %w", err)
	}

	b, err := readBlobData(rdr, int64(h.GetDatasize()))
	if err != nil {
		return nil, fmt.Errorf("error reading blob: %w", err)
	}

	return b, nil
}

// readBlobHeader unmarshals a header from an array of protobuf encoded bytes.
// The header is used when decoding blobs into OSM entities.
func readBlobHeader(rdr io.Reader) (header *pb.BlobHeader, err error) {
	buf := core.NewPooledBuffer()
	defer buf.Close()

	var size uint32

	err = binary.Read(rdr, binary.BigEndian, &size)
	if err != nil {
		return nil, fmt.Errorf("error reading blob size: %w", err)
	}

	if n, err := io.CopyN(buf, rdr, int64(size)); err != nil {
		return nil, fmt.Errorf("error reading blob: %w", err)
	} else if n != int64(size) {
		return nil, fmt.Errorf("error reading blob: expected %d bytes, got %d", size, n)
	}

	header = &pb.BlobHeader{}

	if err := proto.Unmarshal(buf.Bytes(), header); err != nil {
		return nil, fmt.Errorf("error unmarshalling blob header: %w", err)
	}

	return header, nil
}

// readBlobData unmarshals a blob from an array of protobuf encoded bytes.  The
// blob still needs to be decoded into OSM entities.
func readBlobData(rdr io.Reader, size int64) (*pb.Blob, error) {
	buf := core.NewPooledBuffer()
	defer buf.Close()

	if _, err := io.CopyN(buf, rdr, size); err != nil {
		return nil, fmt.Errorf("error reading blob: %w", err)
	}

	blob := &pb.Blob{}

	if err := proto.Unmarshal(buf.Bytes(), blob); err != nil {
		return nil, fmt.Errorf("error unmarshalling blob: %w", err)
	}

	return blob, nil
}
