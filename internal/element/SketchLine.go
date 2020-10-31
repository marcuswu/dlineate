package element

import "math"

// SketchLine represents a line in a 2D sketch in the form
// Ax + By + C = 0. A and B are represented as x and y in the BaseElement
type SketchLine struct {
	elementType Type
	id          uint
	a           float64
	b           float64
	c           float64
}

// NewSketchLine creates a new SketchLine
func NewSketchLine(id uint, a float64, b float64, c float64) *SketchLine {
	return &SketchLine{
		elementType: Line,
		id:          id,
		a:           a,
		b:           b,
		c:           c,
	}
}

// GetID returns the line element identifier
func (l *SketchLine) GetID() uint { return l.id }

// SetID sets the line element identifier
func (l *SketchLine) SetID(id uint) { l.id = id }

// GetA returns A in the formula Ax + By + C = 0
func (l *SketchLine) GetA() float64 { return l.a }

// GetB returns B in the formula Ax + By + C = 0
func (l *SketchLine) GetB() float64 { return l.b }

// GetC returns c in the formula Ax + By + C = 0
func (l *SketchLine) GetC() float64 { return l.c }

// SetC set the c value for the line (Ax + Bx + C = 0)
func (l *SketchLine) SetC(c float64) { l.c = c }

// GetType returns the sketch type
func (l *SketchLine) GetType() Type { return l.elementType }

// Is returns true if the two elements are equal
func (l *SketchLine) Is(o SketchElement) bool {
	return l.id == o.GetID()
}

// SquareDistanceTo returns the squared distance to the other element
func (l *SketchLine) SquareDistanceTo(o SketchElement) float64 {
	d := l.DistanceTo(o)

	return d * d
}

func (l *SketchLine) distanceToPoint(x float64, y float64) float64 {
	// This formula can be found on wikipedia
	// https://en.wikipedia.org/wiki/Distance_from_a_point_to_a_line#Line_defined_by_an_equation
	magnitude := math.Sqrt(l.a*l.a + l.b*l.b)
	return math.Abs((l.GetA()*x)+(l.GetB()*y)+l.GetC()) / magnitude
}

// DistanceTo returns the distance to the other element
func (l *SketchLine) DistanceTo(o SketchElement) float64 {
	switch o.GetType() {
	case Line:
		// Technically I should return 0 if lines aren't parallel
		// Here I am instead comparing min distances to origin
		return l.distanceToPoint(0, 0) - o.(*SketchLine).distanceToPoint(0, 0)
	default:
		return l.distanceToPoint(o.(*SketchPoint).GetX(), o.(*SketchPoint).GetY())
	}
}

// GetOriginDistance returns the distance to the origin for this line
func (l *SketchLine) GetOriginDistance() float64 { return l.distanceToPoint(0, 0) }

// PointNearestOrigin get the point on the line nearest to the origin
func (l *SketchLine) PointNearestOrigin() *SketchPoint {
	squareMagnitude := l.a*l.a + l.b*l.b
	return NewSketchPoint(
		0,
		(-l.GetC()*l.GetA())/squareMagnitude,
		(-l.GetC()*l.GetB())/squareMagnitude)
}

// TranslateDistance translates the line by a distance along its normal
func (l *SketchLine) TranslateDistance(dist float64) {
	// find point nearest to origin
	l.c = l.TranslatedDistance(dist).GetC()
}

// TranslatedDistance returns the line translated by a distance along its normal
func (l *SketchLine) TranslatedDistance(dist float64) *SketchLine {
	// find point nearest to origin
	p := l.PointNearestOrigin()
	move, _ := p.UnitVector()
	move.Scaled(dist)
	return l.Translated(move.GetX(), move.GetY())
}

// Translated returns a line translated by an x and y value
func (l *SketchLine) Translated(tx float64, ty float64) *SketchLine {
	pointOnLine := Vector{0, -l.GetC() / l.GetB()}
	pointOnLine.Translate(tx, ty)
	newC := (-l.GetA() * pointOnLine.GetX()) - (l.GetB() * pointOnLine.GetY())
	return NewSketchLine(l.GetID(), l.GetA(), l.GetB(), newC)
}

// Translate translates the location of this line by an x and y distance
func (l *SketchLine) Translate(tx float64, ty float64) {
	l.c = l.Translated(tx, ty).GetC()
}

