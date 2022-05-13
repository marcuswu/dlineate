package solver

import (
	"math"
	"fmt"

	"github.com/marcuswu/dlineation/internal/constraint"
	el "github.com/marcuswu/dlineation/internal/element"
	"github.com/marcuswu/dlineation/utils"
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

func (ss SolveState) String() string {
	switch ss {
	case UnderConstrained:
		return "under constrained"
	case OverConstrained:
		return "over constrained"
	case NonConvergent:
		return "non-convergent"
	case Solved:
		return "solved"
	default:
		return fmt.Sprintf("%d", int(ss))
	}
}

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
	newP3, state := ConstraintResult(c1, c2)

	if newP3 == nil {
		return state
	}

	switch {
	case c1.Element1.Is(c2.Element1):
		c1.Element1.AsPoint().X = newP3.X
		c1.Element1.AsPoint().Y = newP3.Y
		c2.Element1.AsPoint().X = newP3.X
		c2.Element1.AsPoint().Y = newP3.Y
	case c1.Element2.Is(c2.Element1):
		c1.Element2.AsPoint().X = newP3.X
		c1.Element2.AsPoint().Y = newP3.Y
		c2.Element1.AsPoint().X = newP3.X
		c2.Element1.AsPoint().Y = newP3.Y
	case c1.Element1.Is(c2.Element2):
		c1.Element1.AsPoint().X = newP3.X
		c1.Element1.AsPoint().Y = newP3.Y
		c2.Element2.AsPoint().X = newP3.X
		c2.Element2.AsPoint().Y = newP3.Y
	case c1.Element2.Is(c2.Element2):
		c1.Element2.AsPoint().X = newP3.X
		c1.Element2.AsPoint().Y = newP3.Y
		c2.Element2.AsPoint().X = newP3.X
		c2.Element2.AsPoint().Y = newP3.Y
	}

	return state
}

// ConstraintResult returns the result of solving two constraints sharing one point
func ConstraintResult(c1 *constraint.Constraint, c2 *constraint.Constraint) (*el.SketchPoint, SolveState) {
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

	return nil, NonConvergent
}

// SolveDistanceConstraint solves a distance constraint and returns the solution state
func SolveDistanceConstraint(c *constraint.Constraint) SolveState {
	if c.Type != constraint.Distance {
		return NonConvergent
	}

	var point *el.SketchPoint
	var other el.SketchElement
	if c.Element1.GetType() == el.Point {
		point = c.Element1.(*el.SketchPoint)
		other = c.Element2
	} else {
		point = c.Element2.(*el.SketchPoint)
		other = c.Element1
	}

	// If two points, get distance between them, translate constraint value - distance between
	// If point and line, get distance between them, translate normal to line constraint value - distance between
	dist := point.DistanceTo(other)
	trans, ok := point.VectorTo(other).UnitVector()
	if !ok {
		return NonConvergent
	}
	trans.Scaled(c.GetValue() - dist)
	point.Translate(trans.GetX(), trans.GetY())

	return Solved
}

// MoveLineToPoint solves a constraint between a line and a point where the line needs to move
func MoveLineToPoint(c *constraint.Constraint) SolveState {
	if c.Type != constraint.Distance {
		return NonConvergent
	}

	var point *el.SketchPoint
	var line *el.SketchLine
	var e1Type = c.Element1.GetType()
	var e2Type = c.Element2.GetType()
	if e1Type == e2Type {
		return NonConvergent
	}
	if e1Type == el.Point && e2Type == el.Line {
		point = c.Element1.(*el.SketchPoint)
		line = c.Element2.(*el.SketchLine)
	}
	if e2Type == el.Point && e1Type == el.Line {
		point = c.Element2.(*el.SketchPoint)
		line = c.Element1.(*el.SketchLine)
	}

	// If two points, get distance between them, translate constraint value - distance between
	// If point and line, get distance between them, translate normal to line constraint value - distance between
	dist := line.DistanceTo(point)
	line.TranslateDistance(-(c.GetValue() - dist))

	return Solved
}

