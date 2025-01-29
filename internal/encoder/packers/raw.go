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
