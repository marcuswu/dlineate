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

// SketchLine represents a line in a 2D sketch
type SketchLine struct {
	BaseElement
	originDistance float64
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
		originDistance: d,
	}
}

// GetOriginDistance returns the distance to the origin for this line
func (l *SketchLine) GetOriginDistance() float64 { return l.originDistance }

// SetOriginDistance returns the distance to the origin for this line
func (l *SketchLine) SetOriginDistance(distance float64) { l.originDistance = distance }

// Translate translates the location of this line by an x and y distance
func (l *SketchLine) Translate(tx float64, ty float64) {
	l.originDistance = l.originDistance + (tx * l.y) - (ty * l.x)
}

// TranslateByElement translates the location of this line by another element
func (l *SketchLine) TranslateByElement(e *BaseElement) {
	l.Translate(e.GetX(), e.GetY())
}

// ReverseTranslateByElement translates the location of this line by the inverse of another element
func (l *SketchLine) ReverseTranslateByElement(e *BaseElement) {
	l.Translate(-e.GetX(), -e.GetY())
}

// IdentityMap is a map of id to SketchElement
type IdentityMap = map[uint]SketchElement
