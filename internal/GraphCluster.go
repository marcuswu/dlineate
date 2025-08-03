package core

import (
	"fmt"

	"github.com/marcuswu/dlineate/internal/accessors"
	"github.com/marcuswu/dlineate/internal/constraint"
	el "github.com/marcuswu/dlineate/internal/element"
	"github.com/marcuswu/dlineate/internal/solver"
	"github.com/marcuswu/dlineate/utils"
	"github.com/rs/zerolog"
)

type GraphCluster struct {
	id          int
	constraints []uint
	elements    *utils.Set
	solved      *utils.Set
}

func NewGraphCluster(id int) *GraphCluster {
	g := new(GraphCluster)
	g.id = id
	g.constraints = make([]uint, 0, 10)
	g.elements = utils.NewSet()
	g.solved = utils.NewSet()
	return g
}

func (g *GraphCluster) GetID() int {
	return g.id
}

func (g *GraphCluster) HasElement(eId uint) bool {
	return g.elements.Contains(eId)
}

func (g *GraphCluster) HasConstraint(cId uint) bool {
	for _, c := range g.constraints {
		if cId == c {
			return true
		}
	}
	return false
}

func (g *GraphCluster) AddElement(e uint) {
	if g.HasElement(e) {
		return
	}
	utils.Logger.Debug().
		Int("cluster", g.id).
		Uint("element id", e).
		Msg("Cluster adding element")
	g.elements.Add(e)
}

// AddConstraint adds a constraint to the cluster
func (g *GraphCluster) AddConstraint(c *constraint.Constraint) {
	if g.HasConstraint(c.GetID()) {
		return
	}

	g.constraints = append(g.constraints, c.GetID())
	g.AddElement(c.Element1)
	g.AddElement(c.Element2)
}

// SharedElements returns the shared elements between this and another cluster
func (g *GraphCluster) SharedElements(gc *GraphCluster) *utils.Set {
	var shared *utils.Set = utils.NewSet()

	for _, elementID := range g.elements.Contents() {
		if gc.HasElement(elementID) {
			shared.Add(elementID)
		}
	}

	return shared
}

// Translate translates all elements in the cluster by an x and y value
func (g *GraphCluster) TranslateCluster(acc accessors.ElementAccessor, xDist float64, yDist float64) {
	for _, e := range g.elements.Contents() {
		element, ok := acc.GetElement(g.GetID(), e)
		if !ok {
			continue
		}
		element.Translate(xDist, yDist)
	}
}

// Rotate rotates all elements in the cluster around a point by an angle
func (g *GraphCluster) RotateCluster(acc accessors.ElementAccessor, origin *el.SketchPoint, angle float64) {
	v := el.Vector{X: origin.GetX(), Y: origin.GetY()}
	for _, e := range g.elements.Contents() {
		element, ok := acc.GetElement(g.GetID(), e)
		if !ok {
			continue
		}
		element.Translate(-v.X, -v.Y)
		element.Rotate(angle)
		element.Translate(v.X, v.Y)
	}
}

// func (g *GraphCluster) solvedConstraintsFor(ca ConstraintAccessor, eID uint) []*constraint.Constraint {
// 	var solvedC = make([]*constraint.Constraint, 0)
// 	for _, cId := range g.solved.Contents() {
// 		c, ok := ca.GetConstraint(cId)
// 		if !ok {
// 			continue
// 		}
// 		if c.Element1.GetID() == eID || c.Element2.GetID() == eID {
// 			solvedC = append(solvedC, c)
// 		}
// 	}
// 	return solvedC
// }

// func (g *GraphCluster) unsolvedConstraintsFor(ca ConstraintAccessor, eID uint) constraint.ConstraintList {
// 	var unsolved = make([]*constraint.Constraint, 0)
// 	uc := utils.NewSet()
// 	uc.AddList(g.constraints)
// 	uc = uc.Difference(g.solved)
// 	for _, cId := range uc.Contents() {
// 		c, ok := ca.GetConstraint(cId)
// 		if !ok {
// 			continue
// 		}
// 		if c.Element1.GetID() == eID || c.Element2.GetID() == eID {
// 			unsolved = append(unsolved, c)
// 		}
// 	}
// 	return unsolved
// }

