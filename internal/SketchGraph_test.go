package core

import (
	"math"
	"testing"

	"github.com/marcuswu/dlineation/internal/constraint"
	"github.com/marcuswu/dlineation/internal/solver"
	"github.com/stretchr/testify/assert"
)

func TestGraphBasics(t *testing.T) {
	sketch := NewSketch()

	origin := sketch.AddOrigin(0, 0)
	xAxis := sketch.AddAxis(0, 1, 0)
	yAxis := sketch.AddAxis(1, 0, 0)
	p1 := sketch.AddPoint(1, 0)
	l1 := sketch.AddLine(1, 1, 1)
	c1 := sketch.AddConstraint(constraint.Distance, p1, l1, 2)
	c2 := sketch.AddConstraint(constraint.Angle, xAxis, yAxis, math.Pi/2)
	c3 := sketch.AddConstraint(constraint.Distance, p1, xAxis, 0)

	_, ok := sketch.GetElement(p1.GetID())
	assert.True(t, ok, "Should be able to find element p1")

	_, ok = sketch.GetConstraint(c1.GetID())
	assert.True(t, ok, "Should be able to find constraint c1")

	cList := sketch.FindConstraints(origin.GetID())
	assert.Zero(t, len(cList), "constraint list for origin should be empty")

	c2.Solved = true
	c3.Solved = true
	assert.False(t, sketch.IsElementSolved(origin), "origin should not be solved")
	assert.True(t, sketch.IsElementSolved(xAxis), "xAxis should be solved")
}

func TestCombinePoints(t *testing.T) {
	sketch := NewSketch()
	e1 := sketch.AddPoint(1, 0)
	e2 := sketch.AddOrigin(0, 0)
	newE := sketch.CombinePoints(e1, e2)

	assert.Equal(t, e2.GetID(), newE.GetID(), "Combining points with origin should keep origin's id")
	_, ok := sketch.elements[e2.GetID()]
	assert.False(t, ok, "The eliminated element should no longer exist in the sketch")

	sketch = NewSketch()
	e2 = sketch.AddOrigin(0, 0)
	e3 := sketch.AddPoint(1, 0)
	newE = sketch.CombinePoints(e2, e3)
	assert.Equal(t, e2.GetID(), newE.GetID(), "Ensure kept element id")
	_, ok = sketch.elements[e3.GetID()]
	assert.False(t, ok, "The eliminated element should no longer exist in the sketch")

	sketch = NewSketch()
	e2 = sketch.AddPoint(0, 0)
	e3 = sketch.AddPoint(1, 0)
	c1 := sketch.AddConstraint(constraint.Distance, e2, e1, 1)
	c2 := sketch.AddConstraint(constraint.Distance, e1, e2, 1)
	newE = sketch.CombinePoints(e3, e2)
	assert.Equal(t, e3.GetID(), newE.GetID(), "Ensure kept element id")
	_, ok = sketch.elements[e2.GetID()]
	assert.False(t, ok, "The eliminated element should no longer exist in the sketch")
	assert.True(t, c1.HasElementID(newE.GetID()), "constraints should be updated with the new element")
	assert.True(t, c2.HasElementID(newE.GetID()), "constraints should be updated with the new element")

	sketch = NewSketch()
	e2 = sketch.AddOrigin(0, 0)
	e3 = sketch.AddPoint(1, 0)
	c1 = sketch.AddConstraint(constraint.Distance, e2, e1, 1)
	c2 = sketch.AddConstraint(constraint.Distance, e1, e2, 1)
	newE = sketch.CombinePoints(e3, e2)
	assert.Equal(t, e2.GetID(), newE.GetID(), "Ensure kept element id")
	_, ok = sketch.elements[e2.GetID()]
	assert.False(t, ok, "The eliminated element should no longer exist in the sketch")
	assert.True(t, c1.HasElementID(newE.GetID()), "constraints should be updated with the new element")
	assert.True(t, c2.HasElementID(newE.GetID()), "constraints should be updated with the new element")
}

