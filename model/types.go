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

package model

import (
	"fmt"
	"math"
	"strconv"

	"github.com/golang/geo/s1"
)

const (
	coordinatesPerDegree = 1e-9
)

// Degrees is the decimal degree representation of a longitude or latitude.
type Degrees float64

// Angle represents a 1D angle in radians.
type Angle s1.Angle

// Epsilon is an enumeration of precisions that can be used when comparing Degrees.
type Epsilon float64

// Degrees units.
const (
	Degree           Degrees = 1
	radiansPerPi             = 180
	Radian                   = (radiansPerPi / math.Pi) * Degree
	MinutesPerDegree         = 60
	SecondsPerDegree         = 3600

	E5 Epsilon = 1e-5
	E6 Epsilon = 1e-6
	E7 Epsilon = 1e-7
	E8 Epsilon = 1e-8
	E9 Epsilon = 1e-9

	TenMillionths      = 10_000_000
	Millionths         = 1_000_000
	HundredThousandths = 100_000

	Half = 0.5
)

// Angle returns the equivalent s1.Angle.
func (d Degrees) Angle() Angle { return Angle(float64(d) * float64(s1.Degree)) }

func (d Degrees) String() string {
	var sign string
	if d < 0 {
		sign = "-"
	} else {
		sign = ""
	}

	val := math.Abs(float64(d))
	degrees := int(math.Floor(val))
	minutes := int(math.Floor(MinutesPerDegree * (val - float64(degrees))))
	seconds := SecondsPerDegree * (val - float64(degrees) - (float64(minutes) / MinutesPerDegree))

	return fmt.Sprintf("%s%d\u00B0 %d' %s\"", sign, degrees, minutes, ftoa(seconds))
}

func (d Degrees) MarshalJSON() ([]byte, error) {
	return []byte(ftoa(float64(d))), nil
}

// EqualWithin checks if two degrees are within a specific epsilon.
func (d Degrees) EqualWithin(o Degrees, eps Epsilon) bool {
	return round(float64(d)/float64(eps))-round(float64(o)/float64(eps)) == 0
}

// EqualWithin checks if two angles are within a specific epsilon.
func (d Angle) EqualWithin(o Angle, eps Epsilon) bool {
	return round(float64(d)/float64(eps))-round(float64(o)/float64(eps)) == 0
}

// E5 returns the angle in a hundred thousandths of degrees.
func (d Degrees) E5() int32 { return round(float64(d * HundredThousandths)) }

// E6 returns the angle in millionths of degrees.
func (d Degrees) E6() int32 { return round(float64(d * Millionths)) }

// E7 returns the angle in ten millionths of degrees.
func (d Degrees) E7() int32 { return round(float64(d * TenMillionths)) }

// ToDegrees converts a coordinate into Degrees, given the offset and
// granularity of the coordinate.
func ToDegrees(offset int64, granularity int32, coordinate int64) Degrees {
	return coordinatesPerDegree * Degrees(offset+(int64(granularity)*coordinate))
}

// round returns the value rounded to nearest as an int32.
// This does not match C++ exactly for the case of x.5.
func round(val float64) int32 {
	if val < 0 {
		return int32(val - Half)
	}

	return int32(val + Half)
}

// ParseDegrees converts a string to a Degrees instance.
func ParseDegrees(s string) (Degrees, error) {
	u, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}

	return Degrees(u), nil
}
