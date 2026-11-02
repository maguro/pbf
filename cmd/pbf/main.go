// SPDX-FileCopyrightText: the original author or authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"

	"m4o.io/pbf/v2/cmd/pbf/cli"
	_ "m4o.io/pbf/v2/cmd/pbf/info"
)

func main() {
	if err := cli.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
