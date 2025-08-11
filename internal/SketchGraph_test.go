package core

import (
	"math"
	"math/big"
	"testing"

	"github.com/marcuswu/dlineate/internal/constraint"
	el "github.com/marcuswu/dlineate/internal/element"
	"github.com/marcuswu/dlineate/internal/solver"
	"github.com/marcuswu/dlineate/utils"
	"github.com/stretchr/testify/assert"
)

func TestGraphBasics(t *testing.T) {
	sketch := NewSketch()

	origin := sketch.AddOrigin(big.NewFloat(0), big.NewFloat(0))
	xAxis := sketch.AddAxis(big.NewFloat(0), big.NewFloat(1), big.NewFloat(0))
	yAxis := sketch.AddAxis(big.NewFloat(1), big.NewFloat(0), big.NewFloat(0))
	p1 := sketch.AddPoint(big.NewFloat(1), big.NewFloat(0))
	l1 := sketch.AddLine(big.NewFloat(1), big.NewFloat(1), big.NewFloat(1))
	c1 := sketch.AddConstraint(constraint.Distance, p1, l1, big.NewFloat(2))
	c2 := sketch.AddConstraint(constraint.Angle, xAxis, yAxis, big.NewFloat(math.Pi/2))
	c3 := sketch.AddConstraint(constraint.Distance, p1, xAxis, big.NewFloat(0))

	_, ok := sketch.GetElement(p1.GetID())
	assert.True(t, ok, "Should be able to find element p1")

	_, ok = sketch.GetConstraint(c1.GetID())
	assert.True(t, ok, "Should be able to find constraint c1")

	cList := sketch.constraintAccessor.ConstraintsForElement(origin.GetID())
	assert.Zero(t, len(cList), "constraint list for origin should be empty")

	c2.Solved = true
	c3.Solved = true
	assert.False(t, sketch.IsElementSolved(origin), "origin should not be solved")
	assert.True(t, sketch.IsElementSolved(xAxis), "xAxis should be solved")
}

func TestCombinePoints(t *testing.T) {
	sketch := NewSketch()
	e1 := sketch.AddPoint(big.NewFloat(1), big.NewFloat(0))
	e2 := sketch.AddOrigin(big.NewFloat(0), big.NewFloat(0))
	newE := sketch.CombinePoints(e1, e2)

	assert.Equal(t, e2.GetID(), newE.GetID(), "Combining points with origin should keep origin's id")
	_, ok := sketch.elementAccessor.GetElement(-1, e1.GetID())
	assert.False(t, ok, "The eliminated element should no longer exist in the sketch")

	sketch = NewSketch()
	e2 = sketch.AddOrigin(big.NewFloat(0), big.NewFloat(0))
	e3 := sketch.AddPoint(big.NewFloat(1), big.NewFloat(0))
	newE = sketch.CombinePoints(e2, e3)
	assert.Equal(t, e2.GetID(), newE.GetID(), "Ensure kept element id")
	_, ok = sketch.elementAccessor.GetElement(-1, e3.GetID())
	assert.False(t, ok, "The eliminated element should no longer exist in the sketch")

	sketch = NewSketch()
	e2 = sketch.AddPoint(big.NewFloat(0), big.NewFloat(0))
	e3 = sketch.AddPoint(big.NewFloat(1), big.NewFloat(0))
	c1 := sketch.AddConstraint(constraint.Distance, e2, e1, big.NewFloat(1))
	c2 := sketch.AddConstraint(constraint.Distance, e1, e2, big.NewFloat(1))
	newE = sketch.CombinePoints(e3, e2)
	assert.Equal(t, e3.GetID(), newE.GetID(), "Ensure kept element id")
	_, ok = sketch.elementAccessor.GetElement(-1, e2.GetID())
	assert.False(t, ok, "The eliminated element should no longer exist in the sketch")
	assert.True(t, c1.HasElementID(newE.GetID()), "constraints should be updated with the new element")
	assert.True(t, c2.HasElementID(newE.GetID()), "constraints should be updated with the new element")

	sketch = NewSketch()
	e2 = sketch.AddOrigin(big.NewFloat(0), big.NewFloat(0))
	e3 = sketch.AddPoint(big.NewFloat(1), big.NewFloat(0))
	c1 = sketch.AddConstraint(constraint.Distance, e2, e1, big.NewFloat(1))
	c2 = sketch.AddConstraint(constraint.Distance, e1, e2, big.NewFloat(1))
	newE = sketch.CombinePoints(e3, e2)
	assert.Equal(t, e2.GetID(), newE.GetID(), "Ensure kept element id")
	_, ok = sketch.elementAccessor.GetElement(-1, e3.GetID())
	assert.False(t, ok, "The eliminated element should no longer exist in the sketch")
	assert.True(t, c1.HasElementID(newE.GetID()), "constraints should be updated with the new element")
	assert.True(t, c2.HasElementID(newE.GetID()), "constraints should be updated with the new element")
}

