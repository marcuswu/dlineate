package core

import (
	"github.com/marcuswu/dlineation/internal/constraint"
	el "github.com/marcuswu/dlineation/internal/element"
	"github.com/marcuswu/dlineation/internal/solver"
	"github.com/marcuswu/dlineation/utils"
)

// GraphCluster A cluster within a Graph
type GraphCluster struct {
	constraints    []*constraint.Constraint
	others         []*GraphCluster
	elements       utils.Set
	solvedElements utils.Set
}

// AddConstraint adds a constraint to the cluster
func (g *GraphCluster) AddConstraint(c *constraint.Constraint) {
	g.constraints = append(g.constraints, c)
	g.elements.Add(c.Element1.GetID())
	g.elements.Add(c.Element2.GetID())
}

// HasElementID returns whether this cluster contains an element ID
func (g *GraphCluster) HasElementID(eID uint) bool {
	elementIDs := g.elements.Contents()
	for _, elementID := range g.elements.Contents() {
		if eID == elementID {
			return true
		}
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

// SharedElements returns the shared elements between this and another cluster
func (g *GraphCluster) SharedElements(gc *GraphCluster) *utils.Set {
	var shared *utils.Set = utils.NewSet()

	for _, elementID := range g.elements.Contents() {
		if gc.HasElementID(elementID) {
			shared.Add(elementID)
		}
	}

	return shared
}

// LocalSolve attempts to solve the constraints in the cluster, returns solution state
func (g *GraphCluster) LocalSolve() solver.SolveState {
	// TODO
	return solver.Solved
}

// SolveMerge resolves merging solved child clusters to this one
func (g *GraphCluster) SolveMerge(c1 *GraphCluster, c2 *GraphCluster) solver.SolveState {
	// TODO
	return solver.Solved
}

// Solve solves the cluster and any child clusters associated with it
func (g *GraphCluster) Solve() solver.SolveState {
	// TODO
	return solver.Solved
}
