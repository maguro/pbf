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
	"time"

	"github.com/destel/rill"

	"m4o.io/pbf/v2/internal/decoder"
	"m4o.io/pbf/v2/model"
)

// Decoder reads and decodes OpenStreetMap PBF data from an input stream.
type Decoder struct {
	Header   model.Header
	Entities <-chan rill.Try[[]model.Entity]
	cancel   context.CancelFunc
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

	if hdr, err := decoder.LoadHeader(rdr); err != nil {
		return nil, err
	} else {
		d.Header = hdr
	}

	blobs := rill.FromSeq2(decoder.GenerateBlobReader(ctx, rdr))

	batches := rill.Batch(blobs, cfg.protoBatchSize, time.Second)

	entities := rill.FlatMap(batches, int(cfg.nCPU), decoder.DecodeBatch)

	d.Entities = entities

	return d, nil
}

// Decode reads the next OSM object and returns either a pointer to Node, Way
// or Relation struct representing the underlying OpenStreetMap PBF data, or
// error encountered. The end of the input stream is reported by an io.EOF
// error.
func (d *Decoder) Decode() ([]model.Entity, error) {
	decoded, more := <-d.Entities
	if !more {
		return nil, io.EOF
	}

	return decoded.Value, decoded.Error
}

// Close will cancel the background decoding pipeline.
func (d *Decoder) Close() {
	d.cancel()
	rill.DrainNB(d.Entities)
}
