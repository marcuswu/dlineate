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
	constraints []*Constraint
	others      []*GraphCluster
	elements    map[uint]el.SketchElement
}

// NewGraphCluster constructs a new GraphCluster
func NewGraphCluster() *GraphCluster {
	g := new(GraphCluster)
	g.constraints = make([]*Constraint, 0, 2)
	g.others = make([]*GraphCluster, 0, 0)
	g.elements = make(map[uint]el.SketchElement, 0)
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

func (g *GraphCluster) rebuildMap() {
	g.elements = make(map[uint]el.SketchElement, 0)

	for _, c := range g.constraints {
		g.elements[c.Element1.GetID()] = c.Element1
		g.elements[c.Element2.GetID()] = c.Element2
	}
}

// LocalSolve attempts to solve the constraints in the cluster, returns solution state
func (g *GraphCluster) localSolve() solver.SolveState {
	// Solve all angle constraints first
	var toSolve *utils.Set = utils.NewSet()
	for _, c := range g.constraints {
		if c.Type != constraint.Angle {
			toSolve.Add(c.Element1.GetID())
			toSolve.Add(c.Element2.GetID())
			continue
		}
		solver.SolveAngleConstraint(c)
	}

	constraintsWith := func(list []*Constraint, eID uint) []*Constraint {
		var constraints = make([]*Constraint, 0, 2)
		for _, c := range g.constraints {
			if c.Element1.GetID() == eID {
				constraints = append(constraints, c)
				continue
			}
			if c.Element2.GetID() == eID {
				constraints = append(constraints, c)
			}
		}
		return constraints
	}

	// for each unsolved element
	state := solver.Solved
	for _, e := range toSolve.Contents() {
		// find constraints involving it
		c := constraintsWith(g.constraints, e)
		toSolve.Remove(e)
		switch len(c) {
		case 2:
			if c[0].Type == constraint.Angle && c[1].Type == constraint.Angle {
				// over constrained
				state = solver.OverConstrained
				continue
			}
			if c[0].Type == constraint.Angle || c[1].Type == constraint.Angle {
				dc := c[0]
				if c[1].Type == constraint.Distance {
					dc = c[1]
				}
				other := dc.Element1
				if other.GetID() == e {
					other = dc.Element2
				}
				// angle constraints have been solved, so just move the line into place
				// Move the line to be constraint distance from point
				l := g.elements[e].(*el.SketchLine)
				l.TranslateDistance(dc.Value - l.DistanceTo(other))
				continue
			}
			if s := solver.SolveConstraints(c[0], c[1]); state == solver.Solved {
				state = s
			}
		case 1:
			fallthrough
		case 0:
			state = solver.UnderConstrained
		default:
			state = solver.OverConstrained
		}
	}

	// solver changes element instances in constraints, so rebuild the element map
	g.rebuildMap()

	return state
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
