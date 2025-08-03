package solver

import (
	"math"

	"github.com/marcuswu/dlineate/internal/accessors"
	"github.com/marcuswu/dlineate/internal/constraint"
	el "github.com/marcuswu/dlineate/internal/element"
	"github.com/marcuswu/dlineate/utils"
)

func SolveForLine(cluster int, ea accessors.ElementAccessor, c1 *constraint.Constraint, c2 *constraint.Constraint) SolveState {
	line, solveState := LineResult(cluster, ea, c1, c2)

	if line == nil {
		return solveState
	}

	c1e, _ := ea.GetElement(cluster, line.GetID())
	c1Line := c1e.AsLine()
	c1Line.SetA(line.GetA())
	c1Line.SetB(line.GetB())
	c1Line.SetC(line.GetC())

	c2e, _ := ea.GetElement(cluster, line.GetID())
	c2Line := c2e.AsLine()
	c2Line.SetA(line.GetA())
	c2Line.SetB(line.GetB())
	c2Line.SetC(line.GetC())

	return solveState
}

func LineResult(cluster int, ea accessors.ElementAccessor, c1 *constraint.Constraint, c2 *constraint.Constraint) (*el.SketchLine, SolveState) {
	/*
		There are three possibilities:
		 * three lines -- only possible during a merge and should already be solved
		 * two lines and a point: The two lines must have an angle constraint between them
		 * two points and a line
	*/
	_, numLines := typeCounts(c1, c2, ea)
	// 2 lines, 1 point -> LineFromPointLine
	var line *el.SketchLine = nil
	var solveState SolveState = NonConvergent

	// If all are lines, this is coming from a merge and it's successful if the shared line is already solved
	if numLines == 4 {
		lineId, ok := c1.Shared(c2)
		if !ok {
			return nil, NonConvergent
		}
		lineEl, ok := ea.GetElement(cluster, lineId)
		if !ok {
			return nil, NonConvergent
		}
		line = lineEl.AsLine()
		c1OtherId, ok := c1.Other(lineId)
		if !ok {
			return line, NonConvergent
		}
		c1Other, ok := ea.GetElement(cluster, c1OtherId)
		if !ok {
			return line, NonConvergent
		}
		c2OtherId, ok := c2.Other(lineId)
		if !ok {
			return line, NonConvergent
		}
		c2Other, ok := ea.GetElement(cluster, c2OtherId)
		if !ok {
			return line, NonConvergent
		}
		if !c1.IsMet(line, c1Other) || !c2.IsMet(line, c2Other) {
			return line, NonConvergent
		}
		return line, Solved
	}

	if numLines == 3 {
		line, solveState = LineFromPointLine(cluster, ea, c1, c2)
		if line != nil {
			utils.Logger.Trace().
				Str("result", solveState.String()).
				Str("line", line.String()).
				Msg("LineFromPointLine result")
		}
	}

	// 1 line, 2 points -> LineFromPoints
	if numLines == 2 {
		line, solveState = LineFromPoints(cluster, ea, c1, c2)
		if line != nil {
			utils.Logger.Trace().
				Str("result", solveState.String()).
				Str("line", line.String()).
				Msgf("LineFromPoints result")
		}
	}

	if solveState == Solved {
		c1.Solved = true
		c2.Solved = true
	}

	return line, solveState
}

// MoveLineToPoint solves a constraint between a line and a point where the line needs to move
func MoveLineToPoint(ea accessors.ElementAccessor, c *constraint.Constraint) SolveState {
	if c.Type != constraint.Distance {
		utils.Logger.Error().
			Uint("constraint", c.GetID()).
			Msg("MoveLineToPoint constraint was not Distance type")
		return NonConvergent
	}

	var point *el.SketchPoint
	var line *el.SketchLine
	e1, ok := ea.GetElement(-1, c.Element1)
	if !ok {
		return NonConvergent
	}
	e2, ok := ea.GetElement(-1, c.Element2)
	if !ok {
		return NonConvergent
	}
	var e1Type = e1.GetType()
	var e2Type = e2.GetType()
	if e1Type == e2Type {
		utils.Logger.Error().
			Uint("constraint", c.GetID()).
			Msg("MoveLineToPoint did not have the correct element types")
		return NonConvergent
	}
	if e1Type == el.Point && e2Type == el.Line {
		point = e1.(*el.SketchPoint)
		line = e2.(*el.SketchLine)
	}
	if e2Type == el.Point && e1Type == el.Line {
		point = e2.(*el.SketchPoint)
		line = e1.(*el.SketchLine)
	}

	// If two points, get distance between them, translate constraint value - distance between
	// If point and line, get distance between them, translate normal to line constraint value - distance between
	v := line.VectorTo(point)
	line.Translate(-v.X, -v.Y)
	translate1 := c.GetValue()
	translate2 := -c.GetValue()

	if math.Abs(translate1) < math.Abs(translate2) {
		line.TranslateDistance(translate1)
	} else {
		line.TranslateDistance(translate2)
	}

	c.Solved = true

	return Solved
}

