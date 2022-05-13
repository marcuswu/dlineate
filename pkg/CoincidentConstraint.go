package dlineation

func (s *Sketch) AddCoincidentConstraint(p1 *Element, p2 *Element) *Constraint {
	c := s.AddDistanceConstraint(p1, p2, 0)
	c.constraintType = Coincident
	return c
}

/*
 A point is coincident with a line segment or arc if:
  * The point is coincident with the line or arc /and/
  * The distance from point to the start and end is less than the segment lenth
*/
