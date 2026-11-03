// SPDX-FileCopyrightText: the original author or authors
// SPDX-License-Identifier: Apache-2.0

// Command edgecases writes testdata/edge-cases.osm.pbf from the cases in
// package internal/edgecases. The file is in the repository. Run the command
// again only after a change to the cases:
//
//	go run ./internal/tools/edgecases -o testdata/edge-cases.osm.pbf
//
// The encoder writes the blocks of each entity type in a different order on
// each run. Thus two runs give files with different bytes and the same
// entities.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"

	"m4o.io/pbf/v2"
	"m4o.io/pbf/v2/internal/edgecases"
	"m4o.io/pbf/v2/model"
)

func main() {
	out := flag.String("o", "testdata/edge-cases.osm.pbf", "path of the PBF file to write")
	flag.Parse()

	if err := write(*out); err != nil {
		log.Fatal(err)
	}
}

func write(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)

	header := edgecases.Header()
	enc, err := pbf.NewEncoder(w,
		pbf.WithOptionalFeatures(header.OptionalFeatures...),
		pbf.WithSource(header.Source),
		pbf.WithOsmosisReplicationTimestamp(header.OsmosisReplicationTimestamp),
		pbf.WithOsmosisReplicationSequenceNumber(header.OsmosisReplicationSequenceNumber),
		pbf.WithOsmosisReplicationBaseURL(header.OsmosisReplicationBaseURL),
	)
	if err != nil {
		return fmt.Errorf("create encoder: %w", err)
	}

	cases := edgecases.Cases()
	entities := make([]model.Entity, 0, len(cases))

	for _, c := range cases {
		entities = append(entities, c.Entity)
	}

	if err := enc.EncodeBatch(entities); err != nil {
		enc.Close()

		return fmt.Errorf("encode: %w", err)
	}

	enc.Close()

	if err := w.Flush(); err != nil {
		return fmt.Errorf("flush %s: %w", path, err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("close %s: %w", path, err)
	}

	return nil
}
