package graph

import (
	"fmt"
	"math/big"

	"github.com/marcuswu/dlineate/internal/accessors"
	"github.com/marcuswu/dlineate/internal/constraint"
	el "github.com/marcuswu/dlineate/internal/element"
	"github.com/marcuswu/dlineate/internal/solver"
	"github.com/marcuswu/dlineate/utils"
	"github.com/rs/zerolog"
)

type SketchGraph struct {
	elementAccessor    accessors.ElementAccessor
	constraintAccessor accessors.ConstraintAccessor

	clusters  []*GraphCluster
	freeEdges *utils.Set // Constraints not yet referenced by a cluster
	usedNodes *utils.Set // Elements referenced in a cluster

	state            solver.SolveState
	degreesOfFreedom uint
	conflicting      *utils.Set
	freeEdgeMap      map[uint]*utils.Set
}

func NewSketch() *SketchGraph {
	g := new(SketchGraph)
	g.elementAccessor = accessors.NewElementRepository()
	g.constraintAccessor = accessors.NewConstraintRepository()
	g.clusters = make([]*GraphCluster, 0)
	g.freeEdges = utils.NewSet()
	g.usedNodes = utils.NewSet()
	g.state = solver.None
	g.degreesOfFreedom = 6
	g.conflicting = utils.NewSet()
	g.freeEdgeMap = make(map[uint]*utils.Set)
	return g
}

// GetElement gets an element from the graph
func (g *SketchGraph) GetElement(id uint) (el.SketchElement, bool) {
	return g.elementAccessor.GetElement(-1, id)
}

// GetConstraint gets a constraint from the graph
func (g *SketchGraph) GetConstraint(id uint) (*constraint.Constraint, bool) {
	return g.constraintAccessor.GetConstraint(id)
}

func (g *SketchGraph) MakeFixed(e el.SketchElement) {
	e.SetFixed(true)
}

// AddPoint adds a point to the sketch
func (g *SketchGraph) AddPoint(x *big.Float, y *big.Float) el.SketchElement {
	elementID := g.elementAccessor.NextId()
	utils.Logger.Debug().
		Str("X", x.String()).
		Str("Y", y.String()).
		Uint("id", elementID).
		Msg("Adding point")
	p := el.NewSketchPoint(elementID, x, y)
	// g.freeNodes.Add(elementID)
	g.elementAccessor.AddElement(p)
	return p
}

// AddLine adds a line to the sketch
func (g *SketchGraph) AddLine(a *big.Float, b *big.Float, c *big.Float) el.SketchElement {
	elementID := g.elementAccessor.NextId()
	utils.Logger.Debug().
		Str("A", a.String()).
		Str("B", b.String()).
		Str("C", c.String()).
		Uint("id", elementID).
		Msg("Adding line")
	l := el.NewSketchLine(elementID, a, b, c)
	// g.freeNodes.Add(elementID)
	g.elementAccessor.AddElement(l)
	return l
}

func (g *SketchGraph) AddOrigin(x *big.Float, y *big.Float) el.SketchElement {
	elementID := g.elementAccessor.NextId()
	utils.Logger.Debug().
		Str("X", x.String()).
		Str("Y", y.String()).
		Uint("id", elementID).
		Msgf("Adding origin")
	ax := el.NewSketchPoint(elementID, x, y)
	// g.freeNodes.Add(elementID)
	g.elementAccessor.AddElement(ax)

	g.MakeFixed(ax)

	return ax
}

