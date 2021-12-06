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
