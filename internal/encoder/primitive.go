// Copyright 2025 the original author or authors.
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
	"encoding/binary"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/destel/rill"
	"golang.org/x/exp/constraints"
	"google.golang.org/protobuf/proto"

	"m4o.io/pbf/v2/internal/pb"
	"m4o.io/pbf/v2/model"
)

const (
	DateGranularityMs = 1000
	Granularity       = 100
	LatOffset         = 0
	LonOffset         = 0

	// EntityLimit is the max number of entities in a pb.PrimitiveBlock.
	// Certain programs (e.g. osmosis 0.38) limit the number of entities in
	// each block to 8000 when writing PBF format.
	EntityLimit = 8000
)

func SaveBlock(w io.Writer, bb rill.Try[[]byte]) error {
	if bb.Error != nil {
		return bb.Error
	}

	hdr := &pb.BlobHeader{
		Type:     proto.String("OSMData"),
		Datasize: proto.Int32(int32(len(bb.Value))),
	}

	hb, err := proto.Marshal(hdr)
	if err != nil {
		return fmt.Errorf("could not marshal blob header: %w", err)
	}

	if err = binary.Write(w, binary.BigEndian, uint32(len(hb))); err != nil {
		return fmt.Errorf("could not write header size: %w", err)
	}

	if _, err = w.Write(hb); err != nil {
		return fmt.Errorf("could not write blob header: %w", err)
	}

	if _, err = w.Write(bb.Value); err != nil {
		return fmt.Errorf("could not write blob data: %w", err)
	}

	return nil
}

type blockContext struct {
	table    *Table
	bbox     model.BoundingBox
	entities []model.Entity
}

func newBlockContext(entities []model.Entity) *blockContext {
	strings := NewStrings()

	for _, e := range entities {
		extractTagsAndInfo(strings, e)

		if r, ok := e.(*model.Relation); ok {
			extractMemberRoles(strings, r)
		} else if n, ok := e.(*model.Node); ok {
			strings.Add(n.GetInfo().User)
		}
	}

	return &blockContext{
		table:    strings.CalcTable(),
		entities: entities,
	}
}

func (bc *blockContext) extractPrimitiveBlock() *pb.PrimitiveBlock {
	pg := &pb.PrimitiveGroup{}
	switch bc.entities[0].(type) {
	case *model.Node:
		pg.Dense = bc.extractDenseNodes()
	case *model.Way:
		pg.Ways = bc.extractWays()
	case *model.Relation:
		pg.Relations = bc.extractRelations()
	default:
		panic("unknown type")
	}

	b := &pb.PrimitiveBlock{
		Stringtable: &pb.StringTable{
			S: bc.table.AsArray(),
		},
		Primitivegroup:  []*pb.PrimitiveGroup{pg},
		Granularity:     proto.Int32(Granularity),
		LatOffset:       proto.Int64(LatOffset),
		LonOffset:       proto.Int64(LonOffset),
		DateGranularity: proto.Int32(DateGranularityMs),
	}

	return b
}

func (bc *blockContext) extractDenseNodes() *pb.DenseNodes {
	dn := &pb.DenseNodes{}

	ids := make([]int64, 0)

	lats := make([]int64, 0)
	lons := make([]int64, 0)

	versions := make([]int32, 0)
	uids := make([]int32, 0)
	ts := make([]int64, 0)
	cs := make([]int64, 0)
	usids := make([]int32, 0)

	keyValIDs := make([]int32, 0)

	for _, e := range bc.entities {
		if n, ok := e.(*model.Node); ok {
			ids = append(ids, int64(n.ID))

			lat := n.Lat
			lon := n.Lon

			bc.bbox.ExpandWithLatLng(lat, lon)

			lats = append(lats, model.ToCoordinate(LatOffset, Granularity, lat))
			lons = append(lons, model.ToCoordinate(LonOffset, Granularity, lon))

			info := n.GetInfo()
			versions = append(versions, info.Version)
			uids = append(uids, int32(info.UID))
			ts = append(ts, fromTimestamp(DateGranularityMs, info.Timestamp))
			cs = append(cs, info.Changeset)
			usids = append(usids, bc.table.IndexOf(info.User))

			kIDs, vIDs := calcTagIDs(n.Tags, bc.table)
			for i, k := range kIDs {
				keyValIDs = append(keyValIDs, int32(k))
				keyValIDs = append(keyValIDs, int32(vIDs[i]))
			}

			keyValIDs = append(keyValIDs, 0)
		}
	}

	dn.Id = calcDeltas(ids)
	dn.Denseinfo = &pb.DenseInfo{
		Version:   calcDeltas(versions),
		Timestamp: calcDeltas(ts),
		Changeset: calcDeltas(cs),
		Uid:       calcDeltas(uids),
		UserSid:   calcDeltas(usids),
	}
	dn.Lat = calcDeltas(lats)
	dn.Lon = calcDeltas(lons)
	dn.KeysVals = keyValIDs

	return dn
}

