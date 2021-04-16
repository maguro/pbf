// Copyright 2017-21 the original author or authors.
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

package pbf // import "m4o.io/pbf"

import (
	"bytes"
	"compress/zlib"
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"runtime"
	"time"

	"google.golang.org/protobuf/proto"

	"m4o.io/pbf/protobuf"
)

const (
	// DefaultBufferSize is the default buffer size for protobuf un-marshaling
	DefaultBufferSize = 1024 * 1024

	// DefaultInputChannelLength is the default channel length of raw blobs
	DefaultInputChannelLength = 16

	// DefaultOutputChannelLength is the default channel length of decoded arrays of element
	DefaultOutputChannelLength = 8

	// DefaultDecodedChannelLength is the default channel length of decoded elements coalesced from output channels
	DefaultDecodedChannelLength = 8000
)

// DefaultNCpu provides the default number of CPUs.
func DefaultNCpu() uint16 {
	return uint16(runtime.GOMAXPROCS(-1))
}

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
	Header Header
	pairs  chan pair
	cancel context.CancelFunc
}

// decoderOptions provides optional configuration parameters for Decoder construction.
type decoderOptions struct {
	protoBufferSize      int    // buffer size for protobuf un-marshaling
	inputChannelLength   int    // channel length of raw blobs
	outputChannelLength  int    // channel length of decoded arrays of element
	decodedChannelLength int    // channel length of decoded elements coalesced from output channels
	nCPU                 uint16 // the number of CPUs to use for background processing
}

// DecoderOption configures how we set up the decoder.
type DecoderOption func(*decoderOptions)

// WithProtoBufferSize lets you set the buffer size for protobuf un-marshaling.
func WithProtoBufferSize(s int) DecoderOption {
	return func(o *decoderOptions) {
		o.protoBufferSize = s
	}
}

// WithInputChannelLength lets you set the channel length of raw blobs.
func WithInputChannelLength(l int) DecoderOption {
	return func(o *decoderOptions) {
		o.inputChannelLength = l
	}
}

// WithOutputChannelLength lets you set the channel length of decoded arrays of element.
func WithOutputChannelLength(l int) DecoderOption {
	return func(o *decoderOptions) {
		o.outputChannelLength = l
	}
}

// WithDecodedChannelLength lets you set the channel length of decoded elements coalesced from output channels.
func WithDecodedChannelLength(l int) DecoderOption {
	return func(o *decoderOptions) {
		o.decodedChannelLength = l
	}
}

// WithNCpus lets you set the number of CPUs to use for background processing.
func WithNCpus(n uint16) DecoderOption {
	return func(o *decoderOptions) {
		o.nCPU = n
	}
}

// defaultDecoderConfig provides a default configuration for decoders.
var defaultDecoderConfig = decoderOptions{
	protoBufferSize:      DefaultBufferSize,
	inputChannelLength:   DefaultInputChannelLength,
	outputChannelLength:  DefaultOutputChannelLength,
	decodedChannelLength: DefaultDecodedChannelLength,
	nCPU:                 DefaultNCpu(),
}

// NewDecoder returns a new decoder, configured with cfg, that reads from
// reader.  The decoder is initialized with the OSM header.
func NewDecoder(ctx context.Context, reader io.Reader, opts ...DecoderOption) (*Decoder, error) {
	d := &Decoder{}
	c := defaultDecoderConfig

	for _, opt := range opts {
		opt(&c)
	}

	ctx, d.cancel = context.WithCancel(ctx)

	r := newBlobReader(reader)
	buf := bytes.NewBuffer(make([]byte, 0, c.protoBufferSize))

	h, err := r.readBlobHeader(buf)
	if err != nil {
		return nil, err
	}

	b, err := r.readBlob(buf, h)
	if err != nil {
		return nil, err
	}

	e, err := elements(h, b, bytes.NewBuffer(make([]byte, 0, 1024)))
	if err != nil {
		return nil, err
	}

	if e[0].(*Header) == nil {
		err = fmt.Errorf("expected header data but got %v", reflect.TypeOf(e[0]))
		return nil, err
	}

	d.Header = *e[0].(*Header)

	// create decoding pipelines
	var outputs []chan decoded
	for _, input := range read(ctx, r, c) {
		outputs = append(outputs, decode(input, c))
	}

	d.pairs = coalesce(c, outputs...)

	return d, nil
}

