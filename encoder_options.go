// SPDX-FileCopyrightText: the original author or authors
// SPDX-License-Identifier: Apache-2.0

package pbf

import (
	"errors"
	"fmt"
	"os"
	"path"
	"slices"
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
	w     *os.File

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

// WithRequiredFeatures adds required features to the PBF header. The default
// features stay in the header. The features of an earlier call also stay in
// the header. The encoder keeps the first occurrence of each feature and
// removes the duplicates.
func WithRequiredFeatures(features ...string) EncoderOption {
	return func(o *encoderOptions) {
		o.requiredFeatures = appendFeatures(o.requiredFeatures, features)
	}
}

// WithOptionalFeatures adds optional features to the PBF header. The features
// of an earlier call stay in the header. The encoder keeps the first
// occurrence of each feature and removes the duplicates.
func WithOptionalFeatures(features ...string) EncoderOption {
	return func(o *encoderOptions) {
		o.optionalFeatures = appendFeatures(o.optionalFeatures, features)
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

	w, err := os.Create(path.Join(o.store, tempFileName))
	if err != nil {
		return fmt.Errorf("%w %s: %w", ErrCreateTempFile, path.Join(o.store, tempFileName), err)
	}

	o.w = w

	return nil
}

// appendFeatures returns a new slice with the features of current, followed by
// the features of added that current does not contain. The result has no
// duplicates. appendFeatures never writes to the array of current, because
// current can share its array with defaultEncoderConfig.
func appendFeatures(current, added []string) []string {
	seen := make(map[string]struct{}, len(current)+len(added))
	result := make([]string, 0, len(current)+len(added))

	for _, feature := range slices.Concat(current, added) {
		if _, ok := seen[feature]; ok {
			continue
		}

		seen[feature] = struct{}{}
		result = append(result, feature)
	}

	return result
}