// TranslateByElement translates the location of this line by another element
func (l *SketchLine) TranslateByElement(e SketchElement) {
	var point *SketchPoint
	if e.GetType() == Line {
		point = e.(*SketchLine).PointNearestOrigin()
	} else {
		point = e.(*SketchPoint)
	}
	l.Translate(point.GetX(), point.GetY())
}

// ReverseTranslateByElement translates the location of this line by the inverse of another element
func (l *SketchLine) ReverseTranslateByElement(e SketchElement) {
	var point *SketchPoint
	if e.GetType() == Line {
		point = e.(*SketchLine).PointNearestOrigin()
	} else {
		point = e.(*SketchPoint)
	}
	l.Translate(-point.GetX(), -point.GetY())
}

// GetSlope returns the slope of the line (Ax + By + C = 0)
func (l *SketchLine) GetSlope() float64 {
	return -l.GetA() / l.GetB()
}

// AngleTo returns the angle to another vector in radians
func (l *SketchLine) AngleTo(u Vector) float64 {
	// point [0, -C / B] - point[-C / A, 0]
	lv := Vector{l.GetC() / l.GetA(), -l.GetC() / l.GetB()}
	return lv.AngleTo(u)
}

// AngleToLine returns the angle to another vector in radians
func (l *SketchLine) AngleToLine(o *SketchLine) float64 {
	// point [0, -C / B] - point[-C / A, 0]
	lv := Vector{l.GetC() / l.GetA(), -l.GetC() / l.GetB()}
	ov := Vector{o.GetC() / o.GetA(), -o.GetC() / o.GetB()}
	return lv.AngleTo(ov)
}

// Rotated returns a line representing this line rotated around the origin by angle radians
func (l *SketchLine) Rotated(angle float64) *SketchLine {
	// create vectors with points from the line (x and y intercepts)
	p1 := Vector{-l.GetC() / l.GetA(), 0}
	p2 := Vector{0, -l.GetC() / l.GetB()}
	// rotate those vectors to get points from the rotated line
	p1.Rotate(angle)
	p2.Rotate(angle)
	// -A / B is slope and slope is y diff / x diff, so
	// A is -y diff and B is x diff
	A := -(p2.GetY() - p1.GetY())
	B := p2.GetX() - p1.GetY()
	// Calculate C based on points from the rotated vectors
	// based on the general form line Ax + Bx + C = 0 formula
	C := -((A * p1.GetX()) + (B * p1.GetY()))
	return NewSketchLine(l.GetID(), A, B, C)
}

// Rotate returns a line representing this line rotated around the origin by angle radians
func (l *SketchLine) Rotate(angle float64) {
	rotated := l.Rotated(angle)
	l.a = rotated.GetA()
	l.b = rotated.GetB()
	l.c = rotated.GetC()
}

// Intersection returns the intersection of two lines
func (l *SketchLine) Intersection(l2 *SketchLine) Vector {
	// y := ((l.GetC() / l.GetA()) + (l2.GetC() / l2.GetA())) * (1 - (l2.GetA() / l.GetB()))
	// (x, y)  = [b1c2−b2c1/a1b2−a2b1, a2c1−a1c2/a1b2−a2b1]
	return Vector{
		((l.GetB() * l2.GetC()) - (l2.GetB() * l.GetC())) / ((l.GetA() * l2.GetB()) - (l2.GetA() * l.GetB())),
		((l.GetC() * l2.GetA()) - (l2.GetC() * l.GetA())) / ((l.GetA() * l2.GetB()) - (l2.GetA() * l.GetB()))}
}

// VectorTo returns a Vector to SketchElement o
func (l *SketchLine) VectorTo(o SketchElement) Vector {
	var point *SketchPoint
	var myPoint = l.PointNearestOrigin()
	if o.GetType() == Point {
		point = o.(*SketchPoint)
	} else {
		point = o.(*SketchLine).PointNearestOrigin()
	}

	return Vector{myPoint.GetX() - point.GetX(), myPoint.GetY() - point.GetY()}
}

// AsPoint returns a SketchElement as a *SketchPoint or nil
func (l *SketchLine) AsPoint() *SketchPoint {
	return nil
}

// AsLine returns a SketchElement as a *SketchLine or nil
func (l *SketchLine) AsLine() *SketchLine {
	return l
}
