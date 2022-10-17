package core

import (
	"fmt"

	"github.com/marcuswu/dlineation/internal/constraint"
	el "github.com/marcuswu/dlineation/internal/element"
	"github.com/marcuswu/dlineation/internal/solver"
	"github.com/marcuswu/dlineation/utils"
)

// Constraint a convenient alias for cosntraint.Constraint
type Constraint = constraint.Constraint

// GraphCluster A cluster within a Graph
type GraphCluster struct {
	id          int
	constraints []*Constraint
	// others      []*GraphCluster
	elements   map[uint]el.SketchElement
	eToC       map[uint][]*Constraint
	solveOrder []uint
	solved     *utils.Set
}

// NewGraphCluster constructs a new GraphCluster
func NewGraphCluster(id int) *GraphCluster {
	g := new(GraphCluster)
	g.id = id
	g.constraints = make([]*Constraint, 0)
	// g.others = make([]*GraphCluster, 0)
	g.elements = make(map[uint]el.SketchElement, 0)
	g.eToC = make(map[uint][]*Constraint)
	g.solved = utils.NewSet()
	g.solveOrder = make([]uint, 0)
	return g
}

func (g *GraphCluster) GetID() int {
	return g.id
}

func (g *GraphCluster) AddElement(e el.SketchElement) {
	if _, ok := g.elements[e.GetID()]; ok {
		return
	}
	fmt.Printf("Cluster adding element %d\n", e.GetID())
	g.elements[e.GetID()] = el.CopySketchElement(e)
	g.solveOrder = append(g.solveOrder, e.GetID())
}

// AddConstraint adds a constraint to the cluster
func (g *GraphCluster) AddConstraint(c *Constraint) {
	gc := constraint.CopyConstraint(c)
	g.constraints = append(g.constraints, gc)
	if _, ok := g.elements[gc.Element1.GetID()]; !ok {
		fmt.Printf("Warning: adding constraint %d to cluster before element %d\n", c.GetID(), gc.Element2.GetID())
		g.elements[gc.Element1.GetID()] = gc.Element1
	} else {
		gc.Element1 = g.elements[gc.Element1.GetID()]
	}

	if _, ok := g.elements[gc.Element2.GetID()]; !ok {
		fmt.Printf("Warning: adding constraint %d to cluster before element %d\n", c.GetID(), gc.Element2.GetID())
		g.elements[gc.Element2.GetID()] = gc.Element2
	} else {
		gc.Element2 = g.elements[gc.Element2.GetID()]
	}

	if _, ok := g.eToC[gc.Element1.GetID()]; !ok {
		g.eToC[gc.Element1.GetID()] = make([]*Constraint, 0)
	}
	if _, ok := g.eToC[gc.Element2.GetID()]; !ok {
		g.eToC[gc.Element2.GetID()] = make([]*Constraint, 0)
	}
	g.eToC[gc.Element1.GetID()] = append(g.eToC[gc.Element1.GetID()], gc)
	g.eToC[gc.Element2.GetID()] = append(g.eToC[gc.Element2.GetID()], gc)
}

// HasElementIDDirect returns whether this cluster directly contains an element ID
func (g *GraphCluster) HasElementIDImmediate(eID uint) bool {
	_, e := g.elements[eID]
	return e
}

// HasElementID returns whether this cluster contains an element ID
func (g *GraphCluster) HasElementID(eID uint) bool {
	if _, e := g.elements[eID]; e {
		return true
	}
	// for _, c := range g.others {
	// 	if c.HasElementID(eID) {
	// 		return true
	// 	}
	// }
	return false
}

// HasElement returns whether this cluster contains an element
func (g *GraphCluster) HasElement(e el.SketchElement) bool {
	if e == nil {
		return true
	}
	return g.HasElementID(e.GetID())
}

// GetElement returns the copy of an element represented in this cluster
func (g *GraphCluster) GetElement(eID uint) (el.SketchElement, bool) {
	if element, ok := g.elements[eID]; ok {
		return element, ok
	}
	// for _, c := range g.others {
	// 	if element, ok := c.elements[eID]; ok {
	// 		return element, ok
	// 	}
	// }
	return nil, false
}

// SharedElements returns the shared elements between this and another cluster
func (g *GraphCluster) SharedElements(gc *GraphCluster) *utils.Set {
	var shared *utils.Set = utils.NewSet()

	for elementID := range g.elements {
		if gc.HasElementID(elementID) {
			shared.Add(elementID)
		}
	}

	return shared
}

