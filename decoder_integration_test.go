//go:build integration
// +build integration

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
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeBremen(t *testing.T) {
	publicDecodeOsmPbf(t, "testdata/bremen.osm.pbf", 1640420)
}

func TestDecodeLondon(t *testing.T) {
	publicDecodeOsmPbf(t, "testdata/greater-london.osm.pbf", 3200894)
}

func TestDecoderStop(t *testing.T) {
	in, err := os.Open("testdata/greater-london.osm.pbf")
	if err != nil {
		t.Errorf("Error reading file: %v", err)
	}

	defer in.Close()

	// decode header blob
	decoder, err := NewDecoder(context.Background(), in)
	if err != nil {
		t.Errorf("Error reading blob header: %v", err)
	}

	assert.Equal(t, reflect.TypeOf(Header{}), reflect.TypeOf(decoder.Header))

	cancel := make(chan bool, 1)

	go func() {
		<-cancel
		decoder.Close()
		decoder.Close()
	}()

	// decode elements
	var nEntries int

	for {
		if nEntries == 1000 {
			cancel <- true
		}

		e, err := decoder.Decode()
		if err != nil {
			if err != io.EOF {
				t.Errorf("Error decoding%v", err)
			} else {
				break
			}
		}

		assert.NotEqual(t, reflect.TypeOf(Header{}), reflect.TypeOf(e))

		nEntries++
	}

	assert.True(t, nEntries < 3200894, "Close did not cancel decoding")
}