func (g *GraphCluster) logElements(ea accessors.ElementAccessor, level zerolog.Level) {
	for _, e := range g.elements.Contents() {
		element, ok := ea.GetElement(g.GetID(), e)
		if !ok {
			utils.Logger.WithLevel(level).Msgf("Could not log element with id %d", e)
			continue
		}
		g.logElement(element, level)
	}
}

func (g *GraphCluster) logElement(e el.SketchElement, level zerolog.Level) {
	utils.Logger.WithLevel(level).Msg(e.String())
}

func (c *GraphCluster) ToGraphViz(ea accessors.ElementAccessor, ca accessors.ConstraintAccessor) string {
	edges := ""
	elements := ""
	for _, cId := range c.constraints {
		constraint, ok := ca.GetConstraint(cId)
		if !ok {
			utils.Logger.WithLevel(zerolog.ErrorLevel).Msgf("Could not create graphviz node for constraint with id %d", cId)
			continue
		}
		edges = edges + constraint.ToGraphViz(c.id)
		e1, _ := ea.GetElement(c.id, constraint.Element1)
		if e1 != nil {
			elements = elements + e1.ToGraphViz(c.id)
		}
		e2, _ := ea.GetElement(c.id, constraint.Element2)
		if e2 != nil {
			elements = elements + e2.ToGraphViz(c.id)
		}
	}
	return fmt.Sprintf(`subgraph cluster_%d {
		label = "Cluster %d"
		%s
		%s
	}`, c.id, c.id, edges, elements)
}