func (g *GraphCluster) immediateSharedElements(gc *GraphCluster) *utils.Set {
	var shared *utils.Set = utils.NewSet()

	for elementID := range g.elements {
		if gc.HasElementIDImmediate(elementID) {
			shared.Add(elementID)
		}
	}

	return shared
}

// Translate translates all elements in the cluster by an x and y value
func (g *GraphCluster) Translate(xDist float64, yDist float64) {
	for _, e := range g.elements {
		e.Translate(xDist, yDist)
	}
}

// Rotate rotates all elements in the cluster around a point by an angle
func (g *GraphCluster) Rotate(origin *el.SketchPoint, angle float64) {
	v := el.Vector{X: origin.GetX(), Y: origin.GetY()}
	for _, e := range g.elements {
		e.Translate(-v.X, -v.Y)
		e.Rotate(angle)
		e.Translate(v.X, v.Y)
	}
}

func (g *GraphCluster) rebuildMap() {
	g.elements = make(map[uint]el.SketchElement, 0)

	for _, c := range g.constraints {
		g.elements[c.Element1.GetID()] = c.Element1
		g.elements[c.Element2.GetID()] = c.Element2
	}
}

func (g *GraphCluster) solvedConstraintsFor(eID uint) []*Constraint {
	constraints := g.eToC[eID]
	var solvedC = make([]*Constraint, 0)
	for _, c := range constraints {
		if g.solved.Contains(c.GetID()) {
			solvedC = append(solvedC, c)
		}
	}
	return solvedC
}

func (g *GraphCluster) unsolvedConstraintsFor(eID uint) []*Constraint {
	var constraints = g.eToC[eID]
	var unsolved = make([]*Constraint, 0)
	for _, c := range constraints {
		if g.solved.Contains(c.GetID()) {
			continue
		}
		unsolved = append(unsolved, c)
	}

	return unsolved
}

// LocalSolve attempts to solve the constraints in the cluster, returns solution state
func (g *GraphCluster) localSolve() solver.SolveState {
	// solver changes element instances in constraints, so rebuild the element map
	defer g.rebuildMap()
	// Order constraints for element 0
	if len(g.solveOrder) < 2 {
		return solver.NonConvergent
	}

	state := solver.Solved

	e1 := g.solveOrder[0]
	e2 := g.solveOrder[1]
	g.solveOrder = g.solveOrder[2:]
	fmt.Println("Local Solve Step 0")
	fmt.Printf("Single constraint betw first two elements: %d, %d\n", e1, e2)
	for _, c := range g.constraints {
		if !c.HasElements(e1, e2) {
			continue
		}
		fmt.Println("Solving constraint", c.GetID())
		state = solver.SolveConstraint(c)
		fmt.Printf("State: %v\n", state)
		g.solved.Add(c.GetID())
		break
	}

	/*
		1. Look for point w/ 2 constraints to solved elements -- fall back to point w/ fewest unsolved constraints
		2. Solve the element by those 2 constraints
		3. Solve any lines via translation that are connected to solved point w/ 1 other solved constraint
		4. If there are unsolved elements, go to step 1

		An element is considered solved when it has at least two solved constraints.
		A constraint needs a solved flag or a structure to track solved state
		Need to be able to get constraints for an element
		Need to be able to filter constraint list by solved / unsolved (get by state?)
		Need to be able to quickly determine if an element is solved

		solved = Set of constraint
		map[elementID][constraint]
		isElementSolved(elementID)
	*/

	// Pick 2 from constraintList and solve. If only 1 in constraintList, solve just the one

	//for g.solved.Count() < len(g.constraints) {
	for len(g.solveOrder) > 0 {
		// Step 1
		fmt.Println("Local Solve Step 1")
		fmt.Printf("Solve Order: %v\n", g.solveOrder)
		//e, c := g.findElement()
		e := g.solveOrder[0]
		g.solveOrder = g.solveOrder[1:]
		c := g.unsolvedConstraintsFor(e)

		/*if g.elements[e].GetType() == el.Line {
			fmt.Printf("Skipping line element %d.\n", e)
			continue
		}*/

		if len(g.solvedConstraintsFor(e)) >= 2 {
			fmt.Printf("Element %d already solved. Continuing.\n", e)
			continue
		}

		fmt.Println("Solving for element", e)
		fmt.Printf("Element %d's eligible constraints: {", e)
		for _, constraint := range c {
			fmt.Print(constraint.GetID(), ", ")
		}
		fmt.Println("}")
		if len(c) < 2 {
			fmt.Println("Could not find a constraint to solve with", len(g.constraints)-g.solved.Count(), "constraints left to solve")
			state = solver.NonConvergent
			break
		}

		// Step 2
		fmt.Println("Local Solve Step 2")
		fmt.Println("Solving constraints", c[0].GetID(), c[1].GetID())
		if s := solver.SolveConstraints(c[0], c[1], g.elements[e]); state == solver.Solved {
			fmt.Println("solve state changed to", s)
			fmt.Println("solved element ", g.elements[e])
			element, _ := c[0].Element(e)
			fmt.Printf("solved element in constraint 0: %v\n", element)
			element, _ = c[1].Element(e)
			fmt.Printf("solved element in constraint 1: %v\n", element)
			state = s
			fmt.Printf("State: %v\n", state)
		}
		g.solved.Add(c[0].GetID())
		g.solved.Add(c[1].GetID())

		// Step 3
		// Look for solvable lines that have 1 solved constraint
		fmt.Println("Local Solve Step 3")
		/*		lines := g.solvableLines()
				for _, cs := range lines {
					if len(cs) < 1 {
						continue
					}
					if len(cs) > 1 {
						fmt.Println("Solving constraints", cs[0].GetID(), cs[1].GetID())
						if s := solver.MoveLineToPoints(cs); state == solver.Solved {
							fmt.Println("solve state changed to", s)
							state = s
						}
						g.solved.Add(c[0].GetID())
						g.solved.Add(c[1].GetID())
					} else {
						fmt.Println("Solving constraints", cs[0].GetID())
						if s := solver.MoveLineToPoint(cs[0]); state == solver.Solved {
							fmt.Println("solve state changed to", s)
							state = s
						}
						g.solved.Add(c[0].GetID())
					}
				}*/
		fmt.Printf("Local Solve Step 4 (check for completion) %d / %d solved\n", g.solved.Count(), len(g.constraints))
	}

	fmt.Println("finished with state", state)
	g.logElements()
	return state
}

