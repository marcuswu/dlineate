package dlineate

import "github.com/marcuswu/dlineate/internal/element"

func (s *Sketch) AddCoincidentConstraint(p1 *Element, p2 *Element) *Constraint {
	// If two points are coincident, they are the same point -- make them reference the same element
	if p1.elementType == Point && p2.elementType == Point {
		if p2.element.GetID() == 0 {
			return s.AddCoincidentConstraint(p2, p1)
		}

		newElement := s.sketch.CombinePoints(p1.element, p2.element)
		// Anywhere we referenced p2, we should now reference p1
		s.ReplaceElement(p1.element.GetID(), newElement)
		s.ReplaceElement(p2.element.GetID(), newElement)
		p1.element = newElement
		p2.element = newElement
		// These elements must now reference the same constraints
		for _, c := range s.eToC[p2.id] {
			c.replaceElement(p2, p1)
		}
		s.eToC[p1.id] = append(s.eToC[p1.id], s.eToC[p2.id]...)
		delete(s.eToC, p2.id)
		p2.id = p1.id
		return nil
	}
	c := s.AddDistanceConstraint(p1, p2, 0)
	c.constraintType = Coincident
	return c
}

func (s *Sketch) ReplaceElement(original uint, new element.SketchElement) {
	for _, e := range s.Elements {
		e.replaceElement(original, new)
	}
}

/*
 A point is coincident with a line segment or arc if:
  * The point is coincident with the line or arc /and/
  * The distance from point to the start and end is less than the segment lenth
*/
