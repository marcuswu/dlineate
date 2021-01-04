package core

import (
	"fmt"
	"sort"

	"github.com/marcuswu/dlineate/internal/constraint"
	el "github.com/marcuswu/dlineate/internal/element"
	"github.com/marcuswu/dlineate/internal/solver"
	"github.com/marcuswu/dlineate/utils"
)

// Constraint a convenient alias for cosntraint.Constraint
type Constraint = constraint.Constraint

// GraphCluster A cluster within a Graph
type GraphCluster struct {
	constraints []*Constraint
	others      []*GraphCluster
	elements    map[uint]el.SketchElement
	eToC        map[uint][]*Constraint
	solved      *utils.Set
}

// NewGraphCluster constructs a new GraphCluster
func NewGraphCluster() *GraphCluster {
	g := new(GraphCluster)
	g.constraints = make([]*Constraint, 0, 2)
	g.others = make([]*GraphCluster, 0, 0)
	g.elements = make(map[uint]el.SketchElement, 0)
	g.eToC = make(map[uint][]*Constraint)
	g.solved = utils.NewSet()
	return g
}

// AddConstraint adds a constraint to the cluster
func (g *GraphCluster) AddConstraint(c *Constraint) {
	g.constraints = append(g.constraints, c)
	if _, ok := g.elements[c.Element1.GetID()]; !ok {
		g.elements[c.Element1.GetID()] = c.Element1
	}
	if _, ok := g.elements[c.Element2.GetID()]; !ok {
		g.elements[c.Element2.GetID()] = c.Element2
	}
	if _, ok := g.eToC[c.Element1.GetID()]; !ok {
		g.eToC[c.Element1.GetID()] = make([]*Constraint, 0, 1)
	}
	if _, ok := g.eToC[c.Element2.GetID()]; !ok {
		g.eToC[c.Element2.GetID()] = make([]*Constraint, 0, 1)
	}
	g.eToC[c.Element1.GetID()] = append(g.eToC[c.Element1.GetID()], c)
	g.eToC[c.Element2.GetID()] = append(g.eToC[c.Element2.GetID()], c)
}

// HasElementID returns whether this cluster contains an element ID
func (g *GraphCluster) HasElementID(eID uint) bool {
	_, e := g.elements[eID]
	if e {
		return true
	}
	for _, c := range g.others {
		if c.HasElementID(eID) {
			return true
		}
	}
	return false
}

// HasElement returns whether this cluster contains an element
func (g *GraphCluster) HasElement(e el.SketchElement) bool {
	return g.HasElementID(e.GetID())
}

// GetElement returns the copy of an element represented in this cluster
func (g *GraphCluster) GetElement(eID uint) (el.SketchElement, bool) {
	if element, ok := g.elements[eID]; ok {
		return element, ok
	}
	for _, c := range g.others {
		if element, ok := c.elements[eID]; ok {
			return element, ok
		}
	}
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
	var solvedC = make([]*Constraint, 0, len(constraints))
	for _, c := range constraints {
		if g.solved.Contains(c.GetID()) {
			solvedC = append(solvedC, c)
		}
	}
	return solvedC
}

func (g *GraphCluster) elementSolveCount(eID uint) int {
	solvedC := g.solvedConstraintsFor(eID)
	return len(solvedC)
}

func (g *GraphCluster) connectedPoints(eID uint) []uint {
	elements := utils.NewSet()
	for _, c := range g.eToC[eID] {
		if g.elementSolveCount(c.Element1.GetID()) < 2 && c.Element1.GetType() == el.Point {
			elements.Add(c.Element1.GetID())
		}
		if g.elementSolveCount(c.Element2.GetID()) < 2 && c.Element2.GetType() == el.Point {
			elements.Add(c.Element2.GetID())
		}
	}
	elements.Remove(eID)
	return elements.Contents()
}

func (g *GraphCluster) unsolvedConstraintsFor(eID uint) []*Constraint {
	var constraints = g.eToC[eID]
	var unsolved = make([]*Constraint, 0, len(constraints))
	for _, c := range constraints {
		if g.solved.Contains(c.GetID()) {
			continue
		}
		unsolved = append(unsolved, c)
	}

	return unsolved
}

