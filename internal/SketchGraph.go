package core

import (
	"fmt"
	"sort"

	"github.com/marcuswu/dlineate/internal/accessors"
	"github.com/marcuswu/dlineate/internal/constraint"
	el "github.com/marcuswu/dlineate/internal/element"
	"github.com/marcuswu/dlineate/internal/solver"
	iutils "github.com/marcuswu/dlineate/internal/utils"
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

func (g *SketchGraph) isShared(cId int, eId uint) bool {
	return g.elementAccessor.IsShared(eId)
}

// AddPoint adds a point to the sketch
func (g *SketchGraph) AddPoint(x float64, y float64) el.SketchElement {
	elementID := g.elementAccessor.NextId()
	utils.Logger.Debug().
		Float64("X", x).
		Float64("Y", y).
		Uint("id", elementID).
		Msg("Adding point")
	p := el.NewSketchPoint(elementID, x, y)
	// g.freeNodes.Add(elementID)
	g.elementAccessor.AddElement(p)
	return p
}

// AddLine adds a line to the sketch
func (g *SketchGraph) AddLine(a float64, b float64, c float64) el.SketchElement {
	elementID := g.elementAccessor.NextId()
	utils.Logger.Debug().
		Float64("A", a).
		Float64("B", b).
		Float64("C", c).
		Uint("id", elementID).
		Msg("Adding line")
	l := el.NewSketchLine(elementID, a, b, c)
	// g.freeNodes.Add(elementID)
	g.elementAccessor.AddElement(l)
	return l
}

func (g *SketchGraph) AddOrigin(x float64, y float64) el.SketchElement {
	elementID := g.elementAccessor.NextId()
	utils.Logger.Debug().
		Float64("X", x).
		Float64("Y", y).
		Uint("id", elementID).
		Msgf("Adding origin")
	ax := el.NewSketchPoint(elementID, x, y)
	// g.freeNodes.Add(elementID)
	g.elementAccessor.AddElement(ax)

	g.MakeFixed(ax)

	return ax
}

func (g *SketchGraph) AddAxis(a float64, b float64, c float64) el.SketchElement {
	elementID := g.elementAccessor.NextId()
	utils.Logger.Debug().
		Float64("A", a).
		Float64("B", b).
		Float64("C", c).
		Uint("id", elementID).
		Msg("Adding axis")
	ax := el.NewSketchLine(elementID, a, b, c)
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
func (g *SketchGraph) AddConstraint(t constraint.Type, e1 el.SketchElement, e2 el.SketchElement, value float64) *constraint.Constraint {
	constraintID := g.constraintAccessor.NextId()
	cType := "Distance"
	if t != constraint.Distance {
		cType = "Angle"
	}
	utils.Logger.Debug().
		Str("type", cType).
		Float64("value", value).
		Uint("constraint id", constraintID).
		Msg("Adding constraint")
	constraint := constraint.NewConstraint(constraintID, t, e1.GetID(), e2.GetID(), value, e1.IsFixed() && e2.IsFixed())
	g.constraintAccessor.AddConstraint(constraint)
	g.freeEdges.Add(constraint.GetID())
	return constraint
}

func (g *SketchGraph) addElementToCluster(c *GraphCluster, e uint) {
	c.AddElement(e)
	g.elementAccessor.AddElementToCluster(e, c.GetID())
	g.usedNodes.Add(e)
}

func (g *SketchGraph) addConstraintToCluster(c *GraphCluster, constraint *constraint.Constraint) {
	g.addElementToCluster(c, constraint.Element1)
	g.addElementToCluster(c, constraint.Element2)
	c.AddConstraint(constraint)
	g.freeEdges.Remove(constraint.GetID())
}

// Creates a cluster with an id and an initial constraint.
// Start with an initial constraint (first), add those elements
// Next, find two constraints connected to one element where each of those
// constraints is connected to an element in our new cluster
// Continue doing that until no further constraints can be added
func (g *SketchGraph) createCluster(first uint, id int) *GraphCluster {
	c := NewGraphCluster(id)

	// Add elements connected to other elements in the cluster by two constraints
	clusterNum := len(g.clusters)
	oc, ok := g.GetConstraint(first)
	if !ok {
		utils.Logger.Error().
			Int("cluster", clusterNum).
			Uint("constraint", first).
			Msgf("createCluster(%d): Failed to find initial constraint", clusterNum)
		return nil
	}
	utils.Logger.Debug().
		Uint("first constraint", first).
		Msgf("createCluster(%d): starting", clusterNum)
	g.addConstraintToCluster(c, oc)

	// find a pair of free constraints which is connected to an element (might be in another cluster)
	// and each constraint shares an element with the cluster we're creating
	for cIds, eId, ok := g.findConstraints(c); ok; cIds, eId, ok = g.findConstraints(c) {
		utils.Logger.Debug().
			Int("cluster", clusterNum).
			Uint("element", eId).
			Int("constraint count", len(cIds)).
			Bool("found ok", ok).
			Msgf("createCluster(%d): adding element", clusterNum)
		level := el.FullyConstrained
		if len(cIds) > 2 {
			level = el.OverConstrained
			// These constraints are conflicting, add them to the conflicting list
			g.conflicting.AddList(cIds)
		}
		g.elementAccessor.SetConstraintLevel(eId, level)
		element, _ := g.elementAccessor.GetElement(-1, eId)
		c.AddElement(element.GetID())
		for _, cId := range cIds[:2] {
			utils.Logger.Debug().
				Int("cluster", clusterNum).
				Uint("constraint", cId).
				Msgf("createCluster(%d): adding constraint", clusterNum)
			oc, _ = g.GetConstraint(cId)
			g.addConstraintToCluster(c, oc)
		}
	}

	utils.Logger.Info().
		Int("cluster", clusterNum).
		Int("element count", c.elements.Count()).
		Int("constraint count", len(c.constraints)).
		Msgf("createCluster(%d) completed building cluster", clusterNum)
	g.clusters = append(g.clusters, c)

	return c
}

func (g *SketchGraph) findConstraints(c *GraphCluster) ([]uint, uint, bool) {
	// First, find free constraints connected to the cluster, grouped by an element not in the cluster
	constraints := make(map[uint]*utils.Set)
	for _, cId := range g.freeEdges.Contents() {
		constraint, _ := g.constraintAccessor.GetConstraint(cId)
		if c.HasConstraint(constraint.GetID()) {
			continue
		}
		// Skip constraints with no connection to the cluster
		if !c.HasElement(constraint.First()) && !c.HasElement(constraint.Second()) {
			continue
		}
		// Skip constraints completely contained in the cluster
		// This means the constraint would over-define the cluster
		if c.HasElement(constraint.First()) && c.HasElement(constraint.Second()) {
			utils.Logger.Warn().
				Int("cluster", c.GetID()).
				Uint("element 1", constraint.Element1).
				Uint("element 2", constraint.Element1).
				Msgf("findConstraints(%d) found constraint %d with elements %d and %d already in the cluster where the constraint is not",
					c.GetID(), constraint.GetID(), constraint.Element1, constraint.Element2)
			g.freeEdges.Remove(constraint.GetID())
			g.conflicting.Add(constraint.GetID())
			continue
		}
		other := constraint.First()
		if c.HasElement(other) {
			other = constraint.Second()
		}

		if _, ok := constraints[other]; !ok {
			constraints[other] = utils.NewSet()
		}
		// fmt.Printf("findConstraints: element %d adding constraint %d\n", other, constraint.GetID())
		constraints[other].Add(constraint.GetID())
	}

	var element uint
	var targetConstraints []uint
	for eId, cList := range constraints {
		// Element needs to be connected to the cluster by two constraints
		if cList.Count() < 2 {
			continue
		}
		if cList.Count() == 2 {
			return cList.Contents(), eId, true
		}
		// Consider more than 2 constraints after
		if cList.Count() < len(targetConstraints) {
			continue
		}

		element = eId
		targetConstraints = cList.Contents()
	}

	return targetConstraints, element, len(targetConstraints) > 1
}

// findStartConstraint finds a constraint to start a cluster with. The strategy is
// in order of precedence
//  1. find a constraint where both elements are in other clusters, connecting
//     those clusters with the cluster built from that constraint.
//  2. find a constraint where one element is in another cluster in the hopes that
//     a future third cluster will connect the existing one and the one about to
//     be created.
func (g *SketchGraph) findStartConstraint() uint {
	constraints := make([]uint, 0)
	for _, constraintId := range g.freeEdges.Contents() {
		constraint, _ := g.constraintAccessor.GetConstraint(constraintId)
		// If the constraint's elements are both fixed, use this as a start
		if g.elementAccessor.IsFixed(constraint.Element1) && g.elementAccessor.IsFixed(constraint.Element2) {
			return constraintId
		}
		// If we have a constraint where both elements are used, but the constraint
		// is free, it means each of those elements are in a different cluster.
		// Use this constraint as a start to tie those clusters together.
		if g.usedNodes.Contains(constraint.Element1) &&
			g.usedNodes.Contains(constraint.Element2) {
			return constraintId
		}

		if g.usedNodes.Contains(constraint.Element1) ||
			g.usedNodes.Contains(constraint.Element2) {
			constraints = append(constraints, constraintId)
		}
	}

	// Check unused elements in constraints for highest constraint count
	var retVal uint
	ccount := 0
	sort.Sort(iutils.IdList(constraints))
	for _, constraintId := range constraints {
		constraint, _ := g.constraintAccessor.GetConstraint(constraintId)

		eId := constraint.Element1
		if g.usedNodes.Contains(eId) {
			eId = constraint.Element2
		}
		if len(g.constraintAccessor.ConstraintsForElement(eId)) < ccount {
			continue
		}

		retVal = constraintId
		ccount = len(g.constraintAccessor.ConstraintsForElement(eId))
	}

	return retVal
}

func (g *SketchGraph) createClusters() {
	id := 1
	utils.Logger.Info().
		Int("unassigned constraints", g.freeEdges.Count()).
		Msg("Creating clusters")
	for g.freeEdges.Count() > 0 {
		g.createCluster(g.findStartConstraint(), id)
		id++
		utils.Logger.Info().Str("free constraints", g.freeEdges.String()).Msgf("%d unassigned constraints left\n", g.freeEdges.Count())
		utils.Logger.Debug().Msgf("Total of %d nodes", g.elementAccessor.Count())
		utils.Logger.Debug().Msgf("Total of %d used nodes", g.usedNodes.Count())
	}
	utils.Logger.Info().
		Int("unassigned constraints", g.freeEdges.Count()).
		Msg("Finished Creating clusters")
}

func (g *SketchGraph) logConstraintsElements(level zerolog.Level) {
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

	for _, cId := range g.constraintAccessor.IdSet().Contents() {
		c, _ := g.constraintAccessor.GetConstraint(cId)
		c.Solved = false
		if g.elementAccessor.IsFixed(c.Element1) && g.elementAccessor.IsFixed(c.Element2) {
			c.Solved = true
		}
	}
}

func (g *SketchGraph) BuildClusters() {
	g.logConstraintsElements(zerolog.InfoLevel)
	if len(g.clusters) == 0 {
		g.createClusters()
	}
}

func (g *SketchGraph) Conflicting() *utils.Set {
	return g.conflicting
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
		edges = edges + constraint.ToGraphViz(-1)
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
	if len(g.clusters) == 1 {
		// Drop cluster 0's elements into the main graph
		g.elementAccessor.MergeToRoot(g.clusters[0].GetID())
	}
	return g.state
}

func (g *SketchGraph) mergeClusters() {
	removeCluster := func(g *SketchGraph, cIndex int) {
		last := len(g.clusters) - 1
		g.clusters[cIndex], g.clusters[last] = g.clusters[last], g.clusters[cIndex]
		g.clusters = g.clusters[:last]
	}
	// TODO: is this the best way? are there problems finding merges this way?
	for first, second, third := g.findMerge(); first > 0 && second > 0; first, second, third = g.findMerge() {
		utils.Logger.Debug().
			Int("first cluster", first).
			Int("second cluster", second).
			Int("third cluster", third).
			Msg("Found merge")
		c1 := g.clusters[first]
		c2 := g.clusters[second]
		var c3 *GraphCluster = nil
		if third > 0 {
			c3 = g.clusters[third]
		}
		mergeState := c1.solveMerge(g.elementAccessor, g.constraintAccessor, c2, c3)
		utils.Logger.Debug().
			Str("state", fmt.Sprintf("%v", mergeState)).
			Msg("Completed merge")
		for _, c := range g.clusters {
			c.IsSolved(g.elementAccessor, g.constraintAccessor)
		}
		if g.state != mergeState && mergeState != solver.Solved {
			utils.Logger.Debug().
				Str("graph state", mergeState.String()).
				Msg("Updating state after cluster merge")
			g.state = mergeState
		}
		// Remove second and third clusters
		// TODO: why remove in a particular order? Perhaps the answer is in removeCluster
		ordered := []int{second, third}
		if second < third {
			ordered[0], ordered[1] = ordered[1], ordered[0]
		}
		removeCluster(g, ordered[0])
		if third > 0 {
			removeCluster(g, ordered[1])
		}
	}
	utils.Logger.Info().Msg("Merging with origin and X & Y axes")
	// We should have the fixed cluster and the solved main cluster
	if len(g.clusters) < 2 {
		return
	}
	// Merge the elements to the fixed cluster -- aligning the sketch with origin and axes
	mergeState := solver.Solved
	if len(g.clusters) < 3 {
		mergeState = g.clusters[0].mergeOne(g.elementAccessor, g.clusters[1], false)
	} else {
		first, second, third := g.findMerge()
		if second == 0 {
			first, second = second, first
		}
		if third == 0 {
			_, third = third, first
		}
		c1 := g.clusters[second]
		c2 := g.clusters[third]
		mergeState = g.clusters[0].solveMerge(g.elementAccessor, g.constraintAccessor, c1, c2)
	}
	if g.state != mergeState && mergeState != solver.Solved {
		utils.Logger.Debug().
			Str("graph state", mergeState.String()).
			Msg("Updating state after cluster merge")
		g.state = mergeState
	}
	utils.Logger.Debug().
		Str("graph state", g.state.String()).
		Msg("Final graph state")

	if !g.IsSolved() {
		g.state = solver.NonConvergent
	}
}

func (g *SketchGraph) findMergeForCluster(c *GraphCluster) (int, int) {
	connectedClusters := func(g *SketchGraph, c *GraphCluster) map[int][]uint {
		connected := make(map[int][]uint)
		for i, other := range g.clusters {
			if other.id == c.id {
				continue
			}
			shared := c.SharedElements(other).Contents()
			if len(shared) < 1 {
				continue
			}
			connected[i] = shared
		}
		return connected
	}

	// These are the clusters connected to c
	// We want to find either:
	// * One cluster connected to c by two elements
	// * Two clusters connected to c by one element each
	//   and each other by one
	connected := connectedClusters(g, c)
	for ci, shared := range connected {
		if ci == 0 {
			continue
		}
		utils.Logger.Debug().
			Int("cluster 1", c.id).
			Int("cluster 2", g.clusters[ci].id).
			Msg("Looking for merge")
		if len(shared) == 2 {
			utils.Logger.Debug().
				Int("cluster", g.clusters[ci].id).
				Msg("Found connected cluster for merge")
			return ci, -1
		}

		if len(shared) == 1 {
			// Find another cluster in connected that is connected to g.clusters[ci]
			for oi, oshared := range connected {
				if oi == 0 || ci == oi || len(oshared) != 1 || oshared[0] == shared[0] {
					continue
				}
				utils.Logger.Debug().
					Int("cluster 0", c.id).
					Int("cluster 1", g.clusters[ci].id).
					Int("cluster 2", g.clusters[oi].id).
					Msg("Testing for valid merge for clusters")
				ciOiShared := g.clusters[ci].SharedElements(g.clusters[oi])
				if ciOiShared.Count() == 1 && !ciOiShared.Contains(shared[0]) && !ciOiShared.Contains(oshared[0]) {
					utils.Logger.Debug().
						Int("cluster 0", c.id).
						Int("cluster 1", g.clusters[ci].id).
						Int("cluster 2", g.clusters[oi].id).
						Msg("Found connected clusters for merge")
					return ci, oi
				}
			}
		}
	}

	return -1, -1
}

// Find and return clusters which can be merged.
// This can either be:
//   - Three clusters each sharing an element with one other
//   - Two clusters sharing two elements with each other
//
// Returns the index(es) of the mergable clusters
func (g *SketchGraph) findMerge() (int, int, int) {
	for i, c := range g.clusters {
		// Merge cluster 0 last manually
		utils.Logger.Debug().
			Int("start cluster", c.id).
			Msg("Looking for merge")
		c1, c2 := g.findMergeForCluster(c)
		if c1 >= 0 {
			return i, c1, c2
		}
	}

	utils.Logger.Debug().Msg("No merge found")
	return -1, -1, -1
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
