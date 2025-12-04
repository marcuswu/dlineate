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
		Int("total elements", elements.Count()).
		Int("total constraints", constraints.Count()).
		Msg("Numeric merge: total elements and constraints")

	// Build and solve numeric solver
	numericSolver := numeric.NewSolver()

	elementList := elements.Contents()
	for _, eId := range elementList {
		element, _ := g.elementAccessor.GetElement(-1, eId)
		numericSolver.AddElement(el.CopySketchElement(element))
	}

	for _, cId := range constraints.Contents() {
		constraint, _ := g.constraintAccessor.GetConstraint(cId)
		numericSolver.AddConstraint(constraint)
	}

	solved := numericSolver.Solve(utils.FloatPrecision, utils.MaxNumericIterations)

	utils.Logger.Debug().
		Bool("solved", solved).
		Msg("Numeric merge: solver completed")

	if !solved {
		return solver.NonConvergent
	}

	// for each cluster, find two elements from our solve and translate the cluster accordingly
	for _, c := range g.clusters {
		sharedElements := c.elements.Intersect(elements)
		if sharedElements.Count() < 2 {
			utils.Logger.Debug().
				Int("cluster id", c.id).
				Int("shared elements", sharedElements.Count()).
				Msg("Not enough shared elements for numeric merge")
			return solver.NonConvergent
		}
		sharedElementsList := sharedElements.Contents()
		e1Local, _ := g.elementAccessor.GetElement(c.GetID(), sharedElementsList[0])
		e1Other, _ := numericSolver.GetElement(sharedElementsList[0])
		i := 1
		e2Local, _ := g.elementAccessor.GetElement(c.GetID(), sharedElementsList[i])
		// Make sure that we don't get two lines
		for ; (e1Local.GetType() != el.Point && e2Local.GetType() != el.Point) && i < len(sharedElementsList); i++ {
			e2Local, _ = g.elementAccessor.GetElement(c.GetID(), sharedElementsList[i])
		}
		if e1Local.GetType() == el.Line && e2Local.GetType() == el.Line {
			return solver.NonConvergent
		}
		e2Other, _ := numericSolver.GetElement(sharedElementsList[i])
		solveState := c.translateToElements(g.elementAccessor, e1Local, e1Other, e2Local, e2Other)
		g.elementAccessor.MergeToRoot(c.GetID())
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
