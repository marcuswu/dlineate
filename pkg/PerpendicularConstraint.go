package dlineation

import "fmt"

func (s *Sketch) AddPerpendicularConstraint(p1 *Element, p2 *Element) (*Constraint, error) {
	c, err := s.AddAngleConstraint(p1, p2, 90)
	if err != nil {
		fmt.Printf("error: %s\n", err)
	}
	if c != nil {
		c.constraintType = Perpendicular
	}
	return c, err
}
