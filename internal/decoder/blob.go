// SPDX-FileCopyrightText: the original author or authors
// SPDX-License-Identifier: Apache-2.0

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
func GenerateBlobReader(ctx context.Context, r io.Reader) func(yield func(enc *pb.Blob, err error) bool) {
	return func(yield func(enc *pb.Blob, err error) bool) {
		buffer := core.NewPooledBuffer()
		defer buffer.Close()

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			blob, err := readBlob(r)
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

// readBlob reads a PBF blob from r.
func readBlob(r io.Reader) (*pb.Blob, error) {
	h, err := readBlobHeader(r)
	if err != nil {
		return nil, fmt.Errorf("error reading blob header: %w", err)
	}

	b, err := readBlobData(r, int64(h.GetDatasize()))
	if err != nil {
		return nil, fmt.Errorf("error reading blob: %w", err)
	}

	return b, nil
}

// readBlobHeader unmarshals a header from an array of protobuf encoded bytes.
// The header is used when decoding blobs into OSM entities.
func readBlobHeader(r io.Reader) (header *pb.BlobHeader, err error) {
	buf := core.NewPooledBuffer()
	defer buf.Close()

	var size uint32

	err = binary.Read(r, binary.BigEndian, &size)
	if err != nil {
		return nil, fmt.Errorf("error reading blob size: %w", err)
	}

	if n, err := io.CopyN(buf, r, int64(size)); err != nil {
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
func readBlobData(r io.Reader, size int64) (*pb.Blob, error) {
	buf := core.NewPooledBuffer()
	defer buf.Close()

	if _, err := io.CopyN(buf, r, size); err != nil {
		return nil, fmt.Errorf("error reading blob: %w", err)
	}

	blob := &pb.Blob{}

	if err := proto.Unmarshal(buf.Bytes(), blob); err != nil {
		return nil, fmt.Errorf("error unmarshalling blob: %w", err)
	}

	return blob, nil
}
