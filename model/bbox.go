package model

import (
	"fmt"
)

const (
	MaxLat Degrees = 90.0
	MaxLon Degrees = 180.0
	MinLat Degrees = -90.0
	MinLon Degrees = -180.0
)

// BoundingBox is simply a bounding box.
type BoundingBox struct {
	Top    Degrees `json:"top"`
	Left   Degrees `json:"left"`
	Bottom Degrees `json:"bottom"`
	Right  Degrees `json:"right"`
}

// InitialBoundingBox creates a BoundingBox that is meant to be expanded.
func InitialBoundingBox() *BoundingBox {
	return &BoundingBox{
		Top:    MinLat,
		Left:   MaxLon,
		Bottom: MaxLat,
		Right:  MinLon,
	}
}

// EqualWithin checks if two bounding boxes are within a specific epsilon.
func (b *BoundingBox) EqualWithin(o *BoundingBox, eps Epsilon) bool {
	return b.Left.EqualWithin(o.Left, eps) &&
		b.Right.EqualWithin(o.Right, eps) &&
		b.Top.EqualWithin(o.Top, eps) &&
		b.Bottom.EqualWithin(o.Bottom, eps)
}

// Contains checks if the bounding box contains the lat lng point.
func (b *BoundingBox) Contains(lat Degrees, lon Degrees) bool {
	return b.Left <= lon && lon <= b.Right && b.Bottom <= lat && lat <= b.Top
}

func (b *BoundingBox) ExpandWithLatLng(lat, lon Degrees) {
	if b.Top < lat {
		b.Top = lat
	}

	if b.Bottom > lat {
		b.Bottom = lat
	}

	if b.Left > lon {
		b.Left = lon
	}

	if b.Right < lon {
		b.Right = lon
	}
}

func (b *BoundingBox) ExpandWithBoundingBox(bbox *BoundingBox) {
	if b.Top < bbox.Top {
		b.Top = bbox.Top
	}

	if b.Bottom > bbox.Bottom {
		b.Bottom = bbox.Bottom
	}

	if b.Left > bbox.Left {
		b.Left = bbox.Left
	}

	if b.Right < bbox.Right {
		b.Right = bbox.Right
	}
}

func (b *BoundingBox) String() string {
	return fmt.Sprintf("[(%s, %s) (%s, %s)]",
		ftoa(float64(b.Top)), ftoa(float64(b.Left)),
		ftoa(float64(b.Bottom)), ftoa(float64(b.Right)))
}
