// SPDX-FileCopyrightText: the original author or authors
// SPDX-License-Identifier: Apache-2.0

package packers

import (
	"bytes"
	"sync"

	"github.com/klauspost/compress/zstd"

	"m4o.io/pbf/v2/internal/pb"
)

// zstdEncoders holds idle zstd encoders. An encoder keeps its history buffer
// when it goes back to the pool.
var zstdEncoders = sync.Pool{
	New: func() any {
		w, err := zstd.NewWriter(nil)
		if err != nil {
			panic(err)
		}

		return w
	},
}

type ZstdPacker struct {
	w   *zstd.Encoder
	buf bytes.Buffer
}

func NewZstdPacker() *ZstdPacker {
	p := ZstdPacker{}

	p.w = zstdEncoders.Get().(*zstd.Encoder) //nolint:forcetypeassert // The pool holds only encoders.
	p.w.Reset(&p.buf)

	return &p
}

// Write gives b to the zstd encoder. The encoder writes the compressed data to
// the buffer of the packer.
func (p *ZstdPacker) Write(b []byte) (int, error) {
	return p.w.Write(b)
}

// Close writes the remaining compressed data to the buffer. Then Close puts the
// encoder back in the pool. Do not call Write or Close after Close.
func (p *ZstdPacker) Close() error {
	err := p.w.Close()

	p.w.Reset(nil)
	zstdEncoders.Put(p.w)
	p.w = nil

	return err
}

func (p *ZstdPacker) SaveTo(blob *pb.Blob) {
	blob.Data = &pb.Blob_ZstdData{ZstdData: p.buf.Bytes()}
}
