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
