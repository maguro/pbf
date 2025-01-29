// Copyright 2017-25 the original author or authors.
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
	"fmt"
	"time"

	"google.golang.org/protobuf/proto"

	"m4o.io/pbf/v2/internal/pb"
	"m4o.io/pbf/v2/model"
)

func parsePrimitiveBlock(buf []byte) ([]model.Entity, error) {
	blk := &pb.PrimitiveBlock{}
	if err := proto.Unmarshal(buf, blk); err != nil {
		return nil, fmt.Errorf("unable to unmarshal primitive block: %w", err)
	}

	c := newBlockContext(blk)

	entities := make([]model.Entity, 0)
	for _, pg := range blk.GetPrimitivegroup() {
		entities = append(entities, c.decodeNodes(pg.GetNodes())...)
		entities = append(entities, c.decodeDenseNodes(pg.GetDense())...)
		entities = append(entities, c.decodeWays(pg.GetWays())...)
		entities = append(entities, c.decodeRelations(pg.GetRelations())...)
	}

	return entities, nil
}

type blockContext struct {
	strings         []string
	granularity     int32
	latOffset       int64
	lonOffset       int64
	dateGranularity int32
}

func newBlockContext(pb *pb.PrimitiveBlock) *blockContext {
	return &blockContext{
		strings:         pb.GetStringtable().GetS(),
		granularity:     pb.GetGranularity(),
		latOffset:       pb.GetLatOffset(),
		lonOffset:       pb.GetLonOffset(),
		dateGranularity: pb.GetDateGranularity(),
	}
}

func (c *blockContext) decodeNodes(nodes []*pb.Node) (entities []model.Entity) {
	entities = make([]model.Entity, len(nodes))

	for i, node := range nodes {
		entities[i] = &model.Node{
			ID:   model.ID(node.GetId()),
			Tags: c.decodeTags(node.GetKeys(), node.GetVals()),
			Info: c.decodeInfo(node.GetInfo()),
			Lat:  model.ToDegrees(c.latOffset, c.granularity, node.GetLat()),
			Lon:  model.ToDegrees(c.lonOffset, c.granularity, node.GetLon()),
		}
	}

	return entities
}

func (c *blockContext) decodeDenseNodes(nodes *pb.DenseNodes) []model.Entity {
	ids := nodes.GetId()
	entities := make([]model.Entity, len(ids))

	tic := c.newTagsContext(nodes.GetKeysVals())
	dic := c.newDenseInfoContext(nodes.GetDenseinfo())
	lats := nodes.GetLat()
	lons := nodes.GetLon()

	var id, lat, lon int64
	for i := range ids {
		id += ids[i]
		lat += lats[i]
		lon += lons[i]

		entities[i] = &model.Node{
			ID:   model.ID(id),
			Tags: tic.decodeTags(),
			Info: dic.decodeInfo(i),
			Lat:  model.ToDegrees(c.latOffset, c.granularity, lat),
			Lon:  model.ToDegrees(c.lonOffset, c.granularity, lon),
		}
	}

	return entities
}

func (c *blockContext) decodeWays(nodes []*pb.Way) []model.Entity {
	entities := make([]model.Entity, len(nodes))

	for i, node := range nodes {
		refs := node.GetRefs()
		nodeIDs := make([]model.ID, len(refs))

		var nodeID int64

		for j, delta := range refs {
			nodeID = delta + nodeID
			nodeIDs[j] = model.ID(nodeID)
		}

		entities[i] = &model.Way{
			ID:      model.ID(node.GetId()),
			Tags:    c.decodeTags(node.GetKeys(), node.GetVals()),
			NodeIDs: nodeIDs,
			Info:    c.decodeInfo(node.GetInfo()),
		}
	}

	return entities
}

func (c *blockContext) decodeRelations(nodes []*pb.Relation) []model.Entity {
	entities := make([]model.Entity, len(nodes))

	for i, node := range nodes {
		entities[i] = &model.Relation{
			ID:      model.ID(node.GetId()),
			Tags:    c.decodeTags(node.GetKeys(), node.GetVals()),
			Info:    c.decodeInfo(node.GetInfo()),
			Members: c.decodeMembers(node),
		}
	}

	return entities
}

