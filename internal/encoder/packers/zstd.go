// SPDX-FileCopyrightText: the original author or authors
// SPDX-License-Identifier: Apache-2.0

package packers

import (
	"bytes"

	"github.com/klauspost/compress/zstd"

	"m4o.io/pbf/v2/internal/pb"
)

type ZstdPacker struct {
	*base
	buf bytes.Buffer
}

func NewZstdPacker() *ZstdPacker {
	p := ZstdPacker{}

	w, err := zstd.NewWriter(&p.buf)
	if err != nil {
		panic(err)
	}

	p.base = newBasePacker(w)

	return &p
}

func (p *ZstdPacker) SaveTo(blob *pb.Blob) {
	blob.Data = &pb.Blob_ZstdData{ZstdData: p.buf.Bytes()}
}
