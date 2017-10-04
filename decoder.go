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
	protoBufferSize      int
	inputChannelLength   int
	outputChannelLength  int
	decodedChannelLength int
	ncpu                 uint16
	decoded              chan pair
	done                 chan struct{}
	begin                sync.Once
	end                  sync.Once

	reader io.Reader

	Header *Header
}

// DecoderConfig provides optional configuration parameters for Decoder construction.
type DecoderConfig struct {
	ProtoBufferSize      int    // buffer size for protobuf un-marshaling
	InputChannelLength   int    // channel length of raw blobs
	OutputChannelLength  int    // channel length of decoded arrays of element
	DecodedChannelLength int    // channel length of decoded elements coalesced from output channels
	NCpu                 uint16 // the number of CPUs to use for background processing
}

// DefaultConfig provides a default configuration.
var DefaultConfig = DecoderConfig{}

// NewDecoder returns a new decoder that reads from r.
func NewDecoder(r io.Reader, cfg DecoderConfig) (*Decoder, error) {
	d := &Decoder{
		protoBufferSize:      initialBufferSize,
		inputChannelLength:   16,
		outputChannelLength:  8,
		decodedChannelLength: 8000,
		ncpu:                 uint16(runtime.GOMAXPROCS(-1)),
		reader:               r,
	}

	if cfg.ProtoBufferSize > 0 {
		d.protoBufferSize = cfg.ProtoBufferSize
	}
	if cfg.InputChannelLength > 0 {
		d.inputChannelLength = cfg.InputChannelLength
	}
	if cfg.OutputChannelLength > 0 {
		d.outputChannelLength = cfg.OutputChannelLength
	}
	if cfg.DecodedChannelLength > 0 {
		d.decodedChannelLength = cfg.DecodedChannelLength
	}
	if cfg.NCpu > 0 {
		d.ncpu = cfg.NCpu
	}

	buf := bytes.NewBuffer(make([]byte, 0, d.protoBufferSize))

	h, err := d.readBlobHeader(buf)
	if err != nil {
		return nil, err
	}
	b, err := d.readBlob(buf, h)
	if err != nil {
		return nil, err
	}
	elements, err := decode(h, b, bytes.NewBuffer(make([]byte, 0, 1024)))
	if err != nil {
		return nil, err
	}
	d.Header = elements[0].(*Header)
	if d.Header == nil {
		err = fmt.Errorf("expected header data but got %v", reflect.TypeOf(elements[0]))
		return nil, err
	}

	return d, nil
}

// start begins parsing in the background using n goroutines.  The background
// processing can be canceled by calling Stop.
func (d *Decoder) start() {
	n := int(d.ncpu)

	d.begin.Do(func() {
		inputs := make([]chan<- encoded, n)
		outputs := make([]<-chan decoded, n)
		d.decoded = make(chan pair, d.decodedChannelLength)
		d.done = make(chan struct{})

		// start data decoders
		for i := 0; i < n; i++ {
			input := make(chan encoded, d.inputChannelLength)
			output := make(chan decoded, d.outputChannelLength)

			go func() {
				defer close(output)

				buf := bytes.NewBuffer(make([]byte, 0, d.protoBufferSize))

				for {
					raw, more := <-input
					if !more {
						return
					}

					if raw.err != nil {
						output <- decoded{nil, raw.err}
						return
					}

					elements, err := decode(raw.header, raw.blob, buf)

					output <- decoded{elements, err}
				}
			}()

			inputs[i] = input
			outputs[i] = output
		}

		// read raw blobs and distribute amongst inputs
		go func() {
			defer func() {
				for _, input := range inputs {
					close(input)
				}
			}()

			buffer := bytes.NewBuffer(make([]byte, 0, d.protoBufferSize))
			var i int
			for {
				input := inputs[i]
				i = (i + 1) % n

				h, err := d.readBlobHeader(buffer)
				if err == io.EOF {
					return
				} else if err != nil {
					input <- encoded{err: err}
					return
				}

				b, err := d.readBlob(buffer, h)
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
				output := outputs[i]
				i = (i + 1) % n

				decoded, more := <-output
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
		}()
	})
}

// Stop will cancel the background decoding pipeline.
func (d *Decoder) Stop() {

	d.begin.Do(func() {
		// close decoded so calls to Decode return EOF
		close(d.decoded)
	})

	d.end.Do(func() {
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

	d.start()

	decoded, more := <-d.decoded
	if !more {
		return nil, io.EOF
	}
	return decoded.element, decoded.err
}

// readBlobHeader unmarshals a header from an array of protobuf encoded bytes.
// The header is used when decoding blobs into OSM elements.
func (d *Decoder) readBlobHeader(buffer *bytes.Buffer) (header *protobuf.BlobHeader, err error) {
	var size uint32
	err = binary.Read(d.reader, binary.BigEndian, &size)
	if err != nil {
		return nil, err
	}

	buffer.Reset()
	if _, err := io.CopyN(buffer, d.reader, int64(size)); err != nil {
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
func (d *Decoder) readBlob(buffer *bytes.Buffer, header *protobuf.BlobHeader) (*protobuf.Blob, error) {
	size := header.GetDatasize()

	buffer.Reset()
	if _, err := io.CopyN(buffer, d.reader, int64(size)); err != nil {
		return nil, err
	}

	blob := &protobuf.Blob{}
	if err := proto.Unmarshal(buffer.Bytes(), blob); err != nil {
		return nil, err
	}

	return blob, nil
}

// decode unmarshals an array of OSM elements from an array of protobuf encoded
// bytes.  The bytes could possibly be compressed; zlibBuf is used to facilitate
// decompression.
func decode(header *protobuf.BlobHeader, blob *protobuf.Blob, zlibBuf *bytes.Buffer) ([]interface{}, error) {
	var buf []byte
	switch {
	case blob.Raw != nil:
		buf = blob.GetRaw()

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
		buf = zlibBuf.Bytes()

	default:
		return nil, errors.New("unknown blob data type")
	}

	ht := *header.Type
	if ht == "OSMHeader" {
		h, err := parseOSMHeader(buf)
		if err != nil {
			return nil, err
		}
		return []interface{}{h}, nil
	} else if ht == "OSMData" {
		return parsePrimitiveBlock(buf)
	} else {
		return nil, fmt.Errorf("unknown header type %s", ht)
	}
}

// parseOSMHeader unmarshals the OSM header from an array of protobuf encoded bytes.
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
			Left:   toDegrees(0, 1, hb.Bbox.GetLeft()),
			Right:  toDegrees(0, 1, hb.Bbox.GetRight()),
			Top:    toDegrees(0, 1, hb.Bbox.GetTop()),
			Bottom: toDegrees(0, 1, hb.Bbox.GetBottom()),
		}
	}

	if hb.OsmosisReplicationTimestamp != nil {
		header.OsmosisReplicationTimestamp = time.Unix(*hb.OsmosisReplicationTimestamp, 0)
	}

	return header, nil
}

// toDegrees converts a coordinate into Degrees, given the offset and
// granularity of the coordinate.
func toDegrees(offset int64, granularity int32, coordinate int64) Degrees {
	return 1e-9 * Degrees(offset+(int64(granularity)*coordinate))
}
