// SPDX-FileCopyrightText: the original author or authors
// SPDX-License-Identifier: Apache-2.0

package packers

import (
	"bytes"
	"io"

	"m4o.io/pbf/v2/internal/pb"
)

type RawPacker struct {
	*base
	buf bytes.Buffer
}

type nopCloserWriter struct {
	io.Writer
}

func (w nopCloserWriter) Close() error {
	return nil
}

func NewRawPacker() *RawPacker {
	p := RawPacker{}
	p.base = newBasePacker(nopCloserWriter{&p.buf})

	return &p
}

func (rp *RawPacker) SaveTo(blob *pb.Blob) {
	blob.Data = &pb.Blob_Raw{Raw: rp.buf.Bytes()}
}