func (g *GraphCluster) solvableConstraintsFor(eID uint) []*Constraint {
	var constraints = g.unsolvedConstraintsFor(eID)
	var done = g.solvedConstraintsFor(eID)
	var solvable = make([]*Constraint, 0, len(constraints))
	for _, c := range constraints {
		other := c.Element1.GetID()
		if other == eID {
			other = c.Element2.GetID()
		}
		if g.elementSolveCount(other) > 0 {
			solvable = append(solvable, c)
		}
	}

	return append(solvable, done...)
}

func (g *GraphCluster) orderedConstraintsFor(eID uint) []*Constraint {
	var constraints = g.unsolvedConstraintsFor(eID)
	var done = g.solvedConstraintsFor(eID)
	var solvable = make([]*Constraint, 0, len(constraints))
	var others = make([]*Constraint, 0, len(constraints))
	for _, c := range constraints {
		other := c.Element1.GetID()
		if other == eID {
			other = c.Element2.GetID()
		}
		if g.elementSolveCount(other) > 0 {
			solvable = append(solvable, c)
		} else {
			others = append(others, c)
		}
	}

	return append(append(solvable, done...), others...)
}

func (g *GraphCluster) unsolvedElements() []uint {
	unsolved := make([]uint, 0, len(g.eToC))
	for e := range g.eToC {
		solveCount := g.elementSolveCount(e)
		if solveCount < 2 {
			unsolved = append(unsolved, e)
		}
	}
	return unsolved
}

func (g *GraphCluster) solvedElements() []uint {
	solved := make([]uint, 0, len(g.eToC))
	for e := range g.eToC {
		solveCount := g.elementSolveCount(e)
		if solveCount > 1 {
			solved = append(solved, e)
		}
	}
	return solved
}

type elementSolvability struct {
	ID          uint
	Solvability int
}
type bySolvability []elementSolvability

func (a bySolvability) Len() int           { return len(a) }
func (a bySolvability) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a bySolvability) Less(i, j int) bool { return a[i].Solvability > a[j].Solvability }
func (g *GraphCluster) toSolvability(es []uint) bySolvability {
	result := make([]elementSolvability, 0, 1)
	for _, e := range es {
		if g.elements[e].GetType() == el.Line {
			continue
		}
		total := len(g.eToC[e])
		solvable := len(g.solvableConstraintsFor(e))
		solved := len(g.solvedConstraintsFor(e))
		unsolvable := total - (solvable + solved)
		fmt.Println(e, ": total", total, ", solvable", solvable, ", solved", solved, ", unsolvable", unsolvable)
		result = append(result, elementSolvability{ID: e, Solvability: (solvable + solved) - unsolvable})
	}
	return bySolvability(result)
}

func (g *GraphCluster) startElement() uint {
	unsolved := g.toSolvability(g.unsolvedElements())
	for _, e := range unsolved {
		fmt.Println("unsolved element", e.ID, "has solvability score", e.Solvability)
	}
	sort.Sort(unsolved)
	return unsolved[0].ID
}

func (g *GraphCluster) findElement() (uint, []*Constraint) {
	best := g.startElement()
	score := -len(g.constraints)
	solved := g.solvedElements()
	for _, e := range solved {
		points := g.toSolvability(g.connectedPoints(e))
		if len(points) == 0 {
			continue
		}
		sort.Sort(points)
		if points[0].Solvability > score {
			best = points[0].ID
			score = points[0].Solvability
		}
	}

	return best, g.orderedConstraintsFor(best)
}

func (g *GraphCluster) solvableLines(eID uint) map[uint]*Constraint {
	lines := make(map[uint]*Constraint)
	cList := g.unsolvedConstraintsFor(eID)
	for _, c := range cList {
		other := c.Element1
		if other.GetID() == eID {
			other = c.Element2
		}
		if other.GetType() == el.Point {
			continue
		}
		lineCs := g.unsolvedConstraintsFor(other.GetID())
		solved := g.elementSolveCount(other.GetID())
		var canSolve *Constraint = nil
		for _, lineC := range lineCs {
			if lineC.Element1.GetID() == eID && lineC.Element2.GetID() == other.GetID() {
				canSolve = lineC
				break
			}
			if lineC.Element2.GetID() == eID && lineC.Element1.GetID() == other.GetID() {
				canSolve = lineC
				break
			}
		}
		if solved == 1 && canSolve != nil {
			lines[other.GetID()] = canSolve
		}
	}
	return lines
}

