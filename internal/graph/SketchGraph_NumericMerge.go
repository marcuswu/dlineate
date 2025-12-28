package graph

import (
	"github.com/marcuswu/dlineate/internal/accessors"
	el "github.com/marcuswu/dlineate/internal/element"
	"github.com/marcuswu/dlineate/internal/numeric"
	"github.com/marcuswu/dlineate/internal/solver"
	"github.com/marcuswu/dlineate/utils"
)

func (g *SketchGraph) numericMerge() solver.SolveState {
	// 1. Gather all relevant elements and constraints in the graph
	//   a. Elements shared by multiple clusters
	//   b. Constraints shared by multiple clusters (i.e. free constraints)
	//   c. Constraints connecting used elements in each cluster (to ensure rigidity)
	//   d. Create constraints if necessary to fix elements in each cluster
	//      (are there cases where this is needed?)
	// 2. Add all relevant elements and constraints to the numeric solver
	// 3. Solve the numeric solver
	// 4. If solved, use two elements in each cluster to determine transforms for each cluster
	// 5. Apply transforms to each cluster to merge them together

	elements := g.sharedElements()
	utils.Logger.Debug().
		Int("shared elements", elements.Count()).
		Msg("Numeric merge: gathered shared elements")

	constraints := g.freeEdges
	utils.Logger.Debug().
		Int("free constraints", constraints.Count()).
		Msg("Numeric merge: gathered free constraints")

	for _, c := range constraints.Contents() {
		constraint, _ := g.constraintAccessor.GetConstraint(c)
		elements.Add(constraint.Element1)
		elements.Add(constraint.Element2)
	}

	// Ensure we have constraints within each cluster to make them rigid
	g.EnsureRigid(elements, constraints, g.constraintAccessor)

	utils.Logger.Debug().
		Uints("Id", elements.Contents()).
		Msg("Elements in numeric solver")
	utils.Logger.Debug().
		Msg("Constraints in numeric solver:")
	for _, c := range constraints.Contents() {
		constraint, _ := g.constraintAccessor.GetConstraint(c)
		utils.Logger.Debug().
			Str("constraint", constraint.String()).
			Msg("Constraint")
	}

	utils.Logger.Debug().
		Int("total elements", elements.Count()).
		Int("total constraints", constraints.Count()).
		Msg("Numeric merge: total elements and constraints")

	// Build and solve numeric solver
	numericSolver := numeric.NewSolver()

	elementList := elements.Contents()
	for _, eId := range elementList {
		cId, ok := g.elementAccessor.Cluster(eId)
		cluster := int(cId)
		if !ok {
			cluster = -1
		}
		element, _ := g.elementAccessor.GetElement(cluster, eId)
		var solveElement el.SketchElement
		if element.GetType() == el.Line {
			l := element.AsLine()
			solveElement = numeric.NewSegmentFromLine(element.AsLine())
			solverConstraints := constraints.Contents()
			for _, cId := range solverConstraints {
				constraint, _ := g.constraintAccessor.GetConstraint(cId)
				// Remove constraints on lines that reference their points
				// This is intrinsic to segments
				if constraint.HasElementID(eId) &&
					(constraint.HasElementID(l.Start.GetID()) ||
						constraint.HasElementID(l.End.GetID())) {
					constraints.Remove(cId)
				}
			}
		} else {
			solveElement = el.CopySketchElement(element)
		}
		numericSolver.AddElement(solveElement)
	}

	for _, cId := range constraints.Contents() {
		constraint, _ := g.constraintAccessor.GetConstraint(cId)
		numericSolver.AddConstraint(constraint)
	}

	solved := numericSolver.Solve(utils.StandardCompare, utils.MaxNumericIterations)

	utils.Logger.Debug().
		Bool("solved", solved).
		Msg("Numeric merge: solver completed")

	if !solved {
		return solver.NonConvergent
	}

	// for each cluster, find two elements from our solve and translate the cluster accordingly
	for len(g.clusters) > 0 {
		c := g.clusters[len(g.clusters)-1]
		sharedElements := c.elements.Intersect(elements)
		if sharedElements.Count() < 2 {
			utils.Logger.Debug().
				Int("cluster id", c.id).
				Int("shared elements", sharedElements.Count()).
				Msg("Not enough shared elements for numeric merge")
			return solver.NonConvergent
		}
		sharedElementsList := sharedElements.Contents()
		// If the first element is a line, use its two points
		e1Local, _ := g.elementAccessor.GetElement(c.GetID(), sharedElementsList[0])
		var e2Local el.SketchElement = nil
		if e1Local.GetType() == el.Line {
			e2Local = e1Local.AsLine().End
			e1Local = e1Local.AsLine().Start
		}
		if e2Local == nil {
			e2Local, _ = g.elementAccessor.GetElement(c.GetID(), sharedElementsList[1])
			// If first element was a point and second element is a line, use its two points instead
			if e2Local.GetType() == el.Line {
				e1Local = e2Local.AsLine().Start
				e2Local = e2Local.AsLine().End
			}
		}
		e1Other, _ := numericSolver.GetElement(e1Local.GetID())
		e2Other, _ := numericSolver.GetElement(e2Local.GetID())
		// Should now have two points to translate and rotate to
		solveState := c.translateToElements(g.elementAccessor, e1Local, e1Other, e2Local, e2Other)
		for _, eId := range c.elements.Contents() {
			el, _ := g.elementAccessor.GetElement(c.GetID(), eId)
			g.elementAccessor.ReplaceElement(-1, eId, el)
		}
		g.removeCluster(c.GetID())
		if solveState != solver.Solved {
			return solveState
		}
	}

	return solver.Solved
}

