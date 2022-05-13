package dlineation

func (s *Sketch) AddParallelConstraint(p1 *Element, p2 *Element) (*Constraint, error) {
	c, e := s.AddAngleConstraint(p1, p2, 0)
	c.constraintType = Perpendicular
	return c, e
}
