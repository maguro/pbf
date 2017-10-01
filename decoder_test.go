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
	"runtime/trace"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func BenchmarkLondon(b *testing.B) {
	in, err := os.Open("testdata/greater-london.osm.pbf")
	if err != nil {
		b.Errorf("Error reading file: %v", err)
	}
	defer in.Close()

	if t, err := strconv.ParseBool(os.Getenv("PBF_TRACE")); err == nil {
		if t {
			f, err := os.Create("trace.out")
			if err != nil {
				b.Errorf("Error opening trace file: %v", err)
			} else {
				defer f.Close()
				trace.Start(f)
				defer trace.Stop()
			}
		}
	}

	pbs, _ := strconv.Atoi(os.Getenv("PBF_PROTO_BUFFER_SIZE"))
	ibl, _ := strconv.Atoi(os.Getenv("PBF_INPUT_CHANNEL_LENGTH"))
	ocl, _ := strconv.Atoi(os.Getenv("PBF_OUTPUT_CHANNEL_LENGTH"))
	dcl, _ := strconv.Atoi(os.Getenv("PBF_DECODED_CHANNEL_LENGTH"))
	ncpu, _ := strconv.Atoi(os.Getenv("PBF_NCPU"))

	cfg := DecoderConfig{
		ProtoBufferSize:      pbs,
		InputChannelLength:   ibl,
		OutputChannelLength:  ocl,
		DecodedChannelLength: dcl,
		NCpu:                 uint16(ncpu),
	}

	for n := 0; n < b.N; n++ {
		if _, err = in.Seek(0, 0); err != nil {
			b.Fatal(err)
		}

		if decoder, err := NewDecoder(in, cfg); err != nil {
			b.Fatal(err)
		} else {
			for {
				if _, err := decoder.Decode(); err == io.EOF {
					break
				} else if err != nil {
					b.Fatal(err)
				}
			}
		}
	}
}

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
	decoder, err := NewDecoder(in, DefaultConfig)
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
	decoder, err := NewDecoder(in, DefaultConfig)
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
	decoder, err := NewDecoder(in, DefaultConfig)
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
