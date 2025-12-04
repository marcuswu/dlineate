package solver

import (
	"math/big"

	"github.com/marcuswu/dlineate/internal/accessors"
	"github.com/marcuswu/dlineate/internal/constraint"
	el "github.com/marcuswu/dlineate/internal/element"
	"github.com/marcuswu/dlineate/utils"
)

func SolveConstraint(cluster int, ea accessors.ElementAccessor, c *constraint.Constraint) SolveState {
	if c.Type == constraint.Distance {
		return SolveDistanceConstraint(cluster, ea, c)
	}
	solveElement := c.Element2
	if ea.IsFixed(c.Element2) {
		solveElement = c.Element1
	}

	if ea.IsFixed(solveElement) {
		return Solved
	}

	newLine, state := SolveAngleConstraint(cluster, ea, c, solveElement)

	e, _ := ea.GetElement(cluster, solveElement)
	cl := e.AsLine()
	cl.SetA(newLine.GetA())
	cl.SetB(newLine.GetB())
	cl.SetC(newLine.GetC())

	return state
}

func SolveConstraints(cluster int, ea accessors.ElementAccessor, c1 *constraint.Constraint, c2 *constraint.Constraint, solveFor el.SketchElement) SolveState {
	uniqueElements := utils.NewSet()
	uniqueElements.Add(c1.Element1)
	uniqueElements.Add(c1.Element2)
	uniqueElements.Add(c2.Element1)
	uniqueElements.Add(c2.Element2)
	if uniqueElements.Count() != 3 {
		return OverConstrained
	}

	if solveFor.GetType() == el.Point {
		return SolveForPoint(cluster, ea, c1, c2)
	}

	return SolveForLine(cluster, ea, c1, c2)
}

func ConstraintResult(cluster int, ea accessors.ElementAccessor, c1 *constraint.Constraint, c2 *constraint.Constraint, solveFor el.SketchElement) (el.SketchElement, SolveState) {
	if solveFor.GetType() == el.Point {
		return PointResult(cluster, ea, c1, c2)
	}

	return LineResult(cluster, ea, c1, c2)
}

// SolveConstraints solve two constraints and return the solution state
func SolveForPoint(cluster int, ea accessors.ElementAccessor, c1 *constraint.Constraint, c2 *constraint.Constraint) SolveState {
	newP3, state := PointResult(cluster, ea, c1, c2)

	if newP3 == nil {
		return state
	}

	c1e, _ := ea.GetElement(cluster, newP3.GetID())
	c1p := c1e.AsPoint()
	c1p.X = newP3.X
	c1p.Y = newP3.Y

	c2e, _ := ea.GetElement(cluster, newP3.GetID())
	c2p := c2e.AsPoint()
	c2p.X = newP3.X
	c2p.Y = newP3.Y

	return state
}

// PointResult returns the result of solving two constraints sharing one point
func PointResult(cluster int, ea accessors.ElementAccessor, c1 *constraint.Constraint, c2 *constraint.Constraint) (*el.SketchPoint, SolveState) {
	numPoints, _ := TypeCounts(c1, c2, ea)
	// 4 points -> PointFromPoints
	var point *el.SketchPoint = nil
	var solveState SolveState = NonConvergent
	if numPoints == 4 {
		point, solveState = PointFromPoints(cluster, ea, c1, c2)
	}

	// 3 points, 1 line -> PointFromPointLine
	if numPoints == 3 {
		point, solveState = PointFromPointLine(cluster, ea, c1, c2)
	}
	// 2 points, 2 lines -> PointFromLineLine
	if numPoints == 2 {
		point, solveState = PointFromLineLine(cluster, ea, c1, c2)
	}

	if solveState == Solved {
		c1.Solved = true
		c2.Solved = true
	}

	return point, solveState
}

