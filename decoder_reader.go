// Copyright 2017-24 the original author or authors.
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

package pbf

import (
	"bytes"
	"encoding/binary"
	"io"

	"google.golang.org/protobuf/proto"

	"m4o.io/pbf/protobuf"
)

type blobReader struct {
	r io.Reader
}

func newBlobReader(r io.Reader) blobReader {
	return blobReader{r: r}
}

// readBlobHeader unmarshals a header from an array of protobuf encoded bytes.
// The header is used when decoding blobs into OSM elements.
func (b blobReader) readBlobHeader(buffer *bytes.Buffer) (header *protobuf.BlobHeader, err error) {
	var size uint32

	err = binary.Read(b.r, binary.BigEndian, &size)
	if err != nil {
		return nil, err
	}

	buffer.Reset()

	if _, err := io.CopyN(buffer, b.r, int64(size)); err != nil {
		return nil, err
	}

	header = &protobuf.BlobHeader{}

	if err := proto.Unmarshal(buffer.Bytes(), header); err != nil {
		return nil, err
	}

	return header, nil
}

// readBlob unmarshals a blob from an array of protobuf encoded bytes.  The
// blob still needs to be decoded into OSM elements using decode().
func (b blobReader) readBlob(buffer *bytes.Buffer, header *protobuf.BlobHeader) (*protobuf.Blob, error) {
	size := header.GetDatasize()

	buffer.Reset()

	if _, err := io.CopyN(buffer, b.r, int64(size)); err != nil {
		return nil, err
	}

	blob := &protobuf.Blob{}

	if err := proto.Unmarshal(buffer.Bytes(), blob); err != nil {
		return nil, err
	}

	return blob, nil
}
