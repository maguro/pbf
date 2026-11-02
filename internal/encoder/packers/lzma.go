// SPDX-FileCopyrightText: the original author or authors
// SPDX-License-Identifier: Apache-2.0

package packers

import (
	"bytes"

	"github.com/ulikunitz/xz/lzma"

	"m4o.io/pbf/v2/internal/pb"
)

type LzmaPacker struct {
	*base
	buf bytes.Buffer
}

func NewLzmaPacker() *LzmaPacker {
	p := LzmaPacker{}

	w, err := lzma.NewWriter(&p.buf)
	if err != nil {
		panic(err)
	}

	p.base = newBasePacker(w)

	return &p
}

func (p *LzmaPacker) SaveTo(blob *pb.Blob) {
	blob.Data = &pb.Blob_LzmaData{LzmaData: p.buf.Bytes()}
}
