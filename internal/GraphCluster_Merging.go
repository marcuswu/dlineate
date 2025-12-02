package core

import (
	"fmt"
	"math/big"

	"github.com/marcuswu/dlineate/internal/accessors"
	"github.com/marcuswu/dlineate/internal/constraint"
	el "github.com/marcuswu/dlineate/internal/element"
	"github.com/marcuswu/dlineate/internal/solver"
	"github.com/marcuswu/dlineate/internal/solver/graph"
	"github.com/marcuswu/dlineate/utils"
	"github.com/rs/zerolog"
)

func (c *GraphCluster) moveCluster(ea accessors.ElementAccessor, pivot el.SketchElement, from el.SketchElement, to el.SketchElement) {
	utils.Logger.Trace().
		Str("pivot", pivot.String()).
		Str("from", from.String()).
		Str("to", to.String()).
		Msg("Move Cluster Params")
	if pivot.GetType() == el.Line {
		var neg big.Float
		neg.SetPrec(utils.FloatPrecision).SetFloat64(-1)
		move := from.VectorTo(to)
		move.Scaled(&neg)
		c.TranslateCluster(ea, &move.X, &move.Y)
		return
	}

	// current, desired := pivot.VectorTo(from), pivot.VectorTo(to)
	current, desired := from.VectorTo(pivot), to.VectorTo(pivot)
	angle := desired.AngleTo(current)
	angle2 := current.AngleTo(desired)
	if from.GetType() == el.Line {
		angle = from.AsLine().AngleToLine(to.AsLine())
	}
	new1 := el.CopySketchElement(from)
	new1.Rotate(angle)
	if new1.IsEqual(to) {
		utils.Logger.Trace().
			Str("angle", angle.String()).
			Msg("Cluster Rotation")
		c.RotateCluster(ea, pivot.AsPoint(), angle)
		return
	}
	utils.Logger.Trace().
		Str("angle", angle2.String()).
		Msg("Cluster Rotation")
	c.RotateCluster(ea, pivot.AsPoint(), angle2)
}