// Solve will solve the constraints in the cluster, returns solution state
//  1. Look for point w/ 2 constraints to solved elements -- fall back to point w/ fewest unsolved constraints
//  2. Solve the element by those 2 constraints
//  3. If there are unsolved elements, go to step 1
func (g *GraphCluster) Solve(ea accessors.ElementAccessor, ca accessors.ConstraintAccessor) solver.SolveState {
	// constraints are sorted by solve order
	constraints := make([]uint, 0, len(g.constraints))
	for _, c := range g.constraints {
		if g.solved.Contains(c) {
			continue
		}
		constraints = append(constraints, c)
	}
	c := make(constraint.ConstraintList, 2)

	g.logElements(ea, zerolog.InfoLevel)

	state := solver.Solved

	cId := constraints[0]
	constraints = constraints[1:]

	c1, ok := ca.GetConstraint(cId)
	if !ok {
		utils.Logger.Error().
			Uint("constraint", cId).
			Msg("Could not find constraint")
		return solver.NonConvergent
	}
	utils.Logger.Info().
		Uint("constraint", c1.GetID()).
		Uint("element 1", c1.Element1).
		Uint("element 2", c1.Element2).
		Msg("Solving first constraint")
	isFixed := ea.IsFixed(c1.Element1) && ea.IsFixed(c1.Element2)
	if !isFixed {
		state = solver.SolveConstraint(g.id, ea, c1)
	}
	utils.Logger.Trace().
		Str("state", state.String()).
		Msg("State")
	g.solved.Add(c1.GetID())

	/*
		1. Look for point w/ 2 constraints to solved elements -- fall back to point w/ fewest unsolved constraints
		2. Solve the element by those 2 constraints
		3. If there are unsolved elements, go to step 1

		An element is considered solved when it has at least two solved constraints.
		A constraint needs a solved flag or a structure to track solved state
		Need to be able to get constraints for an element
		Need to be able to filter constraint list by solved / unsolved (get by state?)
		Need to be able to quickly determine if an element is solved

		solved = Set of constraint
		map[elementID][constraint]
		isElementSolved(elementID)
	*/

	// Pick next two constraints and solve. If only 1 in constraintList, solve just the one
	for len(constraints) > 0 {
		// Step 1
		utils.Logger.Debug().Msg("Local Solve Step 1")
		utils.Logger.Debug().
			Str("constraints", fmt.Sprintf("%v", constraints)).
			Msg("Solve Order")

		if len(constraints) < 2 {
			utils.Logger.Error().
				Msg("Incorrect solve graph (odd number of constraints after solving first)")
			return solver.NonConvergent
		}

		c1, ok := ca.GetConstraint(constraints[0])
		if !ok {
			utils.Logger.Error().
				Uint("constraint", constraints[0]).
				Msg("Could not find constraint")
			return solver.NonConvergent
		}
		c2, ok := ca.GetConstraint(constraints[1])
		if !ok {
			utils.Logger.Error().
				Uint("constraint", constraints[1]).
				Msg("Could not find constraint")
			return solver.NonConvergent
		}
		c[0] = c1
		c[1] = c2
		constraints = constraints[2:]
		// c := g.unsolvedConstraintsFor(ca, eId)

		// if len(g.solvedConstraintsFor(ca, eId)) >= 2 {
		// 	utils.Logger.Trace().
		// 		Uint("element", eId).
		// 		Msg("Element already solved. Continuing.")
		// 	continue
		// }
		eId := c1.Element1
		if c2.HasElementID(c1.Element2) {
			eId = c1.Element2
		}
		if !c2.HasElementID(eId) {
			utils.Logger.Error().
				Uint("constraint 1", c[0].GetID()).
				Uint("constraint 2", c[1].GetID()).
				Msg("Could not find common element in constraints by solve order")
			return solver.NonConvergent
		}

		utils.Logger.Debug().Msg("")
		utils.Logger.Debug().
			Uint("element", eId).
			Msg("Solving for element")
		utils.Logger.Trace().
			Array("constraints", c).
			Msg("Solving for constraints")
		element, ok := ea.GetElement(g.GetID(), eId)
		if !ok {
			utils.Logger.Error().
				Uint("element ", eId).
				Msg("Could not find element")
			state = solver.NonConvergent
			break
		}

		// Step 2
		utils.Logger.Debug().Msg("Local Solve Step 2")
		utils.Logger.Debug().
			Uint("constraint 1", c[0].GetID()).
			Uint("constraint 2", c[1].GetID()).
			Msg("Solving constraints")
		s := solver.Solved
		if !element.IsFixed() {
			s = solver.SolveConstraints(g.id, ea, c[0], c[1], element)
		}
		if state == solver.Solved {
			utils.Logger.Trace().
				Str("state", s.String()).
				Msg("solve state changed")
			utils.Logger.Debug().
				Str("element", element.String()).
				Msg("solved element")
			state = s
			utils.Logger.Trace().
				Str("state", state.String()).
				Msg("State")
		}
		g.solved.Add(c[0].GetID())
		g.solved.Add(c[1].GetID())

		utils.Logger.Info().
			Str("solve ratio", fmt.Sprintf("%d / %d", g.solved.Count(), len(g.constraints))).
			Msg("Local Solve Step 3 (check for completion)")
	}

	utils.Logger.Info().
		Str("state", state.String()).
		Msg("finished")
	g.logElements(ea, zerolog.InfoLevel)
	return state
}

