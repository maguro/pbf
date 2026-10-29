// Copyright the original author or authors.
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
	"errors"
	"fmt"
	"os"
	"path"
	"time"

	"m4o.io/pbf/v2/internal/encoder"
)

const (
	DefaultBlobCompression = encoder.ZLIB

	tempFileName = "entities.pbf"
)

var (
	// ErrCreateTempDir is returned when the encoder cannot create a temporary directory.
	ErrCreateTempDir = errors.New("create temporary directory")
	// ErrCreateTempFile is returned when the encoder cannot create a temporary file.
	ErrCreateTempFile = errors.New("create temporary file")
)

// encoderOptions provides optional configuration parameters for Encoder construction.
type encoderOptions struct {
	compression encoder.BlobCompression

	store string
	wrtr  *os.File

	requiredFeatures                 []string
	optionalFeatures                 []string
	writingProgram                   string
	source                           string
	osmosisReplicationTimestamp      time.Time
	osmosisReplicationSequenceNumber int64
	osmosisReplicationBaseURL        string
}

// EncoderOption configures how we set up the encoder.
type EncoderOption func(*encoderOptions)

// WithCompression specifies the compression algorithm to use when encoding
// PBF blobs.  The default is ZLIB.
func WithCompression(compression encoder.BlobCompression) EncoderOption {
	return func(o *encoderOptions) {
		o.compression = compression
	}
}

// WithStorePath lets you specify where to temporarily store entities.
func WithStorePath(dir string) EncoderOption {
	return func(o *encoderOptions) {
		o.store = dir
	}
}

// WithRequiredFeatures sets the required features of the PBF header.
func WithRequiredFeatures(features ...string) EncoderOption {
	return func(o *encoderOptions) {
		o.requiredFeatures = append(o.requiredFeatures, features...)
	}
}

// WithOptionalFeatures sets the optional features of the PBF header.
func WithOptionalFeatures(features ...string) EncoderOption {
	return func(o *encoderOptions) {
		o.optionalFeatures = append(o.optionalFeatures, features...)
	}
}

// WithWritingProgram sets the writing program of the PBF header.
func WithWritingProgram(program string) EncoderOption {
	return func(o *encoderOptions) {
		o.writingProgram = program
	}
}

// WithSource sets the source of the PBF header.
func WithSource(source string) EncoderOption {
	return func(o *encoderOptions) {
		o.source = source
	}
}

// WithOsmosisReplicationTimestamp sets the Osmosis replication timestamp of
// the PBF header.
func WithOsmosisReplicationTimestamp(timestamp time.Time) EncoderOption {
	return func(o *encoderOptions) {
		o.osmosisReplicationTimestamp = timestamp
	}
}

// WithOsmosisReplicationSequenceNumber sets the Osmosis replication sequence
// number of the PBF header.
func WithOsmosisReplicationSequenceNumber(sequenceNumber int64) EncoderOption {
	return func(o *encoderOptions) {
		o.osmosisReplicationSequenceNumber = sequenceNumber
	}
}

// WithOsmosisReplicationBaseURL sets the Osmosis replication base URL of the
// PBF header.
func WithOsmosisReplicationBaseURL(url string) EncoderOption {
	return func(o *encoderOptions) {
		o.osmosisReplicationBaseURL = url
	}
}

// defaultEncoderConfig provides a default configuration for encoders.
var defaultEncoderConfig = encoderOptions{
	compression: DefaultBlobCompression,
	requiredFeatures: []string{
		featureOSMSchema,
		featureDenseNodes,
		featureHistoricalInformation,
	},
}

// initializeTempStore initializes the temporary file that entities are stored
// before being copied, after the header, to the io.Writer passed to the encoder.
func initializeTempStore(o *encoderOptions) error {
	if o.store == "" {
		tmpdir, err := os.MkdirTemp("", "pbf")
		if err != nil {
			return fmt.Errorf("%w: %w", ErrCreateTempDir, err)
		}

		o.store = tmpdir
	}

	wrtr, err := os.Create(path.Join(o.store, tempFileName))
	if err != nil {
		return fmt.Errorf("%w %s: %w", ErrCreateTempFile, path.Join(o.store, tempFileName), err)
	}

	o.wrtr = wrtr

	return nil
}
