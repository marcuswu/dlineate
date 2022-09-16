package core

import (
	"errors"
	"fmt"

	"github.com/marcuswu/dlineation/internal/constraint"
	el "github.com/marcuswu/dlineation/internal/element"
	"github.com/marcuswu/dlineation/internal/solver"
	"github.com/marcuswu/dlineation/utils"
)

// SketchGraph A graph representing a set of 2D sketch elements and constraints
type SketchGraph struct {
	constraints map[uint]*constraint.Constraint
	elements    map[uint]el.SketchElement
	eToC        map[uint][]*constraint.Constraint
	clusters    []*GraphCluster
	freeNodes   *utils.Set
	usedNodes   *utils.Set

	state            solver.SolveState
	degreesOfFreedom uint
	constraintMap    map[uint]int
}

// NewSketch creates a new sketch for solving
func NewSketch() *SketchGraph {
	g := new(SketchGraph)
	g.eToC = make(map[uint][]*Constraint, 0)
	g.constraints = make(map[uint]*Constraint, 0)
	g.elements = make(map[uint]el.SketchElement, 0)
	g.clusters = make([]*GraphCluster, 0, 1)
	g.freeNodes = utils.NewSet()
	g.usedNodes = utils.NewSet()
	g.state = solver.None
	g.degreesOfFreedom = 6
	return g
}

// GetElement gets an element from the graph
func (g *SketchGraph) GetElement(id uint) el.SketchElement {
	return g.elements[id]
}

// GetConstraint gets a constraint from the graph
func (g *SketchGraph) GetConstraint(id uint) (*constraint.Constraint, bool) {
	c, ok := g.constraints[id]
	return c, ok
}

// AddPoint adds a point to the sketch
func (g *SketchGraph) AddPoint(x float64, y float64) el.SketchElement {
	elementID := uint(len(g.elements))
	fmt.Printf("Adding point %f, %f as element %d\n", x, y, elementID)
	p := el.NewSketchPoint(elementID, x, y)
	g.freeNodes.Add(elementID)
	g.elements[elementID] = p
	return g.GetElement(elementID)
}

// AddLine adds a line to the sketch
func (g *SketchGraph) AddLine(a float64, b float64, c float64) el.SketchElement {
	elementID := uint(len(g.elements))
	fmt.Printf("Adding line %f, %f, %f as element %d\n", a, b, c, elementID)
	l := el.NewSketchLine(elementID, a, b, c)
	g.freeNodes.Add(elementID)
	g.elements[elementID] = l
	return g.GetElement(elementID)
}

func (g *SketchGraph) CombinePoints(e1 el.SketchElement, e2 el.SketchElement) el.SketchElement {
	// Look for any constraints referencing e2, replace with e1
	for _, constraint := range g.constraints {
		if constraint.Element1.GetID() == e2.GetID() {
			constraint.Element1 = e1
		}
		if constraint.Element2.GetID() == e2.GetID() {
			constraint.Element2 = e1
		}
	}
	// remove e2 from freenodes, elements
	g.eToC[e1.GetID()] = append(g.eToC[e1.GetID()], g.eToC[e2.GetID()]...)
	g.freeNodes.Remove(e2.GetID())
	delete(g.elements, e2.GetID())
	return e1
}

// AddConstraint adds a constraint to sketch elements
func (g *SketchGraph) AddConstraint(t constraint.Type, e1 el.SketchElement, e2 el.SketchElement, value float64) *constraint.Constraint {
	constraintID := uint(len(g.constraints))
	cType := "Distance"
	if t != constraint.Distance {
		cType = "Angle"
	}
	fmt.Printf("Adding %s constraint with value %f with id %d\n", cType, value, constraintID)
	constraint := constraint.NewConstraint(constraintID, t, e1, e2, value, false)
	g.constraints[constraintID] = constraint
	if _, ok := g.eToC[e1.GetID()]; !ok {
		g.eToC[e1.GetID()] = make([]*Constraint, 0, 1)
	}
	if _, ok := g.eToC[e2.GetID()]; !ok {
		g.eToC[e2.GetID()] = make([]*Constraint, 0, 1)
	}
	g.eToC[e1.GetID()] = append(g.eToC[e1.GetID()], constraint)
	g.eToC[e2.GetID()] = append(g.eToC[e2.GetID()], constraint)
	g.freeNodes.Add(e1.GetID())
	g.freeNodes.Add(e2.GetID())
	return constraint
}

