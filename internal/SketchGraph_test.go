package core

import (
	"math"
	"testing"

	"github.com/marcuswu/dlineation/internal/constraint"
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

}

func TestFindStartConstraint(t *testing.T) {

}

func TestFindConstraints(t *testing.T) {

}

func TestAddConstraintToCluster(t *testing.T) {

}

func TestCreateCluster(t *testing.T) {

}

func TestCreateClusters(t *testing.T) {

}

func TestConstraintLevel(t *testing.T) {

}

func TestTranslateRotate(t *testing.T) {

}

func TestLogConstraintsElements(t *testing.T) {

}

func TestUpdateElements(t *testing.T) {

}

func TestAddClusterConstraints(t *testing.T) {

}

func TestResetClusters(t *testing.T) {

}

func TestBuildClusters(t *testing.T) {

}

func TestSolve(t *testing.T) {

}

func TestFindMergeForCluster(t *testing.T) {

}

func TestFindMerge(t *testing.T) {

}

func TestIsSolved(t *testing.T) {

}

func TestGraphToGraphViz(t *testing.T) {

}
