// SPDX-FileCopyrightText: the original author or authors
// SPDX-License-Identifier: Apache-2.0

package pbf_test

import (
	"testing"

	parser "m4o.io/pbf/v2"
)

// A caller outside the module must be able to name every value that an
// exported option accepts. WithCompression takes a BlobCompression, so the
// type and its constants have to be part of the public API.
func TestBlobCompressionValuesAreExported(t *testing.T) {
	compressions := []parser.BlobCompression{
		parser.RAW,
		parser.ZLIB,
		parser.LZMA,
		parser.LZ4,
		parser.ZSTD,
	}

	if len(compressions) != 5 {
		t.Fatalf("expected 5 compression constants, got: %d", len(compressions))
	}
}
