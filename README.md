# pbf
OpenStreetMap PBF golang encoder/decoder

[![Build Status](https://github.com/maguro/pbf/actions/workflows/ci.yml/badge.svg)](https://github.com/maguro/pbf/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/maguro/pbf)](https://goreportcard.com/report/github.com/maguro/pbf) 
[![Documentation](https://godoc.org/github.com/maguro/pbf?status.svg)](http://godoc.org/github.com/maguro/pbf)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=maguro_pbf&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=maguro_pbf)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

A golang based OpenStreetMap PBF encoder/decoder with a handy command line utility, `pbf`.

## Library Usage

Add the module to your project:

```sh
go get m4o.io/pbf/v2
```

The module path is `m4o.io/pbf/v2`. Do not import `github.com/maguro/pbf`.
The root package is named `pbf`. The entity types are in
`m4o.io/pbf/v2/model`.

### Decode a PBF file

`NewDecoder` reads the file header and starts background decoding workers.
Each call to `Decode` returns the next batch of entities. After the last
batch, `Decode` returns `io.EOF`. Always call `Close`. It stops the workers,
also when you stop before the end of the file.

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"m4o.io/pbf/v2"
	"m4o.io/pbf/v2/model"
)

func main() {
	f, err := os.Open("greater-london.osm.pbf")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	d, err := pbf.NewDecoder(context.Background(), f)
	if err != nil {
		log.Fatal(err)
	}
	defer d.Close()

	var nodes, ways, relations int

	for {
		entities, err := d.Decode()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		for _, e := range entities {
			switch e.(type) {
			case *model.Node:
				nodes++
			case *model.Way:
				ways++
			case *model.Relation:
				relations++
			}
		}
	}

	fmt.Printf("nodes=%d ways=%d relations=%d\n", nodes, ways, relations)
}
```

### Read tags, coordinates, and members

The decoder returns entities as `*model.Node`, `*model.Way`, and
`*model.Relation` pointers:

```go
for _, e := range entities {
	switch v := e.(type) {
	case *model.Node:
		fmt.Println(v.ID, v.Lat, v.Lon, v.Tags["name"])
	case *model.Way:
		fmt.Println(v.ID, v.NodeIDs, v.Tags["highway"])
	case *model.Relation:
		for _, m := range v.Members {
			fmt.Println(v.ID, m.Type, m.ID, m.Role)
		}
	}
}
```

`Lat` and `Lon` are `model.Degrees`, a `float64` type that prints as degrees,
minutes, and seconds. Use `float64(v.Lat)` to get decimal degrees.

All three types implement `model.Entity`, which provides `GetID`, `GetTags`,
and `GetInfo`. `Info` holds the version, changeset, user, and timestamp.

### Read the file header

`NewDecoder` decodes the header before it returns. You can read the header
without decoding entities:

```go
d, err := pbf.NewDecoder(context.Background(), f)
if err != nil {
	log.Fatal(err)
}
defer d.Close()

h := d.Header
if h.BoundingBox != nil { // nil when the file has no bounding box
	fmt.Println(h.BoundingBox.Left, h.BoundingBox.Bottom, h.BoundingBox.Right, h.BoundingBox.Top)
}
fmt.Println(h.RequiredFeatures, h.WritingProgram, h.OsmosisReplicationTimestamp)
```

The decoder supports the `OsmSchema-V0.6`, `DenseNodes`, and
`HistoricalInformation` required features. For other required features,
`NewDecoder` returns an error that wraps `pbf.ErrUnsupportedRequiredFeature`:

```go
if errors.Is(err, pbf.ErrUnsupportedRequiredFeature) {
	// The file needs a feature this decoder does not support.
}
```

### Configure decoding concurrency

```go
d, err := pbf.NewDecoder(context.Background(), f,
	pbf.WithNCpus(4),           // decoding workers, default pbf.DefaultNCpu()
	pbf.WithProtoBatchSize(32), // blobs per decoding batch, default pbf.DefaultBatchSize
)
```

`pbf.DefaultNCpu()` is `GOMAXPROCS` minus one, with a minimum of one.

### Encode a PBF file

`NewEncoder` writes entities to a temporary file. `Close` then writes the
header and the entities to the output. The header gets a bounding box that
contains all nodes. The output is not complete until `Close` returns.

```go
package main

import (
	"log"
	"os"
	"time"

	"m4o.io/pbf/v2"
	"m4o.io/pbf/v2/model"
)

func main() {
	f, err := os.Create("out.osm.pbf")
	if err != nil {
		log.Fatal(err)
	}

	e, err := pbf.NewEncoder(f, pbf.WithWritingProgram("my-app"))
	if err != nil {
		log.Fatal(err)
	}

	info := &model.Info{Version: 1, Timestamp: time.Now(), Visible: true}

	if err := e.EncodeBatch([]model.Entity{
		&model.Node{ID: 1, Lat: 51.5007, Lon: -0.1246, Info: info},
		&model.Node{ID: 2, Lat: 51.5014, Lon: -0.1419, Info: info},
	}); err != nil {
		log.Fatal(err)
	}

	if err := e.Encode(&model.Way{
		ID:      10,
		NodeIDs: []model.ID{1, 2},
		Tags:    map[string]string{"highway": "footway"},
		Info:    info,
	}); err != nil {
		log.Fatal(err)
	}

	e.Close()

	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}
```

Rules for encoding:

- Pass `*model.Node`, `*model.Way`, and `*model.Relation` pointers.
- Set `Info` on every entity. A nil `Info` causes a panic.
- Set `Info.Visible` to `true` for current data. The default header requires
  `HistoricalInformation`, so `Visible: false` marks a deleted entity.
- Do not call `Encode` or `EncodeBatch` after `Close`.

Header options: `WithWritingProgram`, `WithSource`, `WithRequiredFeatures`
(added to the defaults), `WithOptionalFeatures`,
`WithOsmosisReplicationTimestamp`, `WithOsmosisReplicationSequenceNumber`,
and `WithOsmosisReplicationBaseURL`.

## Development

This repository includes a Codex-friendly workflow:

    $ make fmt
    $ make test
    $ make test-race
    $ make test-integration
    $ make lint
    $ make verify

Project-specific agent guidance is documented in `AGENTS.md`.
The Makefile uses a project-local `.cache/` directory for Go and lint caches to
keep local and sandboxed runs reproducible; `.cache/` is intentionally ignored
by git.

## Go Version Support

- Supported window: the Go release from approximately one year ago through latest stable.
- As of February 26, 2026, CI-verified versions are `1.24.x`, `1.25.x`, and `1.26.x`.
- `go.mod` currently declares `go 1.23`; this is a compatibility floor, while active support is the CI-verified window above.
- Bug reports should include `go version` output so compatibility issues can be reproduced quickly.

## pbf Command Line Utility

The `pbf` CLI can be installed using the `go install` command:

    $ go install m4o.io/pbf/v2/cmd/pbf

### pbf info

The `pbf` CLI can be used to obtain summary and extended information about an
OpenStreetMap PBF file:

    $ pbf info -i testdata/greater-london.osm.pbf
    BoundingBox: [(51.69344, -0.511482) (51.28554, 0.335437)]
    RequiredFeatures: OsmSchema-V0.6, DenseNodes
    OptionalFeatures: 
    WritingProgram: Osmium (http://wiki.openstreetmap.org/wiki/Osmium)
    Source: 
    OsmosisReplicationTimestamp: 2014-03-24T21:55:02Z
    OsmosisReplicationSequenceNumber: 0
    OsmosisReplicationBaseURL: 

JSON output can be obtained by adding the `-j` option:

    $ pbf info -j -i testdata/greater-london.osm.pbf | jq
    {
      "BoundingBox": {
        "Left": -0.511482,
        "Right": 0.33543700000000004,
        "Top": 51.69344,
        "Bottom": 51.285540000000005
      },
      "RequiredFeatures": [
        "OsmSchema-V0.6",
        "DenseNodes"
      ],
      "OptionalFeatures": null,
      "WritingProgram": "Osmium (http://wiki.openstreetmap.org/wiki/Osmium)",
      "Source": "",
      "OsmosisReplicationTimestamp": "2014-03-24T14:55:02-07:00",
      "OsmosisReplicationSequenceNumber": 0,
      "OsmosisReplicationBaseURL": ""
    }

Here, [jq](https://stedolan.github.io/jq/) is used to pretty print the compact
JSON.

Extended information about the OpenStreetMap PBF file can be obtained
by using the `-e` option.  This causes the entire file to be scanned, which can
take a very long time; a progress bar is displayed on `stderr`.

    $ pbf info -e -i testdata/greater-london.osm.pbf
    BoundingBox: [(51.69344, -0.511482) (51.28554, 0.335437)]
    RequiredFeatures: OsmSchema-V0.6, DenseNodes
    OptionalFeatures: 
    WritingProgram: Osmium (http://wiki.openstreetmap.org/wiki/Osmium)
    Source: 
    OsmosisReplicationTimestamp: 2014-03-24T21:55:02Z
    OsmosisReplicationSequenceNumber: 0
    OsmosisReplicationBaseURL: 
    NodeCount: 2,729,006
    WayCount: 459,055
    RelationCount: 12,833

Finally, `pbf` can read an OpenStreetMap PBF file from `stdin`:

    $ cat testdata/greater-london.osm.pbf | pbf info -e

In this case, a progress bar is not displayed since there is no way to know,
a priori, what the size of the PBF file is.
