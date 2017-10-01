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

package main

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
	"github.com/spf13/cobra"
	pb "gopkg.in/cheggaaa/pb.v1"
)

var (
	jsonfmt  bool
	extended bool
	cpu      uint16
	progress bool
)

type ExtendedHeader struct {
	pbf.Header

	NodeCount     int64
	WayCount      int64
	RelationCount int64
}

func init() {
	RootCmd.AddCommand(infoCmd)
	infoCmd.Flags().BoolVarP(&jsonfmt, "json", "j", false, "format information in JSON")
	infoCmd.Flags().BoolVarP(&extended, "extended", "e", false, "provide extended information (scans entire file)")
	infoCmd.Flags().Uint16VarP(&cpu, "max-cpu", "m", uint16(runtime.GOMAXPROCS(-1)), "maximum number of CPUs to use for scanning")
}

var infoCmd = &cobra.Command{
	Use:   "info [<OSM file>]",
	Short: "Print information about an OSM file",
	Long:  "Print information about an OSM file",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		var in *os.File
		if len(args) == 1 {
			f, err := os.Open(args[0])
			if err != nil {
				log.Fatal(err)
			}
			in = f
			defer in.Close()

			progress = !jsonfmt
		} else {
			in = os.Stdin
		}

		cfg := pbf.DecoderConfig{NCpu: cpu}

		d, err := pbf.NewDecoder(in, cfg)
		if err != nil {
			log.Fatal(err)
		}

		info := &ExtendedHeader{Header: *d.Header}

		var nc, wc, rc int64
		if extended {
			var bar *pb.ProgressBar
			if progress {
				fi, err := in.Stat()
				if err != nil {
					log.Fatal(err)
				}
				size := int(fi.Size())

				bar = pb.New(size).SetUnits(pb.U_BYTES)
				bar.SetWidth(80)
				bar.Start()

				if _, err = in.Seek(0, 0); err != nil {
					log.Fatal(err)
				}

				reader := bar.NewProxyReader(in)
				d, err = pbf.NewDecoder(reader, cfg)
				if err != nil {
					log.Fatal(err)
				}
			}

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

			if progress {
				bar.NotPrint = true // make sure newline is not printed
				bar.Finish()
				fmt.Printf("\033[2K\r") // clear status bar
			}

			info.NodeCount = nc
			info.WayCount = wc
			info.RelationCount = rc
		}

		if jsonfmt {
			// marshall the smallest struct needed
			var v interface{}
			if extended {
				v = info
			} else {
				v = d.Header
			}
			b, err := json.Marshal(v)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(string(b))
		} else {
			fmt.Printf("BoundingBox: %s\n", info.BoundingBox)
			fmt.Printf("RequiredFeatures: %s\n", strings.Join(info.RequiredFeatures, ", "))
			fmt.Printf("OptionalFeatures: %s\n", strings.Join(info.OptionalFeatures, ", "))
			fmt.Printf("WritingProgram: %s\n", info.WritingProgram)
			fmt.Printf("Source: %s\n", info.Source)
			fmt.Printf("OsmosisReplicationTimestamp: %s\n", info.OsmosisReplicationTimestamp.UTC().Format(time.RFC3339))
			fmt.Printf("OsmosisReplicationSequenceNumber: %d\n", info.OsmosisReplicationSequenceNumber)
			fmt.Printf("OsmosisReplicationBaseURL: %s\n", info.OsmosisReplicationBaseURL)
			if extended {
				fmt.Printf("NodeCount: %s\n", humanize.Comma(info.NodeCount))
				fmt.Printf("WayCount: %s\n", humanize.Comma(info.WayCount))
				fmt.Printf("RelationCount: %s\n", humanize.Comma(info.RelationCount))
			}
		}
	},
}