// MergeOne resolves merging one solved child clusters to this one
func (g *GraphCluster) mergeOne(ea accessors.ElementAccessor, other *GraphCluster, mergeConstraints bool) solver.SolveState {
	if mergeConstraints {
		defer ea.MergeElements(g.GetID(), other.GetID())
	}
	sharedElements := g.SharedElements(other).Contents()

	if g.id == 0 && other.id == 1 && len(sharedElements) > 2 {
		sharedElements = []uint{0, 1}
	}

	if len(sharedElements) != 2 {
		return solver.NonConvergent
	}

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
	other.TranslateCluster(ea, direction.X, direction.Y)

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

func (g *GraphCluster) solveMerge(ea accessors.ElementAccessor, ca accessors.ConstraintAccessor, c1 *GraphCluster, c2 *GraphCluster) solver.SolveState {
	if c2 == nil {
		utils.Logger.Info().Msg("Beginning one cluster merge")
		return g.mergeOne(ea, c1, true)
	}
	// Move constraints / elements from c1, c2 to g when we're done
	defer ea.MergeElements(g.GetID(), c1.GetID())
	defer ea.MergeElements(g.GetID(), c2.GetID())

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

	solveOne := func(ea accessors.ElementAccessor, root *GraphCluster, other *GraphCluster, shared uint) (solver.SolveState, el.Type) {
		e1, _ := ea.GetElement(root.GetID(), shared)
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
			Float64("X", translation.X).
			Float64("y", translation.Y).
			Msg("Cluster translation")

		// translate element into place
		other.TranslateCluster(ea, translation.X, translation.Y)
		if utils.StandardFloatCompare(e2.DistanceTo(e1), 0) != 0 {
			return solver.NonConvergent, eType
		}
		return solver.Solved, eType
	}

	// Solve c1 to g and c2 to g
	state1, _ := solveOne(ea, g, c1, gc1Shared)
	utils.Logger.Info().Msg("moved c1 to g")
	utils.Logger.Info().Msg("g:")
	g.logElements(ea, zerolog.InfoLevel)
	utils.Logger.Info().Msg("c1:")
	c1.logElements(ea, zerolog.InfoLevel)
	state2, _ := solveOne(ea, g, c2, gc2Shared)
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
			utils.Logger.Trace().
				Uint("element 1", anchor).
				Uint("element 2", solveFor).
				Float64("angle", -angle).
				Msg("Creating constraint")
			return constraint.NewConstraint(0, constraint.Angle, anchor, solveFor, -angle, false)
		}
		dist := anchorElement.DistanceTo(solveElement)
		utils.Logger.Trace().
			Uint("element 1", anchor).
			Uint("element 2", solveFor).
			Float64("distance", dist).
			Msg("Creating constraint")
		return constraint.NewConstraint(0, constraint.Distance, anchor, solveFor, dist, false)
	}

	c1Constraint := constraintFor(ea, c1, gc1Shared, c1c2Shared)
	c2Constraint := constraintFor(ea, c2, gc2Shared, c1c2Shared)
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

	moveCluster := func(c *GraphCluster, pivot el.SketchElement, from el.SketchElement, to el.SketchElement) {
		utils.Logger.Trace().
			Str("pivot", pivot.String()).
			Str("from", from.String()).
			Str("to", to.String()).
			Msg("Move Cluster Params")
		if pivot.GetType() == el.Line {
			move := from.VectorTo(to)
			c.TranslateCluster(ea, -move.X, -move.Y)
			return
		}

		// current, desired := pivot.VectorTo(from), pivot.VectorTo(to)
		current, desired := from.VectorTo(pivot), to.VectorTo(pivot)
		angle := desired.AngleTo(current)
		angle2 := current.AngleTo(desired)
		utils.Logger.Trace().
			Float64("angle", angle).
			Float64("angle 2", angle2).
			Msg("Cluster Rotation")
		c.RotateCluster(ea, pivot.AsPoint(), angle2)
	}

	gc1SharedElement, _ := ea.GetElement(c1.GetID(), gc1Shared)
	utils.Logger.Trace().Msg("Pivoting c1")
	moveCluster(c1, gc1SharedElement, c1Shared, newC1C2Shared)
	utils.Logger.Trace().
		Str("c1 shared element final", c1Shared.String()).
		Msgf("c1c2 shared moved")

	gc2SharedElement, _ := ea.GetElement(c2.GetID(), gc2Shared)
	utils.Logger.Trace().Msg("Pivoting c2")
	moveCluster(c2, gc2SharedElement, c2Shared, newC1C2Shared)
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

