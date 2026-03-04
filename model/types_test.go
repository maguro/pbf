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

func TestCoordinateRoundTripWithGranularity(t *testing.T) {
	const (
		offset      int64 = 0
		granularity int32 = 100
		coordinate  int64 = -1374389532
	)

	degrees := model.ToDegrees(offset, granularity, coordinate)
	roundTripped := model.ToCoordinate(offset, granularity, degrees)

	assert.Equal(t, coordinate, roundTripped)
}

func TestCoordinateRoundTripMatrix(t *testing.T) {
	testCases := []struct {
		name        string
		offset      int64
		granularity int32
		coordinates []int64
	}{
		{
			name:        "zero-offset-granularity-100",
			offset:      0,
			granularity: 100,
			coordinates: []int64{-1_800_000_000, -1_374_389_532, -1, 0, 1, 1_374_389_532, 1_800_000_000},
		},
		{
			name:        "zero-offset-granularity-1000",
			offset:      0,
			granularity: 1000,
			coordinates: []int64{-180_000_000, -42_123_456, -1, 0, 1, 42_123_456, 180_000_000},
		},
		{
			name:        "positive-offset-granularity-100",
			offset:      1_234_500,
			granularity: 100,
			coordinates: []int64{-10_000_000, -123_456, 0, 123_456, 10_000_000},
		},
		{
			name:        "negative-offset-granularity-100",
			offset:      -9_876_500,
			granularity: 100,
			coordinates: []int64{-10_000_000, -123_456, 0, 123_456, 10_000_000},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, coordinate := range tc.coordinates {
				degrees := model.ToDegrees(tc.offset, tc.granularity, coordinate)
				roundTripped := model.ToCoordinate(tc.offset, tc.granularity, degrees)
				assert.Equal(t, coordinate, roundTripped, "coordinate=%d", coordinate)
			}
		})
	}
}

func TestCoordinateRoundTripRange(t *testing.T) {
	const (
		offset      int64 = 0
		granularity int32 = 100
		start       int64 = -1_800_000_000
		end         int64 = 1_800_000_000
		step        int64 = 131_071
	)

	for coordinate := start; coordinate <= end; coordinate += step {
		degrees := model.ToDegrees(offset, granularity, coordinate)
		roundTripped := model.ToCoordinate(offset, granularity, degrees)
		assert.Equal(t, coordinate, roundTripped, "coordinate=%d", coordinate)
	}
}
