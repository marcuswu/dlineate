package dlineation

import "github.com/marcuswu/dlineation/utils"

func (s *Sketch) AddParallelConstraint(p1 *Element, p2 *Element) (*Constraint, error) {
	c, e := s.AddAngleConstraint(p1, p2, 0)
	if e != nil {
		utils.Logger.Error().Msgf("error: %s", e)
	}
	if c != nil {
		c.constraintType = Parallel
	}
	return c, e
}
