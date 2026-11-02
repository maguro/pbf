// SPDX-FileCopyrightText: the original author or authors
// SPDX-License-Identifier: Apache-2.0

package encoder

//go:generate stringer -type=BlobCompression

type BlobCompression int

const (
	RAW BlobCompression = iota
	ZLIB
	LZMA
	LZ4
	ZSTD
)
