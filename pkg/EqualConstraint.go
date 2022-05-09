package dlineate

func (s *Sketch) AddEqualConstraint(p1 *Element, p2 *Element) *Constraint {
	c := RatioConstraint(p1, p2)

	if p1.elementType == Point || p2.elementType == Point {
		return nil
	}
	c.dataValue = 1
	s.eToC[p1.element.GetID()] = append(s.eToC[p1.element.GetID()], c)
	s.eToC[p2.element.GetID()] = append(s.eToC[p2.element.GetID()], c)

	s.resolveRatioConstraint(c)

	return c
}