func (g *GraphCluster) solveMerge(ea accessors.ElementAccessor, ca accessors.ConstraintAccessor, mergeData MergeData) solver.SolveState {
	if mergeData.cluster3 == nil {
		utils.Logger.Info().Msg("Beginning one cluster merge")
		return g.mergeOne(ea, ca, mergeData, true)
	}
	// Move constraints / elements from c1, c2 to g when we're done
	defer ea.MergeElements(g.GetID(), mergeData.clusterId2)
	defer ea.MergeElements(g.GetID(), mergeData.clusterId3)

	c1, c2 := mergeData.cluster2, mergeData.cluster3

	utils.Logger.Info().Msg("")
	utils.Logger.Info().Msg("Beginning cluster merge")
	solve := g.IsSolved(ea, ca)
	utils.Logger.Info().Msgf("Checking g solved: %v", solve)
	solve = c1.IsSolved(ea, ca)
	utils.Logger.Info().Msgf("Checking c1 solved: %v", solve)
	solve = c2.IsSolved(ea, ca)
	utils.Logger.Info().Msgf("Checking c2 solved: %v", solve)
	utils.Logger.Info().Msgf("")
	utils.Logger.Debug().Msg("Pre-merge state:")
	utils.Logger.Debug().Msg("g:")
	g.logElements(ea, zerolog.DebugLevel)
	utils.Logger.Debug().Msg("c1:")
	c1.logElements(ea, zerolog.DebugLevel)
	utils.Logger.Debug().Msg("c2:")
	c2.logElements(ea, zerolog.DebugLevel)

	// TODO: Update these checks to allow for free constraints
	sharedSet := ea.SharedElements(g.GetID(), c1.GetID())
	if sharedSet.Count() != 1 {
		return solver.NonConvergent
	}
	gc1Shared := sharedSet.Contents()[0]

	sharedSet = ea.SharedElements(g.GetID(), c2.GetID())
	if sharedSet.Count() != 1 {
		return solver.NonConvergent
	}
	gc2Shared := sharedSet.Contents()[0]

	sharedSet = ea.SharedElements(c1.GetID(), c2.GetID())
	if sharedSet.Count() != 1 {
		return solver.NonConvergent
	}
	c1c2Shared := sharedSet.Contents()[0]

	utils.Logger.Trace().
		Str("elements", fmt.Sprintf("%d, %d, %d", gc1Shared, gc2Shared, c1c2Shared)).
		Msg("Solving for shared elements")

	// TODO: Update to allow for solving by a free constraint
	// Solve c1 to g and c2 to g
	state1, _ := g.solveOne(ea, g, gc1Shared)
	utils.Logger.Info().Msg("moved c1 to g")
	utils.Logger.Info().Msg("g:")
	g.logElements(ea, zerolog.InfoLevel)
	utils.Logger.Info().Msg("c1:")
	c1.logElements(ea, zerolog.InfoLevel)
	state2, _ := g.solveOne(ea, g, gc2Shared)
	utils.Logger.Info().Msg("moved c2 to g")
	utils.Logger.Info().Msg("g:")
	g.logElements(ea, zerolog.InfoLevel)
	utils.Logger.Info().Msg("c2:")
	c2.logElements(ea, zerolog.InfoLevel)
	if state1 == solver.NonConvergent || state2 == solver.NonConvergent {
		return solver.NonConvergent
	}
	ea.CopyToCluster(g.GetID(), c1.GetID(), gc1Shared)
	ea.CopyToCluster(g.GetID(), c2.GetID(), gc2Shared)

	constraintFor := func(ea accessors.ElementAccessor, other *GraphCluster, anchor uint, solveFor uint) *constraint.Constraint {
		anchorElement, _ := ea.GetElement(other.GetID(), anchor)
		solveElement, _ := ea.GetElement(other.GetID(), solveFor)
		if anchorElement.GetType() == el.Line && solveElement.GetType() == el.Line {
			angle := anchorElement.AsLine().AngleToLine(solveElement.AsLine())
			angle.Neg(angle)
			utils.Logger.Trace().
				Uint("element 1", anchor).
				Uint("element 2", solveFor).
				Str("angle", angle.String()).
				Msg("Creating constraint")
			return constraint.NewConstraint(0, constraint.Angle, anchor, solveFor, angle, false)
		}
		dist := anchorElement.DistanceTo(solveElement)
		utils.Logger.Trace().
			Uint("element 1", anchor).
			Uint("element 2", solveFor).
			Str("distance", dist.String()).
			Msg("Creating constraint")
		return constraint.NewConstraint(0, constraint.Distance, anchor, solveFor, dist, false)
	}

	// Update to use free constraints if they exist
	c1Constraint := constraintFor(ea, c1, gc1Shared, c1c2Shared)
	c2Constraint := constraintFor(ea, c2, gc2Shared, c1c2Shared)
	c1Shared, _ := ea.GetElement(c1.GetID(), c1c2Shared)
	c2Shared, _ := ea.GetElement(c2.GetID(), c1c2Shared)

	newC1C2Shared, state := graph.ConstraintResult(g.GetID(), ea, c1Constraint, c2Constraint, c1Shared)
	if state == solver.Solved {
		utils.Logger.Trace().
			Str("shared element", newC1C2Shared.String()).
			Msg("Desired c1 c2 rotate solve")
	}

	if state != solver.Solved {
		utils.Logger.Error().Msg("Final element solve failed")
		return state
	}

	gc1SharedElement, _ := ea.GetElement(c1.GetID(), gc1Shared)
	utils.Logger.Trace().Msg("Pivoting c1")
	c1.moveCluster(ea, gc1SharedElement, c1Shared, newC1C2Shared)
	utils.Logger.Trace().
		Str("c1 shared element final", c1Shared.String()).
		Msgf("c1c2 shared moved")

	gc2SharedElement, _ := ea.GetElement(c2.GetID(), gc2Shared)
	utils.Logger.Trace().Msg("Pivoting c2")
	c2.moveCluster(ea, gc2SharedElement, c2Shared, newC1C2Shared)
	utils.Logger.Trace().
		Str("c2 shared element final", c2Shared.String()).
		Msgf("c1c2 shared moved")

	utils.Logger.Info().Msg("Completed cluster merge")
	utils.Logger.Info().Msg("")
	utils.Logger.Info().Msg("g:")
	g.logElements(ea, zerolog.InfoLevel)
	utils.Logger.Info().Msg("c1:")
	c1.logElements(ea, zerolog.InfoLevel)
	utils.Logger.Info().Msg("c2:")
	c2.logElements(ea, zerolog.InfoLevel)
	utils.Logger.Info().Msg("")

	if !g.SharedElementsEquivalent(ea, c1) {
		utils.Logger.Info().Msg("Returning Non-convergent due to element inequivalancy between g and c1 after merge")
		return solver.NonConvergent
	}
	if !g.SharedElementsEquivalent(ea, c2) {
		utils.Logger.Info().Msg("Returning Non-convergent due to element inequivalancy between g and c2 after merge")
		return solver.NonConvergent
	}
	if !c1.SharedElementsEquivalent(ea, c2) {
		utils.Logger.Info().Msg("Returning Non-convergent due to element inequivalancy between c1 and c2 merge")
		return solver.NonConvergent
	}

	return solver.Solved
}

