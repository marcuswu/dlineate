package dlineate

import (
	"errors"
	el "github.com/marcuswu/dlineate/internal/element"
	c "github.com/marcuswu/dlineate/internal/constraint"
	core "github.com/marcuswu/dlineate/internal/core"
)

// Type of a Constraint(Distance or Angle)
type ElementType uint

// ElementType constants
const (
	Point ElementType = iota
	Line
	Circle
	Arc
)

type Element struct {
	values []float64
	elementType ElementType
	constraints []*c.Constraint
	elements []*el.SketchElement
}

func emptyElement() *Element{
	ec := new(Element)
	ec.values = make([]float64, 0, 2)
	ec.constraints = make([]*c.Constraint, 0, 1)
	ec.elements = make([]*el.SketchElement, 0, 1)
	return &Element{}
}

func (e *Element) addToSketch(s *core.SketchGraph) {
	switch elementType {
	case Point: 
		append(s.elements, s.AddPoint(e.values[0], e.values[1]))
	case Line:
		// calculate a, b, c from point [0, 1] & [2, 3]
		a := e.values[3] - e.values[1] // y' - y
		b := e.values[0] - e.values[2] // x - x'
		c := (-a * e.values[0]) - (b * e.values[1]) // c = -ax - by from ax + by + c = 0
		append(e.elements, s.AddLine(a, b, c)) // AddLine normalizes a, b, c
		append(e.elements, s.AddPoint(e.values[0], e.values[1]))
		append(e.elements, s.AddPoint(e.values[2], e.values[3]))
		append(e.constraints, s.AddConstraint(c.Distance, e.elements[0], e.elements[1], 0.0))
		append(e.constraints, s.AddConstraint(c.Distance, e.elements[0], e.elements[2], 0.0))
	case Circle:
		append(e.elements, s.AddPoint(e.values[0], e.values[1])) 
	case Arc:
		append(e.elements, s.AddPoint(e.values[0], e.values[1]))
		append(e.elements, s.AddPoint(e.values[2], e.values[3]))
		append(e.elements, s.AddPoint(e.values[4], e.values[5]))
	}
}

func (e *Element) valuesFromSketch() error {
	switch elementType {
	case Point:
		p := e.elements[0].AsPoint()
		e.values[0] = p.GetX()
		e.values[1] = p.GetY()
	case Line:
		p1 := e.elements[1].AsPoint()
		p2 := e.elements[2].AsPoint()
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
		c := e.elements[0].AsPoint()
		e.values[0] = c.GetX()
		e.values[1] = c.GetY()
		e.values[2], err = e.getCircleRadius(c.GetX(), c.GetY())
		if err != nil {
			return error
		}
	case Arc:
		c := e.elements[0].AsPoint()
		s := e.elements[1].AsPoint()
		e := e.elements[2].AsPoint()
		e.values[0] = c.GetX()
		e.values[1] = c.GetY()
		e.values[2] = s.GetX()
		e.values[3] = s.GetY()
		e.values[4] = e.GetX()
		e.values[5] = e.GetY()
	}
}

func (e *Element) getCircleRadius(Constraint c) (float64, error) {
	if e.elementType != Circle {
		return 0, errors.New("Can't return radius for a non-circle")
	}
	if c.constraintType == Distance && len(c.elements) == 1 {
		return c.constraints[0].Value, nil
	}
	if c.constraintType == Coincident {
		constraint := c.constraints[0]
		other := constraint.Element1
		if (other == e) { other = constraint.Element2 }

		return other.DistanceTo(e.elements[0].AsPoint()), nil
	}

	return 0, errors.New("Constraint type for circle radius myst be Distance or Coincident")
}

func (e *Element) Values() []float64 {
	e.valuesFromSketch()
	return e.values
}

func NewPoint(x double, y double) *Element {
	p := emptyElement()
	p.elementType = Point
	append(p.values, x)
	append(p.values, y)
	return p
}

func NewLine(x1 double, y1 double, x2 double, y2 double) *Element {
	l := emptyElement()
	l.elementType = Line
	append(l.values, x1)
	append(l.values, y1)
	append(l.values, x2)
	append(l.values, y2)
	return l
}

func NewCircle(x double, y double, r double) *Element {
	c := emptyElement()
	c.elementType = Circle
	append(c.values, x)
	append(c.values, y)
	append(c.values, r)
	return c
}

func NewArc(x1 double, y1 double, x2 double, y2 double, x3 double, y3 double) *Element {
	a := emptyElement()
	a.elementType = Arc
	append(a.values, x1)
	append(a.values, y1)
	append(a.values, x2)
	append(a.values, y2)
	append(a.values, x3)
	append(a.values, y3)
	return a
}