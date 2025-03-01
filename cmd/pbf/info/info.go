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

package info

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"

	"m4o.io/pbf/v2"
	"m4o.io/pbf/v2/cmd/pbf/cli"
	"m4o.io/pbf/v2/model"
)

var (
	in  *os.File
	out io.Writer = os.Stdout
)

type extendedHeader struct {
	model.Header

	NodeCount     int64
	WayCount      int64
	RelationCount int64
}

func init() { //nolint:gochecknoinits
	cli.RootCmd.AddCommand(infoCmd)

	flags := infoCmd.Flags()
	flags.VarP(cli.NewReaderValue(os.Stdin, &in, "<OSM source>"), "in", "i", "input OSM file")
	flags.BoolP("extended", "e", false, "provide extended information (scans entire file)")
	flags.BoolP("json", "j", false, "format information in JSON")
	flags.Uint32P("buffer-length", "b", pbf.DefaultBufferSize, "buffer size for protobuf un-marshaling")
	flags.Uint32P("unprocessed-batch-size", "u", pbf.DefaultBatchSize, "batch size for unprocessed blobs")
	flags.Uint16P("cpu", "c", pbf.DefaultNCpu(), "number of CPUs to use for scanning")
	flags.BoolP("silent", "s", false, "silence progress bar")
}

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Print information about an OSM file",
	Long:  "Print information about an OSM file",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		flags := cmd.Flags()

		silent, err := flags.GetBool("silent")
		if err != nil {
			log.Fatal(err)
		}

		var win io.ReadCloser
		if silent {
			win = in
		} else {
			win, err = cli.WrapInputFile(in)
			if err != nil {
				log.Fatal(err)
			}
		}
		extended, err := flags.GetBool("extended")
		if err != nil {
			log.Fatal(err)
		}

		var opts []pbf.DecoderOption

		ncpu, err := flags.GetUint16("cpu")
		if err != nil {
			log.Fatal(err)
		}

		opts = append(opts, pbf.WithNCpus(ncpu))

		batchSize, err := flags.GetUint32("unprocessed-batch-size")
		if err != nil {
			log.Fatal(err)
		}

		opts = append(opts, pbf.WithProtoBatchSize(int(batchSize)))

		info := runInfo(win, extended, opts...)

		err = win.Close()
		if err != nil {
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

func runInfo(in io.Reader, extended bool, opts ...pbf.DecoderOption) *extendedHeader {
	ctx := context.Background()

	d, err := pbf.NewDecoder(ctx, in, opts...)
	if err != nil {
		log.Fatal(err)
	}

	defer d.Close()

	info := &extendedHeader{Header: d.Header}

	var nc, wc, rc int64

	if extended {
	done:
		for {
			objs, err := d.Decode()
			switch {
			case err == io.EOF:
				break done
			case err != nil:
				panic(err.Error())
			default:
				for _, obj := range objs {
					switch t := obj.(type) {
					case *model.Node:
						// Process Node obj.
						nc++
					case *model.Way:
						// Process Way obj.
						wc++
					case *model.Relation:
						// Process Relation obj.
						rc++
					default:
						panic(fmt.Sprintf("unknown type %T\n", t))
					}
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
	var v any
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