func (g *GraphCluster) logElements() {
	for _, e := range g.elements {
		g.logElement(e)
	}
}

func (g *GraphCluster) logElement(e el.SketchElement) {
	point := e.AsPoint()
	line := e.AsLine()
	if point == nil {
		fmt.Printf("element %d: %fx + %fy + %f = 0\n", line.GetID(), line.GetA(), line.GetB(), line.GetC())
		return
	}
	fmt.Printf("element %d: (%f, %f)\n", point.GetID(), point.GetX(), point.GetY())
}

// MergeOne resolves merging one solved child clusters to this one
func (g *GraphCluster) mergeOne(other *GraphCluster, mergeConstraints bool) solver.SolveState {
	if mergeConstraints {
		defer g.mergeConstraints(other, nil)
	}
	sharedElements := g.immediateSharedElements(other).Contents()

	if len(sharedElements) != 2 {
		return solver.NonConvergent
	}

	// Solve two shared elements
	fmt.Printf("Initial configuration:\n")
	fmt.Printf("Shared elements %v\n", sharedElements)
	g.logElements()
	fmt.Println("")
	other.logElements()
	fmt.Println("")

	first := sharedElements[0]
	second := sharedElements[1]

	if g.elements[first].GetType() == el.Line {
		first, second = second, first
	}

	// If both elements are lines, nonconvergent (I think)
	if g.elements[first].GetType() == el.Line {
		fmt.Println("In a merge one and both shared elements are line type")
		return solver.NonConvergent
	}

	p1 := g.elements[first]
	p2 := other.elements[first]

	// If there's a line, first rotate the lines into the same angle, then match first element
	if g.elements[second].GetType() == el.Line {
		angle := other.elements[second].AsLine().AngleToLine(g.elements[second].AsLine())
		other.Rotate(p1.AsPoint(), angle)
		fmt.Println("Rotated to make line the same angle")
	}

	// Match up the first point
	fmt.Println("matching up the first point")
	direction := p1.VectorTo(p2)
	other.Translate(direction.X, direction.Y)

	// If both are points, rotate other to match the element in g
	if g.elements[second].GetType() == el.Point {
		fmt.Println("both elements were points, rotating to match the points together")
		v1 := g.elements[second].VectorTo(g.elements[first])
		v2 := other.elements[second].VectorTo(other.elements[first])
		angle := v1.AngleTo(v2)
		other.Rotate(p1.AsPoint(), angle)
	}

	return solver.Solved
}

