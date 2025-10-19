package core

import (
	"fmt"

	"github.com/marcuswu/dlineate/internal/solver"
	"github.com/marcuswu/dlineate/utils"
)

func (g *SketchGraph) mergeClusters() {
	getClusterIndex := func(cId int) int {
		for i, c := range g.clusters {
			if c.GetID() == cId {
				return i
			}
		}
		return -1
	}
	removeCluster := func(g *SketchGraph, cId int) {
		cIndex := getClusterIndex(cId)
		if cIndex < 0 || len(g.clusters) == 0 {
			return
		}
		last := len(g.clusters) - 1
		g.clusters[cIndex], g.clusters[last] = g.clusters[last], g.clusters[cIndex]
		g.clusters = g.clusters[:last]
	}
	for first, second, third := g.findMerge(); len(g.clusters) > 1 && first >= 0 && g.state == solver.Solved; first, second, third = g.findMerge() {
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
		g.elementAccessor.MergeElements(c1.GetID(), c2.GetID())
		removeCluster(g, c2.GetID())
		if third > 0 {
			g.elementAccessor.MergeElements(c1.GetID(), c3.GetID())
			removeCluster(g, c3.GetID())
		}
	}

	utils.Logger.Debug().
		Str("graph state", g.state.String()).
		Int("clusters remaining", len(g.clusters)).
		Msg("Final graph state")

	clusters := g.clusters
	for _, c := range clusters {
		g.elementAccessor.MergeToRoot(c.GetID())
		removeCluster(g, c.GetID())
	}

	if !g.IsSolved() {
		g.state = solver.NonConvergent
	}
}

func (g *SketchGraph) findMergeForCluster(c *GraphCluster) (int, int) {
	connectedClusters := func(g *SketchGraph, c *GraphCluster) map[int][]uint {
		connected := make(map[int][]uint)
		for _, other := range g.clusters {
			if other.id == c.id {
				continue
			}
			shared := c.SharedElements(other).Contents()
			if len(shared) < 1 {
				continue
			}
			connected[other.GetID()] = shared
		}
		return connected
	}
	getCluster := func(cId int) *GraphCluster {
		for _, c := range g.clusters {
			if c.GetID() == cId {
				return c
			}
		}
		return nil
	}

	// These are the clusters connected to c
	// We want to find either:
	// * One cluster connected to c by two elements
	// * Two clusters connected to c by one element each
	//   and each other by one
	connected := connectedClusters(g, c)
	for ci, shared := range connected {
		c2 := getCluster(ci)
		if c2 == nil {
			continue
		}
		utils.Logger.Debug().
			Int("cluster 1", c.id).
			Int("cluster 2", ci).
			Msg("Looking for merge")
		if len(shared) == 2 {
			utils.Logger.Debug().
				Int("cluster", ci).
				Msg("Found connected cluster for merge")
			return ci, -1
		}

		if len(shared) == 1 {
			// Find another cluster in connected that is connected to g.clusters[ci]
			for oi, oshared := range connected {
				c3 := getCluster(oi)
				if c3 == nil || ci == oi || len(oshared) != 1 || oshared[0] == shared[0] {
					continue
				}
				utils.Logger.Debug().
					Int("cluster 0", c.id).
					Int("cluster 1", ci).
					Int("cluster 2", oi).
					Msg("Testing for valid merge for clusters")
				ciOiShared := c2.SharedElements(c3)
				if ciOiShared.Count() == 1 && !ciOiShared.Contains(shared[0]) && !ciOiShared.Contains(oshared[0]) {
					utils.Logger.Debug().
						Int("cluster 0", c.id).
						Int("cluster 1", ci).
						Int("cluster 2", oi).
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
