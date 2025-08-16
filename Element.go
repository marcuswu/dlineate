package dlineate

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"

	c "github.com/marcuswu/dlineate/internal/constraint"
	"github.com/marcuswu/dlineate/internal/element"
	el "github.com/marcuswu/dlineate/internal/element"
	"github.com/marcuswu/dlineate/utils"
	"github.com/rs/zerolog/log"
	"github.com/tdewolff/canvas"
)

// Type of a Constraint(Distance or Angle)
type ElementType uint

// ElementType constants
const (
	Point ElementType = iota
	Axis
	Line
	Circle
	Arc
)

func (et ElementType) String() string {
	switch et {
	case Point:
		return "Point"
	case Axis:
		return "Axis"
	case Line:
		return "Line"
	case Circle:
		return "Circle"
	case Arc:
		return "Arc"
	default:
		return fmt.Sprintf("%d", int(et))
	}
}

type Element struct {
	id          uint
	values      []float64
	elementType ElementType
	constraints []*c.Constraint
	element     el.SketchElement
	children    []*Element
	isChild     bool
	valuePass   int
}

func emptyElement() *Element {
	ec := new(Element)
	ec.values = make([]float64, 0)
	ec.constraints = make([]*c.Constraint, 0)
	ec.children = make([]*Element, 0)
	ec.isChild = false
	ec.valuePass = 0
	return ec
}

func (e *Element) ID() uint {
	return e.id
}

func (e *Element) valuesFromSketch(s *Sketch) error {
	switch e.elementType {
	case Point:
		p := e.element.AsPoint()
		e.values[0], _ = p.GetX().Float64()
		e.values[1], _ = p.GetY().Float64()
	case Axis:
		p := e.element.AsLine()
		e.values[0], _ = p.GetA().Float64()
		e.values[1], _ = p.GetB().Float64()
		e.values[2], _ = p.GetC().Float64()
	case Line:
		p1 := e.children[0].element.AsPoint()
		p2 := e.children[1].element.AsPoint()
		e.values[0], _ = p1.GetX().Float64()
		e.values[1], _ = p1.GetY().Float64()
		e.values[2], _ = p2.GetX().Float64()
		e.values[3], _ = p2.GetY().Float64()
	case Circle:
		/*
			Circle radius is determined either by
			  * a distance constraint against the Circle
			  * a coincident constraint against a Circle with the location of the center constrained
		*/
		var err error = nil
		c := e.children[0].element.AsPoint()
		e.values[0], _ = c.GetX().Float64()
		e.values[1], _ = c.GetY().Float64()
		// find distance constraint on e
		constraint, err := s.findConstraint(Distance, e)
		if err != nil {
			constraint, err = s.findConstraint(Coincident, e)
		}
		if err != nil {
			return err
		}
		e.values[2], err = e.getCircleRadius(s, constraint)
		if err != nil {
			return err
		}
	case Arc:
		center := e.children[0].element.AsPoint()
		start := e.children[1].element.AsPoint()
		end := e.children[2].element.AsPoint()
		e.values[0], _ = center.GetX().Float64()
		e.values[1], _ = center.GetY().Float64()
		e.values[2], _ = start.GetX().Float64()
		e.values[3], _ = start.GetY().Float64()
		e.values[4], _ = end.GetX().Float64()
		e.values[5], _ = end.GetY().Float64()
	}
	e.valuePass = s.passes

	return nil
}

func (e *Element) getCircleRadius(s *Sketch, c *Constraint) (float64, error) {
	if e.elementType != Circle {
		return 0, errors.New("can't return radius for a non-circle")
	}
	if c.constraintType == Distance && len(c.elements) == 1 && c.elements[0].id == e.id {
		return c.dataValue, nil
	}
	if c.constraintType == Coincident {
		constraint := c.constraints[0]
		other, _ := s.sketch.GetElement(constraint.Element1)
		if other == e.children[0].element {
			other, _ = s.sketch.GetElement(constraint.Element2)
		}

		dist, _ := other.DistanceTo(e.children[0].element.AsPoint()).Float64()
		return dist, nil
	}

	return 0, errors.New("Constraint type for circle radius must be Distance or Coincident")
}

func (e *Element) Values() []float64 {
	return e.values
}

