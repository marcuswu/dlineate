package core

import (
	"fmt"

	"github.com/marcuswu/dlineate/internal/solver"
	"github.com/marcuswu/dlineate/utils"
)

type MergeData struct {
	clusterId1  int
	clusterId2  int
	clusterId3  int
	cluster1    *GraphCluster
	cluster2    *GraphCluster
	cluster3    *GraphCluster
	constraints []uint
	elements    []uint
}

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
	for mergeData := g.findMerge(); len(g.clusters) > 1 && mergeData.clusterId1 >= 0 && g.state == solver.Solved; mergeData = g.findMerge() {
		first, second, third := mergeData.clusterId1, mergeData.clusterId2, mergeData.clusterId3
		utils.Logger.Debug().
			Int("first cluster", first).
			Int("second cluster", second).
			Int("third cluster", third).
			Msg("Found merge")
		mergeData.cluster1 = g.clusters[first]
		mergeData.cluster2 = g.clusters[second]
		if third > 0 {
			mergeData.cluster3 = g.clusters[third]
		}
		c1, c2, c3 := mergeData.cluster1, mergeData.cluster2, mergeData.cluster3
		mergeState := c1.solveMerge(g.elementAccessor, g.constraintAccessor, mergeData)
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

func (g *SketchGraph) findSharedMergeForCluster(c *GraphCluster, sharedMap []map[int][]uint) MergeData {
	getCluster := func(cId int) *GraphCluster {
		for _, c := range g.clusters {
			if c.GetID() == cId {
				return c
			}
		}
		return nil
	}

	// These are the clusters connected to c
	// We want to find one of these possible scenarios:
	// * Two clusters connected to c by one element each
	//   and each other by one item
	// * One cluster connected to c by two elements
	// Where item may be a shared element or a constraint
	// connected := connectedClusters(g, c)
	connected := sharedMap[c.id]
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
			return MergeData{
				clusterId1:  c.id,
				clusterId2:  ci,
				clusterId3:  -1,
				constraints: []uint{},
				elements:    shared,
			}
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
					return MergeData{
						clusterId1:  c.id,
						clusterId2:  ci,
						clusterId3:  oi,
						constraints: []uint{},
						elements:    append(append(shared, oshared...), ciOiShared.Contents()...),
					}
				}
			}
		}
	}

	return MergeData{
		clusterId1:  -1,
		clusterId2:  -1,
		clusterId3:  -1,
		elements:    []uint{},
		constraints: []uint{},
	}
}

/*
// Find mergeable clusters using constraints
//   - Two clusters sharing one element and one constraint
//   - Two clusters with one distance and one angle constraint
func (g *SketchGraph) findConstraintMerge(c *GraphCluster, sharedMap []map[int][]uint) MergeData {
	constraintIds := g.freeEdgeMap[uint(c.id)]
	for clusterId, freeEdges := range g.freeEdgeMap {
		sharedCs := constraintIds.Intersect(freeEdges)
		sharedEs := sharedMap[c.id][int(clusterId)]
		if sharedCs.Count() == 1 && len(sharedEs) == 1 {
			return MergeData{
				clusterId1:  c.id,
				clusterId2:  int(clusterId),
				clusterId3:  -1,
				elements:    sharedEs,
				constraints: sharedCs.Contents(),
			}
		}

		// >= accounts for overconstrained
		if sharedCs.Count() >= 2 {
			return MergeData{
				clusterId1:  c.id,
				clusterId2:  int(clusterId),
				clusterId3:  -1,
				elements:    sharedEs,
				constraints: sharedCs.Contents(),
			}
		}
	}

	return MergeData{
		clusterId1:  -1,
		clusterId2:  -1,
		clusterId3:  -1,
		elements:    []uint{},
		constraints: []uint{},
	}
}
*/

// Find and return clusters which can be merged.
// This can be:
//   - Three clusters each sharing an element with each other
//   - Two clusters sharing two elements with each other
//   - Two clusters sharing one element and any constraint
//   - Two clusters with one distance and one angle constraint
//
// Where item may be a shared element or a constraint
// Returns the indexes of the mergable clusters
// func (g *SketchGraph) findMerge() (int, int, int) {
func (g *SketchGraph) findMerge() MergeData {
	// double for loop up front to do it once per merge
	shared := make([]map[int][]uint, len(g.clusters))
	for _, c := range g.clusters {
		shared[c.id] = make(map[int][]uint)
		for _, other := range g.clusters {
			if other.id == c.id {
				continue
			}
			sharedEs := c.SharedElements(other).Contents()
			if len(sharedEs) < 1 {
				continue
			}
			shared[c.id][other.GetID()] = sharedEs
		}
	}
	for _, c := range g.clusters {
		// utils.Logger.Debug().
		// 	Int("start cluster", c.id).
		// 	Msg("Looking for free constraint merge")
		// mergeData := g.findConstraintMerge(c, shared)
		// if mergeData.clusterId1 >= 0 {
		// 	return mergeData
		// }
		utils.Logger.Debug().
			Int("start cluster", c.id).
			Msg("Looking for shared element merge")
		mergeData := g.findSharedMergeForCluster(c, shared)
		if mergeData.clusterId1 >= 0 {
			return mergeData
		}
	}

	utils.Logger.Debug().Msg("No merge found")
	return MergeData{
		clusterId1:  -1,
		clusterId2:  -1,
		clusterId3:  -1,
		elements:    []uint{},
		constraints: []uint{},
	}
}