func LineFromPoints(cluster int, ea accessors.ElementAccessor, c1 *constraint.Constraint, c2 *constraint.Constraint) (*el.SketchLine, SolveState) {
	l, ok := c1.Shared(c2)
	if !ok {
		return nil, NonConvergent
	}
	e, ok := ea.GetElement(cluster, l)
	if !ok {
		return nil, NonConvergent
	}
	line := e.AsLine()

	if line == nil {
		utils.Logger.Error().
			Uint("constraint 1", c1.GetID()).
			Uint("constraint 2", c2.GetID()).
			Msg("LineFromPoints could not find the line to work with.")
		return line, NonConvergent
	}

	p1e, _ := c1.Other(line.GetID())
	p2e, _ := c2.Other(line.GetID())
	e, ok = ea.GetElement(cluster, p1e)
	if !ok || e.GetType() != el.Point {
		return line, NonConvergent
	}
	p1 := e.AsPoint()
	e, ok = ea.GetElement(cluster, p2e)
	if !ok || e.GetType() != el.Point {
		return line, NonConvergent
	}
	p2 := e.AsPoint()
	if p1 == nil || p2 == nil {
		utils.Logger.Error().
			Uint("constraint 1", c1.GetID()).
			Uint("constraint 2", c2.GetID()).
			Msg("LineFromPoints could not find the points to work with.")
		return line, NonConvergent
	}
	p1Dist := c1.Value
	p2Dist := c2.Value

	// Special case where distances are both 0, calculate a line through the two points
	if p1Dist == 0 && p2Dist == 0 {
		la1 := p2.Y - p1.Y                  // y' - y
		lb1 := p1.X - p2.X                  // x - x'
		lc1 := (-la1 * p1.X) - (lb1 * p1.Y) // c = -ax - by from ax + by + c = 0
		la2 := p1.Y - p2.Y                  // y' - y
		lb2 := p2.X - p1.X                  // x - x'
		lc2 := (-la2 * p1.X) - (lb2 * p1.Y) // c = -ax - by from ax + by + c = 0
		lineV := &el.Vector{X: line.GetA(), Y: line.GetB()}
		angleTo1 := lineV.AngleTo(&el.Vector{X: la1, Y: lb1})
		angleTo2 := lineV.AngleTo(&el.Vector{X: la2, Y: lb2})
		line.SetA(la1)
		line.SetB(lb1)
		line.SetC(lc1)
		if math.Abs(angleTo2) < math.Abs(angleTo1) {
			line.SetA(la2)
			line.SetB(lb2)
			line.SetC(lc2)
		}
		return line, Solved
	}

	// Rotate line to horizontal (and rotate points the same)
	// Translate line p2Dist so it lies on p2
	// The line must be tangent to the two circles defined by the two points and their distances
	// TODO: fix this check -- this is not true for external tangents!
	externalOnly := false
	if p1.DistanceTo(p2) < p1Dist+p2Dist {
		externalOnly = true
	}

	// Math from https://en.wikipedia.org/wiki/Tangent_lines_to_circles#Analytic_geometry
	d := p1.DistanceTo(p2)
	R := (p2Dist - p1Dist) / d
	X := (p2.X - p1.X) / d
	Y := (p2.Y - p1.Y) / d

	calcTangent := func(X, Y, R, k float64, external bool) (float64, float64, float64) {
		rSquared := R * R

		a := (R * X) - (k * Y * math.Sqrt(1.0-rSquared))
		b := (R * Y) + (k * X * math.Sqrt(1.0-rSquared))
		c := p1Dist - ((a * p1.X) + (b * p1.Y))
		if !external {
			c = p2Dist - ((a * p2.X) + (b * p2.Y))
		}
		mag := math.Sqrt((a * a) + (b * b))
		a = a / mag
		b = b / mag
		c = c / mag
		return a, b, c
	}
	// Internal vs external tangents will be handled by positive or negative distance constraint values
	// Both the same sign will be external, opposing signs will be internal
	// There will be two options aside from internal or external -- plus or minus k
	// Use the one closest to the existing line angle (closest slope)
	tanA := make([]float64, 4)
	tanB := make([]float64, 4)
	tanC := make([]float64, 4)
	a, b, c := calcTangent(X, Y, R, 1, true)
	tanA[0], tanB[0], tanC[0] = a, b, c
	a, b, c = calcTangent(X, Y, R, -1, true)
	tanA[1], tanB[1], tanC[1] = a, b, c

	tanA[2], tanB[2], tanC[2] = tanA[0], tanB[0], tanC[0]
	tanA[3], tanB[3], tanC[3] = tanA[1], tanB[1], tanC[1]

	if !externalOnly {
		R = (p2Dist + p1Dist) / d
		a, b, c = calcTangent(X, Y, R, 1, false)
		tanA[2], tanB[2], tanC[2] = a, b, c
		a, b, c = calcTangent(X, Y, R, -1, false)
		tanA[3], tanB[3], tanC[3] = a, b, c
	}

	// Look for the closest combination of slope and origin distance
	minDifference := math.MaxFloat64
	chosenA, chosenB, chosenC := tanA[0], tanB[0], tanC[0]
	for i, _ := range tanA {
		originalSlope := line.GetB() / line.GetA()
		a, b, c = tanA[i], tanB[i], tanC[i]
		tangentSlope := b / a
		slopeDifference := math.Abs(tangentSlope - originalSlope)
		originDistanceDifference := math.Abs(c - line.GetC())
		averageDifference := (slopeDifference + originDistanceDifference) / 2
		if averageDifference < minDifference {
			minDifference = averageDifference
			chosenA, chosenB, chosenC = a, b, c
		}
	}
	line.SetA(chosenA)
	line.SetB(chosenB)
	line.SetC(chosenC)

	return line, Solved
}