func (c *blockContext) decodeMembers(node *pb.Relation) []model.Member {
	memids := node.GetMemids()
	memtypes := node.GetTypes()
	memroles := node.GetRolesSid()
	members := make([]model.Member, len(memids))

	var memid int64

	for i := range memids {
		memid = memids[i] + memid
		members[i] = model.Member{
			ID:   model.ID(memid),
			Type: decodeMemberType(memtypes[i]),
			Role: c.strings[memroles[i]],
		}
	}

	return members
}

func (c *blockContext) decodeTags(keyIDs, valIDs []uint32) map[string]string {
	tags := make(map[string]string, len(keyIDs))

	for i, keyID := range keyIDs {
		tags[c.strings[keyID]] = c.strings[valIDs[i]]
	}

	return tags
}

func (c *blockContext) decodeInfo(info *pb.Info) *model.Info {
	i := &model.Info{Visible: true}
	if info != nil {
		i.Version = info.GetVersion()
		i.Timestamp = toTimestamp(c.dateGranularity, info.GetTimestamp())
		i.Changeset = info.GetChangeset()
		i.UID = model.UID(info.GetUid())

		i.User = c.strings[info.GetUserSid()]

		if info.Visible != nil {
			i.Visible = info.GetVisible()
		}
	}

	return i
}

func (c *blockContext) newDenseInfoContext(di *pb.DenseInfo) *denseInfoContext {
	uids := make([]model.UID, len(di.GetUid()))
	for i, uid := range di.GetUid() {
		uids[i] = model.UID(uid)
	}

	dic := &denseInfoContext{
		dateGranularity: c.dateGranularity,
		strings:         c.strings,
		versions:        di.GetVersion(),
		uids:            uids,
		timestamps:      di.GetTimestamp(),
		changesets:      di.GetChangeset(),
		userSids:        di.GetUserSid(),
	}

	visibilities := di.GetVisible()
	if visibilities != nil && len(visibilities) == 0 {
		dic.visibilities = nil
	} else {
		dic.visibilities = visibilities
	}

	return dic
}

type denseInfoContext struct {
	version   int32
	timestamp int64
	changeset int64
	uid       model.UID
	userSid   int32

	dateGranularity int32
	strings         []string
	versions        []int32
	uids            []model.UID
	timestamps      []int64
	changesets      []int64
	userSids        []int32
	visibilities    []bool
}

func (dic *denseInfoContext) decodeInfo(i int) *model.Info {
	dic.version += dic.versions[i]
	dic.uid += dic.uids[i]
	dic.timestamp += dic.timestamps[i]
	dic.changeset += dic.changesets[i]
	dic.userSid += dic.userSids[i]

	info := &model.Info{
		Version:   dic.version,
		UID:       dic.uid,
		Timestamp: toTimestamp(dic.dateGranularity, int32(dic.timestamp)),
		Changeset: dic.changeset,
		User:      dic.strings[dic.userSid],
	}

	if dic.visibilities == nil {
		info.Visible = true
	} else {
		info.Visible = dic.visibilities[i]
	}

	return info
}

type tagsContext struct {
	strings []string
	i       int
	keyVals []int32
}

func (c *blockContext) newTagsContext(keyVals []int32) *tagsContext {
	tc := &tagsContext{strings: c.strings}

	if len(keyVals) != 0 {
		tc.keyVals = keyVals
	}

	return tc
}

func (tic *tagsContext) decodeTags() map[string]string {
	if tic.keyVals == nil {
		return map[string]string{}
	}

	tags := make(map[string]string)
	i := tic.i

	for tic.keyVals[i] > 0 {
		tags[tic.strings[tic.keyVals[i]]] = tic.strings[tic.keyVals[i+1]]
		i += 2
	}

	tic.i = i + 1

	return tags
}

// decodeMemberType converts protobuf enum Relation_MemberType to a EntityType.
func decodeMemberType(mt pb.Relation_MemberType) model.EntityType {
	switch mt {
	case pb.Relation_NODE:
		return model.NODE
	case pb.Relation_WAY:
		return model.WAY
	case pb.Relation_RELATION:
		return model.RELATION
	default:
		panic("unrecognized member type")
	}
}

// toTimestamp converts a timestamp with a specific granularity, in units of
// milliseconds, to a UTC timestamp of type Time.
func toTimestamp(granularity int32, timestamp int32) time.Time {
	return time.UnixMilli(int64(timestamp) * int64(granularity)).UTC()
}