func (g *GraphCluster) solveOne(ea accessors.ElementAccessor, other *GraphCluster, shared uint) (solver.SolveState, el.Type) {
	e1, _ := ea.GetElement(g.GetID(), shared)
	e2, _ := ea.GetElement(other.GetID(), shared)
	eType := e1.GetType()
	utils.Logger.Trace().
		Uint("element", shared).
		Str("element 1", e1.String()).
		Str("element 2", e2.String()).
		Str("type", eType.String()).
		Msg("Solving for element")

	// Solve element
	// if element is a line, rotate it into place first
	var translation *el.Vector
	if eType == el.Line {
		other.logElements(ea, zerolog.TraceLevel)
		utils.Logger.Trace().Msg("")
		angle := e1.AsLine().AngleToLine(e2.AsLine())
		other.RotateCluster(ea, e1.AsLine().PointNearestOrigin(), angle)
		translation = e1.VectorTo(e2)
	} else {
		translation = e1.VectorTo(e2)
	}
	utils.Logger.Trace().
		Str("X", translation.X.String()).
		Str("y", translation.Y.String()).
		Msg("Cluster translation")

	// translate element into place
	var zero big.Float
	zero.SetPrec(utils.FloatPrecision).SetFloat64(0)
	other.TranslateCluster(ea, &translation.X, &translation.Y)
	if utils.StandardBigFloatCompare(e2.DistanceTo(e1), &zero) != 0 {
		return solver.NonConvergent, eType
	}
	return solver.Solved, eType
}

// MergeOne resolves merging one solved child clusters to this one
// Can be merged via:
//   - Two shared elements
//   - One shared element and a constraint
//   - One distance constraint and one angle constraint
//
// In the future I may need to support:
//   - Two distance constraints to one element
//   - Three distance constraints to diff elements
func (g *GraphCluster) mergeOne(ea accessors.ElementAccessor, ca accessors.ConstraintAccessor, mergeData MergeData, mergeConstraints bool) solver.SolveState {
	if mergeConstraints {
		defer ea.MergeElements(g.GetID(), mergeData.clusterId2)
	}
	sharedElements := mergeData.elements

	switch len(sharedElements) {
	case 0:
		return g.mergeByConstraints(ea, ca, mergeData)
	case 1:
		return g.mergeOneShared(ea, ca, mergeData)
	default:
		return g.mergeTwoShared(ea, mergeData)
	}
}

