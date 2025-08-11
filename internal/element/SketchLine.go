package element

import (
	"fmt"
	"math/big"

	"github.com/marcuswu/dlineate/utils"
)

// SketchLine represents a line in a 2D sketch in the form
// Ax + By + C = 0. A and B are represented as x and y in the BaseElement
type SketchLine struct {
	elementType     Type
	id              uint
	a               big.Float
	b               big.Float
	c               big.Float
	constraintLevel ConstraintLevel
	fixed           bool
}

// NewSketchLine creates a new SketchLine
func NewSketchLine(id uint, a *big.Float, b *big.Float, c *big.Float) *SketchLine {
	// A & B represent a normal vector for the line. This also determines
	// the direction of the line. C represents a magnitude of the normal
	// vector to reach from origin to the line.
	l := &SketchLine{
		elementType:     Line,
		id:              id,
		a:               *a,
		b:               *b,
		c:               *c,
		constraintLevel: FullyConstrained,
	}
	l.Normalize()
	return l
}

// GetID returns the line element identifier
func (l *SketchLine) GetID() uint { return l.id }

// SetID sets the line element identifier
func (l *SketchLine) SetID(id uint) { l.id = id }

// GetA returns A in the formula Ax + By + C = 0
func (l *SketchLine) GetA() *big.Float {
	var ret big.Float
	return ret.Copy(&l.a)
}

// GetB returns B in the formula Ax + By + C = 0
func (l *SketchLine) GetB() *big.Float {
	var ret big.Float
	return ret.Copy(&l.b)
}

// GetC returns c in the formula Ax + By + C = 0
func (l *SketchLine) GetC() *big.Float {
	var ret big.Float
	return ret.Copy(&l.c)
}

// SetC set the a value for the line (Ax + Bx + C = 0)
func (l *SketchLine) SetA(a *big.Float) { l.a.Set(a) }

// SetC set the b value for the line (Ax + Bx + C = 0)
func (l *SketchLine) SetB(b *big.Float) { l.b.Set(b) }

// SetC set the c value for the line (Ax + Bx + C = 0)
func (l *SketchLine) SetC(c *big.Float) { l.c.Set(c) }

// GetType returns the sketch type
func (l *SketchLine) GetType() Type { return l.elementType }

// Is returns true if the two elements are equal
func (l *SketchLine) Is(o SketchElement) bool {
	return l.id == o.GetID()
}

func (l *SketchLine) magnitude() *big.Float {
	var aSq, bSq, magnitude big.Float
	aSq.Mul(&l.a, &l.a)
	bSq.Mul(&l.b, &l.b)
	return magnitude.Sqrt(aSq.Add(&aSq, &bSq))
}

func (l *SketchLine) Normalize() {
	var magnitude big.Float
	magnitude.Set(l.magnitude())
	l.a.Quo(&l.a, &magnitude)
	l.b.Quo(&l.b, &magnitude)
	l.c.Quo(&l.c, &magnitude)
}

// IsEquivalent returns true if the two lines are equivalent
func (l *SketchLine) IsEquivalent(o *SketchLine) bool {
	return utils.StandardBigFloatCompare(&l.a, &o.a) == 0 &&
		utils.StandardBigFloatCompare(&l.b, &o.b) == 0 &&
		utils.StandardBigFloatCompare(&l.c, &o.c) == 0
}

// SquareDistanceTo returns the squared distance to the other element
func (l *SketchLine) SquareDistanceTo(o SketchElement) *big.Float {
	var ret big.Float
	ret.Set(l.DistanceTo(o))
	return ret.Mul(&ret, &ret)
}

func (l *SketchLine) distanceToPoint(x *big.Float, y *big.Float) *big.Float {
	var a, b, ret big.Float
	a.Mul(&l.a, x)
	b.Mul(&l.b, y)
	ret.Add(&a, &b)
	ret.Add(&ret, &l.c)
	return ret.Abs(&ret)
}

// NearestPoint returns the point on the line nearest the provided point
func (l *SketchLine) NearestPoint(x *big.Float, y *big.Float) *SketchPoint {
	var bx, ay, ac, bc, px, py big.Float
	bx.Mul(&l.b, x)
	ay.Mul(&l.a, y)
	ac.Mul(&l.a, &l.c)
	bc.Mul(&l.b, &l.c)
	px.Sub(&bx, &ay)
	px.Mul(&l.b, &px)
	px.Sub(&px, &ac)
	py.Sub(&ay, &bx)
	py.Mul(&l.a, &py)
	py.Sub(&py, &bc)

	return NewSketchPoint(0, &px, &py)
}