func TestFindStartConstraint(t *testing.T) {
	sketch := NewSketch()
	e1 := sketch.AddPoint(1, 0)
	e2 := sketch.AddOrigin(0, 0)
	e3 := sketch.AddPoint(1, 0)
	e4 := sketch.AddLine(1, 1, 1)
	sketch.AddConstraint(constraint.Distance, e2, e1, 1)
	sketch.AddConstraint(constraint.Distance, e1, e2, 1)
	sketch.AddConstraint(constraint.Distance, e3, e4, 1)

	start := sketch.findStartConstraint()
	assert.Equal(t, uint(1), start)

	sketch.usedNodes.Add(e1.GetID())
	sketch.usedNodes.Add(e2.GetID())

	start = sketch.findStartConstraint()
	assert.Equal(t, uint(0), start)
}

func TestFindConstraints(t *testing.T) {
	sketch := NewSketch()
	e1 := sketch.AddPoint(1, 0)
	e2 := sketch.AddOrigin(0, 0)
	e3 := sketch.AddPoint(1, 0)
	e4 := sketch.AddLine(1, 1, 1)
	e5 := sketch.AddLine(1, 1, 2)
	_ = sketch.AddConstraint(constraint.Distance, e2, e1, 1)
	_ = sketch.AddConstraint(constraint.Distance, e5, e4, 1)
	_ = sketch.AddConstraint(constraint.Distance, e1, e2, 1)
	_ = sketch.AddConstraint(constraint.Distance, e4, e1, 1)
	c3 := sketch.AddConstraint(constraint.Distance, e3, e2, 1)
	c4 := sketch.AddConstraint(constraint.Distance, e1, e3, 1)

	cluster := NewGraphCluster(1)
	cluster.AddElement(e1)
	cluster.AddElement(e2)
	constraints, element, ok := sketch.findConstraints(cluster)
	assert.True(t, ok, "Should find constraints to add to the cluster")
	assert.Equal(t, e3.GetID(), element, "Should find element id for e3")
	assert.Contains(t, constraints, c3.GetID(), "Should contain constraint c3")
	assert.Contains(t, constraints, c4.GetID(), "Should contain constraint c4")

	sketch = NewSketch()
	e1 = sketch.AddPoint(1, 0)
	e2 = sketch.AddOrigin(0, 0)
	e3 = sketch.AddPoint(1, 0)
	e4 = sketch.AddLine(1, 1, 1)
	e5 = sketch.AddLine(1, 1, 2)
	e6 := sketch.AddLine(1, 2, 2)
	_ = sketch.AddConstraint(constraint.Distance, e4, e1, 1)
	_ = sketch.AddConstraint(constraint.Distance, e4, e2, 1)
	_ = sketch.AddConstraint(constraint.Distance, e4, e5, 1)
	_ = sketch.AddConstraint(constraint.Distance, e2, e1, 1)
	_ = sketch.AddConstraint(constraint.Distance, e1, e2, 1)
	c3 = sketch.AddConstraint(constraint.Distance, e3, e2, 1)
	c4 = sketch.AddConstraint(constraint.Distance, e1, e3, 1)
	c5 := sketch.AddConstraint(constraint.Distance, e3, e5, 1)
	c6 := sketch.AddConstraint(constraint.Distance, e3, e6, 1)

	cluster = NewGraphCluster(1)
	cluster.AddElement(e1)
	cluster.AddElement(e2)
	cluster.AddElement(e5)
	cluster.AddElement(e6)
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
	e1 := sketch.AddPoint(1, 0)
	e2 := sketch.AddOrigin(0, 0)
	e3 := sketch.AddPoint(1, 0)
	c3 := sketch.AddConstraint(constraint.Distance, e3, e2, 1)
	c4 := sketch.AddConstraint(constraint.Distance, e1, e3, 1)

	cluster := NewGraphCluster(1)
	sketch.addConstraintToCluster(cluster, c3)
	sketch.addConstraintToCluster(cluster, c4)
	_, ok := cluster.constraints[c3.GetID()]
	assert.True(t, ok, "Cluster has constraint c3")
	_, ok = cluster.constraints[c4.GetID()]
	assert.True(t, ok, "Cluster has constraint c4")
}