func (bc *blockContext) extractWays() []*pb.Way {
	var ways []*pb.Way

	for _, e := range bc.entities {
		if w, ok := e.(*model.Way); ok {
			var refs []int64

			for _, r := range w.NodeIDs {
				refs = append(refs, int64(r))
			}

			keyIDs, valIDs := calcTagIDs(w.Tags, bc.table)

			way := &pb.Way{
				Id:   proto.Int64(int64(w.ID)),
				Keys: keyIDs,
				Vals: valIDs,
				Info: toInfoPb(w.Info, bc.table),
				Refs: calcDeltas(refs),
			}

			ways = append(ways, way)
		}
	}

	return ways
}

func (bc *blockContext) extractRelations() []*pb.Relation {
	var relations []*pb.Relation

	for _, e := range bc.entities {
		if r, ok := e.(*model.Relation); ok {
			keyIDs, valIDs := calcTagIDs(r.Tags, bc.table)
			memids := make([]int64, len(r.Members))
			roleids := make([]int32, len(r.Members))
			types := make([]pb.Relation_MemberType, len(r.Members))

			for i, m := range r.Members {
				memids[i] = int64(m.ID)
				roleids[i] = bc.table.IndexOf(m.Role)
				types[i] = pb.Relation_MemberType(m.Type)
			}

			relation := &pb.Relation{
				Id:       proto.Int64(int64(r.ID)),
				Keys:     keyIDs,
				Vals:     valIDs,
				Info:     toInfoPb(r.Info, bc.table),
				RolesSid: roleids,
				Memids:   calcDeltas(memids),
				Types:    types,
			}

			relations = append(relations, relation)
		}
	}

	return relations
}

func extractMemberRoles(strings *Strings, r *model.Relation) {
	for _, m := range r.Members {
		strings.Add(m.Role)
	}
}

func extractTagsAndInfo(strings *Strings, e model.Entity) {
	for k, v := range e.GetTags() {
		strings.Add(k)
		strings.Add(v)
	}

	if info := e.GetInfo(); info != nil {
		strings.Add(info.User)
	}
}

// calcDeltas calculates the delta-encoding of the values.
func calcDeltas[T interface {
	constraints.Integer | constraints.Float
}](values []T) []T {
	prev := T(0)
	deltas := make([]T, len(values))

	for i, id := range values {
		deltas[i] = id - prev
		prev = id
	}

	return deltas
}

func calcTagIDs(tags map[string]string, table *Table) (keyIDs []uint32, valIDs []uint32) {
	keys := make([]string, 0, len(tags))

	for k := range tags {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		keyIDs = append(keyIDs, uint32(table.IndexOf(k)))
		valIDs = append(valIDs, uint32(table.IndexOf(tags[k])))
	}

	return keyIDs, valIDs
}

func toInfoPb(info *model.Info, table *Table) *pb.Info {
	pbInfo := &pb.Info{
		Version:   proto.Int32(info.Version),
		Timestamp: proto.Int32(int32(info.Timestamp.UTC().UnixMilli() / DateGranularityMs)),
		Changeset: proto.Int64(info.Changeset),
		Uid:       proto.Int32(int32(info.UID)),
		UserSid:   proto.Int32(table.IndexOf(info.User)),
		Visible:   proto.Bool(info.Visible),
	}

	return pbInfo
}

// fromTimestamp converts a timestamp with a specific granularity, in units of
// milliseconds, to a UTC timestamp of type Time.
func fromTimestamp(granularity int32, timestamp time.Time) int64 {
	millis := timestamp.UnixMilli()

	return millis / int64(granularity)
}
