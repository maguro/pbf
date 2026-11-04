// SPDX-FileCopyrightText: the original author or authors
// SPDX-License-Identifier: Apache-2.0

package encoder

import (
	"strings"
	"testing"

	"github.com/klauspost/compress/zstd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"m4o.io/pbf/v2/internal/pb"
)

func TestPackWrapsMarshalError(t *testing.T) {
	// BlobHeader has required fields, so proto.Marshal rejects an empty one.
	_, err := Pack(&pb.BlobHeader{}, RAW)

	require.ErrorIs(t, err, proto.Error)
}

func TestPackZstdReusesEncoders(t *testing.T) {
	t.Parallel()

	dec, err := zstd.NewReader(nil)
	require.NoError(t, err)

	defer dec.Close()

	// Sizes over 128 KiB make the encoder write more than one zstd block.
	for i, size := range []int{300_000, 10, 200_000, 10} {
		msg := &pb.StringTable{S: []string{strings.Repeat(string(rune('a'+i)), size)}}

		want, err := proto.Marshal(msg)
		require.NoError(t, err)

		packed, err := Pack(msg, ZSTD)
		require.NoError(t, err)

		var blob pb.Blob
		require.NoError(t, proto.Unmarshal(packed, &blob))

		got, err := dec.DecodeAll(blob.GetZstdData(), nil)
		require.NoError(t, err, "message %d", i)

		assert.Equal(t, want, got, "message %d", i)
		assert.EqualValues(t, len(want), blob.GetRawSize(), "message %d", i)
	}
}