func TestCreateCluster(t *testing.T) {
	// Fail to find initial constraint
	sketch := NewSketch()
	c := sketch.createCluster(0, 0)
	assert.Nil(t, c, "Return nil when unable to find initial constraint")

	// Overconstrained
	e1 := sketch.AddPoint(1, 0)
	e2 := sketch.AddOrigin(0, 0)
	e3 := sketch.AddPoint(1, 0)
	e4 := sketch.AddLine(1, 1, 1)
	e5 := sketch.AddLine(1, 1, 2)
	e6 := sketch.AddLine(1, 2, 2)
	_ = sketch.AddConstraint(constraint.Distance, e4, e1, 1)
	_ = sketch.AddConstraint(constraint.Distance, e4, e2, 1)
	_ = sketch.AddConstraint(constraint.Distance, e4, e5, 1)
	_ = sketch.AddConstraint(constraint.Distance, e2, e1, 1)
	_ = sketch.AddConstraint(constraint.Distance, e1, e2, 1)
	_ = sketch.AddConstraint(constraint.Distance, e3, e2, 1)
	_ = sketch.AddConstraint(constraint.Distance, e1, e3, 1)
	_ = sketch.AddConstraint(constraint.Distance, e3, e5, 1)
	_ = sketch.AddConstraint(constraint.Distance, e3, e6, 1)

	c = sketch.createCluster(0, 0)
	assert.NotNil(t, c, "Cluster should not be nil")
	assert.Equal(t, 5, len(c.elements), "Cluster should have 5 elements")
	assert.Equal(t, 7, len(c.constraints), "Cluster should have 7 constraints")
}

func TestCreateBuildResetClusters(t *testing.T) {
	sketch := NewSketch()

	e0 := sketch.AddAxis(0, 1, 0)
	e1 := sketch.AddPoint(1, 0)
	e2 := sketch.AddOrigin(0, 0)
	e3 := sketch.AddPoint(1, 0)
	e4 := sketch.AddLine(1, 1, 1)
	e5 := sketch.AddLine(1, 1, 2)
	e6 := sketch.AddLine(1, 2, 2)
	_ = sketch.AddConstraint(constraint.Distance, e0, e2, 0)
	_ = sketch.AddConstraint(constraint.Distance, e4, e1, 1)
	_ = sketch.AddConstraint(constraint.Distance, e4, e2, 1)
	_ = sketch.AddConstraint(constraint.Distance, e4, e5, 1)
	_ = sketch.AddConstraint(constraint.Distance, e2, e1, 1)
	_ = sketch.AddConstraint(constraint.Distance, e1, e2, 1)
	_ = sketch.AddConstraint(constraint.Distance, e3, e2, 1)
	_ = sketch.AddConstraint(constraint.Distance, e1, e3, 1)
	_ = sketch.AddConstraint(constraint.Distance, e3, e5, 1)
	_ = sketch.AddConstraint(constraint.Distance, e3, e6, 1)
	_ = sketch.AddConstraint(constraint.Distance, e4, e0, 1)

	sketch.createClusters()

	assert.Equal(t, 4, len(sketch.clusters), "Should have 4 clusters")
	assert.Equal(t, 2, len(sketch.clusters[0].elements), "cluster 0 should have 2 element, 1 constraints")
	assert.Equal(t, 1, len(sketch.clusters[0].constraints), "cluster 0 should have 2 element, 1 constraints")

	assert.Equal(t, 6, len(sketch.clusters[1].elements), "cluster 1 should have 6 element, 9 constraints")
	assert.Equal(t, 9, len(sketch.clusters[1].constraints), "cluster 1 should have 6 element, 9 constraints")

	assert.Equal(t, 2, len(sketch.clusters[2].elements), "cluster 2 should have 2 element, 1 constraints")
	assert.Equal(t, 1, len(sketch.clusters[2].constraints), "cluster 2 should have 2 element, 1 constraints")

	assert.Equal(t, 2, len(sketch.clusters[3].elements), "cluster 3 should have 2 element, 1 constraints")
	assert.Equal(t, 1, len(sketch.clusters[3].constraints), "cluster 3 should have 2 element, 1 constraints")

	sketch.ResetClusters()
	assert.Equal(t, 1, len(sketch.clusters), "Should have 1 cluster")
	assert.Equal(t, 2, len(sketch.clusters[0].elements), "cluster 0 should have 2 element, 1 constraints")
	assert.Equal(t, 1, len(sketch.clusters[0].constraints), "cluster 0 should have 2 element, 1 constraints")

	sketch.BuildClusters()

	assert.Equal(t, 4, len(sketch.clusters), "Should have 4 clusters")
	assert.Equal(t, 2, len(sketch.clusters[0].elements), "cluster 0 should have 2 element, 1 constraints")
	assert.Equal(t, 1, len(sketch.clusters[0].constraints), "cluster 0 should have 2 element, 1 constraints")

	assert.Equal(t, 6, len(sketch.clusters[1].elements), "cluster 1 should have 6 element, 7 constraints")
	assert.Equal(t, 9, len(sketch.clusters[1].constraints), "cluster 1 should have 6 element, 9 constraints")

	assert.Equal(t, 2, len(sketch.clusters[2].elements), "cluster 2 should have 2 element, 1 constraints")
	assert.Equal(t, 1, len(sketch.clusters[2].constraints), "cluster 2 should have 2 element, 1 constraints")

	assert.Equal(t, 2, len(sketch.clusters[3].elements), "cluster 3 should have 2 element, 1 constraints")
	assert.Equal(t, 1, len(sketch.clusters[3].constraints), "cluster 3 should have 2 element, 1 constraints")
}