/*func (g *SketchGraph) getClusterConstraints(c *GraphCluster) []uint {
	found := make([]uint, 0, 5)
	toAdd := make([]uint, 0, 5)

	// Iterate freeNodes -- for each element
	//   * Get a list of constraints for that element
	//   * Filter constraints for those with an element in c
	//	 * Add constraint to found if len(list) == 2

	/*for el := range g.freeNodes.Contents() {

	}* /

	// Look for any constraints with one element in the cluster
	for id, constraint := range g.constraints {
		if c.HasElement(constraint.Element1) || c.HasElement(constraint.Element2) {
			found = append(found, id)
		}
	}

	fmt.Printf("getClusterConstraints found %d connected constraints\n", len(found))
	for _, id := range found {
		fmt.Printf("getClusterConstraints connected constraint: %d\n", id)
	}

	elementCount := make(map[uint]uint)
	var updateCount = func(elementCount map[uint]uint, elementID uint) {
		_, ok := elementCount[elementID]
		if !ok {
			elementCount[elementID] = 0
		}
		elementCount[elementID]++
		fmt.Printf("getClusterConstraints updating element %d count to %d\n", elementID, elementCount[elementID])
	}

	// Count elements shared with other found constraints
	for _, constraintID := range found {
		fmt.Printf("getClusterConstraints updating count for constraint %d\n", constraintID)
		constraint := g.constraints[constraintID]
		if !c.HasElement(constraint.Element1) {
			// Create entry
			fmt.Printf("getClusterConstraints has Element1\n")
			updateCount(elementCount, constraint.Element1.GetID())
		}
		if !c.HasElement(constraint.Element2) {
			fmt.Printf("getClusterConstraints has Element2\n")
			updateCount(elementCount, constraint.Element2.GetID())
		}
	}

	for id, count := range elementCount {
		fmt.Printf("getClusterConstraints element %d connected to cluster %d times\n", id, count)
	}

	// Remove any constraints without an element in elementCount having a count != 2
	for _, constraintID := range found {
		constraint := g.GetConstraint(constraintID)
		if elementCount[constraint.Element1.GetID()] != 2 &&
			elementCount[constraint.Element2.GetID()] != 2 {
			continue
		}
		toAdd = append(toAdd, constraintID)
	}

	fmt.Printf("getClusterConstraints adding %d constraints\n", len(toAdd))
	return toAdd
}*/

func (g *SketchGraph) findStartConstraint() uint {
	// start with all constraints if g.clusters is empty
	// start with constraints with an element in existing clusters if g.clusters is not empty
	constraints := make([]uint, 0, len(g.constraints))
	if len(g.clusters) == 0 {
		for constraintId := range g.constraints {
			constraints = append(constraints, constraintId)
		}
	} else {
		for constraintId, constraint := range g.constraints {
			if g.usedNodes.Contains(constraint.Element1.GetID()) ||
				g.usedNodes.Contains(constraint.Element2.GetID()) {
				constraints = append(constraints, constraintId)
			}
		}
	}

	// Check unused elements in constraints for highest constraint count
	var retVal uint
	ccount := 0
	for _, constraintId := range constraints {
		eId := g.constraints[constraintId].Element1.GetID()
		if g.usedNodes.Contains(eId) {
			eId = g.constraints[constraintId].Element2.GetID()
		}
		if ccount < len(g.eToC[eId]) {
			continue
		}

		retVal = constraintId
		ccount = len(g.eToC[eId])
	}

	return retVal
}

// Returns element ids which are connected to the cluster by 2 constraints
func (g *SketchGraph) eligibleElements(c *GraphCluster) map[uint][]uint {
	eligible := make(map[uint][]uint)
	for _, eId := range g.freeNodes.Contents() {
		constraints := g.eToC[eId]
		connections := make([]uint, 0, 2)
		for _, constraint := range constraints {
			// skip constraints already in a cluster
			if _, ok := g.constraints[constraint.GetID()]; !ok {
				continue
			}
			if !c.HasElementID(constraint.Element1.GetID()) && !c.HasElementID(constraint.Element2.GetID()) {
				continue
			}
			connections = append(connections, constraint.GetID())
		}

		if len(connections) < 2 {
			continue
		}
		eligible[eId] = connections
	}

	return eligible
}

