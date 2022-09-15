package dlineation

func (s *Sketch) AddCoincidentConstraint(p1 *Element, p2 *Element) *Constraint {
	// If two points are coincident, they are the same point -- make them reference the same element
	if p1.elementType == Point && p2.elementType == Point {
		p1.element = s.sketch.CombinePoints(p1.element, p2.element)
		// search for references to element id, replace with new element id
		s.eToC[p1.element.GetID()] = append(s.eToC[p1.element.GetID()], s.eToC[p2.element.GetID()]...)
		delete(s.eToC, p2.element.GetID())
		p2.element = p1.element
		return nil
	}
	c := s.AddDistanceConstraint(p1, p2, 0)
	c.constraintType = Coincident
	return c
}

/*
 A point is coincident with a line segment or arc if:
  * The point is coincident with the line or arc /and/
  * The distance from point to the start and end is less than the segment lenth
*/
