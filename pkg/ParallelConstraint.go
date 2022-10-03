package dlineation

import "fmt"

func (s *Sketch) AddParallelConstraint(p1 *Element, p2 *Element) (*Constraint, error) {
	c, e := s.AddAngleConstraint(p1, p2, 0)
	if e != nil {
		fmt.Printf("error: %s\n", e)
	}
	if c != nil {
		c.constraintType = Parallel
	}
	return c, e
}
