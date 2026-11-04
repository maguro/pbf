// SPDX-FileCopyrightText: the original author or authors
// SPDX-License-Identifier: Apache-2.0

package pbf

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/destel/rill"

	"m4o.io/pbf/v2/internal/decoder"
	"m4o.io/pbf/v2/model"
)

// ErrUnsupportedRequiredFeature is returned when a PBF header contains a required feature
// that this decoder does not support.
var ErrUnsupportedRequiredFeature = errors.New("unsupported required feature")

// Decoder reads and decodes OpenStreetMap PBF data from an input stream.
type Decoder struct {
	Header   model.Header
	Entities <-chan rill.Try[[]model.Entity]
	cancel   context.CancelFunc
}

// NewDecoder returns a new decoder, configured with opts, that reads from r.
// The decoder is initialized with the OSM header.
func NewDecoder(ctx context.Context, r io.Reader, opts ...DecoderOption) (*Decoder, error) {
	d := &Decoder{}
	cfg := defaultDecoderConfig

	for _, opt := range opts {
		opt(&cfg)
	}

	ctx, d.cancel = context.WithCancel(ctx)

	hdr, err := decoder.LoadHeader(r)
	if err != nil {
		return nil, err
	}

	for _, feature := range hdr.RequiredFeatures {
		if !isSupportedRequiredFeature(feature) {
			return nil, fmt.Errorf("%w: %s", ErrUnsupportedRequiredFeature, feature)
		}
	}

	d.Header = hdr

	blobs := rill.FromSeq2(decoder.GenerateBlobReader(ctx, r))

	batches := rill.Batch(blobs, cfg.protoBatchSize, time.Second)

	entities := rill.OrderedFlatMap(batches, int(cfg.nCPU), decoder.DecodeBatch)

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
	rill.Discard(d.Entities)
}

func isSupportedRequiredFeature(feature string) bool {
	switch feature {
	case featureOSMSchema, featureDenseNodes, featureHistoricalInformation:
		return true
	default:
		return false
	}
}
