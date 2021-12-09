package dlineate

func (s *Sketch) AddCoincidentConstraint(p1 *Element, p2 *Element) *Constraint {
	c := s.AddDistanceConstraint(p1, p2, 0)
	c.constraintType = Coincident
	return c
}
