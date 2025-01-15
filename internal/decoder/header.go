// Copyright 2017-25 the original author or authors.
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

package decoder

import (
	"fmt"
	"io"
	"reflect"

	"m4o.io/pbf/v2/internal/core"
	"m4o.io/pbf/v2/model"
)

func LoadHeader(reader io.Reader) (model.Header, error) {
	buf := core.NewPooledBuffer()
	defer buf.Close()

	h, err := readBlobHeader(buf, reader)
	if err != nil {
		return model.Header{}, err
	}

	b, err := readBlob(buf, reader, h)
	if err != nil {
		return model.Header{}, err
	}

	e, err := extract(h, b)
	if err != nil {
		return model.Header{}, err
	}

	if hdr, ok := e[0].(*model.Header); !ok {
		err = fmt.Errorf("expected header data but got %v", reflect.TypeOf(e[0]))

		return model.Header{}, err
	} else {
		return *hdr, nil
	}
}