func TestFindStartConstraint(t *testing.T) {
	sketch := NewSketch()
	e1 := sketch.AddPoint(big.NewFloat(1), big.NewFloat(0))
	e2 := sketch.AddOrigin(big.NewFloat(0), big.NewFloat(0))
	e3 := sketch.AddPoint(big.NewFloat(1), big.NewFloat(0))
	e4 := sketch.AddLine(big.NewFloat(1), big.NewFloat(1), big.NewFloat(1))
	sketch.AddConstraint(constraint.Distance, e2, e1, big.NewFloat(1))
	sketch.AddConstraint(constraint.Distance, e1, e2, big.NewFloat(1))
	sketch.AddConstraint(constraint.Distance, e3, e4, big.NewFloat(1))

	start := sketch.findStartConstraint()
	assert.Contains(t, []uint{0, 1, 2}, start)

	sketch.freeEdges.Remove(e1.GetID())
	sketch.freeEdges.Remove(e2.GetID())

	start = sketch.findStartConstraint()
	assert.Contains(t, []uint{1, 2}, start)
}

func TestFindConstraints(t *testing.T) {
	sketch := NewSketch()
	e1 := sketch.AddPoint(big.NewFloat(1), big.NewFloat(0))
	e2 := sketch.AddOrigin(big.NewFloat(0), big.NewFloat(0))
	e3 := sketch.AddPoint(big.NewFloat(1), big.NewFloat(0))
	e4 := sketch.AddLine(big.NewFloat(1), big.NewFloat(1), big.NewFloat(1))
	e5 := sketch.AddLine(big.NewFloat(1), big.NewFloat(1), big.NewFloat(2))
	_ = sketch.AddConstraint(constraint.Distance, e2, e1, big.NewFloat(1))
	_ = sketch.AddConstraint(constraint.Distance, e5, e4, big.NewFloat(1))
	_ = sketch.AddConstraint(constraint.Distance, e1, e2, big.NewFloat(1))
	_ = sketch.AddConstraint(constraint.Distance, e4, e1, big.NewFloat(1))
	c3 := sketch.AddConstraint(constraint.Distance, e3, e2, big.NewFloat(1))
	c4 := sketch.AddConstraint(constraint.Distance, e1, e3, big.NewFloat(1))

	cluster := NewGraphCluster(1)
	cluster.AddElement(e1.GetID())
	cluster.AddElement(e2.GetID())
	constraints, element, ok := sketch.findConstraints(cluster)
	assert.True(t, ok, "Should find constraints to add to the cluster")
	assert.Equal(t, e3.GetID(), element, "Should find element id for e3")
	assert.Contains(t, constraints, c3.GetID(), "Should contain constraint c3")
	assert.Contains(t, constraints, c4.GetID(), "Should contain constraint c4")

	sketch = NewSketch()
	e1 = sketch.AddPoint(big.NewFloat(1), big.NewFloat(0))
	e2 = sketch.AddOrigin(big.NewFloat(0), big.NewFloat(0))
	e3 = sketch.AddPoint(big.NewFloat(1), big.NewFloat(0))
	e4 = sketch.AddLine(big.NewFloat(1), big.NewFloat(1), big.NewFloat(1))
	e5 = sketch.AddLine(big.NewFloat(1), big.NewFloat(1), big.NewFloat(2))
	e6 := sketch.AddLine(big.NewFloat(1), big.NewFloat(2), big.NewFloat(2))
	_ = sketch.AddConstraint(constraint.Distance, e4, e1, big.NewFloat(1))
	_ = sketch.AddConstraint(constraint.Distance, e4, e2, big.NewFloat(1))
	_ = sketch.AddConstraint(constraint.Distance, e4, e5, big.NewFloat(1))
	_ = sketch.AddConstraint(constraint.Distance, e2, e1, big.NewFloat(1))
	_ = sketch.AddConstraint(constraint.Distance, e1, e2, big.NewFloat(1))
	c3 = sketch.AddConstraint(constraint.Distance, e3, e2, big.NewFloat(1))
	c4 = sketch.AddConstraint(constraint.Distance, e1, e3, big.NewFloat(1))
	c5 := sketch.AddConstraint(constraint.Distance, e3, e5, big.NewFloat(1))
	c6 := sketch.AddConstraint(constraint.Distance, e3, e6, big.NewFloat(1))

	cluster = NewGraphCluster(1)
	cluster.AddElement(e1.GetID())
	cluster.AddElement(e2.GetID())
	cluster.AddElement(e5.GetID())
	cluster.AddElement(e6.GetID())
	constraints, element, ok = sketch.findConstraints(cluster)
	assert.True(t, ok, "Should find constraints to add to the cluster when over constrained")
	assert.Equal(t, e3.GetID(), element, "Should find element id for e3")
	assert.Contains(t, constraints, c3.GetID(), "Should contain constraint c3")
	assert.Contains(t, constraints, c4.GetID(), "Should contain constraint c4")
	assert.Contains(t, constraints, c5.GetID(), "Should contain constraint c5")
	assert.Contains(t, constraints, c6.GetID(), "Should contain constraint c6")
}

