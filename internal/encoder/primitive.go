// SPDX-FileCopyrightText: the original author or authors
// SPDX-License-Identifier: Apache-2.0

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
		Datasize: proto.Int32(int32(len(bb.Value))), //nolint:gosec // The encoder does not check the 32 MiB blob limit. See maguro/pbf#11.
	}

	hb, err := proto.Marshal(hdr)
	if err != nil {
		return fmt.Errorf("could not marshal blob header: %w", err)
	}

	//nolint:gosec // A BlobHeader is much smaller than the 64 KiB limit.
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

		if v, ok := e.(*model.Relation); ok {
			extractMemberRoles(strings, v)
		}
	}

	return &blockContext{
		table:    strings.CalcTable(),
		entities: entities,
	}
}

// extractPrimitiveBlock returns a block with one primitive group for each run
// of entities of the same type. The groups keep the order of the entities.
func (bc *blockContext) extractPrimitiveBlock() *pb.PrimitiveBlock {
	var groups []*pb.PrimitiveGroup

	for start := 0; start < len(bc.entities); {
		runType := entityType(bc.entities[start])

		end := start + 1
		for end < len(bc.entities) && entityType(bc.entities[end]) == runType {
			end++
		}

		groups = append(groups, bc.extractPrimitiveGroup(bc.entities[start:end]))
		start = end
	}

	b := &pb.PrimitiveBlock{
		Stringtable: &pb.StringTable{
			S: bc.table.AsArray(),
		},
		Primitivegroup:  groups,
		Granularity:     proto.Int32(Granularity),
		LatOffset:       proto.Int64(LatOffset),
		LonOffset:       proto.Int64(LonOffset),
		DateGranularity: proto.Int32(DateGranularityMs),
	}

	return b
}

// extractPrimitiveGroup returns a group that holds run. All entities in run
// have the same type.
func (bc *blockContext) extractPrimitiveGroup(run []model.Entity) *pb.PrimitiveGroup {
	switch run[0].(type) {
	case *model.Node:
		return &pb.PrimitiveGroup{Dense: bc.extractDenseNodes(run)}
	case *model.Way:
		return &pb.PrimitiveGroup{Ways: bc.extractWays(run)}
	case *model.Relation:
		return &pb.PrimitiveGroup{Relations: bc.extractRelations(run)}
	default:
		panic("unknown type")
	}
}

func (bc *blockContext) extractDenseNodes(run []model.Entity) *pb.DenseNodes {
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

	hasInfo := false
	for _, e := range run {
		if e.GetInfo() != nil {
			hasInfo = true
			break
		}
	}

	for _, e := range run {
		n := e.(*model.Node) //nolint:forcetypeassert // All entities in run are nodes.

		ids = append(ids, int64(n.ID))

		lat := n.Lat
		lon := n.Lon

		bc.bbox.ExpandWithLatLng(lat, lon)

		lats = append(lats, model.ToCoordinate(LatOffset, Granularity, lat))
		lons = append(lons, model.ToCoordinate(LonOffset, Granularity, lon))

		if hasInfo {
			info := n.GetInfo()
			if info != nil {
				versions = append(versions, info.Version)
				uids = append(uids, int32(info.UID))
				ts = append(ts, fromTimestamp(DateGranularityMs, info.Timestamp))
				cs = append(cs, info.Changeset)
				usids = append(usids, bc.table.IndexOf(info.User))
			} else {
				versions = append(versions, 0)
				uids = append(uids, 0)
				ts = append(ts, 0)
				cs = append(cs, 0)
				usids = append(usids, bc.table.IndexOf(""))
			}
		}

		kIDs, vIDs := calcTagIDs(n.Tags, bc.table)
		for i, k := range kIDs {
			keyValIDs = append(keyValIDs, int32(k), int32(vIDs[i])) //nolint:gosec // The values are table indexes that IndexOf returned as int32.
		}

		keyValIDs = append(keyValIDs, 0)
	}

	dn.Id = calcDeltas(ids)
	if hasInfo {
		dn.Denseinfo = &pb.DenseInfo{
			Version:   calcDeltas(versions),
			Timestamp: calcDeltas(ts),
			Changeset: calcDeltas(cs),
			Uid:       calcDeltas(uids),
			UserSid:   calcDeltas(usids),
		}
	}
	dn.Lat = calcDeltas(lats)
	dn.Lon = calcDeltas(lons)
	dn.KeysVals = keyValIDs

	return dn
}

func (bc *blockContext) extractWays(run []model.Entity) []*pb.Way {
	var ways []*pb.Way

	for _, e := range run {
		w := e.(*model.Way) //nolint:forcetypeassert // All entities in run are ways.

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

	return ways
}

func (bc *blockContext) extractRelations(run []model.Entity) []*pb.Relation {
	var relations []*pb.Relation

	for _, e := range run {
		r := e.(*model.Relation) //nolint:forcetypeassert // All entities in run are relations.

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

func entityType(e model.Entity) model.EntityType {
	switch e.(type) {
	case *model.Way:
		return model.WAY
	case *model.Relation:
		return model.RELATION
	default:
		return model.NODE
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

func calcTagIDs(tags map[string]string, table *Table) (keyIDs, valIDs []uint32) {
	keys := make([]string, 0, len(tags))

	for k := range tags {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		keyIDs = append(keyIDs, uint32(table.IndexOf(k)))       //nolint:gosec // IndexOf returns an index that is not negative.
		valIDs = append(valIDs, uint32(table.IndexOf(tags[k]))) //nolint:gosec // IndexOf returns an index that is not negative.
	}

	return keyIDs, valIDs
}

func toInfoPb(info *model.Info, table *Table) *pb.Info {
	if info == nil {
		return nil
	}

	pbInfo := &pb.Info{
		Version:   proto.Int32(info.Version),
		Timestamp: proto.Int64(fromTimestamp(DateGranularityMs, info.Timestamp)),
		Changeset: proto.Int64(info.Changeset),
		Uid:       proto.Int32(int32(info.UID)),
		UserSid:   proto.Uint32(uint32(table.IndexOf(info.User))), //nolint:gosec // IndexOf returns an index that is not negative.
	}

	return pbInfo
}

// fromTimestamp converts a timestamp with a specific granularity, in units of
// milliseconds, to a UTC timestamp of type Time.
//
//nolint:unparam // The PBF format lets each block set its date_granularity. The encoder writes only DateGranularityMs.
func fromTimestamp(granularity int32, timestamp time.Time) int64 {
	millis := timestamp.UnixMilli()

	return millis / int64(granularity)
}
