package dlineate

import (
	"math"
	"testing"

	"github.com/marcuswu/dlineate/internal/element"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests record the current public behavior before changing status and
// conflict semantics. Assertions marked as legacy observations should be
// replaced with intended UI-facing semantics as those are implemented.
func TestCharacterizeIsolatedGeometry(t *testing.T) {
	s := NewSketch()
	point := s.AddPoint(2, 3)
	line := s.AddLine(0, 0, 2, 0)
	circle := s.AddCircle(4, 5, 1)
	arc := s.AddArc(7, 8, 8, 8, 7, 9)

	err := s.Solve()
	assert.EqualError(t, err, "failed to solve completely")
	assert.Equal(t, []float64{2, 3}, point.Values())
	assert.Equal(t, []float64{0, 0, 2, 0}, line.Values())
	assert.Equal(t, []float64{4, 5, 1}, circle.Values())
	assert.Equal(t, []float64{7, 8, 8, 8, 7, 9}, arc.Values())

	// Legacy observation: unconstrained geometry currently defaults to fully
	// constrained even though the solve cannot place it from constraints.
	assert.Equal(t, element.FullyConstrained, point.ConstraintLevel())
	assert.Equal(t, element.FullyConstrained, line.ConstraintLevel())
	assert.Equal(t, element.FullyConstrained, circle.ConstraintLevel())
	assert.Equal(t, element.FullyConstrained, arc.ConstraintLevel())
	assert.Empty(t, s.ConflictingConstraints())
}

func TestCharacterizeFixedInconsistentDistance(t *testing.T) {
	s := NewSketch()
	p1 := s.AddPoint(0, 0)
	p2 := s.AddPoint(2, 0)
	s.MakeFixed(p1)
	s.MakeFixed(p2)
	distance := s.AddDistanceConstraint(p1, p2, 1)

	err := s.Solve()
	assert.EqualError(t, err, "failed to solve completely")
	assert.Equal(t, []float64{0, 0}, p1.Values())
	assert.Equal(t, []float64{2, 0}, p2.Values())
	assert.Equal(t, element.FullyConstrained, p1.ConstraintLevel())
	assert.Equal(t, element.FullyConstrained, p2.ConstraintLevel())

	// Legacy observation: a violated fixed-to-fixed distance has no entry in
	// the conflict list, and the generic solve error does not explain why.
	assert.Empty(t, s.ConflictingConstraints())
	assert.NotNil(t, distance)
}

func TestCharacterizeDisconnectedConstrainedComponents(t *testing.T) {
	s := NewSketch()
	leftAnchor := s.AddPoint(-10, 0)
	leftFree := s.AddPoint(-8, 1)
	rightAnchor := s.AddPoint(10, 0)
	rightFree := s.AddPoint(12, -1)
	s.MakeFixed(leftAnchor)
	s.MakeFixed(rightAnchor)
	s.AddDistanceConstraint(leftAnchor, leftFree, 2)
	s.AddDistanceConstraint(rightAnchor, rightFree, 2)

	err := s.Solve()
	// Legacy observation: each component has only one distance from a fixed
	// anchor, but the sketch fails rather than reporting a partial/under-
	// constrained result. The initial geometry is preserved.
	assert.EqualError(t, err, "failed to solve completely")
	assert.InDelta(t, math.Sqrt(5), leftAnchor.DistanceBetweenPoints(leftFree), 1e-6)
	assert.InDelta(t, math.Sqrt(5), rightAnchor.DistanceBetweenPoints(rightFree), 1e-6)
	assert.Empty(t, s.ConflictingConstraints())
}

func TestCharacterizeInconsistentTangentSketch(t *testing.T) {
	s := NewSketch()
	circle := s.AddCircle(0, 0, 3)
	line := s.AddLine(0, 0, 1, 1)
	tangent, err := s.AddTangentConstraint(circle, line)
	assert.NoError(t, err)
	startCoincident := s.AddCoincidentConstraint(line.Start(), s.Origin)
	length := s.AddDistanceConstraint(line, nil, 5)
	endCoincident := s.AddCoincidentConstraint(circle.Center(), line.End())

	err = s.Solve()
	assert.EqualError(t, err, "failed to solve completely")
	assert.Equal(t, element.FullyConstrained, circle.ConstraintLevel())
	assert.Equal(t, element.FullyConstrained, line.ConstraintLevel())

	// Legacy observation: this deliberately inconsistent set is not identified
	// by ConflictingConstraints(), despite Solve returning an error.
	assert.Empty(t, s.ConflictingConstraints())
	assert.NotNil(t, tangent)
	assert.Nil(t, startCoincident, "coincident point pairs are merged rather than represented as constraints")
	assert.NotNil(t, length)
	assert.Nil(t, endCoincident, "coincident point pairs are merged rather than represented as constraints")
}

func TestCharacterizeRedundantConsistentConstraints(t *testing.T) {
	s := NewSketch()
	p1 := s.AddPoint(0, 0)
	p2 := s.AddPoint(2, 0)
	s.MakeFixed(p1)
	s.MakeFixed(p2)
	distance1 := s.AddDistanceConstraint(p1, p2, 2)
	distance2 := s.AddDistanceConstraint(p1, p2, 2)

	err := s.Solve()
	// Legacy observation: duplicate but consistent constraints are currently
	// rejected and reported as conflicts; phase four should correct this.
	assert.EqualError(t, err, "failed to solve completely")
	conflicts := s.ConflictingConstraints()
	assert.Len(t, conflicts, 1)
	assert.True(t, conflicts[0] == distance1 || conflicts[0] == distance2)
	assert.Equal(t, []float64{0, 0}, p1.Values())
	assert.Equal(t, []float64{2, 0}, p2.Values())
}

func TestCharacterizePentagonGraphMerge(t *testing.T) {
	s := NewSketch()
	l1 := s.AddLine(0, 0, 3.13, 0)
	l2 := s.AddLine(3.13, 0, 5.14, 2.27)
	l3 := s.AddLine(5.14, 2.27, 2.28, 4.72)
	l4 := s.AddLine(2.28, 4.72, -1.04, 3.56)
	l5 := s.AddLine(-1.04, 3.56, 0, 0)

	s.AddCoincidentConstraint(s.Origin, l1.Start())
	s.AddParallelConstraint(s.XAxis, l1)
	s.AddCoincidentConstraint(l1.End(), l2.Start())
	s.AddCoincidentConstraint(l2.End(), l3.Start())
	s.AddCoincidentConstraint(l3.End(), l4.Start())
	s.AddCoincidentConstraint(l4.End(), l5.Start())
	s.AddCoincidentConstraint(l5.End(), l1.Start())
	s.AddAngleConstraint(l2, l3, 108, true)
	s.AddAngleConstraint(l3, l4, 108, true)
	s.AddAngleConstraint(l4, l5, 108, true)
	s.AddDistanceConstraint(l1, nil, 4)
	s.AddDistanceConstraint(l2, nil, 4)
	s.AddDistanceConstraint(l4, nil, 4)
	s.AddDistanceConstraint(l5, nil, 4)

	require.NoError(t, s.Solve())
	minX, minY, maxX, maxY := s.calculateRectangle(1)
	assert.InDelta(t, -1.236068, minX, 1e-5)
	assert.InDelta(t, 0, minY, 1e-5)
	assert.InDelta(t, 5.236068, maxX, 1e-5)
	assert.InDelta(t, 6.155367, maxY, 1e-5)
	assert.Empty(t, s.ConflictingConstraints())
	for _, line := range []*Element{l1, l2, l3, l4, l5} {
		assert.Equal(t, element.FullyConstrained, line.ConstraintLevel())
	}
}

func TestCharacterizeModifiedCylinderFallback(t *testing.T) {
	s := NewSketch()
	start := s.AddPoint(6, 0)
	line1 := s.AddLine(6, 0, 9.4, 0)
	line2 := s.AddLine(9.4, 0, 9.4, -1)
	line3 := s.AddLine(9.4, -1, 8.1, -2)
	arc1 := s.AddArc(8.5, -2.4, 8.1, -2, 8.1, -2.7)
	line4 := s.AddLine(8.1, -2.7, 9.4, -3.8)
	line5 := s.AddLine(9.4, -3.8, 9.4, -8)
	line6 := s.AddLine(9.4, -8, 6, -8)
	line7 := s.AddLine(6, -8, 6, 0)
	arc2 := s.AddArc(6, 0, 6, 3, 6, -3)
	arc3 := s.AddArc(6, -5, 6, -8.3, 6, -2)

	s.AddCoincidentConstraint(s.XAxis, start)
	s.AddDistanceConstraint(s.Origin, start, 6)
	s.AddCoincidentConstraint(start, line1.Start())
	s.AddCoincidentConstraint(line1.End(), line2.Start())
	s.AddCoincidentConstraint(line2.End(), line3.Start())
	s.AddCoincidentConstraint(line3.End(), arc1.Start())
	s.AddCoincidentConstraint(arc1.End(), line4.Start())
	s.AddCoincidentConstraint(line4.End(), line5.Start())
	s.AddCoincidentConstraint(line5.End(), line6.Start())
	s.AddCoincidentConstraint(line6.End(), line7.Start())
	s.AddCoincidentConstraint(line7.End(), line1.Start())
	s.AddParallelConstraint(s.XAxis, line1)
	s.AddDistanceConstraint(line1, nil, 3.4)
	s.AddParallelConstraint(s.YAxis, line2)
	s.AddAngleConstraint(line2, line3, 135, false)
	s.AddAngleConstraint(line3, line4, 90, false)
	s.AddDistanceConstraint(arc1, nil, 0.5)
	s.AddDistanceConstraint(arc1.Center(), line7, 2.5)
	s.AddTangentConstraint(arc1, line3)
	s.AddTangentConstraint(arc1, line4)
	s.AddParallelConstraint(s.YAxis, line5)
	s.AddParallelConstraint(s.XAxis, line6)
	s.AddDistanceConstraint(line6, nil, 3.4)
	s.AddParallelConstraint(s.YAxis, line7)
	s.AddDistanceConstraint(line7, nil, 8)
	s.AddCoincidentConstraint(arc2.Center(), line1.Start())
	s.AddCoincidentConstraint(arc2.Start(), line7)
	s.AddCoincidentConstraint(arc2.End(), line7)
	s.AddTangentConstraint(arc2, line3)
	s.AddDistanceConstraint(arc2, nil, 3)
	s.AddCoincidentConstraint(arc3.Center(), line7)
	s.AddCoincidentConstraint(arc3.End(), line7)
	s.AddCoincidentConstraint(arc3.Start(), line7)
	s.AddTangentConstraint(arc3, line4)
	s.AddDistanceConstraint(arc3, nil, 3)

	require.NoError(t, s.Solve())
	assert.InDelta(t, 6, line1.Values()[0], 1e-5)
	assert.InDelta(t, 0, line1.Values()[1], 1e-5)
	assert.InDelta(t, 9.4, line1.Values()[2], 1e-5)
	assert.InDelta(t, 0, line1.Values()[3], 1e-5)
	arcValues := arc1.Values()
	dx := arcValues[0] - arcValues[2]
	dy := arcValues[1] - arcValues[3]
	assert.InDelta(t, 0.5, math.Hypot(dx, dy), 1e-4)
	assert.Empty(t, s.ConflictingConstraints())
}