func TestAddConstraintToCluster(t *testing.T) {
	sketch := NewSketch()
	e1 := sketch.AddPoint(big.NewFloat(1), big.NewFloat(0))
	e2 := sketch.AddOrigin(big.NewFloat(0), big.NewFloat(0))
	e3 := sketch.AddPoint(big.NewFloat(1), big.NewFloat(0))
	c3 := sketch.AddConstraint(constraint.Distance, e3, e2, big.NewFloat(1))
	c4 := sketch.AddConstraint(constraint.Distance, e1, e3, big.NewFloat(1))

	cluster := NewGraphCluster(1)
	sketch.addConstraintToCluster(cluster, c3)
	sketch.addConstraintToCluster(cluster, c4)
	constraintSet := utils.NewSet()
	constraintSet.AddList(cluster.constraints)
	ok := constraintSet.Contains(c3.GetID())
	assert.True(t, ok, "Cluster has constraint c3")
	ok = constraintSet.Contains(c4.GetID())
	assert.True(t, ok, "Cluster has constraint c4")
}

func TestCreateCluster(t *testing.T) {
	// Fail to find initial constraint
	sketch := NewSketch()
	c := sketch.createCluster(0, 0)
	assert.Nil(t, c, "Return nil when unable to find initial constraint")

	// Overconstrained
	e1 := sketch.AddPoint(big.NewFloat(1), big.NewFloat(0))
	e2 := sketch.AddOrigin(big.NewFloat(0), big.NewFloat(0))
	e3 := sketch.AddPoint(big.NewFloat(1), big.NewFloat(0))
	e4 := sketch.AddLine(big.NewFloat(1), big.NewFloat(1), big.NewFloat(1))
	e5 := sketch.AddLine(big.NewFloat(1), big.NewFloat(1), big.NewFloat(2))
	e6 := sketch.AddLine(big.NewFloat(1), big.NewFloat(2), big.NewFloat(2))
	_ = sketch.AddConstraint(constraint.Distance, e4, e1, big.NewFloat(1))
	_ = sketch.AddConstraint(constraint.Distance, e4, e2, big.NewFloat(1))
	_ = sketch.AddConstraint(constraint.Distance, e4, e5, big.NewFloat(1))
	_ = sketch.AddConstraint(constraint.Distance, e2, e1, big.NewFloat(1))
	_ = sketch.AddConstraint(constraint.Distance, e1, e2, big.NewFloat(1))
	_ = sketch.AddConstraint(constraint.Distance, e3, e2, big.NewFloat(1))
	_ = sketch.AddConstraint(constraint.Distance, e1, e3, big.NewFloat(1))
	_ = sketch.AddConstraint(constraint.Distance, e3, e5, big.NewFloat(1))
	_ = sketch.AddConstraint(constraint.Distance, e3, e6, big.NewFloat(1))

	c = sketch.createCluster(0, 0)
	assert.NotNil(t, c, "Cluster should not be nil")
	assert.Equal(t, 5, c.elements.Count(), "Cluster should have 5 elements")
	assert.Equal(t, 7, len(c.constraints), "Cluster should have 7 constraints")
}