// Decode reads the next OSM object and returns either a pointer to Node, Way
// or Relation struct representing the underlying OpenStreetMap PBF data, or
// error encountered. The end of the input stream is reported by an io.EOF
// error.
func (d *Decoder) Decode() (interface{}, error) {
	decoded, more := <-d.pairs
	if !more {
		return nil, io.EOF
	}

	return decoded.element, decoded.err
}

// Close will cancel the background decoding pipeline.
func (d *Decoder) Close() {
	d.cancel()
}

// read obtains OSM blobs and sends them down, in a round-robin manner, a list
// of channels to be decoded.
func read(ctx context.Context, b blobReader, cfg decoderOptions) (inputs []chan encoded) {
	n := cfg.nCPU
	for i := uint16(0); i < n; i++ {
		inputs = append(inputs, make(chan encoded, cfg.inputChannelLength))
	}

	go func() {
		defer func() {
			for _, input := range inputs {
				close(input)
			}
		}()

		buffer := bytes.NewBuffer(make([]byte, 0, cfg.protoBufferSize))

		var i uint16

		for {
			input := inputs[i]
			i = (i + 1) % n

			h, err := b.readBlobHeader(buffer)
			if err == io.EOF {
				return
			} else if err != nil {
				input <- encoded{err: err}
				return
			}

			b, err := b.readBlob(buffer, h)
			if err != nil {
				input <- encoded{err: err}
				return
			}

			select {
			case <-ctx.Done():
				return
			case input <- encoded{header: h, blob: b}:
			}
		}
	}()

	return inputs
}

// decode decodes blob/header pairs into an array of OSM elements.  These
// arrays are placed onto an output channel where they will be coalesced into
// their correct order.
func decode(input <-chan encoded, cfg decoderOptions) (output chan decoded) {
	output = make(chan decoded, cfg.outputChannelLength)

	buf := bytes.NewBuffer(make([]byte, 0, cfg.protoBufferSize))

	go func() {
		defer close(output)

		for {
			raw, more := <-input
			if !more {
				return
			}

			if raw.err != nil {
				output <- decoded{nil, raw.err}
				return
			}

			elements, err := elements(raw.header, raw.blob, buf)

			output <- decoded{elements, err}
		}
	}()

	return
}

// coalesce merges the list of channels in a round-robin manner and sends the
// elements in pairs down a channel of pairs.
func coalesce(cfg decoderOptions, outputs ...chan decoded) (pairs chan pair) {
	pairs = make(chan pair, cfg.decodedChannelLength)

	go func() {
		defer close(pairs)

		n := len(outputs)

		var i int

		for {
			output := outputs[i]
			i = (i + 1) % n

			decoded, more := <-output
			if !more {
				// Since the channels are inspected round-robin, when one channel
				// is done, all subsequent channels are done.
				return
			}

			if decoded.err != nil {
				pairs <- pair{nil, decoded.err}
				return
			}

			for _, e := range decoded.elements {
				pairs <- pair{e, nil}
			}
		}
	}()

	return pairs
}

// elements unmarshals an array of OSM elements from an array of protobuf encoded
// bytes.  The bytes could possibly be compressed; zlibBuf is used to facilitate
// decompression.
func elements(header *protobuf.BlobHeader, blob *protobuf.Blob, zlibBuf *bytes.Buffer) ([]interface{}, error) {
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

	switch ht {
	case "OSMHeader":
		{
			h, err := parseOSMHeader(buf)
			if err != nil {
				return nil, err
			}
			return []interface{}{h}, nil
		}
	case "OSMData":
		return parsePrimitiveBlock(buf)
	default:
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
		RequiredFeatures:                 hb.GetRequiredFeatures(),
		OptionalFeatures:                 hb.GetOptionalFeatures(),
		WritingProgram:                   hb.GetWritingprogram(),
		Source:                           hb.GetSource(),
		OsmosisReplicationBaseURL:        hb.GetOsmosisReplicationBaseUrl(),
		OsmosisReplicationSequenceNumber: hb.GetOsmosisReplicationSequenceNumber(),
	}

	if hb.Bbox != nil {
		header.BoundingBox = BoundingBox{
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
