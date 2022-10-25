package dlineation

import "github.com/marcuswu/dlineation/utils"

func (s *Sketch) AddPerpendicularConstraint(p1 *Element, p2 *Element) (*Constraint, error) {
	c, err := s.AddAngleConstraint(p1, p2, 90)
	if err != nil {
		utils.Logger.Error().Msgf("error: %s", err)
	}
	if c != nil {
		c.constraintType = Perpendicular
	}
	return c, err
}
