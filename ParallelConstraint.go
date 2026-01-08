package dlineate

import (
	"errors"

	"github.com/marcuswu/dlineate/utils"
)

func (s *Sketch) AddParallelConstraint(p1 *Element, p2 *Element) (*Constraint, error) {
	c, e := s.AddAngleConstraint(p1, p2, 0, false)
	if e != nil {
		utils.Logger.Error().Msgf("error: %s", e)
	}
	if c != nil {
		c.constraintType = Parallel
	}
	return c, e
}

func (s *Sketch) AddHorizontalConstraint(p1 *Element) (*Constraint, error) {
	// Check p1's points to see if start is right of end; if so, angle shold be 180 degrees
	if p1.elementType != Line && p1.elementType != Axis {
		return nil, errors.New("incorrect element types for vertical constraint")
	}
	return s.AddParallelConstraint(p1, s.XAxis)
}

func (s *Sketch) AddVerticalConstraint(p1 *Element) (*Constraint, error) {
	// Check p1's points to see if start is below end; if so, angle shold be 180 degrees
	if p1.elementType != Line && p1.elementType != Axis {
		return nil, errors.New("incorrect element types for vertical constraint")
	}
	return s.AddParallelConstraint(p1, s.YAxis)
}