func (g *SketchGraph) AddAxis(a *big.Float, b *big.Float, c *big.Float) el.SketchElement {
	elementID := g.elementAccessor.NextId()
	utils.Logger.Debug().
		Str("A", a.String()).
		Str("B", b.String()).
		Str("C", c.String()).
		Uint("id", elementID).
		Msg("Adding axis")
	ax := el.NewSketchLine(elementID, a, b, c)
	elementID = g.elementAccessor.NextId()
	x := big.NewFloat(0).Mul(a, big.NewFloat(1))
	y := big.NewFloat(0).Mul(b, big.NewFloat(1))
	end := el.NewSketchPoint(elementID, x, y)
	origin, _ := g.elementAccessor.GetElement(-1, 0)
	ax.Start = origin.AsPoint()
	ax.End = end
	// g.freeNodes.Add(elementID)
	g.elementAccessor.AddElement(ax)

	g.MakeFixed(ax)

	return ax
}

func (g *SketchGraph) IsElementSolved(e el.SketchElement) bool {
	constraints := g.constraintAccessor.ConstraintsForElement(e.GetID())
	if len(constraints) < 2 {
		return false
	}

	numSolved := 0
	for _, c := range constraints {
		if c.Solved {
			numSolved++
		}
	}

	return numSolved > 1
}

// CombinePoints combines two sketch elements representing points.
func (g *SketchGraph) CombinePoints(e1 el.SketchElement, e2 el.SketchElement) el.SketchElement {
	keep := e1
	rem := e2

	// If either element is a fixed element, we must keep it
	if rem.IsFixed() {
		keep, rem = rem, keep
	}

	if rem.IsFixed() {
		// Must keep both
		return e1
	}

	g.constraintAccessor.SetConstraintElement(rem.GetID(), keep.GetID())

	// remove e2 from freenodes, elements
	// g.freeNodes.Remove(rem.GetID())
	g.elementAccessor.RemoveElement(rem.GetID())
	return keep
}

// AddConstraint adds a constraint to existing sketch elements
func (g *SketchGraph) AddConstraint(t constraint.Type, e1 el.SketchElement, e2 el.SketchElement, value *big.Float) *constraint.Constraint {
	constraintID := g.constraintAccessor.NextId()
	cType := "Distance"
	if t != constraint.Distance {
		cType = "Angle"
	}
	utils.Logger.Debug().
		Str("type", cType).
		Str("value", value.String()).
		Uint("constraint id", constraintID).
		Uint("element 1", e1.GetID()).
		Uint("element 2", e2.GetID()).
		Msg("Adding constraint")
	constraint := constraint.NewConstraint(constraintID, t, e1.GetID(), e2.GetID(), value, e1.IsFixed() && e2.IsFixed())
	g.constraintAccessor.AddConstraint(constraint)
	g.freeEdges.Add(constraint.GetID())
	return constraint
}

func (g *SketchGraph) logConstraintsElements(level zerolog.Level) {
	numElements := g.elementAccessor.Count()
	numConstraints := g.constraintAccessor.Count()
	expectedConstraints := (numElements * 2) - 3
	utils.Logger.WithLevel(level).
		Bool("fully constrained", expectedConstraints == numConstraints).
		Msgf("C (%d) = 2 * E (%d) - 3", numConstraints, numElements)
	g.elementAccessor.LogElements(level)
	g.constraintAccessor.LogConstraints(level)
	utils.Logger.WithLevel(level).Msg("")
}

func (g *SketchGraph) ResetClusters() {
	g.clusters = make([]*GraphCluster, 0)
	g.freeEdges = g.constraintAccessor.IdSet()
	g.usedNodes.Clear()
	g.state = solver.None
	g.elementAccessor.ClearClusters()
	for k := range g.freeEdgeMap {
		delete(g.freeEdgeMap, k)
	}

	for _, cId := range g.constraintAccessor.IdSet().Contents() {
		c, _ := g.constraintAccessor.GetConstraint(cId)
		c.Solved = false
		if g.elementAccessor.IsFixed(c.Element1) && g.elementAccessor.IsFixed(c.Element2) {
			c.Solved = true
		}
	}
}

func (g *SketchGraph) Conflicting() *utils.Set {
	return g.conflicting
}

