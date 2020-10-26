package core

import (
	"github.com/marcuswu/dlineate/internal/constraint"
	el "github.com/marcuswu/dlineate/internal/element"
	"github.com/marcuswu/dlineate/internal/solver"
	"github.com/marcuswu/dlineate/utils"
)

// Constraint a convenient alias for cosntraint.Constraint
type Constraint = constraint.Constraint

// GraphCluster A cluster within a Graph
type GraphCluster struct {
	constraints    []*Constraint
	others         []*GraphCluster
	elements       map[uint]el.SketchElement
	solvedElements *utils.Set
}

// NewGraphCluster constructs a new GraphCluster
func NewGraphCluster() *GraphCluster {
	g := new(GraphCluster)
	g.constraints = make([]*Constraint, 0, 2)
	g.others = make([]*GraphCluster, 0, 0)
	g.elements = make(map[uint]el.SketchElement, 0)
	g.solvedElements = utils.NewSet()
	return g
}

// AddConstraint adds a constraint to the cluster
func (g *GraphCluster) AddConstraint(c *Constraint) {
	g.constraints = append(g.constraints, c)
	g.elements[c.Element1.GetID()] = c.Element1
	g.elements[c.Element2.GetID()] = c.Element2
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

	for _, c := range g.others {
		shared.AddSet(c.SharedElements(gc))
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
	for _, e := range g.elements {
		e.Translate(-origin.GetX(), -origin.GetY())
		e.Rotate(angle)
		e.Translate(origin.GetX(), origin.GetY())
	}
}

// LocalSolve attempts to solve the constraints in the cluster, returns solution state
func (g *GraphCluster) localSolve() solver.SolveState {
	// A map of element id to constraint which still need to be matched
	// for solving
	available := make(map[uint]*Constraint, len(g.constraints))
	toSolve := make([]*Constraint, len(g.constraints))
	copy(toSolve, g.constraints)
	removeItem := func(list []*Constraint, i int) ([]*Constraint, *Constraint) {
		var item *Constraint
		item, list[i] = list[i], list[len(list)-1]
		return list[:len(list)-1], item
	}
	findConstraint := func() []*Constraint {
		var c *Constraint

		// If we don't have any available elements to create a solve pair, select one
		if len(available) == 0 {
			toSolve, c = removeItem(toSolve, 0)
			available[c.Element1.GetID()] = c
			available[c.Element2.GetID()] = c
		}

		// Find something in toSolve which matches one element in available
		for i, c := range toSolve {
			c2, ok := available[c.Element1.GetID()]
			if ok {
				delete(available, c.Element1.GetID())
				available[c.Element2.GetID()] = c
				toSolve, _ = removeItem(toSolve, i)
				return []*Constraint{c, c2}
			}
			c2, ok = available[c.Element2.GetID()]
			if ok {
				delete(available, c.Element2.GetID())
				available[c.Element1.GetID()] = c
				toSolve, _ = removeItem(toSolve, i)
				return []*Constraint{c, c2}
			}
		}

		return []*Constraint{}
	}
	// Find constraints that share an element, solve that pair of constraints
	// Use findConstraint to get two constraints and solve them
	for len(toSolve) > 0 {
		current := findConstraint()
		if len(current) == 0 && len(toSolve) > 0 {
			return solver.NonConvergent
		}
		// TODO: This isn't right! solver.SolveConstraints may alter
		// already "solved" elements. If that happens, other already
		// solved elements need to have rigid transformations applied
		// to stay in a solved arrangement.
		// TODO: If current[0] and current[1] contain lines, then we
		// need to also find the angle constraint between those lines
		// and solve for it.
		solver.SolveConstraints(current[0], current[1])
	}

	// Continue until toSolve is empty
	return solver.Solved
}

// SolveMerge resolves merging solved child clusters to this one
func (g *GraphCluster) solveMerge(c1 *GraphCluster, c2 *GraphCluster) solver.SolveState {
	// Find the shared element between g and c1
	sharedSet := g.SharedElements(c1)
	if len(sharedSet.Contents()) != 1 {
		return solver.NonConvergent
	}
	elementID := sharedSet.Contents()[0]
	c1SharedG, _ := c1.GetElement(elementID)
	gElement, _ := g.GetElement(elementID)

	// Translate c1 by the difference in position
	translateVector := c1SharedG.VectorTo(gElement)
	c1.Translate(translateVector.X, translateVector.Y)

	// Find the shared element between g and c2
	sharedSet = g.SharedElements(c2)
	if len(sharedSet.Contents()) != 1 {
		return solver.NonConvergent
	}
	elementID = sharedSet.Contents()[0]
	c2SharedG, _ := c2.GetElement(elementID)
	gElement, _ = g.GetElement(elementID)

	// Translate c2 by the difference in position
	translateVector = c2SharedG.VectorTo(gElement)
	c2.Translate(translateVector.X, translateVector.Y)

	// Find the shared element between c1 and c2
	sharedSet = c1.SharedElements(c2)
	if len(sharedSet.Contents()) != 1 {
		return solver.NonConvergent
	}
	elementID = sharedSet.Contents()[0]
	c1P3, _ := c1.GetElement(elementID)
	c2P3, _ := c2.GetElement(elementID)
	p1, p2 := c1SharedG.AsPoint(), c2SharedG.AsPoint()
	if p1 == nil {
		p1 = c1SharedG.AsLine().PointNearestOrigin()
	}
	if p2 == nil {
		p2 = c2SharedG.AsLine().PointNearestOrigin()
	}
	// Find the rotation for c1 and c2 that allows shared element to meet
	// To do that, use the shared elements from g to c1 and c2
	// as p1 and p2 and the distances to the shared element between c1 and c2
	// as p3 as constraint distances. Use GetPointFromPoints to determine
	// the point c1 and c2 rotate to join on their shared element
	c1Dist, c2Dist := p1.DistanceTo(c1P3), p2.DistanceTo(c2P3)
	p3 := c1P3.AsPoint()
	if p3 == nil {
		p3 = c1P3.AsLine().PointNearestOrigin()
	}
	newP3, solved := solver.GetPointFromPoints(p1, p2, p3, c1Dist, c2Dist)
	if solved != solver.Solved {
		return solved
	}
	// Calculate the angle of rotation for c1 and c2 by creating
	// vectors from their points and getting the angle between the vectors
	// Rotate c1 and c2 so the shared element meets
	c1Angle, c1Desired := p1.VectorTo(p3), p1.VectorTo(newP3)
	c2Angle, c2Desired := p2.VectorTo(p3), p2.VectorTo(newP3)
	c1Rotate := c1Angle.AngleTo(c1Desired)
	c2Rotate := c2Angle.AngleTo(c2Desired)
	c1.Rotate(p1, c1Rotate)
	c2.Rotate(p2, c2Rotate)

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
