// SPDX-FileCopyrightText: the original author or authors
// SPDX-License-Identifier: Apache-2.0

package model

import "strconv"

// ftoa converts a float to a string with no trailing zeros.
// Snagged from github.com/dustin/go-humanize to minimize library dependencies.
func ftoa(num float64) string {
	return stripTrailingZeros(strconv.FormatFloat(num, 'f', 6, 64))
}

func stripTrailingZeros(s string) string {
	offset := len(s) - 1
	for offset > 0 {
		if s[offset] == '.' {
			offset--

			break
		}

		if s[offset] != '0' {
			break
		}

		offset--
	}

	return s[:offset+1]
}