// SolveAngleConstraint solve an angle constraint between two lines
func SolveAngleConstraint(c *constraint.Constraint) SolveState {
	if c.Type != constraint.Angle {
		return NonConvergent
	}

	l1 := c.Element1.(*el.SketchLine)
	l2 := c.Element2.(*el.SketchLine)
	angle := l1.AngleToLine(l2)
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
		// TODO: Wrong! Fix this!
		return el.NewSketchPoint(p3.GetID(), 0, 0), Solved
	}

	// Solve for p3
	// translate to p1 (p2 and p3)
	p2.ReverseTranslateByElement(p1)
	p3.ReverseTranslateByElement(p1)
	// rotate p2 and p3 so p2 is on x axis
	angle := p2.AngleTo(&el.Vector{X: 1, Y: 0})
	p2.Rotate(angle)
	p3.Rotate(angle)
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
	actualP3.Rotate(-angle)
	// untranslate actualP3
	actualP3.TranslateByElement(p1)

	// return actualP3
	return actualP3, Solved
}

// PointFromPoints calculates a new p3 representing p3 moved to satisfy
// distance constraints from p1 and p2
func PointFromPoints(c1 *constraint.Constraint, c2 *constraint.Constraint) (*el.SketchPoint, SolveState) {
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

	return GetPointFromPoints(p1, p2, p3, p1Radius, p2Radius)
}

func pointFromPointLine(originalP1 el.SketchElement, originalL2 el.SketchElement, originalP3 el.SketchElement, pointDist float64, lineDist float64) (*el.SketchPoint, SolveState) {
	p1 := el.CopySketchElement(originalP1).(*el.SketchPoint)
	l2 := el.CopySketchElement(originalL2)
	p3 := el.CopySketchElement(originalP3).(*el.SketchPoint)
	distanceDifference := l2.DistanceTo(p1)

	// rotate l2 to X axis
	angle := l2.AngleTo(&el.Vector{X: 1, Y: 0})
	l2.Rotate(angle)
	p1.Rotate(angle)
	p3.Rotate(angle)

	// translate l2 to X axis
	yTranslate := l2.(*el.SketchLine).GetOriginDistance() - lineDist
	if utils.StandardFloatCompare(l2.(*el.SketchLine).GetC()-yTranslate, lineDist) != 0 {
		yTranslate *= -1
	}
	l2.Translate(0, yTranslate)
	// move p1 to Y axis
	xTranslate := p1.GetX()
	p1.Translate(-xTranslate, yTranslate)
	p3.Translate(-xTranslate, yTranslate)

	if pointDist < math.Abs(p1.GetY()) {
		return nil, NonConvergent
	}

	// Find points where circle at p1 with radius pointDist intersects with x axis
	xPos := math.Sqrt(math.Abs((pointDist * pointDist) - (p1.GetY() * p1.GetY())))
	if utils.StandardFloatCompare(distanceDifference, 0) == 0 {
		xPos = pointDist
	}

	newP31 := el.NewSketchPoint(p3.GetID(), xPos, 0)
	newP32 := el.NewSketchPoint(p3.GetID(), -xPos, 0)
	actualP3 := newP31
	if newP32.SquareDistanceTo(p3) < newP31.SquareDistanceTo(p3) {
		actualP3 = newP32
	}
	actualP3.Translate(xTranslate, -yTranslate)
	actualP3.Rotate(-angle)

	return actualP3, Solved
}

// PointFromPointLine construct a point from a point and a line. c2 must contain the line.
func PointFromPointLine(c1 *constraint.Constraint, c2 *constraint.Constraint) (*el.SketchPoint, SolveState) {
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

	return pointFromPointLine(p1, l2, p3, pointDist, lineDist)
}

func pointFromLineLine(l1 *el.SketchLine, l2 *el.SketchLine, p3 *el.SketchPoint, line1Dist float64, line2Dist float64) (*el.SketchPoint, SolveState) {
	// If l1 and l2 are parallel, there is no solution
	if utils.StandardFloatCompare(l1.GetSlope(), l2.GetSlope()) == 0 {
		return nil, NonConvergent
	}
	// Translate l1 line1Dist
	line1Translated := l1.TranslatedDistance(line1Dist)
	// Translate l2 line2Dist
	line2Translated := l2.TranslatedDistance(line2Dist)
	// Return intersection point
	intersection := line1Translated.Intersection(line2Translated)

	return el.NewSketchPoint(p3.GetID(), intersection.GetX(), intersection.GetY()), Solved
}

// PointFromLineLine construct a point from two lines. c2 must contain the point.
func PointFromLineLine(c1 *constraint.Constraint, c2 *constraint.Constraint) (*el.SketchPoint, SolveState) {
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

	return pointFromLineLine(l1.AsLine(), l2.AsLine(), p3.AsPoint(), line1Dist, line2Dist)
}
