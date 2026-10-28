// Copyright 2026 the original author or authors.
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
