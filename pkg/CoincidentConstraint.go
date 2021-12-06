package dlineate

import ic "github.com/marcuswu/dlineate/internal/constraint"

func CoincidentConstraint(p1 *Element, p2 *Element) *Constraint {
	constraint := emptyConstraint()
	append(constraint.elements, p1)
	append(constraint.elements, p2)

	return constraint
}

func (s *Sketch) AddCoincidentConstraint(p1 *Element, p2 *Element) *Constraint {
	c := CoincidentConstraint(p1, p2)

	// Add constraints to sketch
	// TODO: This only works with two points -- Check element types and handle coincidence between other elements
	append(s.constraints, c)
	sketchConstraint := s.sketch.AddConstraint(ic.Distance, p1.elements[0], p2.elements[0], 0)
	append(c.constraints, sketchConstraint)
	return c
}