// LocalSolve attempts to solve the constraints in the cluster, returns solution state
func (g *GraphCluster) localSolve() solver.SolveState {
	// solver changes element instances in constraints, so rebuild the element map
	defer g.rebuildMap()

	fmt.Println("Local Solve Step 0")
	for _, c := range g.constraints {
		if c.Type == constraint.Angle {
			fmt.Println("Solving constraint", c.GetID())
			solver.SolveAngleConstraint(c)
			g.solved.Add(c.GetID())
		}
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

	state := solver.Solved
	// Pick 2 from constraintList and solve. If only 1 in constraintList, solve just the one

	for g.solved.Count() < len(g.constraints) {
		// Step 1
		fmt.Println("Local Solve Step 1")
		e, c := g.findElement()

		fmt.Println("Solving for element", e)
		if len(c) == 0 {
			fmt.Println("Could not find a constraint to solve with", len(g.constraints)-g.solved.Count(), "constraints left to solve")
			state = solver.NonConvergent
			break
		}

		// Step 2
		fmt.Println("Local Solve Step 2")
		fmt.Println("Solving constraints", c[0].GetID(), c[1].GetID())
		if s := solver.SolveConstraints(c[0], c[1]); state == solver.Solved {
			fmt.Println("solve state changed to", s)
			state = s
		}
		g.solved.Add(c[0].GetID())
		g.solved.Add(c[1].GetID())

		// Step 3
		// Look for lines connected to e that have 1 solved constraint
		fmt.Println("Local Solve Step 3")
		lines := g.solvableLines(e)
		fmt.Println("lines found", lines)
		for _, c := range lines {
			fmt.Println("Solving constraint", c.GetID())
			if s := solver.MoveLineToPoint(c); state == solver.Solved {
				fmt.Println("solve state changed to", s)
				state = s
			}
			g.solved.Add(c.GetID())
		}
		fmt.Println("Local Solve Step 4 (check for completion)")
	}

	fmt.Println("finished with state", state)
	return state
}

func (g *GraphCluster) logElements() {
	logElement := func(e el.SketchElement) {
		point := e.AsPoint()
		line := e.AsLine()
		if point == nil {
			fmt.Printf("element %d: %fx + %fy + %f = 0\n", line.GetID(), line.GetA(), line.GetB(), line.GetC())
			return
		}
		fmt.Printf("element %d: (%f, %f)\n", point.GetID(), point.GetX(), point.GetY())
	}

	for _, e := range g.elements {
		logElement(e)
	}
}

// SolveMerge resolves merging solved child clusters to this one
func (g *GraphCluster) solveMerge(c1 *GraphCluster, c2 *GraphCluster) solver.SolveState {
	clusters := []*GraphCluster{g, c1, c2}
	sharedSet := g.SharedElements(c1)
	sharedSet.AddSet(g.SharedElements(c2))
	sharedSet.AddSet(c1.SharedElements(c2))
	sharedElements := sharedSet.Contents()

	clustersFor := func(e uint) []*GraphCluster {
		matching := make([]*GraphCluster, 0, len(clusters))
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

	// Solve two of the elements
	// Prefer solving lines first
	unsolved := 2
	var e [3]uint
	for i, se := range sharedElements {
		parents := clustersFor(se)
		if len(parents) != 2 {
			return solver.NonConvergent
		}

		eType := parents[0].elements[se].GetType()
		if eType != el.Line && i < len(sharedElements)-unsolved {
			e[2] = uint(i)
			continue
		}

		// Solve element
		// if element is a line, rotate it into place first
		ec1 := parents[0].elements[se]
		ec2 := parents[1].elements[se]
		if eType == el.Line {
			angle := ec1.AsLine().AngleToLine(ec2.AsLine())
			parents[1].Rotate(ec2.AsLine().PointNearestOrigin(), -angle)
		}

		// translate element into place
		translation := ec2.VectorTo(ec1)
		parents[1].Translate(-translation.X, -translation.Y)
		e[2-unsolved] = se

		unsolved--
		if unsolved == 0 {
			if i < len(sharedElements) {
				e[2] = uint(len(sharedElements) - 1)
			}
			break
		}
	}

	// Solve the third element in relation to the other two
	parents := clustersFor(e[2])
	c0E2 := parents[0].elements[e[2]]
	c1E2 := parents[1].elements[e[2]]
	e2Type := c0E2.GetType()
	if e2Type == el.Line {
		// We avoid e2 being a line, so if it is one, the other two are also lines.
		// This means e2 should already be placed correctly since the other two are.
		state := solver.Solved
		c0E2 := parents[0].elements[e[2]]
		c1E2 := parents[1].elements[e[2]]
		if !c0E2.AsLine().IsEquivalent(c1E2.AsLine()) {
			state = solver.NonConvergent
		}

		return state
	}

	var constraint1, constraint2 *Constraint
	var e1, e2 el.SketchElement
	var p1, p2 *el.SketchPoint
	if e, ok := parents[0].elements[e[0]]; ok {
		e1 = e
		dist := c0E2.DistanceTo(e)
		constraint1 = constraint.NewConstraint(0, constraint.Distance, c0E2, e, dist)
	}
	if e, ok := parents[1].elements[e[0]]; ok {
		e2 = e
		dist := c0E2.DistanceTo(e)
		constraint1 = constraint.NewConstraint(0, constraint.Distance, c0E2, e, dist)
	}
	if e, ok := parents[0].elements[e[1]]; ok {
		e1 = e
		dist := c0E2.DistanceTo(e)
		constraint2 = constraint.NewConstraint(0, constraint.Distance, c0E2, e, dist)
	}
	if e, ok := parents[1].elements[e[1]]; ok {
		e2 = e
		dist := c0E2.DistanceTo(e)
		constraint2 = constraint.NewConstraint(0, constraint.Distance, c0E2, e, dist)
	}

	newP3, state := solver.ConstraintResult(constraint1, constraint2)

	if state != solver.Solved {
		return state
	}

	p1 = e1.AsPoint()
	p2 = e2.AsPoint()
	if p1 == nil {
		p1 = e1.AsLine().PointNearestOrigin()
	}
	if p2 == nil {
		p2 = e2.AsLine().PointNearestOrigin()
	}
	c0Angle, c0Desired := p1.VectorTo(c0E2), p1.VectorTo(newP3)
	c1Angle, c1Desired := p2.VectorTo(c1E2), p2.VectorTo(newP3)
	c0Rotate := c0Angle.AngleTo(c0Desired)
	c1Rotate := c1Angle.AngleTo(c1Desired)
	fmt.Println("Angle to desired for cluster 1", c0Rotate)
	fmt.Println("Angle to desired for cluster 2", c1Rotate)
	parents[0].Rotate(p1, c0Rotate)
	parents[1].Rotate(p2, c1Rotate)
	// c0E2, _ = parents[0].GetElement(c0E2.GetID())
	// c1E2, _ = parents[1].GetElement(c1E2.GetID())
	c0Angle = p1.VectorTo(c0E2)
	c1Angle = p2.VectorTo(c1E2)
	fmt.Println("Angle to desired for cluster 1", c0Angle.AngleTo(c0Desired))
	fmt.Println("Angle to desired for cluster 2", c1Angle.AngleTo(c1Desired))
	// outputShared()
	g.logElements()
	c1.logElements()
	c2.logElements()

	// Move constraints / elements from c1 to g
	for _, c := range c1.constraints {
		g.AddConstraint(c)
	}
	// Move non-shared elements from c2 to g
	for _, c := range c2.constraints {
		g.AddConstraint(c)
	}

	return solver.Solved
}

// Solve solves the cluster and any child clusters associated with it
func (g *GraphCluster) Solve() solver.SolveState {
	state := g.localSolve()
	if state != solver.Solved {
		return state
	}
	if len(g.others) == 0 {
		return state
	}

	// If there are sub clusters, solve them
	for _, cluster := range g.others {
		state := cluster.Solve()
		if state != solver.Solved {
			return state
		}
	}

	// Now use rigid body transforms to move cluster elements into place
	for len(g.others) > 0 {
		// We will always have pairs added to others
		first := g.others[0]
		second := g.others[1]
		copy(g.others[0:], g.others[2:])
		g.others[len(g.others)-2] = nil
		g.others[len(g.others)-1] = nil
		g.others = g.others[:len(g.others)-2]

		state = g.solveMerge(first, second)
		if state != solver.Solved {
			break
		}
	}

	return state
}
