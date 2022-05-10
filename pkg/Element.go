package dlineate

import (
	"errors"

	c "github.com/marcuswu/dlineate/internal/constraint"
	el "github.com/marcuswu/dlineate/internal/element"
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