func (g *GraphCluster) mergeTwoShared(ea accessors.ElementAccessor, mergeData MergeData) solver.SolveState {
	if len(mergeData.elements) != 2 {
		return solver.NonConvergent
	}

	sharedElements := mergeData.elements
	other := mergeData.cluster2

	// Solve two shared elements
	utils.Logger.Debug().Msg("Initial configuration:")
	utils.Logger.Debug().
		Str("elements", fmt.Sprintf("%v", sharedElements)).
		Msg("Shared elements")
	g.logElements(ea, zerolog.DebugLevel)
	utils.Logger.Debug().Msg("")
	other.logElements(ea, zerolog.DebugLevel)
	utils.Logger.Debug().Msg("")

	first := sharedElements[0]
	firstEl, firstOk := ea.GetElement(g.GetID(), first)
	second := sharedElements[1]
	secondEl, secondOk := ea.GetElement(g.GetID(), second)

	if !firstOk || !secondOk {
		return solver.NonConvergent
	}

	if firstEl.GetType() == el.Line {
		first, second = second, first
		firstEl, secondEl = secondEl, firstEl
	}
	if firstEl.GetType() == el.Line {
		first, second = second, first
		firstEl, secondEl = secondEl, firstEl
	}

	// If both elements are lines, nonconvergent (I think)
	// TODO: line up one. If the other doesn't align, then nonconvergent
	if firstEl.GetType() == el.Line {
		utils.Logger.Error().Msg("In a merge one and both shared elements are line type")
		return solver.NonConvergent
	}

	p1 := firstEl
	p2, _ := ea.GetElement(other.GetID(), first)

	// If there's a line, first rotate the lines into the same angle, then match first element
	if p2.GetType() == el.Line {
		l, _ := ea.GetElement(g.GetID(), second)
		ol, _ := ea.GetElement(other.GetID(), second)
		angle := ol.AsLine().AngleToLine(l.AsLine())
		other.RotateCluster(ea, p1.AsPoint(), angle)
		utils.Logger.Trace().Msg("Rotated to make line the same angle")
	}

	// Match up the first point
	utils.Logger.Trace().Msg("matching up the first point")
	direction := p1.VectorTo(p2)
	other.TranslateCluster(ea, &direction.X, &direction.Y)

	// If both are points, rotate other to match the element in g
	// Use a angle between the two points in both clusters to determine the angle to rotate
	if secondEl.GetType() == el.Point {
		utils.Logger.Trace().Msg("both elements were points, rotating to match the points together")
		otherFirst, _ := ea.GetElement(other.GetID(), first)
		otherSecond, _ := ea.GetElement(other.GetID(), second)
		v1 := secondEl.VectorTo(firstEl)
		v2 := otherSecond.VectorTo(otherFirst)
		angle := v1.AngleTo(v2)
		other.RotateCluster(ea, p1.AsPoint(), angle)
	}

	return solver.Solved
}

// Merge two clusters by one shared element and one constraint
func (g *GraphCluster) mergeOneShared(ea accessors.ElementAccessor, ca accessors.ConstraintAccessor, mergeData MergeData) solver.SolveState {
	if len(mergeData.constraints) < 1 && len(mergeData.elements) < 1 {
		return solver.NonConvergent
	}

	var e1, e2 el.SketchElement
	c, ok := ca.GetConstraint(mergeData.constraints[0])
	if !ok {
		return solver.NonConvergent
	}
	if g.elements.Contains(c.Element1) {
		e1, ok = ea.GetElement(g.id, c.Element1)
		if !ok {
			return solver.NonConvergent
		}
		e2, ok = ea.GetElement(mergeData.clusterId2, c.Element2)
		if !ok {
			return solver.NonConvergent
		}
	} else {
		e2, ok = ea.GetElement(g.id, c.Element1)
		if !ok {
			return solver.NonConvergent
		}
		e1, ok = ea.GetElement(mergeData.clusterId2, c.Element2)
		if !ok {
			return solver.NonConvergent
		}
	}
	shared := mergeData.elements[0]
	sharedE, ok := ea.GetElement(g.id, shared)
	if !ok || sharedE.AsPoint() == nil {
		return solver.NonConvergent
	}
	state, _ := g.solveOne(ea, mergeData.cluster2, shared)
	if !ok {
		state = solver.NonConvergent
	}

	var rotation *big.Float
	if c.Type == constraint.Angle {
		// rotate to match angle
		desiredAngle := c.Value
		currentAngle := e1.AsLine().AngleToLine(e2.AsLine())
		rotation.Sub(&desiredAngle, currentAngle)
	} else {
		r1 := sharedE.DistanceTo(e2)
		r2 := &c.Value
		p3, newState := graph.GetPointFromPoints(e1, sharedE, e2, r1, r2)
		state = newState
		if state == solver.NonConvergent {
			return state
		}
		desired := sharedE.VectorTo(p3)
		current := sharedE.VectorTo(e2)
		rotation = desired.AngleTo(current)
	}
	mergeData.cluster2.RotateCluster(ea, sharedE.AsPoint(), rotation)

	return state
}

