// Copyright 2017-20 the original author or authors.
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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDegreesAngle(t *testing.T) {
	assert.True(t, Angle(0.78539816).EqualWithin(Degrees(45.0).Angle(), E7))
}

func TestDegreesEx(t *testing.T) {
	d := Degrees(53.123456789)

	assert.Equal(t, int32(5312346), d.E5())
	assert.Equal(t, int32(53123457), d.E6())
	assert.Equal(t, int32(531234568), d.E7())
}

func TestDegreesParse(t *testing.T) {
	d, err := ParseDegrees("53.123450")
	if err != nil {
		t.Error(err)
	}

	assert.True(t, Degrees(53.123450).EqualWithin(d, E5))

	_, err = ParseDegrees("abc")
	if err == nil {
		t.Error("Parsing should have failed")
	}
}

func TestDegreesEqualWithin(t *testing.T) {
	assert.True(t, Degrees(53.123450).EqualWithin(Degrees(53.123454), E5))
	assert.False(t, Degrees(53.123450).EqualWithin(Degrees(53.123455), E5))
}

func TestDegreesString(t *testing.T) {
	assert.Equal(t, "53° 7' 24.42\"", Degrees(53.123450).String())
}

func TestBoundingBoxEqualWithin(t *testing.T) {
	bbox := BoundingBox{Left: -0.511482, Right: 0.335437, Top: 51.69344, Bottom: 51.28554}
	assert.True(t, bbox.EqualWithin(bbox, E9))
}

func TestBoundingBoxString(t *testing.T) {
	bbox := BoundingBox{Left: -0.511482, Right: 0.335437, Top: 51.69344, Bottom: 51.28554}
	assert.Equal(t, "[-0.511482, 51.28554, 0.335437, 51.69344]", bbox.String())
}
