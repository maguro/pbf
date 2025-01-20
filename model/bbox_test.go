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

package model_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"m4o.io/pbf/v2/model"
)

func TestInitialBoundingBox(t *testing.T) {
	initial := model.InitialBoundingBox()
	assert.Equal(t, initial.Top, model.MinLat)
	assert.Equal(t, initial.Bottom, model.MaxLat)
	assert.Equal(t, initial.Right, model.MinLon)
	assert.Equal(t, initial.Left, model.MaxLon)
}

func TestBoundingBox_EqualWithin(t *testing.T) {
	bbox_1 := &model.BoundingBox{Top: 51.69344, Left: -0.511482, Bottom: 51.28554, Right: 0.335437}
	bbox_2 := &model.BoundingBox{
		Top:    bbox_1.Top + model.Degrees(model.E6),
		Left:   bbox_1.Left + model.Degrees(model.E6),
		Bottom: bbox_1.Bottom + model.Degrees(model.E6),
		Right:  bbox_1.Right + model.Degrees(model.E6),
	}

	assert.True(t, bbox_1.EqualWithin(bbox_2, model.E5))
	assert.False(t, bbox_1.EqualWithin(bbox_2, model.E7))
}

func TestBoundingBox_Contains(t *testing.T) {
	bbox_1 := &model.BoundingBox{Top: 51.69344, Left: -0.511482, Bottom: 51.28554, Right: 0.335437}

	test_cases := []struct {
		name     string
		lat      model.Degrees
		lng      model.Degrees
		expected bool
	}{
		{"bottom/left", bbox_1.Bottom, bbox_1.Left, true},
		{"top/left", bbox_1.Top, bbox_1.Left, true},
		{"top/right", bbox_1.Top, bbox_1.Right, true},
		{"bottom/right", bbox_1.Bottom, bbox_1.Right, true},

		{"bottom/left-E5", bbox_1.Bottom, bbox_1.Left - model.Degrees(model.E5), false},
		{"bottom-E5/left", bbox_1.Bottom - model.Degrees(model.E5), bbox_1.Left, false},
		{"bottom/left+E5", bbox_1.Bottom, bbox_1.Left + model.Degrees(model.E5), true},
		{"bottom+E5/left", bbox_1.Bottom + model.Degrees(model.E5), bbox_1.Left, true},

		{"top/right+E5", bbox_1.Top, bbox_1.Right + model.Degrees(model.E5), false},
		{"top+E5/right", bbox_1.Top + model.Degrees(model.E5), bbox_1.Right, false},
		{"top/right-E5", bbox_1.Top, bbox_1.Right - model.Degrees(model.E5), true},
		{"top-E5/right", bbox_1.Top - model.Degrees(model.E5), bbox_1.Right, true},
	}

	for _, tc := range test_cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, bbox_1.Contains(tc.lat, tc.lng))
		})
	}
}

func TestBoundingBox_ExpandWithLatLng(t *testing.T) {
	bbox := model.InitialBoundingBox()
	bbox.ExpandWithLatLng(-45, 90)
	bbox.ExpandWithLatLng(45, -90)

	assert.True(t, bbox.Contains(-45, 90))
	assert.True(t, bbox.Contains(45, -90))
	assert.True(t, bbox.Contains(-45, -90))
	assert.True(t, bbox.Contains(45, 90))
}

func TestBoundingBox_ExpandWithBoundingBox(t *testing.T) {
	bbox := model.InitialBoundingBox()
	bbox.ExpandWithBoundingBox(&model.BoundingBox{Top: 45.0, Left: 70.0, Bottom: 20.0, Right: 90.0})
	bbox.ExpandWithBoundingBox(&model.BoundingBox{Top: 20.0, Left: -20.0, Bottom: -20.0, Right: 20.0})
	bbox.ExpandWithBoundingBox(&model.BoundingBox{Top: -25.0, Left: -90.0, Bottom: -45.0, Right: -70.0})

	assert.True(t, bbox.Contains(-45, 90))
	assert.True(t, bbox.Contains(45, -90))
	assert.True(t, bbox.Contains(-45, -90))
	assert.True(t, bbox.Contains(45, 90))
}

func TestBoundingBoxEqualWithin(t *testing.T) {
	bbox := &model.BoundingBox{Top: 51.69344, Left: -0.511482, Bottom: 51.28554, Right: 0.335437}
	assert.True(t, bbox.EqualWithin(bbox, model.E9))
}

func TestBoundingBoxString(t *testing.T) {
	bbox := &model.BoundingBox{Top: 51.69344, Left: -0.511482, Bottom: 51.28554, Right: 0.335437}
	assert.Equal(t, "[(51.69344, -0.511482) (51.28554, 0.335437)]", bbox.String())
}
