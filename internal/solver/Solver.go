package solver

import (
	"math"

	"github.com/marcuswu/dlineate/internal/constraint"
	el "github.com/marcuswu/dlineate/internal/element"
	"github.com/marcuswu/dlineate/utils"
)

// SolveState The state of the sketch graph
type SolveState uint

// SolveState constants
const (
	None SolveState = iota
	UnderConstrained
	OverConstrained
	NonConvergent
	Solved
)

func typeCounts(c1 *constraint.Constraint, c2 *constraint.Constraint) (int, int) {
	numPoints := 0
	numLines := 0
	elements := []el.SketchElement{c1.Element1, c1.Element2, c2.Element1, c2.Element2}

	for _, element := range elements {
		if element.GetType() == el.Point {
			numPoints++
		} else {
			numLines++
		}
	}

	return numPoints, numLines
}

// SolveConstraints solve two constraints and return the solution state
func SolveConstraints(c1 *constraint.Constraint, c2 *constraint.Constraint) SolveState {
	numPoints, _ := typeCounts(c1, c2)
	// 4 points -> PointFromPoints
	if numPoints == 4 {
		return PointFromPoints(c1, c2)
	}

	// 3 points, 1 line -> PointFromPointLine
	if numPoints == 3 {
		return PointFromPointLine(c1, c2)
	}
	// 2 points, 2 lines -> PointFromLineLine
	if numPoints == 2 {
		return PointFromLineLine(c1, c2)
	}

	return NonConvergent
}

// SolveAngleConstraint solve an angle constraint between two lines
func SolveAngleConstraint(c *constraint.Constraint) SolveState {
	if c.Type != constraint.Angle {
		return NonConvergent
	}

	l1 := c.Element1.(*el.SketchLine)
	l2 := c.Element2.(*el.SketchLine)
	angle := l2.AngleToLine(l1)
	rotate := c.Value - angle
	l2.Rotate(rotate)
	return Solved
}

// GetPointFromPoints calculates where a 3rd point exists in relation to two others with
// distance constraints from the first two
func GetPointFromPoints(p1 el.SketchElement, originalP2 el.SketchElement, originalP3 el.SketchElement, p1Radius float64, p2Radius float64) (*el.SketchPoint, SolveState) {
	// Don't mutate the originals
	p2 := el.CopySketchElement(originalP2)
	p3 := el.CopySketchElement(originalP3)
	pointDistance := p1.DistanceTo(p2)
	constraintDist := p1Radius + p2Radius

	if pointDistance > constraintDist {
		return nil, NonConvergent
	}

	if utils.StandardFloatCompare(pointDistance, constraintDist) == 0 {
		return el.NewSketchPoint(p3.GetID(), 0, 0), Solved
	}

	// Solve for p3
	// translate to p1 (p2 and p3)
	p2.ReverseTranslateByElement(p1)
	p3.ReverseTranslateByElement(p1)
	// rotate p2 and p3 so p2 is on x axis
	angle := p2.AngleTo(el.Vector{X: 1, Y: 0})
	p2.Rotate(-angle)
	p3.Rotate(-angle)
	// calculate possible p3s
	p2Dist := p2.(*el.SketchPoint).GetX()

	// https://mathworld.wolfram.com/Circle-CircleIntersection.html
	xDelta := ((-(p2Radius * p2Radius) + (p2Dist * p2Dist)) + (p1Radius * p1Radius)) / (2 * p2Dist)
	yDelta := math.Sqrt((p1Radius * p1Radius) - (xDelta * xDelta))
	p3X := xDelta
	p3Y1 := yDelta
	p3Y2 := -yDelta
	// determine which is closest to the p3 from constraint
	newP31 := el.NewSketchPoint(p3.GetID(), p3X, p3Y1)
	newP32 := el.NewSketchPoint(p3.GetID(), p3X, p3Y2)
	actualP3 := newP31
	if newP32.SquareDistanceTo(p3) < newP31.SquareDistanceTo(p3) {
		actualP3 = newP32
	}
	// unrotate actualP3
	actualP3.Rotate(angle)
	// untranslate actualP3
	actualP3.TranslateByElement(p1)

	// return actualP3
	return actualP3, Solved
}

