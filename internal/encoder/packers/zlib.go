// SPDX-FileCopyrightText: the original author or authors
// SPDX-License-Identifier: Apache-2.0

package packers

import (
	"bytes"
	"compress/zlib"

	"m4o.io/pbf/v2/internal/pb"
)

type ZlibPacker struct {
	*base
	buf bytes.Buffer
}

func NewZlibPacker() *ZlibPacker {
	p := ZlibPacker{}
	p.base = newBasePacker(zlib.NewWriter(&p.buf))

	return &p
}

func (p *ZlibPacker) SaveTo(blob *pb.Blob) {
	blob.Data = &pb.Blob_ZlibData{ZlibData: p.buf.Bytes()}
}
