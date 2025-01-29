package encoder

import (
	"io"

	"github.com/destel/rill"

	"m4o.io/pbf/v2/internal/pb"
	"m4o.io/pbf/v2/model"
)

func Coalesce(in <-chan []model.Entity, size int) <-chan rill.Try[[]model.Entity] {
	nch := make(chan rill.Try[model.Entity])
	rch := make(chan rill.Try[model.Entity])
	wch := make(chan rill.Try[model.Entity])

	go func() {
		defer close(nch)
		defer close(rch)
		defer close(wch)

		for entities := range in {
			for _, e := range entities {
				o := rill.Try[model.Entity]{Value: e}
				nch <- o
				rch <- o
				wch <- o
			}
		}
	}()

	bn := batchEntities[*model.Node](nch, size)
	br := batchEntities[*model.Relation](rch, size)
	bw := batchEntities[*model.Way](wch, size)

	return rill.Merge(bn, br, bw)
}

func ExtractBoundingBoxes(
	in <-chan rill.Try[[]model.Entity],
) (
	<-chan rill.Try[[]model.Entity],
	<-chan rill.Try[*model.BoundingBox],
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

func batchEntities[T model.Entity](in <-chan rill.Try[model.Entity], size int) <-chan rill.Try[[]model.Entity] {
	nodes := rill.OrderedFilter(in, 1, func(object model.Entity) (bool, error) {
		_, ok := object.(T)

		return ok, nil
	})

	return rill.Batch(nodes, size, -1)
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
