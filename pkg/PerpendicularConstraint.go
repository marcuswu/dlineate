package dlineate

func (s *Sketch) AddPerpendicularConstraint(p1 *Element, p2 *Element) (*Constraint, error) {
	c, e := s.AddAngleConstraint(p1, p2, 90)
	c.constraintType = Perpendicular
	return c, e
}
