package solver

import (
	"errors"
	"math"
	"math/big"

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
	var x, y, translate1, translate2, t1, t2 big.Float
	v := line.VectorTo(point)
	x.Neg(&v.X)
	y.Neg(&v.Y)
	line.Translate(&x, &y)
	translate1.Set(c.GetValue())
	translate2.Neg(&translate1)

	t1.Abs(&translate1)
	t2.Abs(&translate2)
	if t1.Cmp(&t2) < 0 {
		line.TranslateDistance(&translate1)
	} else {
		line.TranslateDistance(&translate2)
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
	var zero big.Float
	zero.SetFloat64(0)
	if p1Dist.Cmp(&zero) == 0 && p2Dist.Cmp(&zero) == 0 {
		var la1, lb1, lc1, la2, lb2, lc2, t1, t2 big.Float
		la1.Sub(p2.GetY(), p1.GetY()) // y' - y
		lb1.Sub(p1.GetX(), p2.GetX()) // x - x'
		lc1.Neg(&la1)
		lc1.Mul(&lc1, p1.GetX())
		t1.Mul(&lb1, p1.GetY())
		lc1.Sub(&lc1, &t1)            // c = -ax - by from ax + by + c = 0
		la2.Sub(p1.GetY(), p2.GetY()) // y' - y
		lb2.Sub(p2.GetX(), p1.GetX()) // x - x'
		lc2.Neg(&la2)
		lc2.Mul(&lc2, p1.GetX())
		t1.Mul(&lb2, p1.GetY())
		lc2.Sub(&lc2, &t1) // c = -ax - by from ax + by + c = 0
		t1.Set(line.GetA())
		t2.Set(line.GetB())
		lineV := &el.Vector{X: t1, Y: t2}
		angleTo1 := lineV.AngleTo(&el.Vector{X: la1, Y: lb1})
		angleTo2 := lineV.AngleTo(&el.Vector{X: la2, Y: lb2})
		line.SetA(&la1)
		line.SetB(&lb1)
		line.SetC(&lc1)
		t1.Abs(angleTo2)
		t2.Abs(angleTo1)
		if float64(t1.Cmp(&t2)) < 0 {
			line.SetA(&la2)
			line.SetB(&lb2)
			line.SetC(&lc2)
		}
		return line, Solved
	}

	// Rotate line to horizontal (and rotate points the same)
	// Translate line p2Dist so it lies on p2
	// The line must be tangent to the two circles defined by the two points and their distances
	// TODO: fix this check -- this is not true for external tangents!
	externalOnly := false
	var combinedDistances big.Float
	combinedDistances.Add(&p1Dist, &p2Dist)
	if p1.DistanceTo(p2).Cmp(&combinedDistances) < 0 {
		externalOnly = true
	}

	// Math from https://en.wikipedia.org/wiki/Tangent_lines_to_circles#Analytic_geometry
	var R, X, Y big.Float
	// R := (p2Dist - p1Dist) / d
	d := p1.DistanceTo(p2)
	R.Sub(&p2Dist, &p1Dist)
	R.Quo(&R, d)
	// X := (p2.X - p1.X) / d
	X.Sub(p2.GetX(), p1.GetX())
	X.Quo(&X, d)
	// Y := (p2.Y - p1.Y) / d
	Y.Sub(p2.GetY(), p1.GetY())
	Y.Quo(&Y, d)

	calcTangent := func(X, Y, R, k *big.Float, external bool) (*big.Float, *big.Float, *big.Float, error) {
		var a, b, c, rSquared, one, mag, t1 big.Float
		one.SetFloat64(1)

		rSquared.Mul(R, R)
		rSquared.Sub(&one, &rSquared)
		if rSquared.Sign() < 0 {
			return nil, nil, nil, errors.New("cannot calculate tangent")
		}
		rSquared.Sqrt(&rSquared)

		// a := (R * X) - (k * Y * math.Sqrt(1.0-rSquared))
		a.Mul(R, X)
		t1.Mul(&rSquared, Y)
		t1.Mul(&t1, k)
		a.Sub(&a, &t1)
		// b := (R * Y) + (k * X * math.Sqrt(1.0-rSquared))
		b.Mul(R, Y)
		t1.Mul(&rSquared, X)
		t1.Mul(&t1, k)
		b.Add(&b, &t1)
		// c := p1Dist - ((a * p1.X) + (b * p1.Y))
		c.Mul(&a, p1.GetX())
		t1.Mul(&b, p1.GetY())
		c.Add(&c, &t1)
		c.Sub(&p1Dist, &c)
		if !external {
			// c = p2Dist - ((a * p2.X) + (b * p2.Y))
			c.Mul(&a, p2.GetX())
			t1.Mul(&b, p2.GetY())
			c.Add(&c, &t1)
			c.Sub(&p2Dist, &c)
		}
		// mag := math.Sqrt((a * a) + (b * b))
		mag.Mul(&a, &a)
		t1.Mul(&b, &b)
		mag.Add(&mag, &t1)
		mag.Sqrt(&mag)
		a.Quo(&a, &mag)
		b.Quo(&b, &mag)
		c.Quo(&c, &mag)
		return &a, &b, &c, nil
	}
	// Internal vs external tangents will be handled by positive or negative distance constraint values
	// Both the same sign will be external, opposing signs will be internal
	// There will be two options aside from internal or external -- plus or minus k
	// Use the one closest to the existing line angle (closest slope)
	tanA := make([]big.Float, 4)
	tanB := make([]big.Float, 4)
	tanC := make([]big.Float, 4)
	a, b, c, err := calcTangent(&X, &Y, &R, big.NewFloat(1), true)
	if err != nil {
		return nil, NonConvergent
	}
	tanA[0].Set(a)
	tanB[0].Set(b)
	tanC[0].Set(c)
	a, b, c, err = calcTangent(&X, &Y, &R, big.NewFloat(-1), true)
	if err != nil {
		return nil, NonConvergent
	}
	tanA[1].Set(a)
	tanB[1].Set(b)
	tanC[1].Set(c)

	tanA[2].Set(&tanA[0])
	tanB[2].Set(&tanB[0])
	tanC[2].Set(&tanC[0])
	tanA[3].Set(&tanA[1])
	tanB[3].Set(&tanB[1])
	tanC[3].Set(&tanC[1])

	if !externalOnly {
		// R = (p2Dist + p1Dist) / d
		R.Quo(&combinedDistances, d)
		a, b, c, err = calcTangent(&X, &Y, &R, big.NewFloat(1), false)
		if err != nil {
			return nil, NonConvergent
		}
		// tanA[2], tanB[2], tanC[2] = a, b, c
		tanA[2].Set(a)
		tanB[2].Set(b)
		tanC[2].Set(c)
		a, b, c, err = calcTangent(&X, &Y, &R, big.NewFloat(-1), false)
		if err != nil {
			return nil, NonConvergent
		}
		tanA[3].Set(a)
		tanB[3].Set(b)
		tanC[3].Set(c)
	}

	// Look for the closest combination of slope and origin distance
	var minDifference, originalSlope, tangentSlope, slopeDifference, originDistanceDifference, averageDifference big.Float
	minDifference.SetFloat64(math.MaxFloat64)
	chosenA, chosenB, chosenC := tanA[0], tanB[0], tanC[0]
	for i, _ := range tanA {
		originalSlope.Quo(line.GetB(), line.GetA())
		a, b, c = &tanA[i], &tanB[i], &tanC[i]
		tangentSlope.Quo(b, a)
		slopeDifference.Sub(&tangentSlope, &originalSlope)
		slopeDifference.Abs(&slopeDifference)
		originDistanceDifference.Sub(c, line.GetC())
		originDistanceDifference.Abs(&originDistanceDifference)
		// averageDifference := (slopeDifference + originDistanceDifference) / 2
		averageDifference.Add(&slopeDifference, &originDistanceDifference)
		averageDifference.Quo(&averageDifference, big.NewFloat(2))
		if averageDifference.Cmp(&minDifference) < 0 {
			minDifference.Set(&averageDifference)
			chosenA, chosenB, chosenC = *a, *b, *c
		}
	}
	line.SetA(&chosenA)
	line.SetB(&chosenB)
	line.SetC(&chosenC)

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
	var xT, yT, dist1, dist2, line1Distance, line2Distance big.Float
	xT.Neg(v.GetX())
	yT.Neg(v.GetY())
	newLine.Translate(&xT, &yT)
	dist1.Neg(&distC.Value)
	dist2.Set(&distC.Value)
	line1 := newLine.TranslatedDistance(&dist1)
	line2 := newLine.TranslatedDistance(&dist2)

	line1Distance.Abs(targetLine.DistanceTo(line1))
	line2Distance.Abs(targetLine.DistanceTo(line2))

	if line1Distance.Cmp(&line2Distance) < 0 {
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

	var rotate1, rotate2, reverseRotate1, reverseRotate2 big.Float
	angle1 := l2.AngleToLine(l1)
	rotate1.Add(angle1, &desired)
	rotate2.Add(&desired, angle1)
	reverseRotate1.Sub(angle1, &desired)
	reverseRotate2.Sub(&desired, angle1)

	lines := []*el.SketchLine{
		l2.Rotated(&rotate1),
		l2.Rotated(&rotate2),
		l2.Rotated(&reverseRotate1),
		l2.Rotated(&reverseRotate2),
	}

	var newLine *el.SketchLine = nil
	for _, line := range lines {
		if newLine == nil {
			newLine = line
			continue
		}
		var angle1, angle2 big.Float
		angle1.Abs(line.AngleToLine(l2))
		angle2.Abs(newLine.AngleToLine(l2))
		if angle1.Cmp(&angle2) < 0 {
			newLine = line
		}
	}

	c.Solved = true
	return newLine, Solved
}