func (g *SketchGraph) Solve() solver.SolveState {
	defer g.elementAccessor.LogElements(zerolog.DebugLevel)

	utils.Logger.Info().
		Int("cluster count", len(g.clusters)).
		Int("constraint count", g.constraintAccessor.Count()).
		Msg("Beginning cluster solves")
	for i, c := range g.clusters {
		utils.Logger.Info().
			Int("cluster", i).
			Msg("Starting cluster solve")
		clusterState := c.Solve(g.elementAccessor, g.constraintAccessor)
		utils.Logger.Info().
			Int("cluster", i).
			Str("cluster state", clusterState.String()).
			Str("graph state", g.state.String()).
			Msg("Solved cluster")
		c.logElements(g.elementAccessor, zerolog.TraceLevel)
		if g.state == solver.None || (g.state != clusterState && !(g.state != solver.Solved && clusterState == solver.Solved)) {
			utils.Logger.Info().
				Int("cluster", i).
				Str("new state", clusterState.String()).
				Msg("Updating graph state after cluster solve")
			g.state = clusterState
		}
		utils.Logger.Debug().
			Str("graph state", g.state.String()).
			Msg("Current graph solve state")
	}
	// Merge clusters
	utils.Logger.Info().Msg("Starting Cluster Merges")
	g.mergeClusters()
	utils.Logger.Info().Int("cluster count", len(g.clusters)).Msg("Finished Cluster Merges")
	return g.state
}

func (g *SketchGraph) IsSolved() bool {
	solved := true
	for _, cId := range g.constraintAccessor.IdSet().Contents() {
		c, _ := g.constraintAccessor.GetConstraint(cId)
		if g.constraintAccessor.IsMet(c.GetID(), -1, g.elementAccessor) {
			continue
		}

		utils.Logger.Trace().
			Str("constraint", c.String()).
			Msg("Failed to meet constraint")
		solved = false
	}

	return solved
}

func (g *SketchGraph) ToGraphViz() string {
	edges := ""
	uniqueSharedElements := make(map[string]interface{})
	sharedElements := ""

	// Output clusters
	for _, c := range g.clusters {
		edges = edges + c.ToGraphViz(g.elementAccessor, g.constraintAccessor)
		for _, other := range g.clusters {
			if c.id == other.id {
				continue
			}
			shared := c.SharedElements(other)
			if shared.Count() == 0 {
				continue
			}
			first := c.id
			second := other.id
			if second < first {
				first, second = second, first
			}
			for _, e := range shared.Contents() {
				key := fmt.Sprintf("\t\"%d-%d\" -- \"%d-%d\"\n", first, e, second, e)
				if _, ok := uniqueSharedElements[key]; ok {
					continue
				}
				sharedElements = sharedElements + key
				uniqueSharedElements[key] = 0
			}
		}
	}

	// Output free constraints
	for _, c := range g.freeEdges.Contents() {
		constraint, _ := g.constraintAccessor.GetConstraint(c)
		e1Cluster, e1HasCluster := g.elementAccessor.Cluster(constraint.Element1)
		e2Cluster, e2HasCluster := g.elementAccessor.Cluster(constraint.Element2)
		if !e1HasCluster || !e2HasCluster {
			edges = edges + constraint.ToGraphViz(-1, -1)
		} else {
			edges = edges + constraint.ToGraphViz(int(e1Cluster), int(e2Cluster))
		}
	}

	// Output free elements
	freeNodes := utils.NewSet()
	for _, id := range g.elementAccessor.IdSet().Contents() {
		freeNodes.Add(id)
	}
	freeNodes = freeNodes.Difference(g.usedNodes)
	for _, eId := range freeNodes.Contents() {
		element, _ := g.elementAccessor.GetElement(-1, eId)
		edges = edges + element.ToGraphViz(-1)
	}

	return fmt.Sprintf(`
	graph {
		compound=true
		%s
		%s
	}`, edges, sharedElements)
}
