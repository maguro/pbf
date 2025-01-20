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

//go:generate stringer -type=ElementType

import (
	"time"
)

// UID is the primary key for a user.
type UID int32

// Info represents information common to Node, Way, and Relation elements.
type Info struct {
	Version   int32
	UID       UID
	Timestamp time.Time
	Changeset int64
	User      string
	Visible   bool
}

type Object interface {
	foo()
}

// ID is the primary key of an element.
type ID uint64

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

func (r Node) foo() {}

// Way is an ordered list of between 2 and 2,000 nodes that define a polyline.
type Way struct {
	ID      ID
	Tags    map[string]string
	Info    *Info
	NodeIDs []ID
}

func (r Way) foo() {}

// ElementType is an enumeration of relation types.
type ElementType int

const (
	// NODE denotes that the member is a node.
	NODE ElementType = iota

	// WAY denotes that the member is a way.
	WAY

	// RELATION denotes that the member is a relation.
	RELATION
)

// Member represents an element that.
type Member struct {
	ID   ID
	Type ElementType
	Role string
}

// Relation is a multipurpose data structure that documents a relationship
// between two or more data elements (nodes, ways, and/or other relations).
type Relation struct {
	ID      ID
	Tags    map[string]string
	Info    *Info
	Members []Member
}

func (r Relation) foo() {}
