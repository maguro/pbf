// SPDX-FileCopyrightText: the original author or authors
// SPDX-License-Identifier: Apache-2.0

package decoder

import (
	"bytes"
	"compress/zlib"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/klauspost/compress/zstd"
	"github.com/pierrec/lz4"
	"github.com/ulikunitz/xz/lzma"

	"m4o.io/pbf/v2/internal/core"
	"m4o.io/pbf/v2/internal/pb"
)

var ErrUnknownCompressionType = errors.New("unknown blob compression type")

// zstdDecoders holds idle zstd decoders. A decoder keeps its history buffer
// when it goes back to the pool. A decoder in the pool decodes in the goroutine
// that calls it. It starts no goroutines.
var zstdDecoders = sync.Pool{
	New: func() any {
		d, err := zstd.NewReader(nil, zstd.WithDecoderConcurrency(1))
		if err != nil {
			panic(err)
		}

		return d
	},
}

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
			d := zstdDecoders.Get().(*zstd.Decoder) //nolint:forcetypeassert // The pool holds only decoders.

			if err := d.Reset(bytes.NewReader(b.GetZstdData())); err != nil {
				return nil, fmt.Errorf("could not reset zstd decoder: %w", err)
			}

			return d, nil
		}
	default:
		return nil, ErrUnknownCompressionType
	}

	rawBufferSize := int(blob.GetRawSize() + bytes.MinRead)
	if rawBufferSize > buf.Cap() {
		buf.Grow(rawBufferSize)
	}

	r, err := factory(blob)
	if err != nil {
		return nil, fmt.Errorf("unpacker factory error: %w", err)
	}

	if d, ok := r.(*zstd.Decoder); ok {
		defer releaseZstdDecoder(d)
	}

	if n, err := buf.ReadFrom(r); err != nil {
		return nil, fmt.Errorf("unpacker read error: %w", err)
	} else if n != int64(blob.GetRawSize()) {
		return nil, fmt.Errorf("raw blob data size %d but expected %d", buf.Len(), blob.GetRawSize())
	}

	return buf.Bytes(), nil
}

// releaseZstdDecoder clears the stream of d. Then releaseZstdDecoder puts d
// back in the pool.
func releaseZstdDecoder(d *zstd.Decoder) {
	_ = d.Reset(nil)

	zstdDecoders.Put(d)
}
