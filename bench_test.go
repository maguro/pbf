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
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"runtime/trace"
	"strconv"
	"testing"

	"m4o.io/pbf/v2/internal/encoder"
	"m4o.io/pbf/v2/model"
)

func BenchmarkDecodeLondon(b *testing.B) {
	in, err := os.Open("testdata/greater-london.osm.pbf")
	if err != nil {
		b.Fatalf("Error reading file: %v", err)
	}

	defer in.Close()

	t, err := strconv.ParseBool(os.Getenv("PBF_TRACE"))
	if err == nil && t {
		f, e := os.Create("trace.out")
		if e != nil {
			b.Errorf("Error opening trace file: %v", e)
		} else {
			defer f.Close()
			_ = trace.Start(f)
			defer trace.Stop()
		}
	}

	var opts []DecoderOption

	if pbs, err := strconv.Atoi(os.Getenv("PBF_PROTO_BUFFER_SIZE")); err == nil {
		opts = append(opts, WithProtoBufferSize(pbs))
	}

	if ncpu, err := strconv.ParseUint(os.Getenv("PBF_NCPU"), 10, 16); err == nil && ncpu > 0 {
		opts = append(opts, WithNCpus(uint16(ncpu)))
	}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		if _, err = in.Seek(0, 0); err != nil {
			b.Fatal(err)
		}

		decoder, err := NewDecoder(context.Background(), in, opts...)
		if err != nil {
			b.Fatal(err)
		}

		for {
			if _, err := decoder.Decode(); errors.Is(err, io.EOF) {
				break
			} else if err != nil {
				decoder.Close()
				b.Fatal(err)
			}
		}

		decoder.Close()
	}
}

func BenchmarkEncodeLondon(b *testing.B) {
	in, err := os.Open("testdata/greater-london.osm.pbf")
	if err != nil {
		b.Fatalf("Error reading file: %v", err)
	}

	defer in.Close()

	decoder, err := NewDecoder(context.Background(), in)
	if err != nil {
		b.Fatal(err)
	}

	var entities []model.Entity

	for {
		batch, err := decoder.Decode()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			decoder.Close()
			b.Fatal(err)
		}

		entities = append(entities, batch...)
	}

	decoder.Close()

	store := b.TempDir()

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		enc, err := NewEncoder(io.Discard, WithStorePath(store))
		if err != nil {
			b.Fatal(err)
		}

		if err := enc.EncodeBatch(entities); err != nil {
			enc.Close()
			b.Fatal(err)
		}

		enc.Close()
	}
}

func BenchmarkDecodeLondonZstd(b *testing.B) {
	data := encodeLondon(b, encoder.ZSTD)

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		decoder, err := NewDecoder(context.Background(), bytes.NewReader(data))
		if err != nil {
			b.Fatal(err)
		}

		for {
			if _, err := decoder.Decode(); errors.Is(err, io.EOF) {
				break
			} else if err != nil {
				decoder.Close()
				b.Fatal(err)
			}
		}

		decoder.Close()
	}
}

func BenchmarkEncodeLondonZstd(b *testing.B) {
	entities := decodeLondon(b)
	store := b.TempDir()

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		enc, err := NewEncoder(io.Discard, WithStorePath(store), WithCompression(encoder.ZSTD))
		if err != nil {
			b.Fatal(err)
		}

		if err := enc.EncodeBatch(entities); err != nil {
			enc.Close()
			b.Fatal(err)
		}

		enc.Close()
	}
}

// decodeLondon returns all entities of the Greater London test file.
func decodeLondon(b *testing.B) []model.Entity {
	b.Helper()

	in, err := os.Open("testdata/greater-london.osm.pbf")
	if err != nil {
		b.Fatalf("Error reading file: %v", err)
	}

	defer in.Close()

	decoder, err := NewDecoder(context.Background(), in)
	if err != nil {
		b.Fatal(err)
	}

	defer decoder.Close()

	var entities []model.Entity

	for {
		batch, err := decoder.Decode()
		if errors.Is(err, io.EOF) {
			return entities
		} else if err != nil {
			b.Fatal(err)
		}

		entities = append(entities, batch...)
	}
}

// encodeLondon returns the Greater London test file encoded with compression.
func encodeLondon(b *testing.B, compression encoder.BlobCompression) []byte {
	b.Helper()

	var buf bytes.Buffer

	enc, err := NewEncoder(&buf, WithStorePath(b.TempDir()), WithCompression(compression))
	if err != nil {
		b.Fatal(err)
	}

	if err := enc.EncodeBatch(decodeLondon(b)); err != nil {
		enc.Close()
		b.Fatal(err)
	}

	enc.Close()

	return buf.Bytes()
}
