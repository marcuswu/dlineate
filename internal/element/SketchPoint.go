package element

import (
	"fmt"
	"math/big"

	"github.com/marcuswu/dlineate/utils"
)

// SketchPoint represents a point in a 2D Sketch
type SketchPoint struct {
	Vector
	elementType     Type
	id              uint
	constraintLevel ConstraintLevel
	fixed           bool
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
func (p *SketchPoint) GetX() *big.Float {
	var ret big.Float
	return ret.Copy(&p.X)
}

// GetY gets the x value of the point
func (p *SketchPoint) GetY() *big.Float {
	var ret big.Float
	return ret.Copy(&p.Y)
}

// GetType gets the type of the element
func (p *SketchPoint) GetType() Type {
	return p.elementType
}

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

	var transX, transY big.Float
	transX.Neg(point.GetX())
	transY.Neg(point.GetY())
	p.Translate(&transX, &transY)
}

// Is returns true if the two elements are equal
func (p *SketchPoint) Is(o SketchElement) bool {
	return p.id == o.GetID()
}

func (p *SketchPoint) IsEqual(o SketchElement) bool {
	if p.GetType() != o.GetType() {
		return false
	}
	op := o.AsPoint()
	return utils.StandardBigFloatCompare(&p.X, &op.X) == 0 &&
		utils.StandardBigFloatCompare(&p.Y, &op.Y) == 0
}

// SquareDistanceTo returns the squared distance to the other element
func (p *SketchPoint) SquareDistanceTo(o SketchElement) *big.Float {
	if o.GetType() == Line {
		d := o.(*SketchLine).DistanceTo(p)
		return d.Mul(d, d)
	}
	var a, b big.Float
	a.Sub(o.(*SketchPoint).GetX(), &p.X)
	b.Sub(o.(*SketchPoint).GetY(), &p.Y)
	a.Mul(&a, &a)
	b.Mul(&b, &b)

	return a.Add(&a, &b)
}

// DistanceTo returns the distance to the other element
func (p *SketchPoint) DistanceTo(o SketchElement) *big.Float {
	var ret big.Float
	return ret.Sqrt(p.SquareDistanceTo(o))
}

// SketchPoint represents a point in a 2D sketch

// NewSketchPoint creates a new SketchPoint
func NewSketchPoint(id uint, x *big.Float, y *big.Float) *SketchPoint {
	var myx, myy big.Float
	myx.Copy(x)
	myy.Copy(y)
	return &SketchPoint{
		Vector:          Vector{myx, myy},
		elementType:     Point,
		id:              id,
		constraintLevel: FullyConstrained,
	}
}

func SketchPointFromVector(id uint, v Vector) *SketchPoint {
	return &SketchPoint{
		Vector:          v,
		elementType:     Point,
		id:              id,
		constraintLevel: FullyConstrained,
	}
}

// VectorTo returns a Vector to SketchElement o
func (p *SketchPoint) VectorTo(o SketchElement) *Vector {
	var point *SketchPoint
	if o.GetType() == Point {
		point = o.(*SketchPoint)
	} else {
		point = o.(*SketchLine).NearestPoint(p.GetX(), p.GetY())
	}

	var x, y big.Float
	x.Sub(p.GetX(), point.GetX())
	y.Sub(p.GetY(), point.GetY())
	return &Vector{x, y}
}

// AsPoint returns a SketchElement as a *SketchPoint or nil
func (p *SketchPoint) AsPoint() *SketchPoint {
	return p
}

// AsLine returns a SketchElement as a *SketchLine or nil
func (p *SketchPoint) AsLine() *SketchLine {
	return nil
}

func (p *SketchPoint) ConstraintLevel() ConstraintLevel {
	return p.constraintLevel
}

func (p *SketchPoint) SetConstraintLevel(cl ConstraintLevel) {
	p.constraintLevel = cl
}

func (p *SketchPoint) String() string {
	return fmt.Sprintf("Point(%d) (%s, %s)", p.GetID(), p.X.String(), p.Y.String())
}

func (p *SketchPoint) SetFixed(fixed bool) {
	p.fixed = fixed
}

func (p *SketchPoint) IsFixed() bool {
	return p.fixed
}

func (p *SketchPoint) ToGraphViz(cId int) string {
	return toGraphViz(p, cId)
}
