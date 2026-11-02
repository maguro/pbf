//go:build integration
// +build integration

// SPDX-FileCopyrightText: the original author or authors
// SPDX-License-Identifier: Apache-2.0

package pbf_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	parser "m4o.io/pbf/v2"
	"m4o.io/pbf/v2/model"
)

func Example() {
	in, err := os.Open("testdata/greater-london.osm.pbf")
	if err != nil {
		panic(err.Error())
	}

	defer in.Close()

	const size = 3 * 1024 * 1024

	d, err := parser.NewDecoder(context.Background(), in, parser.WithProtoBufferSize(size), parser.WithNCpus(2))
	if err != nil {
		panic(err.Error())
	}

	defer d.Close()

	var nc, wc, rc uint64

done:
	for {
		entities, err := d.Decode()
		switch {
		case errors.Is(err, io.EOF):
			break done
		case err != nil:
			panic(err.Error())
		default:
			for _, v := range entities {
				switch v := v.(type) {
				case *model.Node:
					// Process Node v.
					nc++
				case *model.Way:
					// Process Way v.
					wc++
				case *model.Relation:
					// Process Relation v.
					rc++
				default:
					panic(fmt.Sprintf("unknown type %T\n", v))
				}
			}
		}
	}

	fmt.Printf("Nodes: %d, Ways: %d, Relations: %d\n", nc, wc, rc)
	// Output:
	// Nodes: 2729006, Ways: 459055, Relations: 12833
}