func (e *Element) ConstraintLevel() el.ConstraintLevel {
	level := e.element.ConstraintLevel()
	var childLevel el.ConstraintLevel
	for _, c := range e.children {
		childLevel = c.element.ConstraintLevel()
		if childLevel < level {
			level = childLevel
		}
	}
	return level
}

func (e *Element) replaceElement(original uint, new element.SketchElement) {
	if e.element.GetID() == original {
		e.element = new
	}
	for _, c := range e.children {
		c.replaceElement(original, new)
	}
}

func (e *Element) minMaxXY() (float64, float64, float64, float64) {
	minX := math.MaxFloat64
	minY := math.MaxFloat64
	maxX := math.MaxFloat64 * -1
	maxY := math.MaxFloat64 * -1

	switch e.elementType {
	case Point:
		if e.values[0] < minX {
			minX = e.values[0]
		}
		if e.values[0] > maxX {
			maxX = e.values[0]
		}
		if e.values[1] < minX {
			minY = e.values[1]
		}
		if e.values[1] > maxX {
			maxY = e.values[1]
		}
	case Line:
		if e.values[0] < minX {
			minX = e.values[0]
		}
		if e.values[0] > maxX {
			maxX = e.values[0]
		}
		if e.values[1] < minX {
			minY = e.values[1]
		}
		if e.values[1] > maxX {
			maxY = e.values[1]
		}
		if e.values[2] < minX {
			minX = e.values[2]
		}
		if e.values[2] > maxX {
			maxX = e.values[2]
		}
		if e.values[3] < minX {
			minY = e.values[3]
		}
		if e.values[3] > maxX {
			maxY = e.values[3]
		}
	case Circle:
		size := e.values[2]
		if e.values[0]-size < minX {
			minX = e.values[0] - size
		}
		if e.values[0]+size > maxX {
			maxX = e.values[0] + size
		}
		if e.values[1]-size < minY {
			minY = e.values[1] - size
		}
		if e.values[1]+size > maxY {
			maxY = e.values[1] + size
		}
	case Arc:
		if e.values[0] < minX {
			minX = e.values[0]
		}
		if e.values[0] > maxX {
			maxX = e.values[0]
		}
		if e.values[1] < minY {
			minY = e.values[1]
		}
		if e.values[1] > maxY {
			maxY = e.values[1]
		}
		if e.values[2] < minX {
			minX = e.values[2]
		}
		if e.values[2] > maxX {
			maxX = e.values[2]
		}
		if e.values[3] < minY {
			minY = e.values[3]
		}
		if e.values[3] > maxY {
			maxY = e.values[3]
		}
		if e.values[4] < minX {
			minX = e.values[4]
		}
		if e.values[4] > maxX {
			maxX = e.values[4]
		}
		if e.values[5] < minY {
			minY = e.values[5]
		}
		if e.values[5] > maxY {
			maxY = e.values[5]
		}
	}
	return minX, minY, maxX, maxY
}

func (e *Element) DrawToSVG(s *Sketch, ctx *canvas.Context, mult float64) {
	ctx.SetStrokeColor(canvas.Blue)
	if e.elementType != Axis && e.ConstraintLevel() == el.FullyConstrained {
		ctx.SetStrokeColor(canvas.Black)
	}
	if e.elementType != Axis && e.ConstraintLevel() == el.OverConstrained {
		ctx.SetStrokeColor(canvas.Red)
	}
	ctx.StrokeWidth = 0.5
	switch e.elementType {
	case Point:
		// May want to draw a small filled circle
		ctx.MoveTo(e.values[0]*mult+0.5, e.values[1]*mult)
		ctx.Arc(0.5, 0.5, 0, 0, 360)
	case Line:
		x1 := e.values[0] * mult
		y1 := e.values[1] * mult
		ctx.MoveTo(x1, y1)
		x2 := e.values[2] * mult
		y2 := e.values[3] * mult
		ctx.LineTo(x2, y2)
	case Circle:
		cx := e.values[0] * mult
		cy := e.values[1] * mult
		// find distance constraint on e
		r := e.values[2] * mult
		ctx.MoveTo(cx, cy)
		ctx.Arc(r, r, 0, 0, 360)
	case Arc:
		cx := e.values[0] * mult
		cy := e.values[1] * mult
		sx := e.values[2] * mult
		sy := e.values[3] * mult
		ex := e.values[4] * mult
		ey := e.values[5] * mult
		r := math.Sqrt(math.Pow(sx-cx, 2) + math.Pow(sy-cy, 2))
		svx := sx - cx
		svy := sy - cy
		evx := ex - cx
		evy := ey - cy
		theta0 := math.Atan2(svx, svy)
		theta1 := math.Atan2(evx, evy)
		dot := evx*svx + evy*svy
		det := evx*svy - evy*svx
		angle := math.Atan2(det, dot)
		large := false
		if angle > math.Pi {
			large = true
		}

		sweep := theta1 < theta0
		ctx.MoveTo(sx, sy)
		ctx.ArcTo(r, r, angle, large, sweep, ex, ey)
	}
	ctx.Stroke()
	e.valuePass = s.passes
}

