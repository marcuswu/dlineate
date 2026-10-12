package dlineate

import (
	"testing"

	"github.com/marcuswu/dlineate/internal/element"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSolveWithReportSolvedSketch(t *testing.T) {
	s := NewSketch()
	circle := s.AddCircle(0, 0, 3)
	s.AddCoincidentConstraint(circle.Center(), s.Origin)

	report := s.SolveWithReport()
	require.Equal(t, SolveStatusSolved, report.Status)
	require.Empty(t, report.Error)
	require.NotEmpty(t, report.Elements)
	require.NotEmpty(t, report.Constraints)
	assert.Empty(t, report.Conflicts)
	assert.Len(t, report.Diagnostics, 1)
	assert.Equal(t, "solve.solved", report.Diagnostics[0].Code)

	byID := make(map[uint]ElementSolveResult, len(report.Elements))
	for _, result := range report.Elements {
		byID[result.ID] = result
	}
	center, ok := byID[circle.Center().ID()]
	require.True(t, ok)
	assert.Equal(t, ElementSolveStatusFullyConstrained, center.Status)
	assert.Equal(t, []float64{0, 0}, center.Values)

	seenIDs := make(map[uint64]bool, len(report.Constraints))
	for _, result := range report.Constraints {
		assert.NotZero(t, result.ID)
		assert.False(t, seenIDs[result.ID], "public constraint IDs must be unique")
		seenIDs[result.ID] = true
	}

	// Reports are snapshots; callers cannot mutate the sketch's cached report
	// by changing slices returned from either solve method.
	report.Elements[0].Values[0] = 12345
	last := s.LastSolveReport()
	assert.NotEqual(t, 12345.0, last.Elements[0].Values[0])
}

func TestSolveWithReportPreservesLegacyError(t *testing.T) {
	s := NewSketch()
	s.AddPoint(2, 3)
	s.AddLine(0, 0, 2, 0)
	s.AddCircle(4, 5, 1)
	s.AddArc(7, 8, 8, 8, 7, 9)

	report := s.SolveWithReport()
	assert.Equal(t, SolveStatusNonConvergent, report.Status)
	assert.Equal(t, "failed to solve completely", report.Error)
	assert.Len(t, report.Diagnostics, 1)
	assert.Equal(t, "solve.non_convergent", report.Diagnostics[0].Code)
	assert.EqualError(t, report.Err(), "failed to solve completely")
	assert.Equal(t, report.Status, s.LastSolveReport().Status)

	legacySketch := NewSketch()
	legacySketch.AddPoint(2, 3)
	legacySketch.AddLine(0, 0, 2, 0)
	legacySketch.AddCircle(4, 5, 1)
	legacySketch.AddArc(7, 8, 8, 8, 7, 9)
	assert.EqualError(t, legacySketch.Solve(), "failed to solve completely")
}

func TestSolveWithReportExposesConflictDetails(t *testing.T) {
	s := NewSketch()
	p1 := s.AddPoint(0, 0)
	p2 := s.AddPoint(2, 0)
	s.MakeFixed(p1)
	s.MakeFixed(p2)
	c1 := s.AddDistanceConstraint(p1, p2, 2)
	c2 := s.AddDistanceConstraint(p1, p2, 2)

	report := s.SolveWithReport()
	require.Equal(t, SolveStatusOverConstrained, report.Status)
	require.Len(t, report.Conflicts, 1)
	assert.Equal(t, "solve.over_constrained", report.Diagnostics[0].Code)
	conflict := report.Conflicts[0]
	assert.True(t, conflict.ID == c1.ID() || conflict.ID == c2.ID())
	assert.True(t, conflict.Conflicting)
	assert.Contains(t, conflict.ElementIDs, p1.ID())
	assert.Contains(t, conflict.ElementIDs, p2.ID())
	assert.NotEmpty(t, conflict.CoreIDs)
	assert.NotEmpty(t, report.Error)

	for _, result := range report.Constraints {
		if result.ID == conflict.ID {
			assert.True(t, result.Conflicting)
		}
	}
}

func TestPublicSolveStatuses(t *testing.T) {
	assert.Equal(t, "solved", SolveStatusSolved.String())
	assert.Equal(t, "under constrained", SolveStatusUnderConstrained.String())
	assert.Equal(t, "over constrained", SolveStatusOverConstrained.String())
	assert.Equal(t, "non-convergent", SolveStatusNonConvergent.String())
	assert.Equal(t, "under constrained", ElementSolveStatusUnderConstrained.String())
	assert.Equal(t, "fully constrained", ElementSolveStatusFullyConstrained.String())
	assert.Equal(t, "over constrained", ElementSolveStatusOverConstrained.String())
	assert.Equal(t, ElementSolveStatusUnderConstrained, publicElementStatus(element.UnderConstrained))
}
