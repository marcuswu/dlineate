package core

import (
	"errors"
	"fmt"
	"sort"

	"github.com/marcuswu/dlineation/internal/constraint"
	el "github.com/marcuswu/dlineation/internal/element"
	"github.com/marcuswu/dlineation/internal/solver"
	iutils "github.com/marcuswu/dlineation/internal/utils"
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
	g.clusters = make([]*GraphCluster, 0, 2)
	g.freeNodes = utils.NewSet()
	g.usedNodes = utils.NewSet()
	g.state = solver.None
	g.degreesOfFreedom = 6

	c := NewGraphCluster()
	g.clusters = append(g.clusters, c)

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

func (g *SketchGraph) AddOrigin(x float64, y float64) el.SketchElement {
	elementID := uint(len(g.elements))
	fmt.Printf("Adding origin %f, %f as element %d\n", x, y, elementID)
	ax := el.NewSketchPoint(elementID, x, y)
	g.freeNodes.Add(elementID)
	g.elements[elementID] = ax

	g.addElementToCluster(g.clusters[0], ax)

	return g.GetElement(elementID)
}

func (g *SketchGraph) AddAxis(a float64, b float64, c float64) el.SketchElement {
	elementID := uint(len(g.elements))
	fmt.Printf("Adding axis %f, %f, %f as element %d\n", a, b, c, elementID)
	ax := el.NewSketchLine(elementID, a, b, c)
	g.freeNodes.Add(elementID)
	g.elements[elementID] = ax

	g.addElementToCluster(g.clusters[0], ax)

	return g.GetElement(elementID)
}

func (g *SketchGraph) CombinePoints(e1 el.SketchElement, e2 el.SketchElement) el.SketchElement {
	fmt.Printf("Combining elements %d and %d (removing %d)\n", e1.GetID(), e2.GetID(), e2.GetID())
	newE2 := e1
	newE1 := e1
	if g.clusters[0].HasElement(e1) {
		newE2 = el.CopySketchElement(e1)
	}
	if g.clusters[0].HasElement(e2) {
		newE1 = el.CopySketchElement(e2)
	}
	// Look for any constraints referencing e2, replace with e1
	for _, constraint := range g.constraints {
		if constraint.Element1.GetID() == e2.GetID() {
			constraint.Element1 = newE2
		}
		if constraint.Element2.GetID() == e2.GetID() {
			constraint.Element2 = newE2
		}
		if constraint.Element1.GetID() == e1.GetID() && !e1.Is(newE1) {
			constraint.Element1 = newE1
		}
		if constraint.Element2.GetID() == e1.GetID() && !e1.Is(newE1) {
			constraint.Element2 = newE1
		}
	}
	// remove e2 from freenodes, elements
	g.eToC[newE1.GetID()] = append(g.eToC[newE1.GetID()], g.eToC[e2.GetID()]...)
	delete(g.eToC, e2.GetID())
	g.freeNodes.Remove(e2.GetID())
	delete(g.elements, e2.GetID())
	return newE1
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
	if g.clusters[0].HasElement(e1) && g.clusters[0].HasElement(e2) {
		g.clusters[0].AddConstraint(constraint)
		g.constraints[constraintID] = constraint
		return constraint
	}
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

func (g *SketchGraph) findStartConstraint() uint {
	// start with all constraints if g.clusters is empty
	// start with constraints with an element in existing clusters if g.clusters is not empty
	constraints := make([]uint, 0, len(g.constraints))
	// if len(g.clusters) < 2 {
	// 	for constraintId := range g.constraints {
	// 		constraints = append(constraints, constraintId)
	// 	}
	// } else {
	for constraintId, constraint := range g.constraints {
		if g.usedNodes.Contains(constraint.Element1.GetID()) ||
			g.usedNodes.Contains(constraint.Element2.GetID()) {
			constraints = append(constraints, constraintId)
		}
	}
	// }

	// Check unused elements in constraints for highest constraint count
	var retVal uint
	ccount := 0
	sort.Sort(iutils.IdList(constraints))
	for _, constraintId := range constraints {
		eId := g.constraints[constraintId].Element1.GetID()
		if g.usedNodes.Contains(eId) {
			eId = g.constraints[constraintId].Element2.GetID()
		}
		if len(g.eToC[eId]) < ccount {
			continue
		}
		if g.constraints[constraintId].Type == constraint.Angle {
			continue
		}

		retVal = constraintId
		ccount = len(g.eToC[eId])
	}

	return retVal
}

// find a pair of free constraints which is connected to a single element (might be in another cluster)
// and each constraint shares an element with the cluster we're creating
func (g *SketchGraph) findConstraints(c *GraphCluster) ([]uint, uint, bool) {
	// First, find free constraints connected to the cluster, grouped by the element not in the cluster
	constraints := make(map[uint]*utils.Set)
	for _, constraint := range g.constraints {
		eligible := c.HasElementID(constraint.First().GetID()) || c.HasElementID(constraint.Second().GetID())
		other := constraint.First().GetID()
		if c.HasElementID(other) {
			other = constraint.Second().GetID()
		}
		if c.HasElementID(other) || !eligible {
			continue
		}

		if _, ok := constraints[other]; !ok {
			constraints[other] = utils.NewSet()
		}
		// fmt.Printf("findConstraints: element %d adding constraint %d\n", other, constraint.GetID())
		constraints[other].Add(constraint.GetID())
	}

	var first bool = true
	var element uint
	var targetConstraints []uint
	for eId, cs := range constraints {
		if cs.Count() < 2 {
			// fmt.Printf("findConstraints: skipping element %d with %d constraints\n", eId, len(cs))
			continue
		}
		if cs.Count() == 2 {
			return cs.Contents(), eId, true
		}
		if !first && cs.Count() < len(targetConstraints) {
			// fmt.Printf("findConstraints: target element has more constraints than %d\n", len(cs))
			continue
		}

		element = eId
		targetConstraints = cs.Contents()
		first = false
		// fmt.Printf("findConstraints: setting target element to %d with %d constraints\n", eId, len(cs))
	}

	// fmt.Printf("findConstraints: returning target element %d with %d constraints\n", element, len(targetConstraints))
	return targetConstraints, element, len(targetConstraints) > 1
}

// Returns the constraints an element is connected to the given cluster by
func (g *SketchGraph) availableConstraints(c *GraphCluster, eId uint) ([]uint, bool) {
	constraints := g.eToC[eId]
	connections := utils.NewSet()
	for _, constraint := range constraints {
		// skip constraints already in a cluster
		if _, ok := g.constraints[constraint.GetID()]; !ok {
			continue
		}
		// skip constraints not connected to the cluster we're building
		if !c.HasElementID(constraint.Element1.GetID()) && !c.HasElementID(constraint.Element2.GetID()) {
			continue
		}
		connections.Add(constraint.GetID())
	}

	if connections.Count() < 2 {
		return connections.Contents(), false
	}
	orderedConnections := connections.Contents()
	sort.Sort(iutils.IdList(orderedConnections))
	return orderedConnections, true
}

func (g *SketchGraph) addElementToCluster(c *GraphCluster, e el.SketchElement) {
	c.AddElement(e)
	g.freeNodes.Remove(e.GetID())
	g.usedNodes.Add(e.GetID())
}

func (g *SketchGraph) addConstraintToCluster(c *GraphCluster, constraint *constraint.Constraint) {
	g.addElementToCluster(c, constraint.Element1)
	g.addElementToCluster(c, constraint.Element2)
	c.AddConstraint(constraint)
	delete(g.constraints, constraint.GetID())
}

func (g *SketchGraph) createCluster(first uint) *GraphCluster {
	c := NewGraphCluster()

	// Add elements connected to other elements in the cluster by two constraints
	clusterNum := len(g.clusters)
	oc, ok := g.GetConstraint(first)
	if !ok {
		fmt.Printf("createCluster(%d): Failed to find initial constraint from first constraint %d\n", clusterNum, first)
		return nil
	}
	// firstConstraint := constraint.CopyConstraint(oc)
	g.addConstraintToCluster(c, constraint.CopyConstraint(oc))

	// find a pair of free constraints which is connected to an element (might be in another cluster)
	// and each constraint shares an element with the cluster we're creating
	for cIds, eId, ok := g.findConstraints(c); ok; cIds, eId, ok = g.findConstraints(c) {
		fmt.Printf("createCluster(%d): adding element %d, %d constraints, ok: %v\n", clusterNum, eId, len(cIds), ok)
		level := el.FullyConstrained
		if len(cIds) > 2 {
			level = el.OverConstrained
		}
		g.elements[eId].SetConstraintLevel(level)
		element := el.CopySketchElement(g.elements[eId])
		// if g.clusters[0].HasElementID(eId) {
		// 	element = el.CopySketchElement(element)
		// }
		c.AddElement(element)
		for _, cId := range cIds[:2] {
			fmt.Printf("createCluster(%d): adding constraint id %d\n", clusterNum, cId)
			oc, ok = g.GetConstraint(cId)
			if !ok {
				fmt.Printf("createCluster(%d): Failed to find constraint %d for element %d\n", clusterNum, cId, eId)
				continue
			}
			//cc := constraint.CopyConstraint(oc)
			g.addConstraintToCluster(c, oc)
		}
	}

	/*for toAdd := g.freeNodes.Contents(); len(toAdd) > 0; toAdd = g.freeNodes.Contents() {
		sort.Sort(iutils.IdList(toAdd))
		for _, eId := range toAdd {
			cIds, ok := g.availableConstraints(c, eId)
			if !ok {
				continue
			}
			fmt.Printf("createCluster(%d): adding element %d\n", clusterNum, eId)
			level := el.FullyConstrained
			if len(cIds) > 2 {
				level = el.OverConstrained
			}
			g.elements[eId].SetConstraintLevel(level)
			c.AddElement(g.elements[eId])
			for _, cId := range cIds {
				fmt.Printf("createCluster(%d): adding constraint id %d\n", clusterNum, cId)
				oc, ok = g.GetConstraint(cId)
				if !ok {
					fmt.Printf("createCluster(%d): Failed to find constraint %d for element %d\n", clusterNum, cId, eId)
					continue
				}
				//cc := constraint.CopyConstraint(oc)
				g.addConstraintToCluster(c, oc)
			}
		}
	}*/
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
		if i == 0 {
			continue
		}
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
		fmt.Printf("%d unassigned constraints left\n", len(g.constraints))
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
	for i := 1; i < len(g.clusters); i++ {
		if g.mergeCluster(i) {
			i = 1
		}
	}
	g.mergeCluster(0)
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
			fmt.Printf("\tPoint(%d) (%f, %f)\n", e.GetID(), e.(*el.SketchPoint).GetX(), e.(*el.SketchPoint).GetY())
		} else {
			fmt.Printf("\tLine(%d) %fx + %fy + %f = 0\n", e.GetID(), e.(*el.SketchLine).GetA(), e.(*el.SketchLine).GetB(), e.(*el.SketchLine).GetC())
		}
	}
	fmt.Println()

	fmt.Printf("Constraints: \n")
	for _, c := range g.constraints {
		fmt.Printf("\tConstraint(%d) type: %v, e1: %d, e2: %d, v: %f\n", c.GetID(), c.Type, c.Element1.GetID(), c.Element2.GetID(), c.Value)
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
	for _, c := range g.clusters[0].constraints {
		delete(g.constraints, c.GetID())
	}
	elements := utils.NewSet()
	for _, c := range g.constraints {
		elements.AddList(c.ElementIDs())
	}
	// TODO: Add back in constraints from cluster where both elements
	// are referenced in constraints not in the cluster
	for _, c := range g.clusters[0].constraints {
		if elements.Contains(c.Element1.GetID()) && elements.Contains(c.Element2.GetID()) {
			fmt.Printf("Adding back in constraint %d with elements %d and %d\n", c.GetID(), c.First().GetID(), c.Second().GetID())
			g.constraints[c.GetID()] = c
		}
	}
	g.logConstraintsElements()
	g.buildConstraintMap()
	if len(g.clusters) == 1 {
		g.createClusters()
	}
	fmt.Printf("Merging clusters beginning with %d clusters\n", len(g.clusters))
	g.mergeClusters()
	if len(g.clusters) > 2 {
		// set state, but attempt to solve as much as possible
		g.state = solver.UnderConstrained
		fmt.Printf("More than two clusters (%d) (under constrained in mergeClusters)\n", len(g.clusters))
	}

	fmt.Printf("Beginning cluster solves with %d clusters and %d unclustered constraints\n", len(g.clusters), len(g.constraints))
	for i, c := range g.clusters {
		if i == 0 {
			continue
		}
		fmt.Printf("Starting cluster %d solve with %d other clusters\n", i, len(c.others))
		clusterState := c.Solve()
		g.updateElements(c)
		fmt.Printf("Solved cluster %d with state %v, current state %v\n", i, clusterState, g.state)
		c.logElements()
		if g.state == solver.None || (g.state != clusterState && !(g.state != solver.Solved && clusterState == solver.Solved)) {
			fmt.Printf("Updating state to %v after cluster %d solve\n", clusterState, i)
			g.state = clusterState
		}
		fmt.Printf("Current graph solve state %v\n", g.state)
	}
	mergeState := g.clusters[0].mergeOne(g.clusters[1])
	if g.state == solver.None || (g.state != mergeState && !(g.state != solver.Solved && mergeState == solver.Solved)) {
		fmt.Printf("Updating state to %v after cluster merge\n", mergeState)
		g.state = mergeState
	}

	return g.state
}

// Test is a test function
func (g *SketchGraph) Test() string {
	return "SketchGraph"
}