func (e *Element) Center() *Element {
	if e.elementType != Arc && e.elementType != Circle {
		return nil
	}
	return e.children[0]
}

func (e *Element) Start() *Element {
	if e.elementType == Arc {
		return e.children[1]
	}
	if e.elementType != Line {
		return nil
	}
	return e.children[0]
}

func (e *Element) End() *Element {
	if e.elementType == Arc {
		return e.children[2]
	}
	if e.elementType != Line {
		return nil
	}
	return e.children[1]
}

func (e *Element) PointVerticalFrom(x, y float64) (float64, float64, bool) {
	if e.elementType != Line && e.elementType != Axis {
		log.Debug().Msg("element is not a line or axis")
		return 0, 0, false
	}
	l := e.element.AsLine()
	if l == nil {
		log.Debug().Msg("element is not castable to Line")
		return 0, 0, false
	}
	start := l.NearestPoint(big.NewFloat(0), big.NewFloat(0))
	var t1 big.Float
	var newY float64
	t1.Neg(l.GetA())
	if utils.StandardBigFloatCompare(&t1, big.NewFloat(0)) == 0 {
		if l.GetB().Cmp(big.NewFloat(0)) == 0 {
			// a and b shouldn't both be 0
			return 0, 0, false
		}
		// horizontal line
		c, _ := l.GetC().Float64()
		b, _ := l.GetB().Float64()
		newY = -c / b
	} else if utils.StandardBigFloatCompare(l.GetB(), big.NewFloat(0)) == 0 {
		log.Debug().Msg("incorrect slope")
		return 0, 0, false // if our element is already a vertical line, a vertical distance constraint makes no sense
	} else {
		// newY = (slope * (x - start.X)) + start.Y
		slope, _ := l.GetSlope().Float64()
		startX, _ := start.X.Float64()
		startY, _ := start.Y.Float64()
		newY = (slope * (x - startX)) + startY
	}
	return x, newY, true
}

func (e *Element) PointHorizontalFrom(x, y float64) (float64, float64, bool) {
	if e.elementType != Line && e.elementType != Axis {
		return 0, 0, false
	}
	l := e.element.AsLine()
	if l == nil {
		return 0, 0, false
	}
	var zero, t1 big.Float
	zero.SetFloat64(0)
	start := l.NearestPoint(&zero, &zero)
	var newX float64
	t1.Neg(l.GetA())
	if utils.StandardBigFloatCompare(&t1, &zero) == 0 {
		return 0, 0, false // if our element is already a vertical line, a vertical distance constraint makes no sense
	} else if utils.StandardBigFloatCompare(l.GetB(), &zero) == 0 {
		// vertical line
		c, _ := l.GetC().Float64()
		a, _ := l.GetA().Float64()
		newX = -c / a
	} else {
		slope, _ := l.GetSlope().Float64()
		startX, _ := start.X.Float64()
		startY, _ := start.Y.Float64()
		newX = ((y - startY) / slope) + startX
	}
	return newX, y, true
}

func (e *Element) DistanceBetweenPoints(other *Element) float64 {
	if e.elementType != Point || other.elementType != Point {
		return math.NaN()
	}
	p := e.element.AsPoint()
	o := other.element.AsPoint()
	dist, _ := p.DistanceTo(o).Float64()
	return dist
}

func (e *Element) String() string {
	values := make([]string, len(e.values))
	for i, value := range e.values {
		values[i] = fmt.Sprintf("%f", value)
	}
	valueString := strings.Join(values, ", ")
	return fmt.Sprintf("Element type %v, internal element: %v, values: %s", e.elementType, e.element, valueString)
}