// SolveMerge resolves merging two solved child clusters to this one
/* TODO: Rewrite this. I originally wrote this when I couldn't solve for a line and had to
solve lines separately and then solve for a point. Now I can solve for a line.

1. Find elements in g shared with c1 and c2
2. Solve c1 and c2 shared elements (moving them to g)
3. Find element shared between c1 and c2 -- this is what we're solving for
4. Construct two constraints from g to c1 and g to c2 based on c1 and c2's shared element
5. Solve the constraint and rotate c1 and c2 to match
*/
/*func (g *GraphCluster) solveMerge(ea accessors.ElementAccessor, ca accessors.ConstraintAccessor, c1 *GraphCluster, c2 *GraphCluster) solver.SolveState {
	if c2 == nil {
		utils.Logger.Info().Msg("Beginning one cluster merge")
		return g.mergeOne(ea, c1, true)
	}
	// Move constraints / elements from c1, c2 to g when we're done
	defer ea.MergeElements(g.GetID(), c1.GetID())
	defer ea.MergeElements(g.GetID(), c2.GetID())
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
	clusters := []*GraphCluster{g, c1, c2}

	sharedSet := ea.SharedElements(g.GetID(), c1.GetID())
	sharedSet.AddSet(ea.SharedElements(g.GetID(), c2.GetID()))
	sharedSet.AddSet(ea.SharedElements(c1.GetID(), c2.GetID()))
	sharedElements := sharedSet.Contents()
	utils.Logger.Trace().
		Str("elements", fmt.Sprintf("%v", sharedElements)).
		Msg("Solving for shared elements")

	orderClustersFor := func(clusters []*GraphCluster, e uint) []*GraphCluster {
		matching := make([]*GraphCluster, 0)
		for _, c := range clusters {
			if !c.HasElement(e) {
				continue
			}
			matching = append(matching, c)
		}
		return matching
	}

	if len(sharedElements) != 3 {
		return solver.NonConvergent
	}

	numSharedLines := func(g *GraphCluster) int {
		lines := 0
		for _, se := range sharedElements {
			if e, ok := ea.GetElement(g.GetID(), se); ok && e.GetType() == el.Line {
				lines++
			}
		}
		return lines
	}

	// Find root cluster
	// Prefer keeping lines on the root cluster (solve lines first)
	rootCluster := g
	sharedLines := numSharedLines(g)
	c1SharedLines := numSharedLines(c1)
	c2SharedLines := numSharedLines(c2)
	if c1SharedLines > sharedLines {
		rootCluster = c1
		sharedLines = c1SharedLines
	}
	if c2SharedLines > sharedLines {
		rootCluster = c2
	}

	// Solve two of the elements
	final := sharedElements[0]
	finalIndex := 0
	for i, ec := range clusters {
		if ec == rootCluster {
			finalIndex = i
			break
		}
	}
	utils.Logger.Trace().
		Int("cluster", finalIndex).
		Msg("root cluster")

	for _, se := range sharedElements {
		parents := orderClustersFor(clusters, se)
		if len(parents) != 2 {
			utils.Logger.Error().
				Uint("element", se).
				Int("number of parents", len(parents)).
				Msg("Shared element should have exactly two parents. Returning Non-Convergent")
			return solver.NonConvergent
		}

		if parents[0] != rootCluster && parents[1] != rootCluster {
			final = se
			continue
		}
		e, _ := ea.GetElement(parents[0].GetID(), se)
		eType := e.GetType()
		utils.Logger.Trace().
			Uint("element", se).
			Str("type", eType.String()).
			Msg("Solving for element")

		// Solve element
		// if element is a line, rotate it into place first
		other := parents[0]
		if other == rootCluster {
			other = parents[1]
		}
		ec1, _ := ea.GetElement(other.GetID(), se)
		ec2, _ := ea.GetElement(rootCluster.GetID(), se)
		var translation *el.Vector
		if eType == el.Line {
			other.logElements(ea, zerolog.TraceLevel)
			utils.Logger.Trace().Msg("")
			angle := ec1.AsLine().AngleToLine(ec2.AsLine())
			other.RotateCluster(ea, ec1.AsLine().PointNearestOrigin(), angle)
			translation = ec1.VectorTo(ec2)
		} else {
			translation = ec2.VectorTo(ec1)
		}

		// translate element into place
		other.TranslateCluster(ea, translation.X, translation.Y)

		utils.Logger.Trace().
			Uint("element", se).
			Msg("Solved for element")
		utils.Logger.Trace().Msg("g:")
		g.logElements(ea, zerolog.TraceLevel)
		utils.Logger.Trace().Msg("c1:")
		c1.logElements(ea, zerolog.TraceLevel)
		utils.Logger.Trace().Msg("c2:")
		c2.logElements(ea, zerolog.TraceLevel)
		utils.Logger.Trace().Msg("")
	}

	var e = [2]uint{sharedElements[0], sharedElements[1]}
	if e[0] == final {
		e[0] = sharedElements[2]
	}
	if e[1] == final {
		e[1] = sharedElements[2]
	}
	utils.Logger.Trace().
		Uint("element 1", e[0]).
		Uint("element 2", e[1]).
		Uint("final unsolved element", final).
		Msg("Solved two elmements")
	g.logElements(ea, zerolog.TraceLevel)
	utils.Logger.Trace().Msg("")
	c1.logElements(ea, zerolog.TraceLevel)
	utils.Logger.Trace().Msg("")
	c2.logElements(ea, zerolog.TraceLevel)
	utils.Logger.Trace().Msg("")

	// Solve the third element in relation to the other two
	parents := orderClustersFor(clusters, final)
	final0, _ := ea.GetElement(parents[0].GetID(), final)
	final1, _ := ea.GetElement(parents[1].GetID(), final)
	finalE := [2]el.SketchElement{final0, final1}
	// p0Final := parents[0].elements[final]
	// p1Final := parents[1].elements[final]
	e2Type := finalE[0].GetType()
	utils.Logger.Trace().
		Str("type", e2Type.String()).
		Msgf("Final element type")
	if e2Type == el.Line {
		// We avoid e2 being a line, so if it is one, the other two are also lines.
		// This means e2 should already be placed correctly since the other two are.
		state := solver.Solved
		if !finalE[0].AsLine().IsEquivalent(finalE[1].AsLine()) {
			utils.Logger.Error().
				Str("line 1", finalE[0].String()).
				Str("line 2", finalE[1].String()).
				Msg("Lines are not equivalent: ")
			state = solver.NonConvergent
		}

		return state
	}

	// var constraint1, constraint2 *Constraint
	// var e1, e2 el.SketchElement
	others := [2]el.SketchElement{nil, nil}
	constraints := [2]*constraint.Constraint{nil, nil}
	for pi := range parents {
		for ei := range e {
			finalElement := finalE[pi]
			otherElement, ok := ea.GetElement(parents[pi].GetID(), e[ei])
			if !ok {
				continue
			}
			others[pi] = otherElement
			dist := finalElement.DistanceTo(otherElement)
			constraints[pi] = constraint.NewConstraint(0, constraint.Distance, finalElement.GetID(), otherElement.GetID(), dist, false)
			utils.Logger.Trace().
				Uint("element 1", finalElement.GetID()).
				Uint("element 2", otherElement.GetID()).
				Float64("distance", dist).
				Msg("Creating constraint")
		}
	}

	newE3, state := solver.ConstraintResult(g.id, ea, constraints[0], constraints[1], finalE[0])
	newP3 := newE3.AsPoint()

	if state != solver.Solved {
		utils.Logger.Error().Msg("Final element solve failed")
		return state
	}

	utils.Logger.Trace().
		Float64("X", newP3.X).
		Float64("Y", newP3.Y).
		Msg("Desired merge point c1 and c2")

	moveCluster := func(c *GraphCluster, pivot el.SketchElement, from *el.SketchPoint, to *el.SketchPoint) {
		if pivot.GetType() == el.Line {
			move := from.VectorTo(to)
			c.TranslateCluster(ea, -move.X, -move.Y)
			return
		}

		current, desired := pivot.VectorTo(from), pivot.VectorTo(to)
		angle := current.AngleTo(desired)
		c.RotateCluster(ea, pivot.AsPoint(), angle)
	}

	utils.Logger.Trace().
		Uint("pivot", others[0].GetID()).
		Str("from", finalE[0].String()).
		Str("to", newP3.String()).
		Msg("Pivoting c0")
	moveCluster(parents[0], others[0], finalE[0].AsPoint(), newP3)
	utils.Logger.Trace().
		Str("parent 0 final", finalE[0].String()).
		Msgf("parent 0 moved")
	utils.Logger.Trace().
		Uint("pivot", others[1].GetID()).
		Str("from", finalE[1].String()).
		Str("to", newP3.String()).
		Msg("Pivoting c1")
	moveCluster(parents[1], others[1], finalE[1].AsPoint(), newP3)
	utils.Logger.Trace().
		Str("parent 1 final", finalE[1].String()).
		Msgf("parent 1 moved")

	utils.Logger.Info().Msg("Completed cluster merge")
	utils.Logger.Info().Msg("")
	utils.Logger.Info().Msg("g:")
	g.logElements(ea, zerolog.InfoLevel)
	utils.Logger.Info().Msg("c1:")
	c1.logElements(ea, zerolog.InfoLevel)
	utils.Logger.Info().Msg("c2:")
	c2.logElements(ea, zerolog.InfoLevel)
	utils.Logger.Info().Msg("")

	if !g.SharedElementsEquivalent(ea, c1) || !g.SharedElementsEquivalent(ea, c2) || !c1.SharedElementsEquivalent(ea, c2) {
		utils.Logger.Info().Msg("Returning Non-convergent due to element inequivalancy after merge")
		return solver.NonConvergent
	}

	return solver.Solved
}*/

