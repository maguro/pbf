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
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/maguro/pbf/protobuf"
)

const (
	initialBufferSize = 1024 * 1024
)

type encoded struct {
	header *protobuf.BlobHeader
	blob   *protobuf.Blob
	err    error
}

type decoded struct {
	elements []interface{}
	err      error
}

type pair struct {
	element interface{}
	err     error
}

// Decoder reads and decodes OpenStreetMap PBF data from an input stream.
type Decoder struct {
	InputBufferSize      int
	ZlibBufferSize       int
	OutputChannelLength  int
	DecodedChannelLength int
	inputs               []chan<- encoded
	outputs              []<-chan decoded
	decoded              chan pair
	done                 chan struct{}
	start                sync.Once
	stop                 sync.Once

	reader io.Reader
	buffer *bytes.Buffer

	Header *Header
}

// NewDecoder returns a new decoder that reads from r.
func NewDecoder(r io.Reader) (*Decoder, error) {
	decoder := &Decoder{
		InputBufferSize:      16,
		ZlibBufferSize:       initialBufferSize,
		OutputChannelLength:  8,
		DecodedChannelLength: 8000,
		reader:               r,
		buffer:               bytes.NewBuffer(make([]byte, 0, initialBufferSize)),
	}

	bh, err := decoder.readBlobHeader()
	if err != nil {
		return nil, err
	}
	b, err := decoder.readBlob(bh)
	if err != nil {
		return nil, err
	}
	elements, err := decode(bh, b, bytes.NewBuffer(make([]byte, 0, 1024)))
	if err != nil {
		return nil, err
	}
	decoder.Header = elements[0].(*Header)
	if decoder.Header == nil {
		err = fmt.Errorf("expected header data but got %v", reflect.TypeOf(elements[0]))
		return nil, err
	}

	return decoder, nil
}

// SetBufferSize sets initial size of decoding buffer. Any value will produce
// valid results; buffer will grow automatically if required.
func (d *Decoder) SetBufferSize(n int) {
	d.buffer = bytes.NewBuffer(make([]byte, 0, n))
}

// Start begins parsing in the background using n goroutines.  The background
// processing can be canceled by calling Stop.
func (d *Decoder) Start(n int) {
	if n < 1 {
		n = 1
	}

	d.start.Do(func() {
		d.inputs = make([]chan<- encoded, n)
		d.outputs = make([]<-chan decoded, n)
		d.decoded = make(chan pair, d.DecodedChannelLength)
		d.done = make(chan struct{})

		// start data decoders
		for i := range d.inputs {
			input := make(chan encoded, d.InputBufferSize)
			output := make(chan decoded, d.OutputChannelLength)

			go func() {
				defer close(output)

				zlibBuffer := bytes.NewBuffer(make([]byte, 0, d.ZlibBufferSize))

				for {
					raw, more := <-input
					if !more {
						return
					}

					if raw.err != nil {
						output <- decoded{nil, raw.err}
						return
					}

					elements, err := decode(raw.header, raw.blob, zlibBuffer)

					output <- decoded{elements, err}
				}
			}()

			d.inputs[i] = input
			d.outputs[i] = output
		}

		// read raw blobs and distribute amongst inputs
		go func() {
			defer func() {
				for _, input := range d.inputs {
					close(input)
				}
			}()

			var i int
			for {
				input := d.inputs[i]
				i = (i + 1) % n

				h, err := d.readBlobHeader()
				if err == io.EOF {
					return
				} else if err != nil {
					input <- encoded{err: err}
					return
				}

				b, err := d.readBlob(h)
				if err != nil {
					input <- encoded{err: err}
					return
				}

				select {
				case <-d.done:
					return
				case input <- encoded{header: h, blob: b}:
				}
			}
		}()

		// coalesce decoded elements
		go func() {
			defer close(d.decoded)

			var i int
			for {
				output := d.outputs[i]
				i = (i + 1) % n

				select {
				case decoded, more := <-output:
					if !more {
						return
					}

					if decoded.err != nil {
						d.decoded <- pair{nil, decoded.err}
						return
					}

					for _, e := range decoded.elements {
						d.decoded <- pair{e, nil}
					}
				}
			}
		}()
	})
}

// Stop will cancel the background decoding pipeline.
func (d *Decoder) Stop() {

	d.start.Do(func() {
		// close decoded so calls to Decode return EOF
		close(d.decoded)
	})

	d.stop.Do(func() {
		if d.done != nil {
			// closing done notifies pipeline to cancel
			close(d.done)
		}
	})
}