// SolveDistanceConstraint solves a distance constraint and returns the solution state
func SolveDistanceConstraint(cluster int, ea accessors.ElementAccessor, c *constraint.Constraint) SolveState {
	if c.Type != constraint.Distance {
		utils.Logger.Error().
			Uint("constraint", c.GetID()).
			Msgf("SolveDistanceConstraint: was not sent a distance constraint")
		return NonConvergent
	}

	var solveElement el.SketchElement
	var other el.SketchElement
	e1, ok := ea.GetElement(cluster, c.Element1)
	if !ok {
		utils.Logger.Error().
			Uint("constraint", c.GetID()).
			Uint("element 1", c.Element1).
			Uint("element 2", c.Element2).
			Msgf("SolveDistanceConstraint: Element 1 not found")
		return NonConvergent
	}
	e2, ok := ea.GetElement(cluster, c.Element2)
	if !ok {
		utils.Logger.Error().
			Uint("constraint", c.GetID()).
			Uint("element 1", c.Element1).
			Uint("element 2", c.Element2).
			Msgf("SolveDistanceConstraint: Element 2 not found")
		return NonConvergent
	}

	if e1.IsFixed() && e2.IsFixed() {
		return Solved
	}

	solveElement, other = e1, e2
	if solveElement.IsFixed() {
		solveElement, other = other, solveElement
	}

	var zero, temp, x, y big.Float
	zero.SetPrec(utils.FloatPrecision).SetFloat64(0)
	direction := other.VectorTo(solveElement)
	dist := direction.Magnitude()
	utils.Logger.Trace().
		Str("distance", dist.String()).
		Msg("Calculated current distance")

	if dist.Cmp(&zero) == 0 && c.GetValue().Cmp(&zero) > 0 {
		utils.Logger.Error().Msg("SolveDistanceConstraint: points are coincident, but they shouldn't be. Infinite solutions.")
		return NonConvergent
	}

	if utils.StandardBigFloatCompare(dist, &zero) == 0 && c.GetValue().Cmp(&zero) == 0 {
		c.Solved = true
		return Solved
	}

	translation, ok := direction.UnitVector()
	if !ok {
		return NonConvergent
	}
	translation.Scaled(temp.Sub(c.GetValue(), dist))

	x.Neg(translation.GetX())
	y.Neg(translation.GetY())
	solveElement.Translate(&x, &y)
	c.Solved = true

	return Solved
}

// GetPointFromPoints calculates where a 3rd point exists in relation to two others with
// distance constraints from the first two
func GetPointFromPoints(p1 el.SketchElement, originalP2 el.SketchElement, originalP3 el.SketchElement, p1Radius *big.Float, p2Radius *big.Float) (*el.SketchPoint, SolveState) {
	// Don't mutate the originals
	p2 := el.CopySketchElement(originalP2)
	p3 := el.CopySketchElement(originalP3)
	pointDistance := p1.DistanceTo(p2)
	var constraintDist, x, y, temp big.Float
	constraintDist.Add(p1Radius, p2Radius)

	if utils.StandardBigFloatCompare(pointDistance, &constraintDist) > 0 {
		utils.Logger.Error().
			Uint("point 1", p1.GetID()).
			Uint("point 2", p2.GetID()).
			Msg("GetPointFromPoints no solution because the points are too far apart")
		return nil, NonConvergent
	}

	if utils.StandardBigFloatCompare(pointDistance, &constraintDist) == 0 {
		translate := p1.VectorTo(p2)
		translate.Scaled(temp.Quo(p1Radius, translate.Magnitude()))
		x.Sub(&p1.AsPoint().X, &translate.X)
		y.Sub(&p1.AsPoint().Y, &translate.Y)
		newP3 := el.NewSketchPoint(p3.GetID(), &x, &y)
		return newP3, Solved
	}

	// Solve for p3
	// translate to p1 (p2 and p3)
	p2.ReverseTranslateByElement(p1)
	p3.ReverseTranslateByElement(p1)
	// rotate p2 and p3 so p2 is on x axis
	x.SetPrec(utils.FloatPrecision).SetFloat64(1)
	y.SetPrec(utils.FloatPrecision).SetFloat64(0)
	angle := p2.AngleTo(&el.Vector{X: x, Y: y})
	p2.Rotate(angle)
	p3.Rotate(angle)
	// calculate possible p3s
	p2Dist := p2.(*el.SketchPoint).GetX()

	var xDelta, yDelta, p3X, p3Y1, p3Y2, p1rSq, temp1, temp2 big.Float
	// https://mathworld.wolfram.com/Circle-CircleIntersection.html
	// xDelta := ((-(p2Radius * p2Radius) + (p2Dist * p2Dist)) + (p1Radius * p1Radius)) / (2 * p2Dist)
	xDelta.Mul(p2Radius, p2Radius)
	xDelta.Neg(&xDelta)
	temp1.Mul(p2Dist, p2Dist)
	xDelta.Add(&xDelta, &temp1)
	p1rSq.Mul(p1Radius, p1Radius)
	xDelta.Add(&xDelta, &p1rSq)
	temp2.SetPrec(utils.FloatPrecision).SetFloat64(2)
	temp1.Mul(p2Dist, &temp2)
	xDelta.Quo(&xDelta, &temp1)

	// yDelta := math.Sqrt((p1Radius * p1Radius) - (xDelta * xDelta))
	temp1.Mul(&xDelta, &xDelta)
	yDelta.Sub(&p1rSq, &temp1)
	yDelta.Sqrt(&yDelta)
	p3X.Set(&xDelta)
	p3Y1.Set(&yDelta)
	p3Y2.Neg(&yDelta)
	// determine which is closest to the p3 from constraint
	newP31 := el.NewSketchPoint(p3.GetID(), &p3X, &p3Y1)
	newP32 := el.NewSketchPoint(p3.GetID(), &p3X, &p3Y2)
	actualP3 := newP31
	if newP32.SquareDistanceTo(p3).Cmp(newP31.SquareDistanceTo(p3)) < 0 {
		actualP3 = newP32
	}
	// unrotate actualP3
	temp1.Neg(angle)
	actualP3.Rotate(&temp1)
	// untranslate actualP3
	actualP3.TranslateByElement(p1)

	// return actualP3
	return actualP3, Solved
}

