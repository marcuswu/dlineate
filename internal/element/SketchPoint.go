package element

import "math"

// SketchPoint represents a point in a 2D Sketch
type SketchPoint struct {
	Vector
	elementType Type
	id          uint
}

// SetID sets the id of the element
func (p *SketchPoint) SetID(id uint) {
	p.id = id
}

// GetID gets the id of the element
func (p *SketchPoint) GetID() uint {
	return p.id
}

// GetX gets the x value of the point
func (p *SketchPoint) GetX() float64 { return p.X }

// GetY gets the x value of the point
func (p *SketchPoint) GetY() float64 { return p.Y }

// GetType gets the type of the element
func (p *SketchPoint) GetType() Type {
	return p.elementType
}

/*
// GetX gets the x coordinate of the element
func (p *BaseElement) GetX() float64 {
	return p.x
}

// GetY gets the y coordinate of the element
func (p *BaseElement) GetY() float64 {
	return p.y
}

// Translate an element by an offset
func (p *BaseElement) Translate(tx float64, ty float64) {
	p.x = p.x + tx
	p.y = p.y + ty
}

// Rotate an element around an origin (tx, ty) by an angle in radians
func (p *BaseElement) Rotate(tx float64, ty float64, angle float64) {
	p.Translate(-tx, -ty)
	sinAngle := math.Sin(angle)
	cosAngle := math.Cos(angle)

	newX := p.x*cosAngle - p.y*sinAngle
	newY := p.x*sinAngle + p.y*cosAngle

	p.x = newX
	p.y = newY

	p.Translate(tx, ty)
}*/

// TranslateByElement translates coordinates by another element's coordinates
func (p *SketchPoint) TranslateByElement(e SketchElement) {
	var point *SketchPoint
	if e.GetType() == Point {
		point = e.(*SketchPoint)
	} else {
		point = e.(*SketchLine).PointNearestOrigin()
	}

	p.Translate(point.GetX(), point.GetY())
}

// ReverseTranslateByElement translates coordinates by the inverse of another element's coordinates
func (p *SketchPoint) ReverseTranslateByElement(e SketchElement) {
	var point *SketchPoint
	if e.GetType() == Point {
		point = e.(*SketchPoint)
	} else {
		point = e.(*SketchLine).PointNearestOrigin()
	}

	p.Translate(-point.GetX(), -point.GetY())
}

// Is returns true if the two elements are equal
func (p *SketchPoint) Is(o SketchElement) bool {
	return p.id == o.GetID()
}

// SquareDistanceTo returns the squared distance to the other element
func (p *SketchPoint) SquareDistanceTo(o SketchElement) float64 {
	if o.GetType() == Line {
		d := o.(*SketchLine).DistanceTo(p)
		return d * d
	}
	a := p.X - o.(*SketchPoint).GetX()
	b := p.Y - o.(*SketchPoint).GetY()

	return (a * a) + (b * b)
}

// DistanceTo returns the distance to the other element
func (p *SketchPoint) DistanceTo(o SketchElement) float64 {
	return math.Sqrt(p.SquareDistanceTo(o))
}

// SketchPoint represents a point in a 2D sketch

// NewSketchPoint creates a new SketchPoint
func NewSketchPoint(id uint, x float64, y float64) *SketchPoint {
	return &SketchPoint{
		Vector:      Vector{x, y},
		elementType: Point,
		id:          id,
	}
}
