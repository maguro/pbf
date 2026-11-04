// SPDX-FileCopyrightText: the original author or authors
// SPDX-License-Identifier: Apache-2.0

package encoder

import (
	"io"

	"github.com/destel/rill"

	"m4o.io/pbf/v2/internal/pb"
	"m4o.io/pbf/v2/model"
)

// Coalesce puts the entities from in into batches of size entities. The last
// batch can hold fewer entities. A batch can hold entities of more than one
// type. The batches keep the order of the entities in in. Coalesce drops a nil
// entity.
func Coalesce(in <-chan []model.Entity, size int) <-chan rill.Try[[]model.Entity] {
	out := make(chan rill.Try[[]model.Entity])

	go func() {
		defer close(out)

		batch := make([]model.Entity, 0, size)

		for entities := range in {
			for _, e := range entities {
				if e == nil {
					continue
				}

				batch = append(batch, e)

				if len(batch) == size {
					out <- rill.Wrap(batch, nil)
					batch = make([]model.Entity, 0, size)
				}
			}
		}

		if len(batch) > 0 {
			out <- rill.Wrap(batch, nil)
		}
	}()

	return out
}

func ExtractBoundingBoxes(
	in <-chan rill.Try[[]model.Entity],
) (
	out <-chan rill.Try[[]model.Entity],
	bboxes <-chan rill.Try[*model.BoundingBox],
) {
	ech := make(chan rill.Try[[]model.Entity])
	bch := make(chan rill.Try[*model.BoundingBox])

	go func() {
		defer close(ech)
		defer close(bch)

		for entities := range in {
			ech <- entities

			bbox := model.InitialBoundingBox()

			for _, e := range entities.Value {
				if n, ok := e.(*model.Node); ok {
					bbox.ExpandWithLatLng(n.Lat, n.Lon)
				}
			}

			bch <- rill.Wrap(bbox, nil)
		}
	}()

	return ech, bch
}

func EncodeBatch(batch []model.Entity) (*pb.PrimitiveBlock, error) {
	return newBlockContext(batch).extractPrimitiveBlock(), nil
}

func SavePacked(w io.Writer, ch <-chan rill.Try[[]byte]) <-chan rill.Try[struct{}] {
	out := make(chan rill.Try[struct{}])

	go func() {
		defer close(out)

		for buf := range ch {
			out <- rill.Wrap(struct{}{}, SaveBlock(w, buf))
		}
	}()

	return out
}

func GenerateBatchPacker(c BlobCompression) func(block *pb.PrimitiveBlock) ([]byte, error) {
	return func(block *pb.PrimitiveBlock) ([]byte, error) {
		return Pack(block, c)
	}
}
