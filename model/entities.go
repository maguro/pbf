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

// Package model contains the shared model for OpenStreetMap PBF encoders/decoders.
package model

//go:generate stringer -type=EntityType

import (
	"time"
)

// UID is the primary key for a user.
type UID int32

// Info represents information common to Node, Way, and Relation entities.
type Info struct {
	Version   int32
	UID       UID
	Timestamp time.Time
	Changeset int64
	User      string
	Visible   bool
}

type Entity interface {
	isEntity() // prevents extensions

	GetID() ID

	GetTags() map[string]string

	GetInfo() *Info
}

// ID is the primary key of an entity.
type ID int64

// Node represents a specific point on the earth's surface defined by its
// latitude and longitude. Each node comprises at least an id number and a
// pair of coordinates.
type Node struct {
	ID   ID
	Tags map[string]string
	Info *Info
	Lat  Degrees
	Lon  Degrees
}

var _ Entity = Node{}

func (r Node) isEntity() {}

func (r Node) GetID() ID {
	return r.ID
}

func (r Node) GetTags() map[string]string {
	return r.Tags
}

func (r Node) GetInfo() *Info {
	return r.Info
}

// Way is an ordered list of between 2 and 2,000 nodes that define a polyline.
type Way struct {
	ID      ID
	Tags    map[string]string
	Info    *Info
	NodeIDs []ID
}

var _ Entity = Way{}

func (w Way) isEntity() {}

func (w Way) GetID() ID {
	return w.ID
}

func (w Way) GetTags() map[string]string {
	return w.Tags
}

func (w Way) GetInfo() *Info {
	return w.Info
}

// EntityType is an enumeration of PBF entity types.
type EntityType int32

const (
	// NODE denotes that the member is a node.
	NODE EntityType = iota

	// WAY denotes that the member is a way.
	WAY

	// RELATION denotes that the member is a relation.
	RELATION
)

// Member represents an entity that.
type Member struct {
	ID   ID
	Type EntityType
	Role string
}

// Relation is a multipurpose data structure that documents a relationship
// between two or more data entities (nodes, ways, and/or other relations).
type Relation struct {
	ID      ID
	Tags    map[string]string
	Info    *Info
	Members []Member
}

var _ Entity = Relation{}

func (r Relation) isEntity() {}

func (r Relation) GetID() ID {
	return r.ID
}

func (r Relation) GetTags() map[string]string {
	return r.Tags
}

func (r Relation) GetInfo() *Info {
	return r.Info
}
