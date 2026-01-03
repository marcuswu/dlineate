package graph

import (
	"fmt"
	"math/big"

	"github.com/marcuswu/dlineate/internal/accessors"
	"github.com/marcuswu/dlineate/internal/constraint"
	el "github.com/marcuswu/dlineate/internal/element"
	"github.com/marcuswu/dlineate/internal/solver"
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

func (g *GraphCluster) mergeConstraint(ea accessors.ElementAccessor, anchor uint, solveFor uint) *constraint.Constraint {
	anchorElement, _ := ea.GetElement(g.GetID(), anchor)
	solveElement, _ := ea.GetElement(g.GetID(), solveFor)
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

func (g *GraphCluster) solveMerge(ea accessors.ElementAccessor, ca accessors.ConstraintAccessor, mergeData MergeData) solver.SolveState {
	if mergeData.cluster3 == nil {
		utils.Logger.Info().Msg("Beginning one cluster merge")
		return g.mergeOne(ea, ca, mergeData)
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

	// Solve c1 to g and c2 to g
	state1, _ := g.solveOne(ea, c1, gc1Shared)
	utils.Logger.Info().Msg("moved c1 to g")
	utils.Logger.Info().Msg("g:")
	g.logElements(ea, zerolog.InfoLevel)
	utils.Logger.Info().Msg("c1:")
	c1.logElements(ea, zerolog.InfoLevel)
	state2, _ := g.solveOne(ea, c2, gc2Shared)
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

	// Update to use free constraints if they exist
	c1Constraint := c1.mergeConstraint(ea, gc1Shared, c1c2Shared)
	c2Constraint := c2.mergeConstraint(ea, gc2Shared, c1c2Shared)
	c1Shared, _ := ea.GetElement(c1.GetID(), c1c2Shared)
	c2Shared, _ := ea.GetElement(c2.GetID(), c1c2Shared)

	newC1C2Shared, state := solver.ConstraintResult(g.GetID(), ea, c1Constraint, c2Constraint, c1Shared)
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
		Int("cluster 1", g.GetID()).
		Int("cluster 2", other.GetID()).
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
func (g *GraphCluster) mergeOne(ea accessors.ElementAccessor, ca accessors.ConstraintAccessor, mergeData MergeData) solver.SolveState {
	defer ea.MergeElements(g.GetID(), mergeData.clusterId2)
	sharedElements := mergeData.elements

	if len(mergeData.elements) != 2 {
		return solver.NonConvergent
	}

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
	firstOther, firstOtherOk := ea.GetElement(other.GetID(), first)
	second := sharedElements[1]
	secondEl, secondOk := ea.GetElement(g.GetID(), second)
	secondOther, secondOtherOk := ea.GetElement(other.GetID(), second)

	if !firstOk || !secondOk || !firstOtherOk || !secondOtherOk {
		return solver.NonConvergent
	}

	return other.translateToElements(ea, firstOther, firstEl, secondOther, secondEl)
}

func (g *GraphCluster) translateToElements(ea accessors.ElementAccessor, e1Local, e1Other, e2Local, e2Other el.SketchElement) solver.SolveState {
	if e1Local.GetType() == el.Line {
		e1Local, e2Local = e2Local, e1Local
		e1Other, e2Other = e2Other, e1Other
	}

	// If both elements are lines, nonconvergent (I think)
	// TODO: line up one. If the other doesn't align, then nonconvergent
	if e1Local.GetType() == el.Line {
		utils.Logger.Error().Msg("In a merge one and both shared elements are line type")
		return solver.NonConvergent
	}

	p1 := e1Local
	p2 := e1Other

	// If there's a line, first rotate the lines into the same angle, then match first element
	if e2Local.GetType() == el.Line {
		l := e2Local
		ol := e2Other
		// This was different from original implementation
		angle := ol.AsLine().AngleToLine(l.AsLine())
		g.RotateCluster(ea, p1.AsPoint(), angle)
		utils.Logger.Trace().Msg("Rotated to make line the same angle")
	}

	// Match up the first point
	utils.Logger.Trace().Msg("matching up the first point")
	// This was different from original implementation
	direction := p1.VectorTo(p2)
	g.TranslateCluster(ea, &direction.X, &direction.Y)

	// If both are points, rotate other to match the element in g
	// Use a angle between the two points in both clusters to determine the angle to rotate
	if e2Local.GetType() == el.Point {
		utils.Logger.Trace().Msg("both elements were points, rotating to match the points together")
		v1 := e2Local.VectorTo(e1Local)
		v2 := e2Other.VectorTo(e1Other)
		// This was different from original implementation
		angle := v1.AngleTo(v2)
		g.RotateCluster(ea, p1.AsPoint(), angle)
	}

	return solver.Solved
}
