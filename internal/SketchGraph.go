package core

import (
	"errors"

	"github.com/marcuswu/dlineate/internal/constraint"
	el "github.com/marcuswu/dlineate/internal/element"
	"github.com/marcuswu/dlineate/internal/solver"
	"github.com/marcuswu/dlineate/utils"
)

// SketchGraph A graph representing a set of 2D sketch elements and constraints
type SketchGraph struct {
	constraints map[uint]*constraint.Constraint
	elements    map[uint]el.SketchElement
	clusters    []*GraphCluster
	freeNodes   *utils.Set

	state            solver.SolveState
	degreesOfFreedom uint
	constraintMap    []int
}

// NewSketch creates a new sketch for solving
func NewSketch() *SketchGraph {
	g := new(SketchGraph)
	g.constraints = make(map[uint]*Constraint, 0)
	g.elements = make(map[uint]el.SketchElement, 0)
	g.clusters = make([]*GraphCluster, 0, 1)
	g.freeNodes = utils.NewSet()
	g.state = solver.None
	g.degreesOfFreedom = 6
	return g
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
	g.constraints[constraintID] = constraint
	g.freeNodes.Add(e1.GetID())
	g.freeNodes.Add(e2.GetID())
	return g.GetConstraint(constraintID)
}

func (g *SketchGraph) getClusterConstraints(c *GraphCluster) []uint {
	found := make([]uint, 0, 5)
	toAdd := make([]uint, 0, 5)

	// Look for any constraints with one element in the cluster
	for id, constraint := range g.constraints {
		if c.HasElement(constraint.Element1) || c.HasElement(constraint.Element2) {
			found = append(found, id)
		}
	}

	elementCount := make(map[uint]uint)
	var updateCount = func(elementCount map[uint]uint, elementID uint) {
		_, ok := elementCount[elementID]
		if !ok {
			elementCount[elementID] = 0
		}
		elementCount[elementID]++
	}

	// Count elements shared with other found constraints
	for _, constraintID := range found {
		constraint := g.constraints[constraintID]
		if !c.HasElement(constraint.Element1) {
			// Create entry
			updateCount(elementCount, constraint.Element1.GetID())
		}
		if !c.HasElement(constraint.Element2) {
			updateCount(elementCount, constraint.Element2.GetID())
		}
	}
	// Remove any constraints without an element in elementCount having a count != 2
	for _, constraintID := range found {
		constraint := g.GetConstraint(constraintID)
		if elementCount[constraint.Element1.GetID()] != 2 {
			continue
		}
		if elementCount[constraint.Element2.GetID()] != 2 {
			continue
		}
		toAdd = append(toAdd, constraintID)
	}

	return toAdd
}

func (g *SketchGraph) createCluster(first uint) *GraphCluster {
	c := NewGraphCluster()

	toAdd := make([]uint, 0, 5)
	toAdd = append(toAdd, first)
	for len(toAdd) > 0 {
		for _, constraintID := range toAdd {
			constraint := constraint.CopyConstraint(g.GetConstraint(constraintID))
			c.AddConstraint(constraint)
			g.freeNodes.Remove(constraint.Element1.GetID())
			g.freeNodes.Remove(constraint.Element2.GetID())
			delete(g.constraints, constraintID)
		}
		toAdd = g.getClusterConstraints(c)
	}
	g.clusters = append(g.clusters, c)

	return c
}

func (g *SketchGraph) mergeCluster(index int) bool {
	// Look for 3 clusters sharing one element with each of the others
	// A mention at the end of https://www.cs.purdue.edu/homes/cmh/electrobook/our_solver1.html
	// indicates that, "If other clusters have two elements in common with the new cluster,
	// they can be merged into it as well."
	cluster := g.clusters[index]
	connected := make([]uint, 2)
	found := 0

	for i := range g.clusters {
		if found >= 2 {
			break
		}

		if i == index {
			continue
		}

		if cluster.SharedElements(g.clusters[i]).Count() == 1 {
			connected[found] = uint(i)
			found++
		}
	}

	if found != 2 {
		return false
	}

	cluster.others = append(cluster.others, g.clusters[connected[0]])
	cluster.others = append(cluster.others, g.clusters[connected[1]])

	// remove connected[0] and connected[1] from g.clusters
	copy(g.clusters[connected[0]:], g.clusters[connected[0]+1:])
	g.clusters[len(g.clusters)-1] = nil
	g.clusters = g.clusters[:len(g.clusters)-1]
	copy(g.clusters[connected[1]:], g.clusters[connected[1]+1:])
	g.clusters[len(g.clusters)-1] = nil
	g.clusters = g.clusters[:len(g.clusters)-1]
	return true
}

func (g *SketchGraph) createClusters() {
	for len(g.constraints) > 0 {
		g.createCluster(0)
	}
}

func (g *SketchGraph) mergeClusters() {
	for i := 0; i < len(g.clusters); i++ {
		if g.mergeCluster(i) {
			i = 0
		}
	}
}

func (g *SketchGraph) buildConstraintMap() {
	g.constraintMap = make([]int, len(g.elements))
	// Create an []int slice where indices represent element ids and values represent number of constraints
	for _, c := range g.constraints {
		g.constraintMap[c.Element1.GetID()]++
		g.constraintMap[c.Element2.GetID()]++
	}
}

func (g *SketchGraph) ConstraintLevel(e el.SketchElement) el.ConstraintLevel {
	numConstraints := g.constraintMap[e.GetID()]
	switch {
	case numConstraints < 2:
		return el.UnderConstrained
	case numConstraints > 2:
		return el.OverConstrained
	default:
		return el.FullyConstrained
	}
}

func (g *SketchGraph) UpdateConstraintLevels() {
	for _, e := range g.elements {
		e.SetConstraintLevel(g.ConstraintLevel(e))
	}
}

func (g *SketchGraph) Translate(x float64, y float64) error {
	if len(g.clusters) > 1 {
		return errors.New("Solve the sketch before translating it")
	}

	g.clusters[0].Translate(x, y)

	return nil
}

func (g *SketchGraph) Rotate(origin *el.SketchPoint, angle float64) error {
	if len(g.clusters) > 1 {
		return errors.New("Solve the sketch before translating it")
	}

	g.clusters[0].Rotate(angle)

	return nil
}

// Solve builds the graph and solves the sketch
func (g *SketchGraph) Solve() solver.SolveState {
	g.buildConstraintMap()
	g.createClusters()
	g.mergeClusters()
	if len(g.clusters) > 1 {
		g.state = solver.UnderConstrained
		// attempt to solve as much as possible
		//return g.state
	}

	clusterState := g.clusters[0].Solve()
	if g.state == solver.Solved && clusterState != solver.Solved {
		g.state = clusterState
	}
	g.UpdateConstraintLevels()
	return g.state
}

// Test is a test function
func (g *SketchGraph) Test() string {
	return "SketchGraph"
}
