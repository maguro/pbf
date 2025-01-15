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
	"runtime"
)

const (
	// DefaultBufferSize is the default buffer size for protobuf un-marshaling.
	DefaultBufferSize = 1024 * 1024

	// DefaultBatchSize is the default batch size for unprocessed blobs.
	DefaultBatchSize = 16
)

// DefaultNCpu provides the default number of CPUs.
func DefaultNCpu() uint16 {
	cpus := uint16(runtime.GOMAXPROCS(-1))

	return max(cpus-1, 1)
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