func (g *SketchGraph) createCluster(first uint) *GraphCluster {
	c := NewGraphCluster()

	// Add elements connected to other elements in the cluster by two constraints
	clusterNum := len(g.clusters)
	oc, ok := g.GetConstraint(first)
	if !ok {
		fmt.Printf("createCluster(%d): Failed to find initial constraint %d\n", clusterNum, first)
		return nil
	}
	// firstConstraint := constraint.CopyConstraint(oc)
	c.AddConstraint(oc)
	g.freeNodes.Remove(oc.Element1.GetID())
	g.freeNodes.Remove(oc.Element2.GetID())
	g.usedNodes.Add(oc.Element1.GetID())
	g.usedNodes.Add(oc.Element2.GetID())
	delete(g.constraints, first)
	for toAdd := g.eligibleElements(c); len(toAdd) > 0; toAdd = g.eligibleElements(c) {
		for eId, cIds := range toAdd {
			fmt.Printf("createCluster(%d): adding element %d\n", clusterNum, eId)
			level := el.FullyConstrained
			if len(cIds) > 2 {
				level = el.OverConstrained
			}
			g.elements[eId].SetConstraintLevel(level)
			for _, cId := range cIds {
				fmt.Printf("createCluster(%d): adding constraint id %d\n", clusterNum, cId)
				oc, ok = g.GetConstraint(cId)
				if !ok {
					fmt.Printf("createCluster(%d): Failed to find constraint %d for element %d\n", clusterNum, cId, eId)
					continue
				}
				//cc := constraint.CopyConstraint(oc)
				c.AddConstraint(oc)
				g.freeNodes.Remove(oc.Element1.GetID())
				g.freeNodes.Remove(oc.Element2.GetID())
				g.usedNodes.Add(oc.Element1.GetID())
				g.usedNodes.Add(oc.Element2.GetID())
				delete(g.constraints, cId)
			}
		}
	}
	fmt.Printf("createCluster(%d) completed building cluster with %d elements and %d constraints\n", clusterNum, len(c.elements), len(c.constraints))
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
	if connected[1] < uint(len(g.clusters)-1) {
		copy(g.clusters[connected[1]:], g.clusters[connected[1]+1:])
	}
	g.clusters[len(g.clusters)-1] = nil
	g.clusters = g.clusters[:len(g.clusters)-1]
	return true
}

func (g *SketchGraph) createClusters() {
	fmt.Printf("Creating clusters -- number of unassigned constraints: %d\n", len(g.constraints))
	for lastLen := len(g.constraints) + 1; len(g.constraints) > 0 && lastLen != len(g.constraints); {
		lastLen = len(g.constraints)
		// Find constraint to begin new cluster
		g.createCluster(g.findStartConstraint())
	}
	for _, c := range g.constraints {
		if !g.usedNodes.Contains(c.Element1.GetID()) {
			g.elements[c.Element1.GetID()].SetConstraintLevel(el.UnderConstrained)
		}
		if !g.usedNodes.Contains(c.Element2.GetID()) {
			g.elements[c.Element1.GetID()].SetConstraintLevel(el.UnderConstrained)
		}
	}
	fmt.Printf("Created clusters -- number of unassigned constraints: %d\n", len(g.constraints))
}

func (g *SketchGraph) mergeClusters() {
	for i := 0; i < len(g.clusters); i++ {
		if g.mergeCluster(i) {
			i = 0
		}
	}
}

func (g *SketchGraph) buildConstraintMap() {
	g.constraintMap = make(map[uint]int, len(g.elements))
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

	g.clusters[0].Rotate(origin, angle)

	return nil
}

func (g *SketchGraph) logConstraintsElements() {
	fmt.Printf("Elements: \n")
	for _, e := range g.elements {
		if e.GetType() == el.Point {
			fmt.Printf("\tPoint(%d) %f, %f\n", e.GetID(), e.(*el.SketchPoint).GetX(), e.(*el.SketchPoint).GetY())
		} else {
			fmt.Printf("\tLine(%d) %f, %f, %f\n", e.GetID(), e.(*el.SketchLine).GetA(), e.(*el.SketchLine).GetB(), e.(*el.SketchLine).GetC())
		}
	}
	fmt.Println()

	fmt.Printf("Constraints: \n")
	for _, c := range g.constraints {
		fmt.Printf("\tConstraint(%d) type %d, e1 %d, e2 %d\n", c.GetID(), c.Type, c.Element1.GetID(), c.Element2.GetID())
	}
	fmt.Println()
	fmt.Println()
}

func (g *SketchGraph) updateElements(c *GraphCluster) {
	for eId, e := range c.elements {
		g.elements[eId] = e
	}
}

// Solve builds the graph and solves the sketch
func (g *SketchGraph) Solve() solver.SolveState {
	g.logConstraintsElements()
	g.buildConstraintMap()
	if len(g.clusters) == 0 {
		g.createClusters()
	}
	fmt.Printf("Merging clusters beginning with %d clusters\n", len(g.clusters))
	g.mergeClusters()
	if len(g.clusters) > 1 {
		// set state, but attempt to solve as much as possible
		g.state = solver.UnderConstrained
		fmt.Printf("More than one cluster (under constrained in mergeClusters)\n")
	}

	fmt.Printf("Running cluster solves with %d clusters and %d unclustered constraints\n", len(g.clusters), len(g.constraints))
	fmt.Printf("The first cluster has %d constraints\n", len(g.clusters[0].constraints))
	for _, c := range g.clusters {
		clusterState := c.Solve()
		g.updateElements(c)
		fmt.Printf("Solved cluster with state %d\n", clusterState)
		c.logElements()
		if g.state == solver.None || (g.state != clusterState && !(g.state != solver.Solved && clusterState == solver.Solved)) {
			g.state = clusterState
		}
		fmt.Printf("Current graph solve state %d\n", g.state)
	}
	return g.state
}

// Test is a test function
func (g *SketchGraph) Test() string {
	return "SketchGraph"
}
