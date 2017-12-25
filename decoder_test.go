// Copyright 2017 the original author or authors.
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
	"io"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDetailedDecodeBremen(t *testing.T) {
	detailedDecodeOsmPbf(t, "testdata/bremen.osm.pbf", 207, 1640420)
}

func TestDetailedDecodeSample(t *testing.T) {
	detailedDecodeOsmPbf(t, "testdata/sample.osm.pbf", 3, 339)
}

func TestDetailedDecodeLondon(t *testing.T) {
	detailedDecodeOsmPbf(t, "testdata/greater-london.osm.pbf", 401, 3200894)
}

func TestPubicDecodeBremen(t *testing.T) {
	publicDecodeOsmPbf(t, "testdata/bremen.osm.pbf", 1640420)
}

func TestPublicDecodeSample(t *testing.T) {
	publicDecodeOsmPbf(t, "testdata/sample.osm.pbf", 339)
}

func TestPublicDecodeLondon(t *testing.T) {
	publicDecodeOsmPbf(t, "testdata/greater-london.osm.pbf", 3200894)
}

func TestDecoderStop(t *testing.T) {
	in, err := os.Open("testdata/greater-london.osm.pbf")
	if err != nil {
		t.Errorf("Error reading file: %v", err)
	}
	defer in.Close()

	// decode header blob
	decoder, err := NewDecoder(in, DfltDecoderConfig)
	if err != nil {
		t.Errorf("Error reading blob header: %v", err)
	}

	assert.Equal(t, reflect.TypeOf(&Header{}), reflect.TypeOf(decoder.Header))

	timer := time.NewTimer(time.Millisecond * 250)
	go func() {
		<-timer.C
		decoder.Stop()
		decoder.Stop()
	}()

	// decode elements
	var nEntries int
	for {
		e, err := decoder.Decode()
		if err != nil {
			if err != io.EOF {
				t.Errorf("Error decoding%v", err)
			} else {
				break
			}
		}
		assert.NotEqual(t, reflect.TypeOf(&Header{}), reflect.TypeOf(e))

		nEntries++
	}

	assert.True(t, nEntries < 3200894, "Stop did not cancel decoding")
}

func detailedDecodeOsmPbf(t *testing.T, file string, expectedBlobs int, expectedEntries int) {
	in, err := os.Open(file)
	if err != nil {
		t.Errorf("Error reading file: %v", err)
	}
	defer in.Close()

	// decode header blob
	decoder, err := NewDecoder(in, DfltDecoderConfig)
	if err != nil {
		t.Errorf("Error reading blob header: %v", err)
	}

	assert.Equal(t, reflect.TypeOf(&Header{}), reflect.TypeOf(decoder.Header))

	// decode data blobs
	var nBlobs, nEntries int
	buf := bytes.NewBuffer(make([]byte, 0, initialBufferSize))
	zbuf := bytes.NewBuffer(make([]byte, 0, initialBufferSize))
	for {
		h, err := decoder.readBlobHeader(buf)
		if err != nil {
			if err != io.EOF {
				t.Errorf("Error reading header: %v", err)
			} else {
				break
			}
		}

		b, err := decoder.readBlob(buf, h)
		if err != nil {
			t.Errorf("Error reading blob: %v", err)
		}
		nBlobs++

		entries, err := decode(h, b, zbuf)
		if err != nil {
			t.Errorf("Error reading elements: %v", err)
		}
		assert.True(t, len(entries) > 0)
		for _, entry := range entries {
			assert.NotEqual(t, reflect.TypeOf(&Header{}), reflect.TypeOf(entry))
		}
		nEntries += len(entries)
	}

	assert.Equal(t, expectedBlobs, nBlobs, "Incorrect number of blobs")
	assert.Equal(t, expectedEntries, nEntries, "Incorrect number of elements")
}

func publicDecodeOsmPbf(t *testing.T, file string, expectedEntries int) {
	in, err := os.Open(file)
	if err != nil {
		t.Errorf("Error reading file: %v", err)
	}
	defer in.Close()

	// decode header blob
	decoder, err := NewDecoder(in, DfltDecoderConfig)
	if err != nil {
		t.Errorf("Error reading blob header: %v", err)
	}

	assert.Equal(t, reflect.TypeOf(&Header{}), reflect.TypeOf(decoder.Header))

	// decode elements
	var nEntries int
	for {
		e, err := decoder.Decode()
		if err != nil {
			if err != io.EOF {
				t.Errorf("Error decoding%v", err)
			} else {
				break
			}
		}
		assert.NotEqual(t, reflect.TypeOf(&Header{}), reflect.TypeOf(e))

		nEntries++
	}

	assert.Equal(t, expectedEntries, nEntries, "Incorrect number of elements")
}