// PointFromPoints calculates a new p3 representing p3 moved to satisfy
// distance constraints from p1 and p2
func PointFromPoints(cluster int, ea accessors.ElementAccessor, c1 *constraint.Constraint, c2 *constraint.Constraint) (*el.SketchPoint, SolveState) {
	c1e1, _ := ea.GetElement(cluster, c1.Element1)
	c1e2, _ := ea.GetElement(cluster, c1.Element2)
	c2e1, _ := ea.GetElement(cluster, c2.Element1)
	c2e2, _ := ea.GetElement(cluster, c2.Element2)
	p1 := c1e1
	p2 := c2e1
	p3 := c1e2
	p1Radius := c1.GetValue()
	p2Radius := c2.GetValue()

	// p3 should always be the shared point
	switch {
	case c1.Element1 == c2.Element1:
		p3, p1, p2 = c1e1, c1e2, c2e2
	case c1.Element2 == c2.Element1:
		p3, p1, p2 = c1e2, c1e1, c2e2
	case c1.Element1 == c2.Element2:
		p3, p1, p2 = c1e1, c1e2, c2e1
	case c1.Element2 == c2.Element2:
		break
	}

	return GetPointFromPoints(p1, p2, p3, p1Radius, p2Radius)
}

func pointFromPointLine(originalP1 el.SketchElement, originalL2 el.SketchElement, originalP3 el.SketchElement, pointDist *big.Float, lineDist *big.Float) (*el.SketchPoint, SolveState) {
	p1 := el.CopySketchElement(originalP1).(*el.SketchPoint)
	l2 := el.CopySketchElement(originalL2).(*el.SketchLine)
	p3 := el.CopySketchElement(originalP3).(*el.SketchPoint)

	// 1. Rotate l2 to be parallel with the x axis. Repeat rotation with p1 and p3
	// 2. Find whether + or - lineDist places l2 closer to p3 and translate l2 towards p3
	// 3. Translate l2 to x axis, repeating with p1 and p3. Combine with the translation from step 2 for later reversal
	// 4. Translate p1 to the y axis, repeating with p3
	// 5. Find point newP3 on altered l2 and pointDist from p1
	// 6. Translate newP3 reverse x translate from 4 and reverse y translate from 3

	// 1. rotate l2 to X axis, repeating with p1 and p3
	// Rotation of the line will also normalize it making l2.C the distance to x axis
	var x, y, xTranslate, yTranslate, temp1, temp2 big.Float
	x.SetPrec(utils.FloatPrecision).SetFloat64(1)
	y.SetPrec(utils.FloatPrecision).SetFloat64(0)
	angle := l2.AngleTo(&el.Vector{X: x, Y: y})
	l2.Rotate(angle)
	p1.Rotate(angle)
	p3.Rotate(angle)

	// 2. Determine whether to use + or - lineDist
	x.SetPrec(utils.FloatPrecision).SetFloat64(0)
	y.Copy(lineDist)
	l2TransPos := l2.Translated(&x, &y)
	y.Neg(lineDist)
	l2TransNeg := l2.Translated(&x, &y)
	l2 = l2TransPos
	if l2TransNeg.DistanceTo(p3).Cmp(l2TransPos.DistanceTo(p3)) < 0 {
		l2 = l2TransNeg
	}

	// 3. Translate l2 to X axis
	yTranslate.Set(l2.GetC())
	x.SetPrec(utils.FloatPrecision).SetFloat64(0)
	l2.Translate(&x, &yTranslate)

	// 4. Translate p1 to Y axis
	xTranslate.Neg(p1.GetX())
	p1.Translate(&xTranslate, &yTranslate)
	p3.Translate(&xTranslate, &yTranslate)

	temp1.Abs(p1.GetY())
	if utils.StandardBigFloatCompare(pointDist, &temp1) < 0 {
		utils.Logger.Error().
			Str("point distance", pointDist.String()).
			Str("p1.y", temp1.String()).
			Msg("pointFromPointLine: Nonconvergent")
		return nil, NonConvergent
	}

	// 5. Find points where circle at p1 with radius pointDist intersects with x axis
	// xPos := math.Sqrt(math.Abs((pointDist * pointDist) - (p1.GetY() * p1.GetY())))
	var xPos, negXPos big.Float
	temp1.Mul(pointDist, pointDist)
	temp2.Mul(p1.GetY(), p1.GetY())
	xPos.Sub(&temp1, &temp2)
	xPos.Abs(&xPos)
	xPos.Sqrt(&xPos)
	negXPos.Neg(&xPos)
	y.SetPrec(utils.FloatPrecision).SetFloat64(0)

	newP31 := el.NewSketchPoint(p3.GetID(), &xPos, &y)
	newP32 := el.NewSketchPoint(p3.GetID(), &negXPos, &y)
	actualP3 := newP31
	if newP32.SquareDistanceTo(p3).Cmp(newP31.SquareDistanceTo(p3)) < 0 {
		actualP3 = newP32
	}

	// 6. Reverse translate new P3
	xTranslate.Neg(&xTranslate)
	yTranslate.Neg(&yTranslate)
	angle.Neg(angle)
	actualP3.Translate(&xTranslate, &yTranslate)
	actualP3.Rotate(angle)

	utils.Logger.Debug().
		Str("p3", actualP3.String()).
		Msg("pointFromPointLine: Final")

	return actualP3, Solved
}

