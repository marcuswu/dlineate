package dlineate

func (s *Sketch) AddEqualConstraint(p1 *Element, p2 *Element) *Constraint {
	c := RatioConstraint(p1, p2)

	if p1.elementType == Point || p2.elementType == Point {
		return nil
	}
	c.dataValue = 1

	s.resolveRatioConstraint(c)

	return c
}
