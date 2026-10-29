// Copyright the original author or authors.
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

package decoder

import (
	"log/slog"

	"github.com/destel/rill"

	"m4o.io/pbf/v2/internal/core"
	"m4o.io/pbf/v2/internal/pb"
	"m4o.io/pbf/v2/model"
)

// DecodeBatch unpacks a batch of primitive blobs and parses them into
// primitive blocks which are subsequently sent down the out channel.
func DecodeBatch(array []*pb.Blob) (out <-chan rill.Try[[]model.Entity]) {
	ch := make(chan rill.Try[[]model.Entity])
	out = ch

	buf := core.NewPooledBuffer()

	go func() {
		defer close(ch)
		defer buf.Close()

		for _, blob := range array {
			buf.Reset()

			unpacked, err := unpack(buf, blob)
			if err != nil {
				slog.Error("unable to unpack blob", "error", err)
				ch <- rill.Try[[]model.Entity]{Error: err}

				return
			}

			entities, err := parsePrimitiveBlock(unpacked)
			if err != nil {
				slog.Error("unable to parse block", "error", err)
				ch <- rill.Try[[]model.Entity]{Error: err}

				return
			}

			ch <- rill.Try[[]model.Entity]{Value: entities}
		}
	}()

	return out
}
