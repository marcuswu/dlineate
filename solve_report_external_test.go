package dlineate_test

import (
	"testing"

	"github.com/marcuswu/dlineate"
	"github.com/stretchr/testify/require"
)

// This external-package test ensures the reporting API can be consumed without
// importing any dlineate/internal packages.
func TestPublicSolveReportAPI(t *testing.T) {
	sketch := dlineate.NewSketch()
	circle := sketch.AddCircle(0, 0, 3)
	sketch.AddCoincidentConstraint(circle.Center(), sketch.Origin)

	report := sketch.SolveWithReport()
	require.Equal(t, dlineate.SolveStatusSolved, report.Status)
	require.NoError(t, report.Err())
	require.NotEmpty(t, report.Elements)
	require.NotEmpty(t, report.Constraints)
	require.Empty(t, report.Conflicts)
	for _, constraint := range report.Constraints {
		require.NotZero(t, constraint.ID)
	}

	var foundPoint bool
	for _, element := range report.Elements {
		if element.ID == circle.Center().ID() {
			foundPoint = true
			require.Equal(t, dlineate.ElementSolveStatusFullyConstrained, element.Status)
			require.False(t, element.DegreesOfFreedomKnown)
		}
	}
	require.True(t, foundPoint)
}