// PointFromPointLine construct a point from a point and a line. c2 must contain the line.
func PointFromPointLine(cluster int, ea accessors.ElementAccessor, c1 *constraint.Constraint, c2 *constraint.Constraint) (*el.SketchPoint, SolveState) {
	c1e1, _ := ea.GetElement(cluster, c1.Element1)
	c1e2, _ := ea.GetElement(cluster, c1.Element2)
	c2e1, _ := ea.GetElement(cluster, c2.Element1)
	c2e2, _ := ea.GetElement(cluster, c2.Element2)
	p1 := c1e1
	l2 := c2e1
	p3 := c1e2
	pointDist := c1.GetValue()
	lineDist := c2.GetValue()

	switch {
	case c1.Element1 == c2.Element1:
		p3 = c1e1
		p1 = c1e2
		l2 = c2e2
	case c1.Element2 == c2.Element1:
		p3 = c1e2
		p1 = c1e1
		l2 = c2e2
	case c1.Element1 == c2.Element2:
		p3 = c1e1
		p1 = c1e2
		l2 = c2e1
	case c1.Element2 == c2.Element2:
		break
	}

	if p1.GetType() == el.Line && l2.GetType() == el.Point {
		p1, l2 = l2, p1
		pointDist, lineDist = lineDist, pointDist
	}

	return pointFromPointLine(p1, l2, p3, pointDist, lineDist)
}