// DistanceTo returns the distance to the other element
func (l *SketchLine) DistanceTo(o SketchElement) *big.Float {
	switch o.GetType() {
	case Line:
		var slope, oSlope, oNeg big.Float
		slope.Set(l.GetSlope())
		oSlope.Set(o.(*SketchLine).GetSlope())
		oNeg.Neg(&oSlope)
		if utils.StandardBigFloatCompare(&slope, &oSlope) == 0 || utils.StandardBigFloatCompare(&slope, &oNeg) == 0 {
			p1 := l.PointNearestOrigin()
			p2 := o.(*SketchLine).NearestPoint(&p1.X, &p1.Y)
			return p1.DistanceTo(p2)
		}
		// Technically, non-parallel line distances should be 0. I am instead comparing min distances to origin
		var zero, res big.Float
		zero.SetFloat64(0)
		res.Sub(l.distanceToPoint(&zero, &zero), o.(*SketchLine).distanceToPoint(&zero, &zero))
		return res.Abs(&res)
	default:
		return l.distanceToPoint(o.(*SketchPoint).GetX(), o.(*SketchPoint).GetY())
	}
}

// GetOriginDistance returns the distance to the origin for this line
func (l *SketchLine) GetOriginDistance() *big.Float {
	var zero big.Float
	zero.SetFloat64(0)
	return l.distanceToPoint(&zero, &zero)
}

// PointNearestOrigin get the point on the line nearest to the origin
func (l *SketchLine) PointNearestOrigin() *SketchPoint {
	var one, x, y big.Float
	one.SetFloat64(1)
	if utils.StandardBigFloatCompare(l.magnitude(), &one) != 0 {
		l.Normalize()
	}
	x.Neg(l.GetC())
	x.Mul(&x, l.GetA())
	y.Neg(l.GetC())
	y.Mul(&y, l.GetB())
	return NewSketchPoint(0, &x, &y)
}

// TranslateDistance translates the line by a distance along its normal
func (l *SketchLine) TranslateDistance(dist *big.Float) {
	// find point nearest to origin
	newC := l.TranslatedDistance(dist).GetC()
	l.c.Set(newC)
}

// TranslatedDistance returns the line translated by a distance along its normal
func (l *SketchLine) TranslatedDistance(dist *big.Float) *SketchLine {
	var one, c big.Float
	one.SetFloat64(1)
	if utils.StandardBigFloatCompare(l.magnitude(), &one) != 0 {
		l.Normalize()
	}
	c.Sub(l.GetC(), dist)
	return &SketchLine{Line, l.GetID(), *l.GetA(), *l.GetB(), c, l.constraintLevel, l.fixed}
}

// Translated returns a line translated by an x and y value
func (l *SketchLine) Translated(tx *big.Float, ty *big.Float) *SketchLine {
	var one, x, y, newc big.Float
	one.SetFloat64(1)
	if utils.StandardBigFloatCompare(l.magnitude(), &one) != 0 {
		l.Normalize()
	}
	newc.Neg(l.GetC())
	x.Mul(l.GetA(), &newc)
	y.Mul(l.GetB(), &newc)
	pointOnLine := Vector{x, y}
	pointOnLine.Translate(tx, ty)
	newc.Neg(l.GetA())
	newc.Mul(&newc, pointOnLine.GetX())
	newc.Sub(&newc, y.Mul(l.GetB(), pointOnLine.GetY()))
	x.Set(l.GetA())
	y.Set(l.GetB())
	// If (A, B) is a unit vector normal to the line,
	// C is the magnitude of the vector to the line,
	// and (tx, ty) is a vector to translate the line,
	// then the dot product of the vectors is the change to C to move the line by tx, ty
	// newC := l.GetC() + (l.GetA() * tx) + (l.GetB() * ty)
	return &SketchLine{Line, l.GetID(), x, y, newc, l.constraintLevel, l.fixed}
}

// Translate translates the location of this line by an x and y distance
func (l *SketchLine) Translate(tx *big.Float, ty *big.Float) {
	l.c.Set(l.Translated(tx, ty).GetC())
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
	var x, y big.Float
	x.Neg(point.GetX())
	y.Neg(point.GetY())
	l.Translate(&x, &y)
}

// GetSlope returns the slope of the line (Ax + By + C = 0)
func (l *SketchLine) GetSlope() *big.Float {
	var res big.Float
	res.Neg(l.GetA())
	return res.Quo(&res, l.GetB())
}

// AngleTo returns the angle to another vector in radians
func (l *SketchLine) AngleTo(u *Vector) *big.Float {
	// point [0, -C / B] - point[-C / A, 0]
	var x, y big.Float
	x.Set(l.GetB())
	y.Neg(l.GetA())
	lv := &Vector{x, y}
	return lv.AngleTo(u)
}

