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
