// Copyright 2025 the original author or authors.
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

// Encoder wr and decodes OpenStreetMap PBF data to an input stream.
type Encoder struct {
	Header   model.Header
	Entities chan<- []model.Entity

	cfg  *encoderOptions
	wrtr io.Writer

	err   error
	close sync.Once

	completed sync.WaitGroup
	closed    sync.WaitGroup
}

// NewEncoder returns a new encoder, configured with options, that reads from
// reader.  The decoder is initialized with the OSM header.
func NewEncoder(wrtr io.Writer, opts ...EncoderOption) (*Encoder, error) {
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

		cfg:  &cfg,
		wrtr: wrtr,
	}

	entities := make(chan []model.Entity)

	e.Entities = entities

	coalesced := encoder.Coalesce(entities, encoder.EntityLimit)
	inspected, bboxes := encoder.ExtractBoundingBoxes(coalesced)
	encoded := rill.OrderedMap(inspected, singleCPU, encoder.EncodeBatch)
	packed := rill.OrderedMap(encoded, singleCPU, encoder.GenerateBatchPacker(cfg.compression))
	statuses := encoder.SavePacked(cfg.wrtr, packed)

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
Loop:
	for {
		select {
		case bbox, ok := <-bboxes:
			if !ok {
				break Loop
			}
			e.Header.BoundingBox.ExpandWithBoundingBox(bbox.Value)
		}
	}
}

func (e *Encoder) consumeStatuses(statuses <-chan rill.Try[struct{}]) {
	defer e.completed.Done()
Loop:
	for {
		select {
		case status, ok := <-statuses:
			if !ok {
				break Loop
			} else if status.Error != nil {
				slog.Error("Got status error", "status", status)
				e.doClose(status.Error)
			}
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

	if err := e.cfg.wrtr.Sync(); err != nil {
		panic(fmt.Errorf("cannot sync batch: %w", err))
	}

	if offset, err := e.cfg.wrtr.Seek(0, io.SeekStart); err != nil {
		panic(fmt.Errorf("cannot seek to beginning of file: %w", err))
	} else if offset != 0 {
		panic("cannot seek to beginning of file")
	}

	if err := encoder.SaveHeader(e.wrtr, e.Header, e.cfg.compression); err != nil {
		panic(fmt.Sprintf("error writing header: %v", err))
	}

	if _, err := io.Copy(e.wrtr, e.cfg.wrtr); err != nil {
		panic(fmt.Sprintf("error copying entities file: %v", err))
	}
}
