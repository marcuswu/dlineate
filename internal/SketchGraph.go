package core

import (
	"github.com/marcuswu/dlineation/internal/constraint"
	el "github.com/marcuswu/dlineation/internal/element"
	"github.com/marcuswu/dlineation/internal/solver"
	"github.com/marcuswu/dlineation/utils"
)

// SketchGraph A graph representing a set of 2D sketch elements and constraints
type SketchGraph struct {
	constraints map[uint]*constraint.Constraint
	elements    map[uint]el.SketchElement
	clusters    []*GraphCluster
	freeNodes   *utils.Set

	state            solver.SolveState
	degreesOfFreedom uint
}

// GetElement gets an element from the graph
func (g *SketchGraph) GetElement(id uint) el.SketchElement {
	return g.elements[id]
}

// GetConstraint gets a constraint from the graph
func (g *SketchGraph) GetConstraint(id uint) *constraint.Constraint {
	return g.constraints[id]
}

// AddPoint adds a point to the sketch
func (g *SketchGraph) AddPoint(x float64, y float64) el.SketchElement {
	elementID := uint(len(g.elements))
	p := el.NewSketchPoint(elementID, x, y)
	g.freeNodes.Add(elementID)
	g.elements[elementID] = p
	return g.GetElement(elementID)
}

// AddLine adds a line to the sketch
func (g *SketchGraph) AddLine(a float64, b float64, c float64) el.SketchElement {
	elementID := uint(len(g.elements))
	l := el.NewSketchLine(elementID, a, b, c)
	g.freeNodes.Add(elementID)
	g.elements[elementID] = l
	return g.GetElement(elementID)
}

// AddConstraint adds a constraint to sketch elements
func (g *SketchGraph) AddConstraint(t constraint.Type, e1 el.SketchElement, e2 el.SketchElement, value float64) *constraint.Constraint {
	constraintID := uint(len(g.constraints))
	constraint := constraint.NewConstraint(constraintID, t, e1, e2, value)
	g.constraints[constraintID] = &constraint
	g.freeNodes.Add(e1.GetID())
	g.freeNodes.Add(e2.GetID())
	return g.GetConstraint(constraintID)
}

func (g *SketchGraph) createClusters() {
	for len(g.constraints) > 0 {
		createCluster(0)
	}
}

func (g *SketchGraph) mergeClusters() {
	for i := 0; i < len(g.clusters); i++ {
		if mergeCluster(i) {
			i = 0
		}
	}
}

func (g *SketchGraph) getClusterConstraints(c *GraphCluster) {
	//found := make([]uint, len(g.constraints))
	//toAdd := make([]uint, toAdd)

	// Look for any constraints with one element in the cluster
}

// Solve builds the graph and solves the sketch
func (g *SketchGraph) Solve() solver.SolveState {
	g.createClusters()
	g.mergeClusters()
	if len(g.clusters) > 1 {
		g.state = solver.UnderConstrained
		return g.state
	}

	g.clusters[0].solve()
	return g.state
}

// Test is a test function
func (g *SketchGraph) Test() string {
	return "SketchGraph"
}