func TestCreateBuildResetClusters(t *testing.T) {
	sketch := NewSketch()

	e0 := sketch.AddAxis(big.NewFloat(0), big.NewFloat(1), big.NewFloat(0))
	e1 := sketch.AddPoint(big.NewFloat(1), big.NewFloat(0))
	e2 := sketch.AddOrigin(big.NewFloat(0), big.NewFloat(0))
	e3 := sketch.AddPoint(big.NewFloat(1), big.NewFloat(0))
	e4 := sketch.AddLine(big.NewFloat(1), big.NewFloat(1), big.NewFloat(1))
	e5 := sketch.AddLine(big.NewFloat(1), big.NewFloat(1), big.NewFloat(2))
	e6 := sketch.AddLine(big.NewFloat(1), big.NewFloat(2), big.NewFloat(2))
	_ = sketch.AddConstraint(constraint.Distance, e0, e2, big.NewFloat(0))
	_ = sketch.AddConstraint(constraint.Distance, e4, e1, big.NewFloat(1))
	_ = sketch.AddConstraint(constraint.Distance, e4, e2, big.NewFloat(1))
	_ = sketch.AddConstraint(constraint.Distance, e4, e5, big.NewFloat(1))
	_ = sketch.AddConstraint(constraint.Distance, e2, e1, big.NewFloat(1))
	_ = sketch.AddConstraint(constraint.Distance, e1, e2, big.NewFloat(1))
	_ = sketch.AddConstraint(constraint.Distance, e3, e2, big.NewFloat(1))
	_ = sketch.AddConstraint(constraint.Distance, e1, e3, big.NewFloat(1))
	_ = sketch.AddConstraint(constraint.Distance, e3, e5, big.NewFloat(1))
	_ = sketch.AddConstraint(constraint.Distance, e3, e6, big.NewFloat(1))
	_ = sketch.AddConstraint(constraint.Distance, e4, e0, big.NewFloat(1))

	sketch.createClusters()

	assert.Equal(t, 2, len(sketch.clusters), "Should have 2 clusters")
	assert.Equal(t, 6, sketch.clusters[0].elements.Count(), "cluster 0 should have 6 element, 9 constraints")
	assert.Equal(t, 9, len(sketch.clusters[0].constraints), "cluster 0 should have 6 element, 9 constraints")

	assert.Equal(t, 2, sketch.clusters[1].elements.Count(), "cluster 1 should have 2 element, 1 constraints")
	assert.Equal(t, 1, len(sketch.clusters[1].constraints), "cluster 1 should have 2 element, 1 constraints")

	sketch.ResetClusters()
	assert.Equal(t, 0, len(sketch.clusters), "Should have 0 clusters")

	sketch.BuildClusters()

	assert.Equal(t, 2, len(sketch.clusters), "Should have 4 clusters")
	assert.Equal(t, 6, sketch.clusters[0].elements.Count(), "cluster 0 should have 6 element, 9 constraints")
	assert.Equal(t, 9, len(sketch.clusters[0].constraints), "cluster 0 should have 6 element, 9 constraints")

	assert.Equal(t, 2, sketch.clusters[1].elements.Count(), "cluster 1 should have 2 element, 1 constraints")
	assert.Equal(t, 1, len(sketch.clusters[1].constraints), "cluster 1 should have 2 element, 1 constraints")

	// assert.Equal(t, 2, sketch.clusters[2].elements.Count(), "cluster 2 should have 2 element, 1 constraints")
	// assert.Equal(t, 1, len(sketch.clusters[2].constraints), "cluster 2 should have 2 element, 1 constraints")

	// assert.Equal(t, 2, sketch.clusters[3].elements.Count(), "cluster 3 should have 2 element, 1 constraints")
	// assert.Equal(t, 1, len(sketch.clusters[3].constraints), "cluster 3 should have 2 element, 1 constraints")
}