// AngleToLine returns the angle the line needs to rotate to be equivalent to to another line in radians
func (l *SketchLine) AngleToLine(o *SketchLine) *big.Float {
	var lx, ly, ox, oy big.Float
	lx.Set(l.GetB())
	ly.Neg(l.GetA())
	lv := &Vector{lx, ly}
	ox.Set(o.GetB())
	oy.Neg(o.GetA())
	ov := &Vector{ox, oy}
	return lv.AngleTo(ov)
}

// Rotated returns a line representing this line rotated around the origin by angle radians
func (l *SketchLine) Rotated(angle *big.Float) *SketchLine {
	// create vectors with points from the line (x and y intercepts)
	l.Normalize()
	var x, y big.Float
	x.Set(l.GetA())
	y.Set(l.GetB())
	n := &Vector{x, y}
	n.Rotate(angle)
	return NewSketchLine(l.GetID(), n.GetX(), n.GetY(), l.GetC())
}

// Rotate returns a line representing this line rotated around the origin by angle radians
func (l *SketchLine) Rotate(angle *big.Float) {
	rotated := l.Rotated(angle)
	l.a.Set(rotated.GetA())
	l.b.Set(rotated.GetB())
	l.c.Set(rotated.GetC())
}

// Intersection returns the intersection of two lines
func (l *SketchLine) Intersection(o *SketchLine) Vector {
	var x, y, temp1, temp2, zero big.Float
	zero.SetFloat64(0)
	// y := ((l.a * o.c) - (l.c * o.a)) / ((l.b * o.a) - (l.a * o.b))
	y.Mul(&l.a, &o.c)
	temp1.Mul(&l.c, &o.a)
	y.Sub(&y, &temp1)
	temp1.Mul(&l.b, &o.a)
	temp2.Mul(&l.a, &o.b)
	temp1.Sub(&temp1, &temp2)
	y.Quo(&y, &temp1)

	x.Set(&zero)
	if utils.StandardBigFloatCompare(&o.a, &zero) == 0 {
		// x = ((l.b * y) + l.c) / -l.a
		x.Mul(&l.b, &y)
		x.Add(&x, &l.c)
		x.Quo(&x, temp1.Neg(&l.a))
	} else {
		// x = ((o.b * y) + o.c) / -o.a
		x.Mul(&o.b, &y)
		x.Add(&x, &o.c)
		x.Quo(&x, temp1.Neg(&o.a))
	}

	return Vector{x, y}
}

// VectorTo returns a Vector to SketchElement o
func (l *SketchLine) VectorTo(o SketchElement) *Vector {
	var point *SketchPoint
	var myPoint *SketchPoint
	var one, x, y big.Float
	one.SetFloat64(1)
	if utils.StandardBigFloatCompare(l.magnitude(), &one) != 0 {
		l.Normalize()
	}
	if o.GetType() == Point {
		point = o.(*SketchPoint)
		myPoint = l.NearestPoint(point.GetX(), point.GetY())
	} else {
		oline := o.AsLine()
		if utils.StandardBigFloatCompare(oline.magnitude(), &one) != 0 {
			oline.Normalize()
		}
		x.Mul(&oline.a, &oline.c)
		y.Mul(&oline.b, &oline.c)
		point = NewSketchPoint(0, &x, &y)
		x.Mul(&l.a, &l.c)
		y.Mul(&l.b, &l.c)
		myPoint = NewSketchPoint(0, &x, &y)
	}

	x.Sub(myPoint.GetX(), point.GetX())
	y.Sub(myPoint.GetY(), point.GetY())
	return &Vector{x, y}
}

// AsPoint returns a SketchElement as a *SketchPoint or nil
func (l *SketchLine) AsPoint() *SketchPoint {
	return nil
}

// AsLine returns a SketchElement as a *SketchLine or nil
func (l *SketchLine) AsLine() *SketchLine {
	return l
}

func (l *SketchLine) ConstraintLevel() ConstraintLevel {
	return l.constraintLevel
}

func (l *SketchLine) SetConstraintLevel(cl ConstraintLevel) {
	l.constraintLevel = cl
}

func (l *SketchLine) String() string {
	return fmt.Sprintf("Line(%d) %sx + %sy + %s = 0", l.id, l.a.String(), l.b.String(), l.c.String())
}

func (l *SketchLine) SetFixed(fixed bool) {
	l.fixed = fixed
}

func (l *SketchLine) IsFixed() bool {
	return l.fixed
}

func (l *SketchLine) ToGraphViz(cId int) string {
	return toGraphViz(l, cId)
}