func (g *SketchGraph) sharedElements() *utils.Set {
	shared := utils.NewSet()
	for _, c := range g.clusters {
		for _, other := range g.clusters {
			if c.id == other.id {
				continue
			}
			shared.AddSet(c.SharedElements(other))
			if shared.Count() == 0 {
				continue
			}
		}
	}
	return shared
}

// Ensure we have constraints within each cluster to make them rigid
func (g *SketchGraph) EnsureRigid(elements *utils.Set, constraints *utils.Set, ca accessors.ConstraintAccessor) {
	// For each cluster, ensure that every elements in 'elements' has a constraint between it and the others
	// Build a map of existing constraints for quick lookup
	for _, c := range g.clusters {
		clusterElements := c.elements.Intersect(elements).Contents()
		for i, eId1 := range clusterElements {
			for j := i + 1; j < len(clusterElements); j++ {
				eId2 := clusterElements[j]

				// Ensure that there is a constraint between these two elements
				hasConstraint := false
				eId1Constraints := ca.ConstraintsForElement(eId1)
				for _, constraint := range eId1Constraints {
					if (constraint.Element1 == eId1 && constraint.Element2 == eId2) ||
						(constraint.Element1 == eId2 && constraint.Element2 == eId1) {
						utils.Logger.Debug().
							Uint("constraint id", constraint.GetID()).
							Uint("element 1", eId1).
							Uint("element 2", eId2).
							Str("value", constraint.GetValue().String()).
							Msg("Found constraint")
						constraints.Add(constraint.GetID())
						hasConstraint = true
						break
					}
				}
				if !hasConstraint {
					// Create a constraint between these two elements
					newConstraint := c.mergeConstraint(g.elementAccessor, eId1, eId2)
					constraints.Add(newConstraint.GetID())
					ca.AddConstraint(newConstraint)
					utils.Logger.Debug().
						Int("cluster id", c.id).
						Uint("element 1", eId1).
						Uint("element 2", eId2).
						Str("value", newConstraint.GetValue().String()).
						Msg("Added constraint")
				}
			}
		}
	}
}
