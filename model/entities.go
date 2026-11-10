// SPDX-FileCopyrightText: the original author or authors
// SPDX-License-Identifier: Apache-2.0

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
	// Visible is false for a deleted version in a history file. The encoder
	// does not write the visible flag. Thus each entity in a file that the
	// encoder writes decodes with Visible set to true.
	Visible bool
}

// Entity is a Node, a Way, or a Relation. A type outside package model cannot implement it.
type Entity interface {
	isEntity() // prevents extensions

	// GetID returns the ID of the entity.
	GetID() ID

	// GetTags returns the Tags of the entity. The returned map is the map
	// of the entity, not a copy. It can be nil.
	GetTags() map[string]string

	// GetInfo returns the Info of the entity. It is nil if the entity has
	// no Info.
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

// GetID returns the ID of the node. It implements Entity.
func (r Node) GetID() ID {
	return r.ID
}

// GetTags returns the Tags of the node. It implements Entity.
func (r Node) GetTags() map[string]string {
	return r.Tags
}

// GetInfo returns the Info of the node. It implements Entity.
func (r Node) GetInfo() *Info {
	return r.Info
}

func (r Node) isEntity() {}

// Way is an ordered list of between 2 and 2,000 nodes that define a polyline.
type Way struct {
	ID      ID
	Tags    map[string]string
	Info    *Info
	NodeIDs []ID
}

var _ Entity = Way{}

// GetID returns the ID of the way. It implements Entity.
func (w Way) GetID() ID {
	return w.ID
}

// GetTags returns the Tags of the way. It implements Entity.
func (w Way) GetTags() map[string]string {
	return w.Tags
}

// GetInfo returns the Info of the way. It implements Entity.
func (w Way) GetInfo() *Info {
	return w.Info
}

func (w Way) isEntity() {}

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

// Member is one entity in the member list of a Relation. ID and Type
// identify the entity. Role states the function of the entity in the
// relation, for example "outer" or "inner". Role can be empty.
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

// GetID returns the ID of the relation. It implements Entity.
func (r Relation) GetID() ID {
	return r.ID
}

// GetTags returns the Tags of the relation. It implements Entity.
func (r Relation) GetTags() map[string]string {
	return r.Tags
}

// GetInfo returns the Info of the relation. It implements Entity.
func (r Relation) GetInfo() *Info {
	return r.Info
}

func (r Relation) isEntity() {}
