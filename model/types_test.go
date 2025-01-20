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

func TestDegreesAngle(t *testing.T) {
	assert.True(t, model.Angle(0.78539816).EqualWithin(model.Degrees(45.0).Angle(), model.E7))
}

func TestDegreesEx(t *testing.T) {
	d := model.Degrees(53.123456789)

	assert.Equal(t, int32(5312346), d.E5())
	assert.Equal(t, int32(53123457), d.E6())
	assert.Equal(t, int32(531234568), d.E7())
}

func TestDegreesParse(t *testing.T) {
	d, err := model.ParseDegrees("53.123450")
	if err != nil {
		t.Error(err)
	}

	assert.True(t, model.Degrees(53.123450).EqualWithin(d, model.E5))

	_, err = model.ParseDegrees("abc")
	if err == nil {
		t.Error("Parsing should have failed")
	}
}

func TestDegreesEqualWithin(t *testing.T) {
	assert.True(t, model.Degrees(53.123450).EqualWithin(model.Degrees(53.123454), model.E5))
	assert.False(t, model.Degrees(53.123450).EqualWithin(model.Degrees(53.123455), model.E5))
}

func TestDegreesString(t *testing.T) {
	assert.Equal(t, "53° 7' 24.42\"", model.Degrees(53.123450).String())
}
