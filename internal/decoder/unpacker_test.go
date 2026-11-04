// SPDX-FileCopyrightText: the original author or authors
// SPDX-License-Identifier: Apache-2.0

package decoder

import (
	"strings"
	"testing"

	"github.com/klauspost/compress/zstd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"m4o.io/pbf/v2/internal/core"
	"m4o.io/pbf/v2/internal/pb"
)

func TestUnpackZstdReusesDecoders(t *testing.T) {
	t.Parallel()

	enc, err := zstd.NewWriter(nil)
	require.NoError(t, err)

	defer enc.Close()

	buf := core.NewPooledBuffer()

	defer buf.Close()

	// Sizes over 128 KiB make a frame with more than one zstd block.
	for i, size := range []int{300_000, 10, 200_000, 10} {
		raw := []byte(strings.Repeat(string(rune('a'+i)), size))
		blob := &pb.Blob{
			RawSize: proto.Int32(int32(size)), //nolint:gosec // The test sizes fit in int32.
			Data:    &pb.Blob_ZstdData{ZstdData: enc.EncodeAll(raw, nil)},
		}

		buf.Reset()

		got, err := unpack(buf, blob)
		require.NoError(t, err, "blob %d", i)

		assert.Equal(t, raw, got, "blob %d", i)
	}
}