// PointFromPoints calculates a new p3 representing p3 moved to satisfy
// distance constraints from p1 and p2
func PointFromPoints(c1 *constraint.Constraint, c2 *constraint.Constraint) SolveState {
	p1 := c1.Element1
	p2 := c2.Element1
	p3 := c1.Element2
	p1Radius := c1.GetValue()
	p2Radius := c2.GetValue()

	switch {
	case c1.Element1.Is(c2.Element1):
		p3, p1, p2 = c1.Element1, c1.Element2, c2.Element2
	case c1.Element2.Is(c2.Element1):
		p3, p1, p2 = c1.Element2, c1.Element1, c2.Element2
	case c1.Element1.Is(c2.Element2):
		p3, p1, p2 = c1.Element1, c1.Element2, c2.Element1
	case c1.Element2.Is(c2.Element2):
		break
	}

	newP3, solved := GetPointFromPoints(p1, p2, p3, p1Radius, p2Radius)

	switch {
	case c1.Element1.Is(c2.Element1):
		c1.Element1 = newP3
		c2.Element1 = newP3
	case c1.Element2.Is(c2.Element1):
		c1.Element2 = newP3
		c2.Element1 = newP3
	case c1.Element1.Is(c2.Element2):
		c1.Element1 = newP3
		c2.Element2 = newP3
	case c1.Element2.Is(c2.Element2):
		c1.Element2 = newP3
		c2.Element2 = newP3
	}

	return solved
}

func pointFromPointLine(originalP1 el.SketchElement, originalL2 el.SketchElement, originalP3 el.SketchElement, pointDist float64, lineDist float64) (*el.SketchPoint, SolveState) {
	p1 := el.CopySketchElement(originalP1).(*el.SketchPoint)
	l2 := el.CopySketchElement(originalL2)
	p3 := el.CopySketchElement(originalP3).(*el.SketchPoint)
	distanceDifference := l2.DistanceTo(p1)

	if distanceDifference+pointDist < lineDist+pointDist {
		return nil, NonConvergent
	}

	if distanceDifference > lineDist+pointDist {
		return nil, NonConvergent
	}

	if distanceDifference == lineDist {
		return el.NewSketchPoint(p3.GetID(), p1.GetX(), p1.GetY()-pointDist), Solved
	}

	// rotate l2 to X axis
	angle := l2.AngleTo(el.Vector{X: 1, Y: 0})
	l2.Rotate(-angle)
	p1.Rotate(-angle)
	p3.Rotate(-angle)
	// translate l2 to X axis
	yTranslate := l2.(*el.SketchLine).GetOriginDistance() - lineDist
	l2.Translate(0, yTranslate)
	// move p1 to Y axis
	xTranslate := -p1.GetX()
	p1.Translate(xTranslate, yTranslate)
	p3.Translate(xTranslate, yTranslate)

	// Find points where circle at p1 with radius pointDist intersects with x axis
	xPos := math.Sqrt((pointDist * pointDist) - (p1.GetY() * p1.GetY()))
	newP31 := el.NewSketchPoint(p3.GetID(), xPos, 0)
	newP32 := el.NewSketchPoint(p3.GetID(), -xPos, 0)
	actualP3 := newP31
	if newP32.SquareDistanceTo(p3) < newP31.SquareDistanceTo(p3) {
		actualP3 = newP32
	}
	actualP3.Translate(-xTranslate, -yTranslate)
	actualP3.Rotate(angle)

	return actualP3, Solved
}