func TestSolve(t *testing.T) {
	s := NewSketch()
	origin := s.AddOrigin(big.NewFloat(0), big.NewFloat(0))
	xAxis := s.AddAxis(big.NewFloat(0), big.NewFloat(-1), big.NewFloat(0))
	yAxis := s.AddAxis(big.NewFloat(1), big.NewFloat(0), big.NewFloat(0))
	s.AddConstraint(constraint.Angle, xAxis, yAxis, big.NewFloat(math.Pi/2))
	s.AddConstraint(constraint.Distance, origin, xAxis, big.NewFloat(0))
	s.AddConstraint(constraint.Distance, origin, yAxis, big.NewFloat(0))

	p1 := origin
	p2 := s.AddPoint(big.NewFloat(4), big.NewFloat(0))                // 3
	p3 := s.AddPoint(big.NewFloat(5.236068), big.NewFloat(3.804226))  // 4
	p4 := s.AddPoint(big.NewFloat(2), big.NewFloat(6.155367))         // 5
	p5 := s.AddPoint(big.NewFloat(-1.236068), big.NewFloat(3.804226)) // 6

	l1 := s.AddLine(big.NewFloat(0), big.NewFloat(-1), big.NewFloat(0))                       // 7
	l2 := s.AddLine(big.NewFloat(0.951057), big.NewFloat(-0.309017), big.NewFloat(-3.804226)) // 8
	l3 := s.AddLine(big.NewFloat(0.587785), big.NewFloat(0.809017), big.NewFloat(-6.155367))  // 9
	l4 := s.AddLine(big.NewFloat(-0.587785), big.NewFloat(0.809017), big.NewFloat(-3.804226)) // 10
	l5 := s.AddLine(big.NewFloat(-0.951057), big.NewFloat(-0.309017), big.NewFloat(0))        // 11

	s.AddConstraint(constraint.Distance, l1, p1, big.NewFloat(0))
	s.AddConstraint(constraint.Distance, l1, p2, big.NewFloat(0))
	s.AddConstraint(constraint.Distance, l2, p2, big.NewFloat(0))
	s.AddConstraint(constraint.Distance, l2, p3, big.NewFloat(0))
	s.AddConstraint(constraint.Distance, l3, p3, big.NewFloat(0))
	s.AddConstraint(constraint.Distance, l3, p4, big.NewFloat(0))
	s.AddConstraint(constraint.Distance, l4, p4, big.NewFloat(0))
	s.AddConstraint(constraint.Distance, l4, p5, big.NewFloat(0))
	s.AddConstraint(constraint.Distance, l5, p5, big.NewFloat(0))
	c1 := s.AddConstraint(constraint.Distance, l5, p1, big.NewFloat(0))

	s.AddConstraint(constraint.Angle, l2, l3, big.NewFloat((72.0/180.0)*math.Pi))
	s.AddConstraint(constraint.Angle, l3, l4, big.NewFloat((72.0/180.0)*math.Pi))
	s.AddConstraint(constraint.Angle, l4, l5, big.NewFloat((72.0/180.0)*math.Pi))

	c2 := s.AddConstraint(constraint.Distance, p1, p2, big.NewFloat(4))
	s.AddConstraint(constraint.Distance, p2, p3, big.NewFloat(4))
	s.AddConstraint(constraint.Distance, p3, p4, big.NewFloat(4))
	c3 := s.AddConstraint(constraint.Distance, p4, p5, big.NewFloat(4))

	c4 := s.AddConstraint(constraint.Angle, l1, xAxis, big.NewFloat(0))

	s.ResetClusters()
	s.BuildClusters()
	state := s.Solve()

	assert.Equal(t, solver.Solved, state, "Graph should be solved")

	s.ResetClusters()
	c, _ := s.constraintAccessor.GetConstraint(c1.GetID())
	c.Value.SetFloat64(0)
	c, _ = s.constraintAccessor.GetConstraint(c2.GetID())
	c.Value.SetFloat64(1)
	c, _ = s.constraintAccessor.GetConstraint(c3.GetID())
	c.Value.SetFloat64(8)
	s.BuildClusters()

	state = s.Solve()

	assert.Equal(t, solver.NonConvergent, state, "Graph should be non-convergent")

	s.ResetClusters()
	c, _ = s.constraintAccessor.GetConstraint(c1.GetID())
	c.Value.SetFloat64(0)
	c, _ = s.constraintAccessor.GetConstraint(c2.GetID())
	c.Value.SetFloat64(4)
	c, _ = s.constraintAccessor.GetConstraint(c3.GetID())
	c.Value.SetFloat64(4)
	s.constraintAccessor.RemoveConstraint(c4.GetID())
	s.BuildClusters()

	state = s.Solve()

	assert.Equal(t, solver.NonConvergent, state, "Graph should be non-convergent")
}

