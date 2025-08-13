package dlineate

import (
	"math/big"

	el "github.com/marcuswu/dlineate/internal/element"
	"github.com/marcuswu/dlineate/utils"
)

func (e *Element) isLineOrArc() bool {
	return e.elementType == Line || e.elementType == Arc
}

func MidpointConstraint(p1 *Element, p2 *Element) *Constraint {
	constraint := emptyConstraint()
	constraint.elements = append(constraint.elements, p1)
	constraint.elements = append(constraint.elements, p2)
	constraint.constraintType = Midpoint
	constraint.state = Unresolved

	return constraint
}

/*
 * A midpoint is coincident AND half the distance away from one end.
 * Only applies to a line or an arc
 */
func (s *Sketch) AddMidpointConstraint(p1 *Element, p2 *Element) *Constraint {
	c := MidpointConstraint(p1, p2)

	if p1.elementType != Point && p2.elementType != Point {
		return nil
	}
	if !p1.isLineOrArc() && !p2.isLineOrArc() {
		return nil
	}
	s.eToC[p1.id] = append(s.eToC[p1.id], c)
	s.eToC[p2.id] = append(s.eToC[p2.id], c)
	s.constraints = append(s.constraints, c)

	s.resolveMidpointConstraint(c)

	return c
}

func (s *Sketch) resolveMidpointConstraint(c *Constraint) bool {
	/*
	 * The line or arc must be fully constrained and solved first
	 */
	point := c.elements[0]
	other := c.elements[1]
	if c.elements[1].elementType == Point {
		point = c.elements[1]
		other = c.elements[0]
	}

	if other.elementType == Line {
		return s.resolveLineMidpoint(c, point, other)
	}

	return s.resolveArcMidpoint(c, point, other)
}

func (s *Sketch) resolveLineMidpoint(c *Constraint, point *Element, other *Element) bool {
	// Line tests
	dist, ok := s.resolveLineLength(other)
	if !ok {
		return false
	}
	// coincident with line
	constraint := s.addDistanceConstraint(other, point, 0)
	if constraint != nil {
		utils.Logger.Debug().
			Uint("constraint", constraint.GetID()).
			Msg("resolveMidpointConstraint: added constraint")
		other.constraints = append(other.constraints, constraint)
		point.constraints = append(point.constraints, constraint)
		c.constraints = append(c.constraints, constraint)
	}
	// distance from start
	constraint = s.addDistanceConstraint(other.children[0], point, dist/2.0)
	if constraint != nil {
		utils.Logger.Debug().
			Uint("constraint", constraint.GetID()).
			Msg("resolveMidpointConstraint: added constraint")
		other.children[0].constraints = append(other.children[0].constraints, constraint)
		point.constraints = append(point.constraints, constraint)
		c.constraints = append(c.constraints, constraint)
	}
	s.constraints = append(s.constraints, c)
	c.state = Resolved

	return c.state == Resolved
}

func (s *Sketch) resolveArcMidpoint(c *Constraint, point *Element, other *Element) bool {
	// Ensure start, end, and center of arc is fully constrained and solved
	// calculate angle between lines formed from center to start and center to end
	// calculate line through center with half that angle
	// place midpoint at radius distance from center along calculated line
	centerSolved := s.isElementSolved(other.children[0])
	startSolved := s.isElementSolved(other.children[1])
	endSolved := s.isElementSolved(other.children[2])
	constrainedAndSolved := centerSolved && startSolved && endSolved
	if !constrainedAndSolved {
		return c.state == Resolved
	}

	var centerX, centerY, startX, startY, endX, endY big.Float
	centerX.SetPrec(utils.FloatPrecision).SetFloat64(other.children[0].values[0])
	centerY.SetPrec(utils.FloatPrecision).SetFloat64(other.children[0].values[1])
	startX.SetPrec(utils.FloatPrecision).SetFloat64(other.children[1].values[0])
	startY.SetPrec(utils.FloatPrecision).SetFloat64(other.children[1].values[1])
	endX.SetPrec(utils.FloatPrecision).SetFloat64(other.children[2].values[0])
	endY.SetPrec(utils.FloatPrecision).SetFloat64(other.children[2].values[1])
	// Calculate vector from center to start
	var x1, y1, x2, y2 big.Float
	x1.Sub(&startX, &centerX)
	y1.Sub(&startY, &centerY)
	start := el.Vector{X: x1, Y: y1}
	// Calculate vector from center to end
	x2.Sub(&endX, &centerX)
	y2.Sub(&endY, &centerY)
	end := el.Vector{X: x2, Y: y2}

	// Calculate center vector
	var two big.Float
	two.SetFloat64(2)
	halfAngle := start.AngleTo(&end)
	halfAngle.Quo(halfAngle, &two)
	start.Rotate(halfAngle)
	midPoint := start.Translated(&centerX, &centerY)

	// Calculate distance from point to start / end
	var a, b, midDist, t1 big.Float
	t1.Copy(&startX)
	a.Sub(&midPoint.X, &t1)
	t1.Copy(&startY)
	b.Sub(&midPoint.Y, &t1)
	midDist.Mul(&a, &a)
	t1.Mul(&b, &b)
	midDist.Add(&midDist, &t1)
	// Set coincident and distance constraints
	dist, _ := midDist.Float64()
	constraint := s.addDistanceConstraint(other.children[1], point, dist)
	if constraint != nil {
		utils.Logger.Debug().
			Uint("constraint", constraint.GetID()).
			Msg("resolveMidpointConstraint: added constraint")
		other.children[1].constraints = append(other.children[1].constraints, constraint)
		point.constraints = append(point.constraints, constraint)
		c.constraints = append(c.constraints, constraint)
	}
	arcRadius := other.children[1].element.DistanceTo(other.children[0].element)
	radius, _ := arcRadius.Float64()
	constraint = s.addDistanceConstraint(point, other.Center(), radius)
	if constraint != nil {
		utils.Logger.Debug().
			Uint("constraint", constraint.GetID()).
			Msg("resolveMidpointConstraint: added constraint")
		other.constraints = append(other.constraints, constraint)
		point.constraints = append(point.constraints, constraint)
		c.constraints = append(c.constraints, constraint)
	}
	s.constraints = append(s.constraints, c)
	c.state = Resolved

	return c.state == Resolved
}
