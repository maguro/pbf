// Copyright 2017 the original author or authors.
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

package info

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/maguro/pbf"
	"github.com/maguro/pbf/cmd/pbf/cli"
	"github.com/spf13/cobra"
)

var out io.Writer = os.Stdout

type extendedHeader struct {
	pbf.Header

	NodeCount     int64
	WayCount      int64
	RelationCount int64
}

func init() {
	cli.RootCmd.AddCommand(infoCmd)

	flags := infoCmd.Flags()
	flags.BoolP("json", "j", false, "format information in JSON")
	flags.Uint16P("cpu", "c", uint16(runtime.GOMAXPROCS(-1)), "number of CPUs to use for scanning")
	flags.BoolP("extended", "e", false, "provide extended information (scans entire file)")
}

var infoCmd = &cobra.Command{
	Use:   "info [<OSM file>]",
	Short: "Print information about an OSM file",
	Long:  "Print information about an OSM file",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		var f *os.File
		var err error
		if len(args) == 1 {
			f, err = os.Open(args[0])
			if err != nil {
				log.Fatal(err)
			}
		} else {
			f = os.Stdin
		}

		in, err := cli.WrapInputFile(f)
		if err != nil {
			log.Fatal(err)
		}

		flags := cmd.Flags()

		ncpu, err := flags.GetUint16("cpu")
		if err != nil {
			log.Fatal(err)
		}

		extended, err := flags.GetBool("extended")
		if err != nil {
			log.Fatal(err)
		}

		info := runInfo(in, ncpu, extended)

		if err := in.Close(); err != nil {
			log.Fatal(err)
		}

		jsonfmt, err := flags.GetBool("json")
		if err != nil {
			log.Fatal(err)
		}
		if jsonfmt {
			renderJSON(info, extended)
		} else {
			renderTxt(info, extended)
		}
	},
}

func runInfo(in io.Reader, ncpu uint16, extended bool) *extendedHeader {

	cfg := pbf.DecoderConfig{NCpu: ncpu}

	d, err := pbf.NewDecoder(in, cfg)
	if err != nil {
		log.Fatal(err)
	}

	info := &extendedHeader{Header: *d.Header}

	var nc, wc, rc int64
	if extended {
		for {
			if v, err := d.Decode(); err == io.EOF {
				break
			} else if err != nil {
				log.Fatal(err)
			} else {
				switch v := v.(type) {
				case *pbf.Node:
					// Process Node v.
					nc++
				case *pbf.Way:
					// Process Way v.
					wc++
				case *pbf.Relation:
					// Process Relation v.
					rc++
				default:
					log.Fatalf("unknown type %T\n", v)
				}
			}
		}

		info.NodeCount = nc
		info.WayCount = wc
		info.RelationCount = rc
	}

	return info
}

func renderJSON(info *extendedHeader, extended bool) {
	// marshall the smallest struct needed
	var v interface{}
	if extended {
		v = info
	} else {
		v = info.Header
	}
	b, err := json.Marshal(v)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprint(out, string(b))
}

func renderTxt(info *extendedHeader, extended bool) {
	fmt.Fprintf(out, "BoundingBox: %s\n", info.BoundingBox)
	fmt.Fprintf(out, "RequiredFeatures: %s\n", strings.Join(info.RequiredFeatures, ", "))
	fmt.Fprintf(out, "OptionalFeatures: %v\n", strings.Join(info.OptionalFeatures, ", "))
	fmt.Fprintf(out, "WritingProgram: %s\n", info.WritingProgram)
	fmt.Fprintf(out, "Source: %s\n", info.Source)
	fmt.Fprintf(out, "OsmosisReplicationTimestamp: %s\n", info.OsmosisReplicationTimestamp.UTC().Format(time.RFC3339))
	fmt.Fprintf(out, "OsmosisReplicationSequenceNumber: %d\n", info.OsmosisReplicationSequenceNumber)
	fmt.Fprintf(out, "OsmosisReplicationBaseURL: %s\n", info.OsmosisReplicationBaseURL)
	if extended {
		fmt.Fprintf(out, "NodeCount: %s\n", humanize.Comma(info.NodeCount))
		fmt.Fprintf(out, "WayCount: %s\n", humanize.Comma(info.WayCount))
		fmt.Fprintf(out, "RelationCount: %s\n", humanize.Comma(info.RelationCount))
	}
}
