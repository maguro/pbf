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

package pbf

import (
	"context"
	"errors"
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"m4o.io/pbf/v2/model"
)

func TestDecodeSample(t *testing.T) {
	publicDecodeOsmPbf(t, "testdata/sample.osm.pbf", 339)
}

func publicDecodeOsmPbf(t *testing.T, file string, expectedEntries int) {
	in, err := os.Open(file)
	if err != nil {
		t.Errorf("Error reading file: %v", err)
	}

	defer in.Close()

	// decode header blob
	decoder, err := NewDecoder(context.Background(), in)
	assert.NoError(t, err)

	assert.Equal(t, reflect.TypeOf(model.Header{}), reflect.TypeOf(decoder.Header))

	// decode entities
	var nEntries int

	for {
		objs, err := decoder.Decode()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				t.Errorf("Error decoding%v", err)
			} else {
				break
			}
		}

		for _, obj := range objs {
			assert.NotEqual(t, reflect.TypeOf(model.Header{}), reflect.TypeOf(obj))
		}

		nEntries = nEntries + len(objs)
	}

	assert.Equal(t, expectedEntries, nEntries, "Incorrect number of entities")
}