func pointFromLineLine(l1 *el.SketchLine, l2 *el.SketchLine, p3 *el.SketchPoint, line1Dist *big.Float, line2Dist *big.Float) (*el.SketchPoint, SolveState) {
	sameSlope := utils.StandardBigFloatCompare(l1.GetA(), l2.GetA()) == 0 && utils.StandardBigFloatCompare(l1.GetB(), l2.GetB()) == 0
	// If l1 and l2 are parallel, and the distance between the lines isn't line1Dist + line2Dist, we can't solve
	distanceBetween := l1.DistanceTo(l2)
	var combinedDistances big.Float
	combinedDistances.Add(line1Dist, line2Dist)
	if sameSlope &&
		utils.StandardBigFloatCompare(&combinedDistances, distanceBetween) != 0 {
		utils.Logger.Error().
			Uint("line 1", l1.GetID()).
			Uint("line 2", l2.GetID()).
			Msg("pointFromLineLine no solution to find a point because the lines are parallel")
		return nil, NonConvergent
	}

	// If l1 & l2 are parallel and it's solvable, there are infinite solutions
	// Choose the one closest to the current point location
	if sameSlope {
		var scale, x, y big.Float
		translate := l1.VectorTo(p3)
		scale.Sub(p3.DistanceTo(l1), line1Dist)
		scale.Quo(&scale, translate.Magnitude())
		translate.Scaled(&scale)
		x.Add(p3.GetX(), &translate.X)
		y.Add(p3.GetY(), &translate.Y)
		return el.NewSketchPoint(p3.GetID(), &x, &y), Solved
	}
	// Translate l1 line1Dist
	var neg big.Float
	neg.Neg(line1Dist)
	line1TranslatePos := l1.TranslatedDistance(line1Dist)
	line1TranslateNeg := l1.TranslatedDistance(&neg)
	// Translate l2 line2Dist
	neg.Neg(line2Dist)
	line2TranslatedPos := l2.TranslatedDistance(line2Dist)
	line2TranslatedNeg := l2.TranslatedDistance(&neg)

	// If line1 and line2 are the same line,
	intersect1 := el.SketchPointFromVector(p3.GetID(), line1TranslatePos.Intersection(line2TranslatedPos))
	intersect2 := el.SketchPointFromVector(p3.GetID(), line1TranslatePos.Intersection(line2TranslatedNeg))
	intersect3 := el.SketchPointFromVector(p3.GetID(), line1TranslateNeg.Intersection(line2TranslatedPos))
	intersect4 := el.SketchPointFromVector(p3.GetID(), line1TranslateNeg.Intersection(line2TranslatedNeg))

	// Return closest intersection point
	closest := intersect1
	dist := p3.DistanceTo(intersect1)
	if next := p3.DistanceTo(intersect2); next.Cmp(dist) < 0 {
		dist = next
		closest = intersect2
	}
	if next := p3.DistanceTo(intersect3); next.Cmp(dist) < 0 {
		dist = next
		closest = intersect3
	}
	if next := p3.DistanceTo(intersect4); next.Cmp(dist) < 0 {
		closest = intersect4
	}

	return closest, Solved
}

// PointFromLineLine construct a point from two lines. c2 must contain the point.
func PointFromLineLine(cluster int, ea accessors.ElementAccessor, c1 *constraint.Constraint, c2 *constraint.Constraint) (*el.SketchPoint, SolveState) {
	c1e1, _ := ea.GetElement(cluster, c1.Element1)
	c1e2, _ := ea.GetElement(cluster, c1.Element2)
	c2e1, _ := ea.GetElement(cluster, c2.Element1)
	c2e2, _ := ea.GetElement(cluster, c2.Element2)
	l1 := c1e1
	l2 := c2e1
	p3 := c1e2
	line1Dist := c1.GetValue()
	line2Dist := c2.GetValue()

	switch {
	case c1.Element1 == c2.Element1:
		p3 = c1e1
		l1 = c1e2
		l2 = c2e2
	case c1.Element2 == c2.Element1:
		p3 = c1e2
		l1 = c1e1
		l2 = c2e2
	case c1.Element1 == c2.Element2:
		p3 = c1e1
		l1 = c1e2
		l2 = c2e1
	case c1.Element2 == c2.Element2:
		break
	}

	return pointFromLineLine(l1.AsLine(), l2.AsLine(), p3.AsPoint(), line1Dist, line2Dist)
}