func TestSolve(t *testing.T) {
	s := NewSketch()
	origin := s.AddOrigin(0, 0)
	xAxis := s.AddAxis(0, -1, 0)
	yAxis := s.AddAxis(1, 0, 0)
	s.AddConstraint(constraint.Angle, xAxis, yAxis, math.Pi/2)
	s.AddConstraint(constraint.Distance, origin, xAxis, 0)
	s.AddConstraint(constraint.Distance, origin, yAxis, 0)

	p1 := s.AddPoint(0, 0)
	p2 := s.AddPoint(4, 0)
	p3 := s.AddPoint(5.236068, 3.804226)
	p4 := s.AddPoint(2, 6.155367)
	p5 := s.AddPoint(-1.236068, 3.804226)

	l1 := s.AddLine(0, -1, 0)
	l2 := s.AddLine(0.951057, -0.309017, -3.804226)
	l3 := s.AddLine(0.587785, 0.809017, -6.155367)
	l4 := s.AddLine(-0.587785, 0.809017, -3.804226)
	l5 := s.AddLine(-0.951057, -0.309017, 0)

	s.AddConstraint(constraint.Distance, l1, p1, 0)
	s.AddConstraint(constraint.Distance, l1, p2, 0)
	s.AddConstraint(constraint.Distance, l2, p2, 0)
	s.AddConstraint(constraint.Distance, l2, p3, 0)
	s.AddConstraint(constraint.Distance, l3, p3, 0)
	s.AddConstraint(constraint.Distance, l3, p4, 0)
	s.AddConstraint(constraint.Distance, l4, p4, 0)
	s.AddConstraint(constraint.Distance, l4, p5, 0)
	s.AddConstraint(constraint.Distance, l5, p5, 0)
	s.AddConstraint(constraint.Distance, l5, p1, 0)

	s.AddConstraint(constraint.Angle, l2, l3, (72.0/180.0)*math.Pi)
	s.AddConstraint(constraint.Angle, l3, l4, (72.0/180.0)*math.Pi)
	s.AddConstraint(constraint.Angle, l4, l5, (72.0/180.0)*math.Pi)

	s.AddConstraint(constraint.Distance, p1, p2, 4)
	s.AddConstraint(constraint.Distance, p2, p3, 4)
	s.AddConstraint(constraint.Distance, p3, p4, 4)
	s.AddConstraint(constraint.Distance, p4, p5, 4)

	s.AddConstraint(constraint.Angle, l1, xAxis, 0)
	s.AddConstraint(constraint.Distance, p1, origin, 0)

	s.ResetClusters()
	s.BuildClusters()
	state := s.Solve()

	assert.Equal(t, solver.Solved, state, "Graph should be solved")
}

func TestFindMergeForCluster(t *testing.T) {

}

func TestFindMerge(t *testing.T) {

}

func TestIsSolved(t *testing.T) {

}

func TestGraphToGraphViz(t *testing.T) {

}
