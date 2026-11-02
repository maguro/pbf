// SPDX-FileCopyrightText: the original author or authors
// SPDX-License-Identifier: Apache-2.0

package packers

import (
	"bytes"

	"github.com/pierrec/lz4"

	"m4o.io/pbf/v2/internal/pb"
)

type Lz4Packer struct {
	*base
	buf bytes.Buffer
}

func NewLz4Packer() *Lz4Packer {
	p := Lz4Packer{}
	p.base = newBasePacker(lz4.NewWriter(&p.buf))

	return &p
}

func (p *Lz4Packer) SaveTo(blob *pb.Blob) {
	blob.Data = &pb.Blob_Lz4Data{Lz4Data: p.buf.Bytes()}
}