func TestFindMergeForCluster(t *testing.T) {
	s := NewSketch()
	origin := s.AddOrigin(big.NewFloat(0), big.NewFloat(0))
	xAxis := s.AddAxis(big.NewFloat(0), big.NewFloat(-1), big.NewFloat(0))
	yAxis := s.AddAxis(big.NewFloat(1), big.NewFloat(0), big.NewFloat(0))
	s.AddConstraint(constraint.Angle, xAxis, yAxis, big.NewFloat(math.Pi/2))
	s.AddConstraint(constraint.Distance, origin, xAxis, big.NewFloat(0))
	s.AddConstraint(constraint.Distance, origin, yAxis, big.NewFloat(0))

	p1 := origin                                          //s.AddPoint(big.NewFloat(0), big.NewFloat( 0))    // 7
	p2 := s.AddPoint(big.NewFloat(3.13), big.NewFloat(0)) // 4
	p3 := s.AddPoint(big.NewFloat(5.14), big.NewFloat(2.27))
	p4 := s.AddPoint(big.NewFloat(2.28), big.NewFloat(4.72))
	p5 := s.AddPoint(big.NewFloat(-1.04), big.NewFloat(3.56))

	l1 := s.AddLine(big.NewFloat(0), big.NewFloat(-1), big.NewFloat(0))
	l2 := s.AddLine(big.NewFloat(2.27), big.NewFloat(-2.01), big.NewFloat(7.1051))
	l3 := s.AddLine(big.NewFloat(2.45), big.NewFloat(2.86), big.NewFloat(19.0852))
	l4 := s.AddLine(big.NewFloat(-1.16), big.NewFloat(3.32), big.NewFloat(13.0256))
	l5 := s.AddLine(big.NewFloat(-3.56), big.NewFloat(-1.04), big.NewFloat(0)) // 12

	s.AddConstraint(constraint.Distance, l1, p1, big.NewFloat(0))
	s.AddConstraint(constraint.Distance, l1, p2, big.NewFloat(0))
	s.AddConstraint(constraint.Distance, l2, p2, big.NewFloat(0))
	s.AddConstraint(constraint.Distance, l2, p3, big.NewFloat(0))
	s.AddConstraint(constraint.Distance, l3, p3, big.NewFloat(0))
	s.AddConstraint(constraint.Distance, l3, p4, big.NewFloat(0))
	s.AddConstraint(constraint.Distance, l4, p4, big.NewFloat(0))
	s.AddConstraint(constraint.Distance, l4, p5, big.NewFloat(0))
	s.AddConstraint(constraint.Distance, l5, p5, big.NewFloat(0))
	s.AddConstraint(constraint.Distance, l5, p1, big.NewFloat(0))

	s.AddConstraint(constraint.Angle, l2, l3, big.NewFloat((72.0/180.0)*math.Pi))
	s.AddConstraint(constraint.Angle, l3, l4, big.NewFloat((72.0/180.0)*math.Pi))
	s.AddConstraint(constraint.Angle, l4, l5, big.NewFloat((72.0/180.0)*math.Pi))

	s.AddConstraint(constraint.Distance, p1, p2, big.NewFloat(4))
	s.AddConstraint(constraint.Distance, p2, p3, big.NewFloat(4))
	s.AddConstraint(constraint.Distance, p3, p4, big.NewFloat(4))
	s.AddConstraint(constraint.Distance, p4, p5, big.NewFloat(4))

	s.AddConstraint(constraint.Angle, l1, xAxis, big.NewFloat(0))

	s.ResetClusters()
	s.BuildClusters()

	for _, c := range s.clusters {
		c.Solve(s.elementAccessor, s.constraintAccessor)
	}
	a, b := s.findMergeForCluster(s.clusters[0])

	assert.Contains(t, []int{1, 2}, a, "First merge cluster is 5 or 6")
	assert.Contains(t, []int{1, 2}, b, "Second merge cluster is 5 or 6")
}

