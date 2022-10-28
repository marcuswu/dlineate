package solver

import (
	"math"
	"testing"

	"github.com/marcuswu/dlineation/internal/constraint"
	el "github.com/marcuswu/dlineation/internal/element"
	"github.com/marcuswu/dlineation/utils"
	"github.com/stretchr/testify/assert"
)

func TestSolveAngleConstraint(t *testing.T) {
	tests := []struct {
		name       string
		constraint *constraint.Constraint
		solveFor   uint
		solveState SolveState
	}{
		{
			"Test -108º constraint",
			constraint.NewConstraint(
				0,
				constraint.Angle,
				el.NewSketchLine(0, 0, 1, 0),
				el.NewSketchLine(1, -0.951057, 0.309017, 0),
				-(108.0/180.0)*math.Pi,
				false,
			),
			0,
			Solved,
		},
		{
			"Test -108º constraint 2",
			constraint.NewConstraint(
				0,
				constraint.Angle,
				el.NewSketchLine(2, -0.506732, -0.862104, 0),
				el.NewSketchLine(3, -0.506732, -0.862104, 0),
				-(108.0/180.0)*math.Pi,
				false,
			),
			3,
			Solved,
		},
		{
			"Test incorrect Constraint",
			constraint.NewConstraint(
				0,
				constraint.Distance,
				el.NewSketchPoint(0, 0, 0),
				el.NewSketchPoint(1, 1, 1),
				2,
				false,
			),
			0,
			NonConvergent,
		},
	}
	for _, tt := range tests {
		newLine, status := SolveAngleConstraint(tt.constraint, tt.solveFor)
		if tt.solveState == Solved {
			first := tt.constraint.Element1.AsLine()
			second := tt.constraint.Element2.AsLine()
			if first.GetID() == newLine.GetID() {
				first = newLine
			} else {
				second = newLine
			}
			assert.Equal(t, tt.solveFor, newLine.GetID(), tt.name)
			assert.InDelta(t, tt.constraint.Value, first.AngleToLine(second), utils.StandardCompare, tt.name)
		} else {
			assert.Nil(t, newLine, tt.name)
		}
		assert.Equal(t, tt.solveState, status, tt.name)
	}
}
