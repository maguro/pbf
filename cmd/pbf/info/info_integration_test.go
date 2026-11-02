//go:build integration
// +build integration

// SPDX-FileCopyrightText: the original author or authors
// SPDX-License-Identifier: Apache-2.0

package info

import (
	"testing"
)

func TestRunInfoExtended(t *testing.T) {
	testRunInfoWith(t, true, 2729006, 459055, 12833)
}