func TestGraphToGraphViz(t *testing.T) {
	s := NewSketch()

	e1 := el.NewSketchPoint(0, big.NewFloat(0), big.NewFloat(1))
	e2 := el.NewSketchPoint(1, big.NewFloat(2), big.NewFloat(1))
	e3 := el.NewSketchLine(2, big.NewFloat(2), big.NewFloat(1), big.NewFloat(-1))
	e4 := el.NewSketchLine(3, big.NewFloat(2), big.NewFloat(2), big.NewFloat(-0))
	c1 := constraint.NewConstraint(0, constraint.Distance, e1.GetID(), e2.GetID(), big.NewFloat(5), false)
	c2 := constraint.NewConstraint(1, constraint.Distance, e2.GetID(), e3.GetID(), big.NewFloat(7), false)
	c3 := constraint.NewConstraint(2, constraint.Distance, e3.GetID(), e4.GetID(), big.NewFloat(2), false)

	g := NewGraphCluster(0)
	s.constraintAccessor.AddConstraint(c1)
	s.constraintAccessor.AddConstraint(c2)
	s.constraintAccessor.AddConstraint(c3)
	s.elementAccessor.AddElement(e1)
	s.elementAccessor.AddElement(e2)
	s.elementAccessor.AddElement(e3)
	s.elementAccessor.AddElement(e4)
	g.AddConstraint(c1)
	g.AddElement(c1.Element1)
	g.AddElement(c1.Element2)
	g.AddConstraint(c2)
	g.AddElement(c2.Element1)
	g.AddElement(c2.Element2)
	g.AddConstraint(c3)
	g.AddElement(c3.Element1)
	g.AddElement(c3.Element2)
	s.clusters = append(s.clusters, g)

	o := NewGraphCluster(1)
	p1 := el.NewSketchPoint(0, big.NewFloat(0), big.NewFloat(0))
	p2 := el.NewSketchPoint(5, big.NewFloat(4), big.NewFloat(0))
	l1 := el.NewSketchLine(6, big.NewFloat(0), big.NewFloat(-1), big.NewFloat(0))
	c4 := constraint.NewConstraint(3, constraint.Distance, l1.GetID(), p1.GetID(), big.NewFloat(0), false)
	c5 := constraint.NewConstraint(4, constraint.Distance, l1.GetID(), p2.GetID(), big.NewFloat(0), false)
	s.constraintAccessor.AddConstraint(c4)
	s.constraintAccessor.AddConstraint(c5)
	s.elementAccessor.AddElement(p1)
	s.elementAccessor.AddElement(p2)
	s.elementAccessor.AddElement(l1)
	o.AddConstraint(c4)
	o.AddElement(c4.Element1)
	o.AddElement(c4.Element2)
	o.AddConstraint(c5)
	o.AddElement(c5.Element1)
	o.AddElement(c5.Element2)
	s.clusters = append(s.clusters, o)

	e5 := el.NewSketchPoint(7, big.NewFloat(1), big.NewFloat(1))
	e6 := el.NewSketchPoint(8, big.NewFloat(2), big.NewFloat(1))
	c6 := constraint.NewConstraint(5, constraint.Distance, e5.GetID(), e6.GetID(), big.NewFloat(1), false)
	s.elementAccessor.AddElement(e5)
	s.elementAccessor.AddElement(e6)
	// s.elements[e5.GetID()] = e5
	// s.freeNodes.Add(e5.GetID())
	// s.elements[e6.GetID()] = e6
	// s.freeNodes.Add(e6.GetID())
	// s.constraints[c6.GetID()] = c6
	s.constraintAccessor.AddConstraint(c6)
	s.freeEdges.Add(c6.GetID())

	gvString := s.ToGraphViz()
	assert.Contains(t, gvString, "subgraph cluster_0")
	assert.Contains(t, gvString, "label = \"Cluster 0\"")
	assert.Contains(t, gvString, c1.ToGraphViz(0), "GraphViz output contains constraint 1")
	assert.Contains(t, gvString, c2.ToGraphViz(0), "GraphViz output contains constraint 2")
	assert.Contains(t, gvString, c3.ToGraphViz(0), "GraphViz output contains constraint 3")
	assert.Contains(t, gvString, e1.ToGraphViz(0), "GraphViz output contains element 1")
	assert.Contains(t, gvString, e2.ToGraphViz(0), "GraphViz output contains element 2")
	assert.Contains(t, gvString, e3.ToGraphViz(0), "GraphViz output contains element 3")
	assert.Contains(t, gvString, e4.ToGraphViz(0), "GraphViz output contains element 4")

	assert.Contains(t, gvString, "subgraph cluster_1")
	assert.Contains(t, gvString, "label = \"Cluster 1\"")
	assert.Contains(t, gvString, c4.ToGraphViz(1), "GraphViz output contains constraint 4")
	assert.Contains(t, gvString, c5.ToGraphViz(1), "GraphViz output contains constraint 5")
	assert.Contains(t, gvString, p1.ToGraphViz(1), "GraphViz output contains point 1")
	assert.Contains(t, gvString, p2.ToGraphViz(1), "GraphViz output contains point 2")
	assert.Contains(t, gvString, l1.ToGraphViz(1), "GraphViz output contains line 1")

	assert.Contains(t, gvString, e5.ToGraphViz(-1), "GraphViz output contains element 5")
	assert.Contains(t, gvString, e6.ToGraphViz(-1), "GraphViz output contains element 6")
	assert.Contains(t, gvString, c6.ToGraphViz(-1), "GraphViz output contains constraint 6")
}
