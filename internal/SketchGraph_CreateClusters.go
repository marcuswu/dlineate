package core

import (
	"sort"

	"github.com/marcuswu/dlineate/internal/constraint"
	el "github.com/marcuswu/dlineate/internal/element"
	iutils "github.com/marcuswu/dlineate/internal/utils"
	"github.com/marcuswu/dlineate/utils"
	"github.com/rs/zerolog"
)

func (g *SketchGraph) BuildClusters() {
	g.logConstraintsElements(zerolog.InfoLevel)
	if len(g.clusters) == 0 {
		g.createClusters()
	}
	g.buildFreeEdgeMap()
}

func (g *SketchGraph) buildFreeEdgeMap() {
	for _, c := range g.freeEdges.Contents() {
		constraint, _ := g.constraintAccessor.GetConstraint(c)
		e1Cluster, e1HasCluster := g.elementAccessor.Cluster(constraint.Element1)
		e2Cluster, e2HasCluster := g.elementAccessor.Cluster(constraint.Element2)
		if e1HasCluster && e2HasCluster {
			s, ok := g.freeEdgeMap[e1Cluster]
			if !ok {
				s = utils.NewSet()
				g.freeEdgeMap[e1Cluster] = s
			}
			s.Add(c)
			s, ok = g.freeEdgeMap[e2Cluster]
			if !ok {
				s = utils.NewSet()
				g.freeEdgeMap[e2Cluster] = s
			}
			s.Add(c)
		}
	}
}

func (g *SketchGraph) createClusters() {
	id := 0
	utils.Logger.Info().
		Int("unassigned constraints", g.freeEdges.Count()).
		Msg("Creating clusters")
	// iterations := 0
	skipConstraints := utils.NewSet()
	for g.usedNodes.Count() < g.elementAccessor.Count() {
		// iterations++
		// if iterations > 10 {
		// 	break
		// }
		utils.Logger.Info().
			Str("assigned nodes", g.usedNodes.String()).
			Msg("Creating a cluster")
		// for g.freeEdges.Count() > 0 {
		startFrom := g.findStartConstraint(skipConstraints)
		cluster := g.createCluster(startFrom, id)
		if cluster == nil {
			/*g.state = solver.NonConvergent
			break*/
			skipConstraints.Add(startFrom)
			continue
		}
		id++
		utils.Logger.Info().Str("free constraints", g.freeEdges.String()).Msgf("%d unassigned constraints left\n", g.freeEdges.Count())
		utils.Logger.Debug().Msgf("Total of %d nodes", g.elementAccessor.Count())
		utils.Logger.Debug().Msgf("Total of %d used nodes", g.usedNodes.Count())
	}
	utils.Logger.Info().
		Int("unassigned constraints", g.freeEdges.Count()).
		Msg("Finished Creating clusters")
}

// findStartConstraint finds a constraint to start a cluster with. The strategy is
// in order of precedence
//  1. find a constraint where both elements are in other clusters, connecting
//     those clusters with the cluster built from that constraint.
//  2. find a constraint where one element is in another cluster in the hopes that
//     a future third cluster will connect the existing one and the one about to
//     be created.
func (g *SketchGraph) findStartConstraint(skip *utils.Set) uint {
	constraints := make([]uint, 0)
	for _, constraintId := range g.freeEdges.Contents() {
		if skip.Contains(constraintId) {
			continue
		}
		constraint, ok := g.constraintAccessor.GetConstraint(constraintId)
		if !ok {
			continue
		}
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
	var retVal uint = 0
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
	if retVal == 0 && (len(constraints) == 0 || constraints[0] > 0) {
		if g.freeEdges.Count() > 0 {
			return g.freeEdges.Contents()[0]
		}
	}

	return retVal
}

// Creates a cluster with two parameters: an id and an initial constraint (first).
//  1. Start with an initial constraint, add those elements
//  2. Next, find two constraints connected to one element where each of them
//     is connected to an element in our new cluster
//  3. Continue doing that until no further constraints can be added
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
			cidSet := utils.NewSet()
			cidSet.AddList(cIds)
			level = el.OverConstrained
			// These constraints are conflicting, add them to the conflicting list
			utils.Logger.Error().
				Int("cluster", clusterNum).
				Str("constraints", cidSet.String()).
				Msgf("createCluster(%d): found conflicting constraints", clusterNum)
			g.conflicting.AddList(cIds)
		}
		g.elementAccessor.SetConstraintLevel(eId, level)
		c.AddElement(eId)
		for _, cId := range cIds[:2] {
			utils.Logger.Debug().
				Int("cluster", clusterNum).
				Uint("constraint", cId).
				Msgf("createCluster(%d): adding constraint", clusterNum)
			oc, _ = g.GetConstraint(cId)
			g.addConstraintToCluster(c, oc)
		}
	}

	// We were unable to add anything other than the initial constraint
	if len(c.constraints) == 1 {
		utils.Logger.Debug().Msgf("createCluster(%d): cancelling cluster", clusterNum)
		g.cancelCluster(c)
		return nil
	}

	utils.Logger.Info().
		Int("cluster", clusterNum).
		Int("element count", c.elements.Count()).
		Int("constraint count", len(c.constraints)).
		Msgf("createCluster(%d) completed building cluster", clusterNum)
	g.clusters = append(g.clusters, c)

	return c
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

func (g *SketchGraph) cancelCluster(c *GraphCluster) {
	for _, e := range c.elements.Contents() {
		if !g.elementAccessor.IsShared(e) {
			g.usedNodes.Remove(e)
		}
	}
	g.elementAccessor.MergeToRoot(c.GetID())
	for _, c := range c.constraints {
		g.freeEdges.Add(c)
	}
}

// findConstraints finds two constraints to add to the specified cluster.
// Returns the constraints, the shared element, and whether or not the algorithm was successful
func (g *SketchGraph) findConstraints(c *GraphCluster) ([]uint, uint, bool) {
	// First, find free constraints connected to the cluster, grouped by an element not in the cluster
	constraints := make(map[uint]*utils.Set)
	for _, cId := range g.freeEdges.Contents() {
		constraint, ok := g.constraintAccessor.GetConstraint(cId)
		if !ok {
			continue
		}
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
