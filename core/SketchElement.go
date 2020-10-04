package core

import "math"

// ElementType of a SketchElement (Point or Line)
type ElementType uint

// SolveState constants
const (
	Point ElementType = iota
	Line
)

// SketchElement A 2D element within a Sketch
type SketchElement interface {
	SetID(uint)
	GetID() uint
	GetType() ElementType
	GetX() float64
	GetY() float64
	AngleTo(Vector) float64
	Translate(tx float64, ty float64)
	TranslateByElement(SketchElement)
	ReverseTranslateByElement(SketchElement)
	Rotate(tx float64)
	Equals(SketchElement) bool
	SquareDistanceTo(SketchElement) float64
	DistanceTo(SketchElement) float64
	// TODO: fill this out

}

// ElementList is a list of SketchElements
type ElementList []SketchElement

func (e ElementList) Len() int           { return len(e) }
func (e ElementList) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }
func (e ElementList) Less(i, j int) bool { return e[i].GetID() < e[j].GetID() }

// BaseElement is the base type for elements in a 2D sketch
type BaseElement struct {
	Vector
	elementType ElementType
	id          uint
}

// SetID sets the id of the element
func (p *BaseElement) SetID(id uint) {
	p.id = id
}

// GetID gets the id of the element
func (p *BaseElement) GetID() uint {
	return p.id
}

// GetType gets the type of the element
func (p *BaseElement) GetType() ElementType {
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
func (p *BaseElement) TranslateByElement(e SketchElement) {
	p.Translate(p.x+e.GetX(), p.y+e.GetY())
}

// ReverseTranslateByElement translates coordinates by the inverse of another element's coordinates
func (p *BaseElement) ReverseTranslateByElement(e SketchElement) {
	p.Translate(p.x-e.GetX(), p.y-e.GetY())
}

// Equals returns true if the two elements are equal
func (p *BaseElement) Equals(o SketchElement) bool {
	return p.id == o.GetID()
}

// SquareDistanceTo returns the squared distance to the other element
func (p *BaseElement) SquareDistanceTo(o SketchElement) float64 {
	a := p.x - o.GetX()
	b := p.y - o.GetY()

	return (a * a) + (b * b)
}

// DistanceTo returns the distance to the other element
func (p *BaseElement) DistanceTo(o SketchElement) float64 {
	return math.Sqrt(p.SquareDistanceTo(o))
}

// SketchPoint represents a point in a 2D sketch
type SketchPoint struct {
	BaseElement
}

// NewSketchPoint creates a new SketchPoint
func NewSketchPoint(id uint, x float64, y float64) *SketchPoint {
	return &SketchPoint{BaseElement: BaseElement{
		Vector:      Vector{x, y},
		elementType: Point,
		id:          id,
	}}
}

// SketchLine represents a line in a 2D sketch in the form
// Ax + By + C = 0. A and B are represented as x and y in the BaseElement
type SketchLine struct {
	BaseElement
	c float64
}

// NewSketchLine creates a new SketchLine
func NewSketchLine(id uint, x float64, y float64, d float64) *SketchLine {
	return &SketchLine{
		BaseElement: BaseElement{
			// In a line, this vector represents the signed unit normal indicating direction
			Vector:      Vector{x, y},
			elementType: Line,
			id:          id,
		},
		// Distance from the origin
		c: d,
	}
}

// GetA returns A in the formula Ax + By + C = 0
func (l *SketchLine) GetA() float64 { return l.GetX() }

// GetB returns B in the formula Ax + By + C = 0
func (l *SketchLine) GetB() float64 { return l.GetY() }

// GetC returns c in the formula Ax + By + C = 0
func (l *SketchLine) GetC() float64 { return l.c }

// SetC set the c value for the line (Ax + Bx + C = 0)
func (l *SketchLine) SetC(c float64) { l.c = c }

// SquareDistanceTo returns the squared distance to the other element
func (l *SketchLine) SquareDistanceTo(o SketchElement) float64 {
	d := l.DistanceTo(o)

	return d * d
}

func (l *SketchLine) distanceToPoint(x float64, y float64) float64 {
	return math.Abs((l.GetA()*x)+(l.GetB()*y)+l.GetC()) / l.Magnitude()
}

// DistanceTo returns the distance to the other element
func (l *SketchLine) DistanceTo(o SketchElement) float64 {
	switch o.GetType() {
	case Line:
		return l.distanceToPoint(0, 0) - o.(*SketchLine).distanceToPoint(0, 0)
	default:
		return l.distanceToPoint(o.GetX(), o.GetY())
	}
}

// GetOriginDistance returns the distance to the origin for this line
func (l *SketchLine) GetOriginDistance() float64 { return l.distanceToPoint(0, 0) }

// PointNearestOrigin get the point on the line nearest to the origin
func (l *SketchLine) PointNearestOrigin() *SketchPoint {
	return NewSketchPoint(
		0,
		(-l.GetC()*l.GetA())/l.SquareMagnitude(),
		(-l.GetC()*l.GetB())/l.SquareMagnitude())
}

/*// TranslateDistance translates the line by a distance along its normal
func (l *SketchLine) TranslateDistance(dist float64) *SketchLine {
	// find point nearest to origin
	p := l.PointNearestOrigin()
	p.Translate()
}*/

// Translated returns a line translated by an x and y value
func (l *SketchLine) Translated(tx float64, ty float64) *SketchLine {
	pointOnLine := Vector{0, -l.GetC() / l.GetB()}
	pointOnLine.Translate(tx, ty)
	newC := (-l.GetA() * tx) - (l.GetB() * ty)
	return NewSketchLine(l.GetID(), l.GetX(), l.GetY(), newC)
}

// Translate translates the location of this line by an x and y distance
func (l *SketchLine) Translate(tx float64, ty float64) {
	l.c = l.Translated(tx, ty).GetC()
}

// TranslateByElement translates the location of this line by another element
func (l *SketchLine) TranslateByElement(e SketchElement) {
	point := e
	if e.GetType() == Line {
		point = e.(*SketchLine).PointNearestOrigin()
	}
	l.Translate(point.GetX(), point.GetY())
}

// ReverseTranslateByElement translates the location of this line by the inverse of another element
func (l *SketchLine) ReverseTranslateByElement(e SketchElement) {
	point := e
	if e.GetType() == Line {
		point = e.(*SketchLine).PointNearestOrigin()
	}
	l.Translate(-point.GetX(), -point.GetY())
}

// GetSlope returns the slope of the line (Ax + By + C = 0)
func (l *SketchLine) GetSlope() float64 {
	return -l.GetA() / l.GetB()
}

// IdentityMap is a map of id to SketchElement
type IdentityMap = map[uint]SketchElement
