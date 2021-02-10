// Copyright 2017-21 the original author or authors.
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

package pbf

import (
	"time"

	"github.com/golang/protobuf/proto"

	"m4o.io/pbf/protobuf"
)

func parsePrimitiveBlock(buffer []byte) ([]interface{}, error) {
	pb := &protobuf.PrimitiveBlock{}
	if err := proto.Unmarshal(buffer, pb); err != nil {
		return nil, err
	}

	c := newBlockContext(pb)

	elements := make([]interface{}, 0)
	for _, pg := range pb.GetPrimitivegroup() {
		elements = append(elements, c.decodeNodes(pg.GetNodes())...)
		elements = append(elements, c.decodeDenseNodes(pg.GetDense())...)
		elements = append(elements, c.decodeWays(pg.GetWays())...)
		elements = append(elements, c.decodeRelations(pg.GetRelations())...)
	}

	return elements, nil
}

func (c *blockContext) decodeNodes(nodes []*protobuf.Node) (elements []interface{}) {
	elements = make([]interface{}, len(nodes))

	for i, node := range nodes {
		elements[i] = &Node{
			ID:   uint64(node.GetId()),
			Tags: c.decodeTags(node.GetKeys(), node.GetVals()),
			Info: c.decodeInfo(node.GetInfo()),
			Lat:  toDegrees(c.latOffset, c.granularity, node.GetLat()),
			Lon:  toDegrees(c.lonOffset, c.granularity, node.GetLon()),
		}
	}

	return elements
}

func (c *blockContext) decodeDenseNodes(nodes *protobuf.DenseNodes) []interface{} {
	ids := nodes.GetId()
	elements := make([]interface{}, len(ids))

	tic := c.newTagsContext(nodes.GetKeysVals())
	dic := c.newDenseInfoContext(nodes.GetDenseinfo())
	lats := nodes.GetLat()
	lons := nodes.GetLon()

	var id, lat, lon int64
	for i := range ids {
		id = ids[i] + id
		lat = lats[i] + lat
		lon = lons[i] + lon

		elements[i] = &Node{
			ID:   uint64(id),
			Tags: tic.decodeTags(),
			Info: dic.decodeInfo(i),
			Lat:  toDegrees(c.latOffset, c.granularity, lat),
			Lon:  toDegrees(c.lonOffset, c.granularity, lon),
		}
	}

	return elements
}

func (c *blockContext) decodeWays(nodes []*protobuf.Way) []interface{} {
	elements := make([]interface{}, len(nodes))

	for i, node := range nodes {
		refs := node.GetRefs()
		nodeIDs := make([]uint64, len(refs))

		var nodeID int64

		for j, delta := range refs {
			nodeID = delta + nodeID
			nodeIDs[j] = uint64(nodeID)
		}

		elements[i] = &Way{
			ID:      uint64(node.GetId()),
			Tags:    c.decodeTags(node.GetKeys(), node.GetVals()),
			NodeIDs: nodeIDs,
			Info:    c.decodeInfo(node.GetInfo()),
		}
	}

	return elements
}

func (c *blockContext) decodeRelations(nodes []*protobuf.Relation) []interface{} {
	elements := make([]interface{}, len(nodes))

	for i, node := range nodes {
		elements[i] = &Relation{
			ID:      uint64(node.GetId()),
			Tags:    c.decodeTags(node.GetKeys(), node.GetVals()),
			Info:    c.decodeInfo(node.GetInfo()),
			Members: c.decodeMembers(node),
		}
	}

	return elements
}

func (c *blockContext) decodeMembers(node *protobuf.Relation) []Member {
	memids := node.GetMemids()
	memtypes := node.GetTypes()
	memroles := node.GetRolesSid()
	members := make([]Member, len(memids))

	var memid int64

	for i := range memids {
		memid = memids[i] + memid
		members[i] = Member{
			ID:   uint64(memid),
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

func (c *blockContext) decodeInfo(info *protobuf.Info) *Info {
	i := &Info{Visible: true}
	if info != nil {
		i.Version = info.GetVersion()
		i.Timestamp = toTimestamp(c.dateGranularity, info.GetTimestamp())
		i.Changeset = info.GetChangeset()
		i.UID = info.GetUid()

		i.User = c.strings[info.GetUserSid()]

		if info.Visible != nil {
			i.Visible = info.GetVisible()
		}
	}

	return i
}

type blockContext struct {
	strings         []string
	granularity     int32
	latOffset       int64
	lonOffset       int64
	dateGranularity int32
}

func newBlockContext(pb *protobuf.PrimitiveBlock) *blockContext {
	return &blockContext{
		strings:         pb.GetStringtable().GetS(),
		granularity:     pb.GetGranularity(),
		latOffset:       pb.GetLatOffset(),
		lonOffset:       pb.GetLonOffset(),
		dateGranularity: pb.GetDateGranularity(),
	}
}

type denseInfoContext struct {
	version   int32
	timestamp int64
	changeset int64
	uid       int32
	userSid   int32

	dateGranularity int32
	strings         []string
	versions        []int32
	uids            []int32
	timestamps      []int64
	changesets      []int64
	userSids        []int32
	visibilities    []bool
}

func (c *blockContext) newDenseInfoContext(di *protobuf.DenseInfo) *denseInfoContext {
	dic := &denseInfoContext{
		dateGranularity: c.dateGranularity,
		strings:         c.strings,
		versions:        di.GetVersion(),
		uids:            di.GetUid(),
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

func (dic *denseInfoContext) decodeInfo(i int) *Info {
	dic.version = dic.versions[i] + dic.version
	dic.uid = dic.uids[i] + dic.uid
	dic.timestamp = dic.timestamps[i] + dic.timestamp
	dic.changeset = dic.changesets[i] + dic.changeset
	dic.userSid = dic.userSids[i] + dic.userSid

	info := &Info{
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

// decodeMemberType converts protobuf enum Relation_MemberType to a ElementType.
func decodeMemberType(mt protobuf.Relation_MemberType) ElementType {
	switch mt {
	case protobuf.Relation_NODE:
		return NODE
	case protobuf.Relation_WAY:
		return WAY
	case protobuf.Relation_RELATION:
		return RELATION
	default:
		panic("unrecognized member type")
	}
}

// toTimestamp converts a timestamp with a specific granularity, in units of
// milliseconds, to a UTC timestamp of type Time.
func toTimestamp(granularity int32, timestamp int32) time.Time {
	ms := time.Duration(timestamp*granularity) * time.Millisecond
	return time.Unix(0, ms.Nanoseconds()).UTC()
}
