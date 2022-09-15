package dlineation

import (
	"fmt"
	"math"

	el "github.com/marcuswu/dlineation/internal/element"
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
	c := RatioConstraint(p1, p2)

	if p1.elementType != Point || p2.elementType != Point {
		return nil
	}
	if !p1.isLineOrArc() && !p2.isLineOrArc() {
		return nil
	}
	s.eToC[p1.element.GetID()] = append(s.eToC[p1.element.GetID()], c)
	s.eToC[p2.element.GetID()] = append(s.eToC[p2.element.GetID()], c)

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

	// Line tests
	dist, ok := s.resolveLineLength(other)
	if ok {
		// coincident with line
		constraint := s.addDistanceConstraint(other, point, 0)
		fmt.Printf("resolveMidpointConstraint: added constraint id %d\n", constraint.GetID())
		other.constraints = append(other.constraints, constraint)
		point.constraints = append(point.constraints, constraint)
		c.constraints = append(c.constraints, constraint)
		// distance from start
		constraint = s.addDistanceConstraint(other.children[0], point, dist/2.0)
		fmt.Printf("resolveMidpointConstraint: added constraint id %d\n", constraint.GetID())
		other.children[0].constraints = append(other.children[0].constraints, constraint)
		point.constraints = append(point.constraints, constraint)
		c.constraints = append(c.constraints, constraint)
		s.constraints = append(s.constraints, c)
		c.state = Resolved

		return c.state == Resolved
	}

	// Ensure start, end, and center of arc is fully constrained and solved
	// calculate angle between lines formed from center to start and center to end
	// calculate line through center with half that angle
	// place midpoint at radius distance from center along calculated line
	centerSolved := s.isElementSolved(other.children[0])
	startSolved := s.isElementSolved(other.children[1])
	endSolved := s.isElementSolved(other.children[2])
	if centerSolved && startSolved && endSolved {
		centerX := other.children[0].values[0]
		centerY := other.children[0].values[1]
		startX := other.children[1].values[0]
		startY := other.children[1].values[1]
		endX := other.children[2].values[0]
		endY := other.children[2].values[1]
		// Calculate vector from center to start
		start := el.Vector{X: startX - centerX, Y: startY - centerY}
		// Calculate vector from center to end
		end := el.Vector{X: endX - centerX, Y: endY - centerY}

		// Calculate center vector
		halfAngle := start.AngleTo(&end) / 2.0
		start.Rotate(halfAngle)
		midPoint := start.Translated(centerX, centerY)

		// Calculate distance from point to start / end
		a := midPoint.X - startX
		b := midPoint.Y - startY
		midDist := math.Sqrt((a * a) + (b * b))
		// Set coincident and distance constraints
		constraint := s.addDistanceConstraint(other.children[1], point, midDist)
		fmt.Printf("resolveMidpointConstraint: added constraint id %d\n", constraint.GetID())
		other.children[1].constraints = append(other.children[1].constraints, constraint)
		point.constraints = append(point.constraints, constraint)
		c.constraints = append(c.constraints, constraint)
		constraint = s.addDistanceConstraint(point, other, 0)
		fmt.Printf("resolveMidpointConstraint: added constraint id %d\n", constraint.GetID())
		other.constraints = append(other.constraints, constraint)
		point.constraints = append(point.constraints, constraint)
		c.constraints = append(c.constraints, constraint)
		s.constraints = append(s.constraints, c)
		c.state = Resolved

		return c.state == Resolved
	}

	return c.state == Resolved
}
