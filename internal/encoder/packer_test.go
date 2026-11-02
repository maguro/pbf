// SPDX-FileCopyrightText: the original author or authors
// SPDX-License-Identifier: Apache-2.0

package encoder

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"m4o.io/pbf/v2/internal/pb"
)

func TestPackWrapsMarshalError(t *testing.T) {
	// BlobHeader has required fields, so proto.Marshal rejects an empty one.
	_, err := Pack(&pb.BlobHeader{}, RAW)

	require.ErrorIs(t, err, proto.Error)
}