// Merge two clusters by one distance constraint and one angle constraint
func (g *GraphCluster) mergeByConstraints(ea accessors.ElementAccessor, ca accessors.ConstraintAccessor, mergeData MergeData) solver.SolveState {
	if len(mergeData.constraints) < 2 {
		return solver.NonConvergent
	}

	cId1 := mergeData.constraints[0]
	cId2 := mergeData.constraints[1]
	constraint1, ok := ca.GetConstraint(cId1)
	if !ok {
		return solver.NonConvergent
	}
	constraint2, ok := ca.GetConstraint(cId2)
	if !ok {
		return solver.NonConvergent
	}
	angleC := constraint1
	distC := constraint2
	if angleC.Type != constraint.Angle {
		angleC, distC = distC, angleC
	}
	if angleC.Type != constraint.Angle || distC.Type != constraint.Distance {
		return solver.NonConvergent
	}

	// Rotate cluster 2 to match the angle constraint
	desiredAngle := angleC.Value
	e1Cluster := mergeData.cluster1
	if mergeData.cluster2.HasElement(angleC.Element1) {
		e1Cluster = mergeData.cluster2
	}
	e1, ok := ea.GetElement(e1Cluster.id, angleC.Element1)
	if !ok {
		return solver.NonConvergent
	}
	e2Cluster := mergeData.cluster1
	if mergeData.cluster2.HasElement(angleC.Element1) {
		e2Cluster = mergeData.cluster2
	}
	e2, ok := ea.GetElement(e2Cluster.id, angleC.Element2)
	if !ok {
		return solver.NonConvergent
	}
	if e1Cluster.id == mergeData.clusterId2 {
		e1, e2 = e2, e1
	}
	currentAngle := e1.AsLine().AngleToLine(e2.AsLine())
	var rotation big.Float
	rotation.Sub(currentAngle, &desiredAngle)
	e2Cluster.RotateCluster(ea, e2Cluster.firstPoint(ea), &rotation)

	// Translate to fit the distance constraint
	// desiredDist := distC.Value
	if mergeData.cluster2.HasElement(distC.Element1) {
		e1Cluster = mergeData.cluster2
	}
	e1, ok = ea.GetElement(e1Cluster.id, distC.Element1)
	if !ok {
		return solver.NonConvergent
	}
	e2Cluster = mergeData.cluster1
	if mergeData.cluster2.HasElement(distC.Element1) {
		e2Cluster = mergeData.cluster2
	}
	e2, ok = ea.GetElement(e2Cluster.id, angleC.Element2)
	if !ok {
		return solver.NonConvergent
	}
	if e1Cluster.id == mergeData.clusterId2 {
		e1, e2 = e2, e1
	}
	var xTranslate, yTranslate big.Float
	xTranslate.Copy(&e2.AsPoint().X)
	xTranslate.Sub(&e2.AsPoint().X, &e1.AsPoint().X)
	yTranslate.Copy(&e2.AsPoint().Y)
	yTranslate.Sub(&e2.AsPoint().Y, &e1.AsPoint().Y)
	e2Cluster.TranslateCluster(ea, &xTranslate, &yTranslate)

	return solver.Solved
}