func (g *GraphCluster) SharedElementsEquivalent(ea accessors.ElementAccessor, o *GraphCluster) bool {
	compareElement := func(e1 el.SketchElement, e2 el.SketchElement) bool {
		if e1.GetType() != e2.GetType() {
			return false
		}

		if e1.AsLine() != nil {
			l1 := e1.AsLine()
			l2 := e2.AsLine()
			return utils.StandardFloatCompare(l1.GetA(), l2.GetA()) == 0 &&
				utils.StandardFloatCompare(l1.GetB(), l2.GetB()) == 0 &&
				utils.StandardFloatCompare(l1.GetC(), l2.GetC()) == 0
		}

		p1 := e1.AsPoint()
		p2 := e2.AsPoint()

		return utils.StandardFloatCompare(p1.X, p2.X) == 0 &&
			utils.StandardFloatCompare(p1.Y, p2.Y) == 0
	}
	equal := true
	shared := ea.SharedElements(g.GetID(), o.GetID())
	for _, e := range shared.Contents() {
		e1, _ := ea.GetElement(g.GetID(), e)
		e2, _ := ea.GetElement(o.GetID(), e)
		equal = equal && compareElement(e1, e2)
	}

	return equal
}

func (g *GraphCluster) IsSolved(ea accessors.ElementAccessor, ca accessors.ConstraintAccessor) bool {
	solved := true
	for _, cId := range g.constraints {
		c, _ := ca.GetConstraint(cId)
		if ca.IsMet(c.GetID(), g.id, ea) {
			continue
		}

		utils.Logger.Trace().
			Str("constraint", c.String()).
			Msg("Failed to meet")
		solved = false
	}

	return solved
}