// Decode reads the next OSM object and returns either a pointer to Node, Way
// or Relation struct representing the underlying OpenStreetMap PBF data, or
// error encountered. The end of the input stream is reported by an io.EOF
// error.
//
// If background parsing was not begun by calling Start, Start is called with
// the maximum number of CPUs returned by GOMAXPROCS.
func (d *Decoder) Decode() (interface{}, error) {

	d.Start(runtime.GOMAXPROCS(-1))

	decoded, more := <-d.decoded
	if !more {
		return nil, io.EOF
	}
	return decoded.element, decoded.err
}

func (d *Decoder) readBlobHeader() (header *protobuf.BlobHeader, err error) {
	var size uint32
	err = binary.Read(d.reader, binary.BigEndian, &size)
	if err != nil {
		return nil, err
	}

	d.buffer.Reset()
	if _, err := io.CopyN(d.buffer, d.reader, int64(size)); err != nil {
		return nil, err
	}

	header = &protobuf.BlobHeader{}
	if err := proto.Unmarshal(d.buffer.Bytes(), header); err != nil {
		return nil, err
	}

	return header, nil
}

func (d *Decoder) readBlob(header *protobuf.BlobHeader) (*protobuf.Blob, error) {
	size := header.GetDatasize()

	d.buffer.Reset()
	if _, err := io.CopyN(d.buffer, d.reader, int64(size)); err != nil {
		return nil, err
	}

	blob := &protobuf.Blob{}
	if err := proto.Unmarshal(d.buffer.Bytes(), blob); err != nil {
		return nil, err
	}

	return blob, nil
}

func decode(header *protobuf.BlobHeader, blob *protobuf.Blob, zlibBuf *bytes.Buffer) ([]interface{}, error) {
	var buffer []byte
	switch {
	case blob.Raw != nil:
		buffer = blob.GetRaw()

	case blob.ZlibData != nil:
		r, err := zlib.NewReader(bytes.NewReader(blob.GetZlibData()))
		if err != nil {
			return nil, err
		}
		zlibBuf.Reset()
		rawBufferSize := int(blob.GetRawSize() + bytes.MinRead)
		if rawBufferSize > zlibBuf.Cap() {
			zlibBuf.Grow(rawBufferSize)
		}
		_, err = zlibBuf.ReadFrom(r)
		if err != nil {
			return nil, err
		}
		if zlibBuf.Len() != int(blob.GetRawSize()) {
			err = fmt.Errorf("raw blob data size %d but expected %d", zlibBuf.Len(), blob.GetRawSize())
			return nil, err
		}
		buffer = zlibBuf.Bytes()

	default:
		return nil, errors.New("unknown blob data type")
	}

	ht := *header.Type
	if ht == "OSMHeader" {
		h, err := parseOSMHeader(buffer)
		if err != nil {
			return nil, err
		}

		return []interface{}{h}, nil
	} else if ht == "OSMData" {
		return parsePrimitiveBlock(buffer)
	} else {
		return nil, fmt.Errorf("unknown header type %s", ht)
	}
}

func parseOSMHeader(buffer []byte) (*Header, error) {
	hb := &protobuf.HeaderBlock{}
	if err := proto.Unmarshal(buffer, hb); err != nil {
		return nil, err
	}

	header := &Header{
		RequiredFeatures: hb.GetRequiredFeatures(),
		OptionalFeatures: hb.GetOptionalFeatures(),
		WritingProgram:   hb.GetWritingprogram(),
		Source:           hb.GetSource(),
		OsmosisReplicationBaseURL:        hb.GetOsmosisReplicationBaseUrl(),
		OsmosisReplicationSequenceNumber: hb.GetOsmosisReplicationSequenceNumber(),
	}

	if hb.Bbox != nil {
		header.BoundingBox = &BoundingBox{
			Left:   toDecimalDegrees(0, 1, hb.Bbox.GetLeft()),
			Right:  toDecimalDegrees(0, 1, hb.Bbox.GetRight()),
			Top:    toDecimalDegrees(0, 1, hb.Bbox.GetTop()),
			Bottom: toDecimalDegrees(0, 1, hb.Bbox.GetBottom()),
		}
	}

	if hb.OsmosisReplicationTimestamp != nil {
		header.OsmosisReplicationTimestamp = time.Unix(*hb.OsmosisReplicationTimestamp, 0)
	}

	return header, nil
}

// toDecimalDegrees converts a coordinate into DecimalDegrees, given the offset and
// granularity of the coordinate.
func toDecimalDegrees(offset int64, granularity int32, coordinate int64) DecimalDegrees {
	return NanoDecimalDegrees * DecimalDegrees(offset+(int64(granularity)*coordinate))
}
