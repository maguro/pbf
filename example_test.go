// Copyright 2017-21 the original author or authors.
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

package pbf_test

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	parser "m4o.io/pbf"
)

func Example() {
	in, err := os.Open("testdata/greater-london.osm.pbf")
	if err != nil {
		log.Fatal(err)
	}

	defer in.Close()

	const size = 3 * 1024 * 1024

	d, err := parser.NewDecoder(context.Background(), in, parser.WithProtoBufferSize(size), parser.WithNCpus(2))
	if err != nil {
		log.Fatal(err)
	}

	defer d.Close()

	var nc, wc, rc uint64

done:
	for {
		v, err := d.Decode()
		switch {
		case err == io.EOF:
			break done
		case err != nil:
			log.Fatal(err)
		default:
			switch v := v.(type) {
			case *parser.Node:
				// Process Node v.
				nc++
			case *parser.Way:
				// Process Way v.
				wc++
			case *parser.Relation:
				// Process Relation v.
				rc++
			default:
				log.Fatalf("unknown type %T\n", v)
			}
		}
	}

	fmt.Printf("Nodes: %d, Ways: %d, Relations: %d\n", nc, wc, rc)
	// Output:
	// Nodes: 2729006, Ways: 459055, Relations: 12833
}