// PointFromPointLine construct a point from a point and a line. c2 must contain the line.
func PointFromPointLine(c1 *constraint.Constraint, c2 *constraint.Constraint) SolveState {
	p1 := c1.Element1
	l2 := c2.Element1
	p3 := c1.Element2
	pointDist := c1.GetValue()
	lineDist := c2.GetValue()

	switch {
	case c1.Element1.Is(c2.Element1):
		p3 = c1.Element1
		p1 = c1.Element2
		l2 = c2.Element2
	case c1.Element2.Is(c2.Element1):
		p3 = c1.Element2
		p1 = c1.Element1
		l2 = c2.Element2
	case c1.Element1.Is(c2.Element2):
		p3 = c1.Element1
		p1 = c1.Element2
		l2 = c2.Element1
	case c1.Element2.Is(c2.Element2):
		break
	}

	if p1.GetType() == el.Line && l2.GetType() == el.Point {
		p1, l2 = l2, p1
		pointDist, lineDist = lineDist, pointDist
	}

	newP3, solved := pointFromPointLine(p1, l2, p3, pointDist, lineDist)

	switch {
	case c1.Element1.Is(c2.Element1):
		c1.Element1 = newP3
		c2.Element1 = newP3
	case c1.Element2.Is(c2.Element1):
		c1.Element2 = newP3
		c2.Element1 = newP3
	case c1.Element1.Is(c2.Element2):
		c1.Element1 = newP3
		c2.Element2 = newP3
	case c1.Element2.Is(c2.Element2):
		c1.Element2 = newP3
		c2.Element2 = newP3
	}

	return solved
}

func pointFromLineLine(originalL1 el.SketchElement, originalL2 el.SketchElement, originalP3 el.SketchElement, line1Dist float64, line2Dist float64) (*el.SketchPoint, SolveState) {
	l1 := el.CopySketchElement(originalL1)
	l2 := el.CopySketchElement(originalL2)
	p3 := el.CopySketchElement(originalP3)
	// If l1 and l2 are parallel, there is no solution
	line1, line2 := l1.(*el.SketchLine), l2.(*el.SketchLine)
	if line1.GetSlope() == line2.GetSlope() {
		return nil, NonConvergent
	}
	// Translate l1 line1Dist
	line1Translated := line1.TranslatedDistance(line1Dist)
	// Translate l2 line2Dist
	line2Translated := line2.TranslatedDistance(line2Dist)
	// Return intersection point
	intersection := line1Translated.Intersection(line2Translated)

	return el.NewSketchPoint(p3.GetID(), intersection.GetX(), intersection.GetY()), Solved
}

// PointFromLineLine construct a point from two lines. c2 must contain the point.
func PointFromLineLine(c1 *constraint.Constraint, c2 *constraint.Constraint) SolveState {
	l1 := c1.Element1
	l2 := c2.Element1
	p3 := c1.Element2
	line1Dist := c1.GetValue()
	line2Dist := c2.GetValue()

	switch {
	case c1.Element1.Is(c2.Element1):
		p3 = c1.Element1
		l1 = c1.Element2
		l2 = c2.Element2
	case c1.Element2.Is(c2.Element1):
		p3 = c1.Element2
		l1 = c1.Element1
		l2 = c2.Element2
	case c1.Element1.Is(c2.Element2):
		p3 = c1.Element1
		l1 = c1.Element2
		l2 = c2.Element1
	case c1.Element2.Is(c2.Element2):
		break
	}

	newP3, solved := pointFromLineLine(l1, l2, p3, line1Dist, line2Dist)

	switch {
	case c1.Element1.Is(c2.Element1):
		c1.Element1 = newP3
		c2.Element1 = newP3
	case c1.Element2.Is(c2.Element1):
		c1.Element2 = newP3
		c2.Element1 = newP3
	case c1.Element1.Is(c2.Element2):
		c1.Element1 = newP3
		c2.Element2 = newP3
	case c1.Element2.Is(c2.Element2):
		c1.Element2 = newP3
		c2.Element2 = newP3
	}

	return solved
}
