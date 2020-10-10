package core

import (
	"math"

	"github.com/marcuswu/dlineation/utils"
)

func pointFromPoints(p1 SketchElement, originalP2 SketchElement, originalP3 SketchElement, p1Radius float64, p2Radius float64) (SketchElement, SolveState) {
	// Don't mutate the originals
	p2 := CopySketchElement(originalP2)
	p3 := CopySketchElement(originalP3)
	pointDistance := p1.DistanceTo(p2)
	constraintDist := p1Radius + p2Radius

	if pointDistance > constraintDist {
		return nil, NonConvergent
	}

	if utils.StandardFloatCompare(pointDistance, constraintDist) == 0 {
		return &SketchPoint{BaseElement{
			Vector:      Vector{0, 0},
			elementType: Point,
			id:          p3.GetID(),
		}}, Solved
	}

	// Solve for p3
	// translate to p1 (p2 and p3)
	p2.ReverseTranslateByElement(p1)
	p3.ReverseTranslateByElement(p1)
	// rotate p2 and p3 so p2 is on x axis
	angle := p2.AngleTo(Vector{1, 0})
	p2.Rotate(-angle)
	p3.Rotate(-angle)
	// calculate possible p3s
	p2Dist := p2.GetX()

	// https://mathworld.wolfram.com/Circle-CircleIntersection.html
	xDelta := ((-(p2Radius * p2Radius) + (p2Dist * p2Dist)) + (p1Radius * p1Radius)) / (2 * p2Dist)
	yDelta := math.Sqrt((p1Radius * p1Radius) - (xDelta * xDelta))
	p3X := xDelta
	p3Y1 := yDelta
	p3Y2 := -yDelta
	// determine which is closest to the p3 from constraint
	newP31 := &SketchPoint{BaseElement{Vector: Vector{p3X, p3Y1}, elementType: Point, id: p3.GetID()}}
	newP32 := &SketchPoint{BaseElement{Vector: Vector{p3X, p3Y2}, elementType: Point, id: p3.GetID()}}
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
func PointFromPoints(c1 Constraint, c2 Constraint) (SketchElement, SolveState) {
	p1 := c1.element1
	p2 := c2.element1
	p3 := c1.element2
	p1Radius := c1.GetValue()
	p2Radius := c2.GetValue()

	switch {
	case c1.element1.Is(c2.element1):
		p3, p1, p2 = c1.element1, c1.element2, c2.element2
	case c1.element2.Is(c2.element1):
		p3, p1, p2 = c1.element2, c1.element1, c2.element2
	case c1.element1.Is(c2.element2):
		p3, p1, p2 = c1.element1, c1.element2, c2.element1
	case c1.element2.Is(c2.element2):
		break
	}

	return pointFromPoints(p1, p2, p3, p1Radius, p2Radius)
}

func pointFromPointLine(originalP1 SketchElement, originalL2 SketchElement, originalP3 SketchElement, pointDist float64, lineDist float64) (SketchElement, SolveState) {
	p1 := CopySketchElement(originalP1)
	l2 := CopySketchElement(originalL2)
	p3 := CopySketchElement(originalP3)
	distanceDifference := l2.DistanceTo(p1)

	if distanceDifference+pointDist < lineDist+pointDist {
		return nil, NonConvergent
	}

	if distanceDifference > lineDist+pointDist {
		return nil, NonConvergent
	}

	if distanceDifference == lineDist {
		return &SketchPoint{BaseElement{
			Vector:      Vector{p1.GetX(), p1.GetY() - pointDist},
			elementType: Point,
			id:          p3.GetID(),
		}}, Solved
	}

	// TODO the rest...
	// rotate l2 to X axis
	angle := l2.AngleTo(Vector{1, 0})
	l2.Rotate(-angle)
	p1.Rotate(-angle)
	p3.Rotate(-angle)
	// translate l2 to X axis
	yTranslate := l2.(*SketchLine).GetOriginDistance() - lineDist
	l2.Translate(0, yTranslate)
	// move p1 to Y axis
	xTranslate := -p1.GetX()
	p1.Translate(xTranslate, yTranslate)
	p3.Translate(xTranslate, yTranslate)

	// Find points where circle at p1 with radius pointDist intersects with x axis
	xPos := math.Sqrt((pointDist * pointDist) - (p1.GetY() * p1.GetY()))
	newP31 := NewSketchPoint(p3.GetID(), xPos, 0)
	newP32 := NewSketchPoint(p3.GetID(), -xPos, 0)
	actualP3 := newP31
	if newP32.SquareDistanceTo(p3) < newP31.SquareDistanceTo(p3) {
		actualP3 = newP32
	}
	actualP3.Translate(-xTranslate, -yTranslate)
	actualP3.Rotate(angle)

	return actualP3, Solved
}

// PointFromPointLine construct a point from a point and a line. c2 must contain the line.
func PointFromPointLine(c1 Constraint, c2 Constraint) (SketchElement, SolveState) {
	p1 := c1.element1
	l2 := c2.element1
	p3 := c1.element2
	pointDist := c1.GetValue()
	lineDist := c2.GetValue()

	switch {
	case c1.element1.Is(c2.element1):
		p3 = c1.element1
		p1 = c1.element2
		l2 = c2.element2
	case c1.element2.Is(c2.element1):
		p3 = c1.element2
		p1 = c1.element1
		l2 = c2.element2
	case c1.element1.Is(c2.element2):
		p3 = c1.element1
		p1 = c1.element2
		l2 = c2.element1
	case c1.element2.Is(c2.element2):
		break
	}

	if p1.GetType() == Line && l2.GetType() == Point {
		p1, l2 = l2, p1
		pointDist, lineDist = lineDist, pointDist
	}

	return pointFromPointLine(p1, l2, p3, pointDist, lineDist)
}

func pointFromLineLine(originalL1 SketchElement, originalL2 SketchElement, originalP3 SketchElement, line1Dist float64, line2Dist float64) (SketchElement, SolveState) {
	l1 := CopySketchElement(originalL1)
	l2 := CopySketchElement(originalL2)
	p3 := CopySketchElement(originalP3)
	// If l1 and l2 are parallel, there is no solution
	line1, line2 := l1.(*SketchLine), l2.(*SketchLine)
	if line1.GetSlope() == line2.GetSlope() {
		return nil, NonConvergent
	}
	// Translate l1 line1Dist
	line1Translated := line1.TranslateDistance(line1Dist)
	// Translate l2 line2Dist
	line2Translated := line2.TranslateDistance(line2Dist)
	// Return intersection point
	intersection := line1Translated.Intersection(line2Translated)

	return NewSketchPoint(p3.GetID(), intersection.GetX(), intersection.GetY()), Solved
}

// PointFromLineLine construct a point from two lines. c2 must contain the point.
func PointFromLineLine(c1 Constraint, c2 Constraint) (SketchElement, SolveState) {
	l1 := c1.element1
	l2 := c2.element1
	p3 := c1.element2
	line1Dist := c1.GetValue()
	line2Dist := c2.GetValue()

	switch {
	case c1.element1.Is(c2.element1):
		p3 = c1.element1
		l1 = c1.element2
		l2 = c2.element2
	case c1.element2.Is(c2.element1):
		p3 = c1.element2
		l1 = c1.element1
		l2 = c2.element2
	case c1.element1.Is(c2.element2):
		p3 = c1.element1
		l1 = c1.element2
		l2 = c2.element1
	case c1.element2.Is(c2.element2):
		break
	}

	return pointFromLineLine(l1, l2, p3, line1Dist, line2Dist)
}
