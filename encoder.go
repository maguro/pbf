// SPDX-FileCopyrightText: the original author or authors
// SPDX-License-Identifier: Apache-2.0

package pbf

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"

	"github.com/destel/rill"

	"m4o.io/pbf/v2/internal/encoder"
	"m4o.io/pbf/v2/model"
)

const (
	numConsumers = 2

	singleCPU = 5
)

// Encoder encodes OpenStreetMap PBF data to an output stream.
//
// The encoder writes the entities in the order that the caller gives them,
// and it does not sort them. Thus a file is sorted only if the caller gives
// sorted entities. To declare that the entities are sorted by type and then by
// ID, add the optional feature "Sort.Type_then_ID" with WithOptionalFeatures.
// The encoder does not check that feature.
type Encoder struct {
	Header   model.Header
	Entities chan<- []model.Entity

	cfg *encoderOptions
	w   io.Writer

	err   error
	close sync.Once

	completed sync.WaitGroup
	closed    sync.WaitGroup
}

// NewEncoder returns a new encoder, configured with opts, that writes to w.
// The encoder writes the OSM header before the data.
func NewEncoder(w io.Writer, opts ...EncoderOption) (*Encoder, error) {
	cfg := defaultEncoderConfig

	for _, opt := range opts {
		opt(&cfg)
	}

	if err := initializeTempStore(&cfg); err != nil {
		return nil, err
	}

	e := &Encoder{
		Header: model.Header{
			BoundingBox:                      model.InitialBoundingBox(),
			RequiredFeatures:                 cfg.requiredFeatures,
			OptionalFeatures:                 cfg.optionalFeatures,
			WritingProgram:                   cfg.writingProgram,
			Source:                           cfg.source,
			OsmosisReplicationTimestamp:      cfg.osmosisReplicationTimestamp,
			OsmosisReplicationSequenceNumber: cfg.osmosisReplicationSequenceNumber,
			OsmosisReplicationBaseURL:        cfg.osmosisReplicationBaseURL,
		},

		cfg: &cfg,
		w:   w,
	}

	entities := make(chan []model.Entity)

	e.Entities = entities

	coalesced := encoder.Coalesce(entities, encoder.EntityLimit)
	inspected, bboxes := encoder.ExtractBoundingBoxes(coalesced)
	encoded := rill.OrderedMap(inspected, singleCPU, encoder.EncodeBatch)
	packed := rill.OrderedMap(encoded, singleCPU, encoder.GenerateBatchPacker(cfg.compression))
	statuses := encoder.SavePacked(cfg.w, packed)

	// writeHeaderAndBody() will wait for these two consumers to complete
	e.completed.Add(numConsumers)
	go e.consumeBBoxes(bboxes)
	go e.consumeStatuses(statuses)

	// Close() will wait for the header and body to be written
	e.closed.Add(1)
	go e.writeHeaderAndBody()

	return e, nil
}

// Encode writes an entity into a PBF Blob.
func (e *Encoder) Encode(entity model.Entity) error {
	return e.EncodeBatch([]model.Entity{entity})
}

// EncodeBatch writes an array of entities into a PBF Blob.
func (e *Encoder) EncodeBatch(entities []model.Entity) error {
	e.Entities <- entities

	return nil
}

// Close will cancel the background encoding pipeline.
func (e *Encoder) Close() {
	e.doClose(io.EOF)
	e.closed.Wait()
}

// Close will cancel the background encoding pipeline.
func (e *Encoder) doClose(err error) {
	e.close.Do(func() {
		e.err = err
		close(e.Entities)
	})
}

func (e *Encoder) consumeBBoxes(bboxes <-chan rill.Try[*model.BoundingBox]) {
	defer e.completed.Done()

	for bbox := range bboxes {
		e.Header.BoundingBox.ExpandWithBoundingBox(bbox.Value)
	}
}

func (e *Encoder) consumeStatuses(statuses <-chan rill.Try[struct{}]) {
	defer e.completed.Done()

	for status := range statuses {
		if status.Error != nil {
			slog.Error("Got status error", "status", status)
			e.doClose(status.Error)
		}
	}
}

func (e *Encoder) writeHeaderAndBody() {
	defer e.closed.Done()
	defer func() {
		if err := os.RemoveAll(e.cfg.store); err != nil {
			slog.Error("error removing temp store", "error", err)
		}
	}()

	e.completed.Wait()

	if err := e.cfg.w.Sync(); err != nil {
		panic(fmt.Errorf("cannot sync batch: %w", err))
	}

	if offset, err := e.cfg.w.Seek(0, io.SeekStart); err != nil {
		panic(fmt.Errorf("cannot seek to beginning of file: %w", err))
	} else if offset != 0 {
		panic("cannot seek to beginning of file")
	}

	if err := encoder.SaveHeader(e.w, &e.Header, e.cfg.compression); err != nil {
		panic(fmt.Sprintf("error writing header: %v", err))
	}

	if _, err := io.Copy(e.w, e.cfg.w); err != nil {
		panic(fmt.Sprintf("error copying entities file: %v", err))
	}
}
