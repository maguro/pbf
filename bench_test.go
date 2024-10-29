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
	"io"
	"os"
	"runtime/trace"
	"strconv"
	"testing"
)

func BenchmarkLondon(b *testing.B) {
	in, err := os.Open("testdata/greater-london.osm.pbf")
	if err != nil {
		b.Errorf("Error reading file: %v", err)
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

	pbs, _ := strconv.Atoi(os.Getenv("PBF_PROTO_BUFFER_SIZE"))
	ncpu, _ := strconv.Atoi(os.Getenv("PBF_NCPU"))

	for n := 0; n < b.N; n++ {
		if _, err = in.Seek(0, 0); err != nil {
			b.Fatal(err)
		}

		if decoder, err := NewDecoder(context.Background(), in,
			WithProtoBufferSize(pbs),
			WithNCpus(uint16(ncpu))); err != nil {
			b.Fatal(err)
		} else {
			for {
				if _, err := decoder.Decode(); err == io.EOF {
					break
				} else if err != nil {
					b.Fatal(err)
				}
			}
		}
	}
}
