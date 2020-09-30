package core

import "github.com/marcuswu/dlineation/utils"

// SolveState The state of the sketch graph
type SolveState uint

// SolveState constants
const (
	None SolveState = iota
	UnderConstrained
	OverConstrained
	NonConvergent
	Solved
)

// SketchGraph A graph representing a set of 2D sketch elements and constraints
type SketchGraph struct {
	constraints map[uint]Constraint
	elements    map[uint]SketchElement
	clusters    []GraphCluster
	freeNodes   utils.Set

	state            SolveState
	degreesOfFreedom uint
}

func (g *SketchGraph) addPoint(x float64, y float64) {
	elementId := uint(len(g.elements))

}

func (g *SketchGraph) getClusterConstraints(c *GraphCluster) {
	found := make([]uint, len(g.constraints))
	//toAdd := make([]uint, toAdd)

	// Look for any constraints with one element in the cluster
}

// Test is a test function
func (g *SketchGraph) Test() string {
	return "SketchGraph"
}