func LineFromPointLine(cluster int, ea accessors.ElementAccessor, c1 *constraint.Constraint, c2 *constraint.Constraint) (*el.SketchLine, SolveState) {
	var targetLine *el.SketchLine = nil
	var point *el.SketchPoint = nil
	distC := c1
	angleC := c2
	if c1.Type == constraint.Angle {
		angleC = c1
		distC = c2
	}

	// The distance constraint with have the point and the shared line
	e, ok := ea.GetElement(cluster, distC.First())
	if !ok {
		return nil, NonConvergent
	}
	targetLine = e.AsLine()
	if targetLine == nil {
		point = e.AsPoint()
	}
	e, ok = ea.GetElement(cluster, distC.Second())
	if !ok {
		return nil, NonConvergent
	}
	if point == nil {
		point = e.AsPoint()
	} else {
		targetLine = e.AsLine()
	}

	// Solve angle
	newLine, state := SolveAngleConstraint(cluster, ea, angleC, targetLine.GetID())

	if state != Solved {
		return newLine, state
	}

	// Translate to distC.Value from the point
	v := newLine.VectorTo(point)
	newLine.Translate(-v.X, -v.Y)
	dist1 := -distC.Value
	dist2 := distC.Value
	line1 := newLine.TranslatedDistance(dist1)
	line2 := newLine.TranslatedDistance(dist2)

	line1Distance := targetLine.DistanceTo(line1)
	line2Distance := targetLine.DistanceTo(line2)

	if math.Abs(line1Distance) < math.Abs(line2Distance) {
		return line1, state
	}

	return line2, state
}

// SolveAngleConstraint solve an angle constraint between two lines
func SolveAngleConstraint(cluster int, ea accessors.ElementAccessor, c *constraint.Constraint, e uint) (*el.SketchLine, SolveState) {
	if c.Type != constraint.Angle {
		utils.Logger.Error().
			Uint("constraint", c.GetID()).
			Msgf("SolveAngleConstraint was not sent an angle constraint")
		return nil, NonConvergent
	}

	element, ok := ea.GetElement(cluster, c.Element1)
	if !ok {
		return nil, NonConvergent
	}
	l1 := element.(*el.SketchLine)
	element, ok = ea.GetElement(cluster, c.Element2)
	if !ok {
		return nil, NonConvergent
	}
	l2 := element.(*el.SketchLine)
	desired := c.Value
	if l1.GetID() == e {
		l1, l2 = l2, l1
	}

	angle1 := l2.AngleToLine(l1)
	rotate1 := angle1 + desired
	rotate2 := desired + angle1
	reverseRotate1 := angle1 - desired
	reverseRotate2 := desired - angle1

	lines := []*el.SketchLine{
		l2.Rotated(rotate1),
		l2.Rotated(rotate2),
		l2.Rotated(reverseRotate1),
		l2.Rotated(reverseRotate2),
	}

	var newLine *el.SketchLine = nil
	for _, line := range lines {
		if newLine == nil || math.Abs(line.AngleToLine(l2)) < math.Abs(newLine.AngleToLine(l2)) {
			newLine = line
		}
	}

	c.Solved = true
	return newLine, Solved
}
