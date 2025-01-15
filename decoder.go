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
	"context"
	"fmt"
	"io"
	"reflect"
	"runtime"
	"time"

	"github.com/destel/rill"

	"m4o.io/pbf/v2/internal/core"
)

const (
	// DefaultBufferSize is the default buffer size for protobuf un-marshaling.
	DefaultBufferSize = 1024 * 1024

	// DefaultBatchSize is the default batch size for unprocessed blobs.
	DefaultBatchSize = 16

	coordinatesPerDegree = 1e-9
)

// DefaultNCpu provides the default number of CPUs.
func DefaultNCpu() uint16 {
	cpus := uint16(runtime.GOMAXPROCS(-1))

	return max(cpus-1, 1)
}

// Decoder reads and decodes OpenStreetMap PBF data from an input stream.
type Decoder struct {
	Header  Header
	Objects <-chan rill.Try[[]Object]
	cancel  context.CancelFunc
}

// decoderOptions provides optional configuration parameters for Decoder construction.
type decoderOptions struct {
	protoBufferSize int    // buffer size for protobuf un-marshaling
	protoBatchSize  int    // batch size for protobuf un-marshaling
	nCPU            uint16 // the number of CPUs to use for background processing
}

// DecoderOption configures how we set up the decoder.
type DecoderOption func(*decoderOptions)

// WithProtoBufferSize lets you set the buffer size for protobuf un-marshaling.
func WithProtoBufferSize(s int) DecoderOption {
	return func(o *decoderOptions) {
		o.protoBufferSize = s
	}
}

// WithProtoBatchSize lets you set the buffer size for protobuf un-marshaling.
func WithProtoBatchSize(s int) DecoderOption {
	return func(o *decoderOptions) {
		o.protoBatchSize = s
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
	protoBufferSize: DefaultBufferSize,
	protoBatchSize:  DefaultBatchSize,
	nCPU:            DefaultNCpu(),
}

// NewDecoder returns a new decoder, configured with cfg, that reads from
// reader.  The decoder is initialized with the OSM header.
func NewDecoder(ctx context.Context, rdr io.Reader, opts ...DecoderOption) (*Decoder, error) {
	d := &Decoder{}
	cfg := defaultDecoderConfig

	for _, opt := range opts {
		opt(&cfg)
	}

	ctx, d.cancel = context.WithCancel(ctx)

	if err := d.loadHeader(rdr); err != nil {
		return nil, err
	}

	blobs := rill.FromSeq2(generate(ctx, rdr))

	batches := rill.Batch(blobs, cfg.protoBatchSize, time.Second)

	objects := rill.FlatMap(batches, int(cfg.nCPU), decode)

	d.Objects = objects

	return d, nil
}

// Decode reads the next OSM object and returns either a pointer to Node, Way
// or Relation struct representing the underlying OpenStreetMap PBF data, or
// error encountered. The end of the input stream is reported by an io.EOF
// error.
func (d *Decoder) Decode() ([]Object, error) {
	decoded, more := <-d.Objects
	if !more {
		return nil, io.EOF
	}

	return decoded.Value, decoded.Error
}

// Close will cancel the background decoding pipeline.
func (d *Decoder) Close() {
	rill.DrainNB(d.Objects)

	d.cancel()
}

func (d *Decoder) loadHeader(reader io.Reader) error {
	buf := core.NewPooledBuffer()
	defer buf.Close()

	h, err := readBlobHeader(buf, reader)
	if err != nil {
		return err
	}

	b, err := readBlob(buf, reader, h)
	if err != nil {
		return err
	}

	e, err := extract(h, b)
	if err != nil {
		return err
	}

	if hdr, ok := e[0].(*Header); !ok {
		err = fmt.Errorf("expected header data but got %v", reflect.TypeOf(e[0]))

		return err
	} else {
		d.Header = *hdr
	}

	return nil
}