func (g *GraphCluster) mergeConstraints(c1 *GraphCluster, c2 *GraphCluster) {
	if c1 != nil {
		for _, c := range c1.constraints {
			g.AddConstraint(c)
		}
	}
	if c2 != nil {
		for _, c := range c2.constraints {
			g.AddConstraint(c)
		}
	}
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
func (g *GraphCluster) solveMerge(c1 *GraphCluster, c2 *GraphCluster) solver.SolveState {
	if c2 == nil {
		fmt.Println("Beginning one cluster merge")
		return g.mergeOne(c1, true)
	}
	// Move constraints / elements from c1, c2 to g when we're done
	defer g.mergeConstraints(c1, c2)
	fmt.Println()
	fmt.Println("Beginning cluster merge")
	solve := g.IsSolved()
	fmt.Printf("Checking g solved: %v\n", solve)
	solve = c1.IsSolved()
	fmt.Printf("Checking c1 solved: %v\n", solve)
	solve = c2.IsSolved()
	fmt.Printf("Checking c2 solved: %v\n", solve)
	fmt.Println()
	fmt.Println("Pre-merge state:")
	fmt.Println("g:")
	g.logElements()
	fmt.Println("c1:")
	c1.logElements()
	fmt.Println("c2:")
	c2.logElements()
	clusters := []*GraphCluster{g, c1, c2}
	sharedSet := g.immediateSharedElements(c1)
	sharedSet.AddSet(g.immediateSharedElements(c2))
	sharedSet.AddSet(c1.immediateSharedElements(c2))
	sharedElements := sharedSet.Contents()
	fmt.Printf("Solving for shared elements %v\n", sharedElements)

	clustersFor := func(e uint) []*GraphCluster {
		matching := make([]*GraphCluster, 0)
		for _, c := range clusters {
			if _, ok := c.elements[e]; !ok {
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
			if e, ok := g.elements[se]; ok && e.GetType() == el.Line {
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
	fmt.Printf("root cluster is g%d\n", finalIndex)
	// fmt.Printf("Initial configuration:\n")
	// g.logElements()
	// c1.logElements()
	// c2.logElements()
	// fmt.Println("")

	for _, se := range sharedElements {
		parents := clustersFor(se)
		if len(parents) != 2 {
			fmt.Printf("Shared element %d only has %d parents. Returning Non-Convergent\n", se, len(parents))
			return solver.NonConvergent
		}

		if parents[0] != rootCluster && parents[1] != rootCluster {
			final = se
			continue
		}
		eType := parents[0].elements[se].GetType()
		fmt.Printf("Solving for element %d (%v)\n", se, eType)

		// Solve element
		// if element is a line, rotate it into place first
		other := parents[0]
		if other == rootCluster {
			other = parents[1]
		}
		ec1 := other.elements[se]
		ec2 := rootCluster.elements[se]
		var translation *el.Vector
		if eType == el.Line {
			other.logElements()
			fmt.Println()
			angle := ec1.AsLine().AngleToLine(ec2.AsLine())
			// fmt.Printf("Before rotate:\n")
			// fmt.Printf("Element 1: %v\n", ec1)
			// fmt.Printf("Element 2: %v\n", ec2)
			// other.logElements()
			// fmt.Printf("Calculated angle %f\n", angle)
			other.Rotate(ec1.AsLine().PointNearestOrigin(), angle)
			// fmt.Printf("After rotate:\n")
			// fmt.Printf("Element 1: %v\n", ec1)
			// fmt.Printf("Element 2: %v\n", ec2)
			// other.logElements()
			// fmt.Println()
			translation = ec1.VectorTo(ec2)
		} else {
			translation = ec2.VectorTo(ec1)
		}

		// translate element into place
		other.Translate(translation.X, translation.Y)
		// if eType == el.Line {
		// 	fmt.Printf("After translate distance %f, %f:\n", translation.X, translation.Y)
		// 	other.logElements()
		// 	fmt.Println()
		// }

		fmt.Printf("Solved for element %d:\n", se)
		g.logElements()
		c1.logElements()
		c2.logElements()
		fmt.Println("")
	}

	var e = [2]uint{sharedElements[0], sharedElements[1]}
	if e[0] == final {
		e[0] = sharedElements[2]
	}
	if e[1] == final {
		e[1] = sharedElements[2]
	}
	fmt.Printf("Solved all but final element (%d): %d, %d\n", final, e[0], e[1])
	g.logElements()
	fmt.Println("")
	c1.logElements()
	fmt.Println("")
	c2.logElements()
	fmt.Println("")

	// Solve the third element in relation to the other two
	parents := clustersFor(final)
	c0E2 := parents[0].elements[final]
	c1E2 := parents[1].elements[final]
	e2Type := c0E2.GetType()
	fmt.Printf("Final element type: %v\n", e2Type)
	if e2Type == el.Line {
		// We avoid e2 being a line, so if it is one, the other two are also lines.
		// This means e2 should already be placed correctly since the other two are.
		state := solver.Solved
		c0E2 := parents[0].elements[final]
		c1E2 := parents[1].elements[final]
		if !c0E2.AsLine().IsEquivalent(c1E2.AsLine()) {
			fmt.Println("Lines are not equivalent: ")
			fmt.Printf("\t(%d): %v\n", c0E2.GetID(), c0E2)
			fmt.Printf("\t(%d): %v\n", c1E2.GetID(), c1E2)
			state = solver.NonConvergent
		}

		return state
	}

	var constraint1, constraint2 *Constraint
	var e1, e2 el.SketchElement
	if e, ok := parents[0].elements[e[0]]; ok {
		e1 = e
		dist := c0E2.DistanceTo(e)
		constraint1 = constraint.NewConstraint(0, constraint.Distance, c0E2, e, dist, false)
		fmt.Printf("Creating constraint from %d to %d with distance %f\n", c0E2.GetID(), e.GetID(), dist)
	}
	if e, ok := parents[1].elements[e[0]]; ok {
		e2 = e
		dist := c1E2.DistanceTo(e)
		constraint1 = constraint.NewConstraint(0, constraint.Distance, c1E2, e, dist, false)
		fmt.Printf("Creating constraint from %d to %d with distance %f\n", c1E2.GetID(), e.GetID(), dist)
	}
	if e, ok := parents[0].elements[e[1]]; ok {
		e1 = e
		dist := c0E2.DistanceTo(e)
		constraint2 = constraint.NewConstraint(0, constraint.Distance, c0E2, e, dist, false)
		fmt.Printf("Creating constraint from %d to %d with distance %f\n", c0E2.GetID(), e.GetID(), dist)
	}
	if e, ok := parents[1].elements[e[1]]; ok {
		e2 = e
		dist := c1E2.DistanceTo(e)
		constraint2 = constraint.NewConstraint(0, constraint.Distance, c1E2, e, dist, false)
		fmt.Printf("Creating constraint from %d to %d with distance %f\n", c1E2.GetID(), e.GetID(), dist)
	}

	newE3, state := solver.ConstraintResult(constraint1, constraint2, c0E2)
	newP3 := newE3.AsPoint()

	if state != solver.Solved {
		fmt.Println("Final element solve failed")
		return state
	}

	fmt.Printf("Desired merge point for c1 and c2: %f, %f\n", newP3.GetX(), newP3.GetY())

	moveCluster := func(c *GraphCluster, pivot el.SketchElement, from *el.SketchPoint, to *el.SketchPoint) {
		if pivot.GetType() == el.Line {
			move := from.VectorTo(to)
			c.Translate(-move.X, -move.Y)
			return
		}

		current, desired := pivot.VectorTo(from), pivot.VectorTo(to)
		angle := current.AngleTo(desired)
		c.Rotate(pivot.AsPoint(), angle)
	}

	fmt.Printf("Pivoting c0 on %d from %v to %v\n", e1.GetID(), c0E2, newP3)
	moveCluster(parents[0], e1, c0E2.AsPoint(), newP3)
	fmt.Printf("c0E2 moved to %v\n", c0E2)
	fmt.Printf("Pivoting c1 on %d from %v to %v\n", e2.GetID(), c1E2, newP3)
	moveCluster(parents[1], e2, c1E2.AsPoint(), newP3)
	fmt.Printf("c1E2 moved to %v\n", c1E2)

	g.logElements()
	println()
	c1.logElements()
	println()
	c2.logElements()
	println()

	println("Completed parent cluster")
	g.logElements()
	println()

	return solver.Solved
}

// Solve solves the cluster and any child clusters associated with it
func (g *GraphCluster) Solve() solver.SolveState {
	// fmt.Printf("Solving cluster with %d other clusters\n", len(g.others))
	fmt.Printf("Solving cluster %d\n", g.id)
	state := g.localSolve()
	// if len(g.others) == 0 {
	// 	return state
	// }

	// If there are sub clusters, solve them
	// for i, cluster := range g.others {
	// 	// attempt as much of a solve as possible even if non-convergent
	// 	fmt.Printf("Solving other cluster %d\n", i)
	// 	otherState := cluster.Solve()
	// 	if state == solver.Solved && otherState != solver.Solved {
	// 		state = otherState
	// 	}
	// }

	// Now use rigid body transforms to move cluster elements into place
	// for len(g.others) > 0 {
	// 	// Find clusters which can me merged
	// 	first, second := g.findMerge()
	// 	if first < 0 && second < 0 {
	// 		// Need to copy elements out of any clusters remaining
	// 		for _, other := range g.others {
	// 			for _, c := range other.constraints {
	// 				g.AddConstraint(c)
	// 			}
	// 		}
	// 		return solver.NonConvergent
	// 	}
	// 	var secondC *GraphCluster = nil
	// 	if second >= 0 {
	// 		secondC = g.others[second]
	// 		swapTo := len(g.others) - 2
	// 		g.others[swapTo], g.others[second] = g.others[second], g.others[swapTo]
	// 	}
	// 	firstC := g.others[first]

	// 	swapTo := len(g.others) - 1
	// 	g.others[swapTo], g.others[first] = g.others[first], g.others[swapTo]

	// 	remove := 1
	// 	if second >= 0 {
	// 		remove = 2
	// 		g.others[len(g.others)-2] = nil
	// 	}
	// 	g.others[len(g.others)-1] = nil
	// 	g.others = g.others[:len(g.others)-remove]

	// 	mergeState := g.solveMerge(firstC, secondC)
	// 	if state == solver.Solved && mergeState != solver.Solved {
	// 		state = mergeState
	// 	}
	// }

	return state
}
func (c *GraphCluster) IsSolved() bool {
	solved := true
	for _, c := range c.constraints {
		if c.IsMet() {
			continue
		}

		fmt.Printf("Failed to meet %v\n", c)
		solved = false
	}

	return solved
}

// TODO: This infinite loops!
func (c *GraphCluster) ToGraphViz() string {
	edges := ""
	elements := ""
	for _, constraint := range c.constraints {
		edges = edges + constraint.ToGraphViz(c.id)
		elements = elements + constraint.Element1.ToGraphViz(c.id)
		elements = elements + constraint.Element2.ToGraphViz(c.id)
		// edges = edges + constraint.ToGraphViz("")
	}
	/*for _, other := range c.others {
		others = others + other.ToGraphViz()
		// edges = edges + other.ToGraphViz(fmt.Sprintf("%d", oid))
		for e := range c.eToC {
			if !other.HasElementID(e) {
				continue
			}
			sharedEdges = sharedEdges + fmt.Sprintf("\t\"%d-%d\" -- \"%d-%d\"\n", c.id, e, other.id, e)
		}

		for _, oother := range c.others {
			if oother.id == other.id {
				continue
			}
			for e := range other.eToC {
				if !oother.HasElementID(e) {
					continue
				}
				reverseEdge := fmt.Sprintf("\t\"%d-%d\" -- \"%d-%d\"\n", oother.id, e, other.id, e)
				edge := fmt.Sprintf("\t\"%d-%d\" -- \"%d-%d\"\n", other.id, e, oother.id, e)
				if betweenEdges.Contains(reverseEdge) || betweenEdges.Contains(edge) {
					continue
				}
				fmt.Printf("Adding edge: \n%sWith reverse edge: \n%s\n", edge, reverseEdge)
				betweenEdges.Add(edge)
			}
		}
	}*/
	return fmt.Sprintf(`subgraph cluster_%d {
		label = "Cluster %d"
		%s
		%s
	}`, c.id, c.id, edges, elements)
}
