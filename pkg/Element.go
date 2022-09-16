package dlineation

import (
	"errors"
	"fmt"
	"math"

	svg "github.com/ajstarks/svgo"
	c "github.com/marcuswu/dlineation/internal/constraint"
	el "github.com/marcuswu/dlineation/internal/element"
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

type Element struct {
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
	ec.values = make([]float64, 0, 2)
	ec.constraints = make([]*c.Constraint, 0, 1)
	ec.children = make([]*Element, 0, 1)
	ec.isChild = false
	ec.valuePass = 0
	return &Element{}
}

func (e *Element) valuesFromSketch(s *Sketch) error {
	switch e.elementType {
	case Point:
		p := e.element.AsPoint()
		e.values[0] = p.GetX()
		e.values[1] = p.GetY()
	case Axis:
		p := e.element.AsLine()
		e.values[0] = p.GetA()
		e.values[1] = p.GetB()
		e.values[2] = p.GetC()
	case Line:
		p1 := e.children[0].element.AsPoint()
		p2 := e.children[1].element.AsPoint()
		e.values[0] = p1.GetX()
		e.values[1] = p1.GetY()
		e.values[2] = p2.GetX()
		e.values[3] = p2.GetY()
	case Circle:
		/*
			Circle radius is determined either by
			  * a distance constraint against the Circle
			  * a coincident constraint against a Circle with the location of the center constrained
		*/
		var err error = nil
		c := e.children[0].element.AsPoint()
		e.values[0] = c.GetX()
		e.values[1] = c.GetY()
		// find distance constraint on e
		constraint, err := s.findConstraint(Distance, e)
		if err != nil {
			return err
		}
		e.values[2], err = e.getCircleRadius(constraint)
		if err != nil {
			return err
		}
	case Arc:
		center := e.children[0].element.AsPoint()
		start := e.children[1].element.AsPoint()
		end := e.children[2].element.AsPoint()
		e.values[0] = center.GetX()
		e.values[1] = center.GetY()
		e.values[2] = start.GetX()
		e.values[3] = start.GetY()
		e.values[4] = end.GetX()
		e.values[5] = end.GetY()
	}
	e.valuePass = s.passes

	return nil
}

func (e *Element) getCircleRadius(c *Constraint) (float64, error) {
	if e.elementType != Circle {
		return 0, errors.New("can't return radius for a non-circle")
	}
	if c.constraintType == Distance && len(c.elements) == 1 {
		return c.constraints[0].Value, nil
	}
	if c.constraintType == Coincident {
		constraint := c.constraints[0]
		other := constraint.Element1
		if other == e.children[0].element {
			other = constraint.Element2
		}

		return other.DistanceTo(e.children[0].element.AsPoint()), nil
	}

	return 0, errors.New("Constraint type for circle radius myst be Distance or Coincident")
}

func (e *Element) Values(s *Sketch) []float64 {
	if e.valuePass != s.passes {
		e.valuesFromSketch(s)
	}
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

func (e *Element) DrawToSVG(s *Sketch, img *svg.SVG, mult float64) {
	color := "blue"
	if e.elementType == Axis {
		color = "gray"
	}
	if e.elementType != Axis && e.ConstraintLevel() == el.FullyConstrained {
		color = "black"
	}
	switch e.elementType {
	case Point:
		// May want to draw a small filled circle
	case Axis:
		// drawing handled in Solver
	case Line:
		p1 := e.children[0].element.AsPoint()
		p2 := e.children[1].element.AsPoint()
		e.values[0] = p1.GetX()
		e.values[1] = p1.GetY()
		e.values[2] = p2.GetX()
		e.values[3] = p2.GetY()
		x1 := int(e.values[0] * mult)
		y1 := int(e.values[1] * mult)
		x2 := int(e.values[2] * mult)
		y2 := int(e.values[3] * mult)
		img.Line(x1, y1, x2, y2, fmt.Sprintf("fill:none;stroke:%s", color))
	case Circle:
		cx := int(e.values[0] * mult)
		cy := int(e.values[1] * mult)
		// find distance constraint on e
		r := int(e.values[2] * mult)
		img.Circle(cx, cy, r, fmt.Sprintf("fill: none;stroke:%s", color))
	case Arc:
		cx := e.values[0]
		cy := e.values[1]
		sx := e.values[2]
		sy := e.values[3]
		ex := e.values[4]
		ey := e.values[5]
		r := math.Sqrt(math.Pow(sx-cx, 2) + math.Pow(sy-cy, 2))
		svx := sx - cx
		svy := sy - cy
		evx := ex - cx
		evy := ey - cy
		dot := evx*svx + evy*svy
		det := evx*svy - evy*svx
		angle := math.Atan2(det, dot)
		large := false
		if angle > math.Pi {
			large = true
		}

		img.Arc(
			int(sx*mult),
			int(sy*mult),
			int(r*mult),
			int(r*mult),
			0,
			large,
			true,
			int(ex*mult),
			int(ey*mult),
			fmt.Sprintf("fill: none; stroke: %s", color),
		)
	}
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
