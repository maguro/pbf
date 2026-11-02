// SPDX-FileCopyrightText: the original author or authors
// SPDX-License-Identifier: Apache-2.0

package cli

import "github.com/spf13/cobra"

// RootCmd is the root cobra command for pbf.
var RootCmd = &cobra.Command{
	Use:   "pbf",
	Short: "An OpenStreetMap data utility",
}
