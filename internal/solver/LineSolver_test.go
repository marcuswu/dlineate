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
				(108.0/180.0)*math.Pi,
				false,
			),
			3,
			Solved,
		},
		{
			"Test reverse rotation",
			constraint.NewConstraint(
				0,
				constraint.Angle,
				el.NewSketchLine(0, 1.5, 0.3, 0.1),
				el.NewSketchLine(1, 0.3, 1.5, -0.1),
				(70.0/180.0)*math.Pi,
				false,
			),
			1,
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
			assert.InDelta(t, math.Abs(tt.constraint.Value), math.Abs(first.AngleToLine(second)), utils.StandardCompare, tt.name)
		} else {
			assert.Nil(t, newLine, tt.name)
		}
		assert.Equal(t, tt.solveState, status, tt.name)
	}
}

func TestLineFromPointLine(t *testing.T) {
	tests := []struct {
		name    string
		c1      *constraint.Constraint
		c2      *constraint.Constraint
		desired *el.SketchLine
		state   SolveState
	}{
		{
			"test 1",
			constraint.NewConstraint(0, constraint.Angle, el.NewSketchLine(0, 1.5, 0.3, 0.1), el.NewSketchLine(1, 0.3, 1.5, -0.1), (70.0/180.0)*math.Pi, false),
			constraint.NewConstraint(0, constraint.Distance, el.NewSketchPoint(0, 1, 1), el.NewSketchLine(1, 0.3, 1.5, -0.1), 1, false),
			el.NewSketchLine(1, 0.151089, 0.988520, -0.139610),
			Solved,
		},
		{
			"test 2",
			constraint.NewConstraint(0, constraint.Angle, el.NewSketchLine(0, 1.5, 0.3, 0.1), el.NewSketchLine(1, 0.3, 1.5, -0.1), (70.0/180.0)*math.Pi, false),
			constraint.NewConstraint(0, constraint.Distance, el.NewSketchPoint(0, -1, -1), el.NewSketchLine(1, 0.3, 1.5, -0.1), 1, false),
			el.NewSketchLine(1, 0.151089, 0.988520, 0.139610),
			Solved,
		},
	}
	for _, tt := range tests {
		newLine, state := LineFromPointLine(tt.c1, tt.c2)
		assert.Equal(t, state, tt.state, tt.name)
		if tt.state == NonConvergent {
			assert.Nil(t, newLine)
		} else {
			c1Line, _ := tt.c1.Element(newLine.GetID())
			c2Line, _ := tt.c2.Element(newLine.GetID())
			c1Line.AsLine().SetA(newLine.GetA())
			c1Line.AsLine().SetB(newLine.GetB())
			c1Line.AsLine().SetC(newLine.GetC())
			c2Line.AsLine().SetA(newLine.GetA())
			c2Line.AsLine().SetB(newLine.GetB())
			c2Line.AsLine().SetC(newLine.GetC())
			assert.True(t, tt.c1.IsMet(), tt.name)
			assert.True(t, tt.c2.IsMet(), tt.name)
			assert.Equal(t, tt.desired.GetID(), newLine.GetID(), tt.name)
			assert.InDelta(t, tt.desired.GetA(), newLine.GetA(), utils.StandardCompare, tt.name)
			assert.InDelta(t, tt.desired.GetB(), newLine.GetB(), utils.StandardCompare, tt.name)
			assert.InDelta(t, tt.desired.GetC(), newLine.GetC(), utils.StandardCompare, tt.name)
		}
	}
}

func TestLineFromPoints(t *testing.T) {
	tests := []struct {
		name    string
		c1      *constraint.Constraint
		c2      *constraint.Constraint
		desired *el.SketchLine
		state   SolveState
	}{
		{
			"Can't find line",
			constraint.NewConstraint(0, constraint.Distance, el.NewSketchPoint(0, 1.5, 0.3), el.NewSketchPoint(1, 0.3, 1.5), 1, false),
			constraint.NewConstraint(0, constraint.Distance, el.NewSketchPoint(2, 1, 1), el.NewSketchPoint(1, 0.3, 1.5), 1, false),
			nil,
			NonConvergent,
		},
		{
			"Can't find points",
			constraint.NewConstraint(0, constraint.Distance, el.NewSketchLine(0, 1.1, 0.1, 0.1), el.NewSketchLine(1, 0.3, 1.5, -0.1), (70.0/180.0)*math.Pi, false),
			constraint.NewConstraint(0, constraint.Distance, el.NewSketchLine(2, -1, -1, 0.0), el.NewSketchLine(1, 0.3, 1.5, -0.1), 1, false),
			el.NewSketchLine(0, 1.1, 0.1, 0.1),
			NonConvergent,
		},
		{
			"Coincident to both points",
			constraint.NewConstraint(0, constraint.Distance, el.NewSketchPoint(0, 1.5, 0.3), el.NewSketchLine(1, 0.3, 1.5, -0.1), 0, false),
			constraint.NewConstraint(0, constraint.Distance, el.NewSketchPoint(2, 1, 1), el.NewSketchLine(1, 0.3, 1.5, -0.1), 0, false),
			el.NewSketchLine(1, 0.7, 0.5, -1.2),
			Solved,
		},
		{
			"Coincident to both points alt slope",
			constraint.NewConstraint(0, constraint.Distance, el.NewSketchPoint(0, 1, 1), el.NewSketchLine(1, 0.3, 1.5, -0.1), 0, false),
			constraint.NewConstraint(0, constraint.Distance, el.NewSketchPoint(2, 1.5, 0.3), el.NewSketchLine(1, 0.3, 1.5, -0.1), 0, false),
			el.NewSketchLine(1, 0.7, 0.5, -1.2),
			Solved,
		},
		{
			"Distance between points is too large",
			constraint.NewConstraint(0, constraint.Distance, el.NewSketchPoint(0, 1.5, 0.3), el.NewSketchLine(1, 0.3, 1.5, -0.1), 1, false),
			constraint.NewConstraint(0, constraint.Distance, el.NewSketchPoint(2, 1, 1), el.NewSketchLine(1, 0.3, 1.5, -0.1), 1, false),
			el.NewSketchLine(1, 0.3, 1.5, -0.1),
			NonConvergent,
		},
		{
			"Solve by tangent external",
			constraint.NewConstraint(0, constraint.Distance, el.NewSketchPoint(0, 1.5, 0.3), el.NewSketchLine(1, 0.3, 1.5, -0.1), 0.25, false),
			constraint.NewConstraint(0, constraint.Distance, el.NewSketchPoint(2, -1, 1), el.NewSketchLine(1, 0.3, 1.5, -0.1), 0.25, false),
			el.NewSketchLine(1, 0.269630, 0.962964, -0.443334),
			Solved,
		},
		{
			"Solve by tangent internal",
			constraint.NewConstraint(0, constraint.Distance, el.NewSketchPoint(0, 1.5, 0.3), el.NewSketchLine(1, 0.3, 1.5, -0.1), 0.25, false),
			constraint.NewConstraint(0, constraint.Distance, el.NewSketchPoint(2, -1, 0), el.NewSketchLine(1, 0.3, 1.5, -0.1), 0.25, false),
			el.NewSketchLine(1, 0.080388, 0.996764, -0.169612),
			Solved,
		},
	}
	for _, tt := range tests {
		newLine, state := LineFromPoints(tt.c1, tt.c2)
		assert.Equal(t, state, tt.state, tt.name)
		if tt.desired == nil {
			assert.Nil(t, newLine, tt.name)
		} else {
			newLine.Normalize()
			c1Line, _ := tt.c1.Element(newLine.GetID())
			c2Line, _ := tt.c2.Element(newLine.GetID())
			c1Line.AsLine().SetA(newLine.GetA())
			c1Line.AsLine().SetB(newLine.GetB())
			c1Line.AsLine().SetC(newLine.GetC())
			c2Line.AsLine().SetA(newLine.GetA())
			c2Line.AsLine().SetB(newLine.GetB())
			c2Line.AsLine().SetC(newLine.GetC())
			assert.Equal(t, tt.desired.GetID(), newLine.GetID(), tt.name)
			assert.InDelta(t, tt.desired.GetA(), newLine.GetA(), utils.StandardCompare, tt.name)
			assert.InDelta(t, tt.desired.GetB(), newLine.GetB(), utils.StandardCompare, tt.name)
			assert.InDelta(t, tt.desired.GetC(), newLine.GetC(), utils.StandardCompare, tt.name)
		}

		if tt.state == Solved {
			assert.True(t, tt.c1.IsMet(), tt.name)
			assert.True(t, tt.c2.IsMet(), tt.name)
		}
	}
}

func TestMoveLineToPoint(t *testing.T) {
	tests := []struct {
		name    string
		c1      *constraint.Constraint
		desired *el.SketchLine
		state   SolveState
	}{
		{
			"The constraint should be a distance constraint",
			constraint.NewConstraint(0, constraint.Angle, el.NewSketchPoint(0, 1.5, 0.3), el.NewSketchPoint(1, 0.3, 1.5), 1, false),
			nil,
			NonConvergent,
		},
		{
			"The constraint should have a point and a line",
			constraint.NewConstraint(0, constraint.Distance, el.NewSketchPoint(0, 1.5, 0.3), el.NewSketchPoint(1, 0.3, 1.5), 1, false),
			nil,
			NonConvergent,
		},
		{
			"The constraint should be met",
			constraint.NewConstraint(0, constraint.Distance, el.NewSketchPoint(0, 1.5, 0.3), el.NewSketchLine(1, 0.3, 1.5, 1), 1, false),
			el.NewSketchLine(1, 0.3, 1.5, 0.629706),
			Solved,
		},
		{
			"The constraint should be met",
			constraint.NewConstraint(0, constraint.Distance, el.NewSketchLine(1, 0.3, 1.5, 1), el.NewSketchPoint(0, 1.5, -2), 1, false),
			el.NewSketchLine(1, 0.3, 1.5, 1.020294),
			Solved,
		},
	}
	for _, tt := range tests {
		state := MoveLineToPoint(tt.c1)
		assert.Equal(t, tt.state, state, tt.name)
		if tt.state != Solved {
			continue
		}
		assert.True(t, tt.c1.IsMet(), tt.name)
		newLine := tt.c1.Element1.AsLine()
		if newLine == nil {
			newLine = tt.c1.Element2.AsLine()
		}
		assert.Equal(t, tt.desired.GetID(), newLine.GetID(), tt.name)
		assert.InDelta(t, tt.desired.GetA(), newLine.GetA(), utils.StandardCompare, tt.name)
		assert.InDelta(t, tt.desired.GetB(), newLine.GetB(), utils.StandardCompare, tt.name)
		assert.InDelta(t, tt.desired.GetC(), newLine.GetC(), utils.StandardCompare, tt.name)
	}
}

func TestLineResult(t *testing.T) {
	tests := []struct {
		name    string
		c1      *constraint.Constraint
		c2      *constraint.Constraint
		desired *el.SketchLine
		state   SolveState
	}{
		{
			"Test Line From Points",
			constraint.NewConstraint(0, constraint.Distance, el.NewSketchPoint(0, 1.5, 0.3), el.NewSketchLine(1, 0.3, 1.5, -0.1), 0, false),
			constraint.NewConstraint(0, constraint.Distance, el.NewSketchPoint(2, 1, 1), el.NewSketchLine(1, 0.3, 1.5, -0.1), 0, false),
			el.NewSketchLine(1, 0.7, 0.5, -1.2),
			Solved,
		},
		{
			"Test Line From Point Line",
			constraint.NewConstraint(0, constraint.Angle, el.NewSketchLine(0, 1.5, 0.3, 0.1), el.NewSketchLine(1, 0.3, 1.5, -0.1), (70.0/180.0)*math.Pi, false),
			constraint.NewConstraint(0, constraint.Distance, el.NewSketchPoint(0, 1, 1), el.NewSketchLine(1, 0.3, 1.5, -0.1), 1, false),
			el.NewSketchLine(1, 0.151089, 0.988520, -0.139610),
			Solved,
		},
	}
	for _, tt := range tests {
		newLine, state := LineResult(tt.c1, tt.c2)
		assert.Equal(t, state, tt.state, tt.name)
		if tt.desired == nil {
			assert.Nil(t, newLine, tt.name)
		} else {
			newLine.Normalize()
			c1Line, _ := tt.c1.Element(newLine.GetID())
			c2Line, _ := tt.c2.Element(newLine.GetID())
			c1Line.AsLine().SetA(newLine.GetA())
			c1Line.AsLine().SetB(newLine.GetB())
			c1Line.AsLine().SetC(newLine.GetC())
			c2Line.AsLine().SetA(newLine.GetA())
			c2Line.AsLine().SetB(newLine.GetB())
			c2Line.AsLine().SetC(newLine.GetC())
			assert.Equal(t, tt.desired.GetID(), newLine.GetID(), tt.name)
			assert.InDelta(t, tt.desired.GetA(), newLine.GetA(), utils.StandardCompare, tt.name)
			assert.InDelta(t, tt.desired.GetB(), newLine.GetB(), utils.StandardCompare, tt.name)
			assert.InDelta(t, tt.desired.GetC(), newLine.GetC(), utils.StandardCompare, tt.name)
		}

		if tt.state == Solved {
			assert.True(t, tt.c1.IsMet(), tt.name)
			assert.True(t, tt.c2.IsMet(), tt.name)
		}
	}
}

func TestSolveForLine(t *testing.T) {
	tests := []struct {
		name    string
		c1      *constraint.Constraint
		c2      *constraint.Constraint
		desired *el.SketchLine
		state   SolveState
	}{
		{
			"Test Nonconvergent",
			constraint.NewConstraint(0, constraint.Distance, el.NewSketchPoint(0, 1.5, 0.3), el.NewSketchPoint(1, 0.3, 1.5), 0, false),
			constraint.NewConstraint(0, constraint.Distance, el.NewSketchPoint(2, 1, 1), el.NewSketchPoint(1, 0.3, 1.5), 0, false),
			nil,
			NonConvergent,
		},
		{
			"Test Line From Points",
			constraint.NewConstraint(0, constraint.Distance, el.NewSketchPoint(0, 1.5, 0.3), el.NewSketchLine(1, 0.3, 1.5, -0.1), 0, false),
			constraint.NewConstraint(0, constraint.Distance, el.NewSketchPoint(2, 1, 1), el.NewSketchLine(1, 0.3, 1.5, -0.1), 0, false),
			el.NewSketchLine(1, 0.7, 0.5, -1.2),
			Solved,
		},
		{
			"Test Line From Point Line",
			constraint.NewConstraint(0, constraint.Angle, el.NewSketchLine(0, 1.5, 0.3, 0.1), el.NewSketchLine(1, 0.3, 1.5, -0.1), (70.0/180.0)*math.Pi, false),
			constraint.NewConstraint(0, constraint.Distance, el.NewSketchPoint(2, 1, 1), el.NewSketchLine(1, 0.3, 1.5, -0.1), 1, false),
			el.NewSketchLine(1, 0.151089, 0.988520, -0.139610),
			Solved,
		},
	}
	for _, tt := range tests {
		state := SolveForLine(tt.c1, tt.c2)
		assert.Equal(t, state, tt.state, tt.name)
		if tt.state == Solved {
			assert.True(t, tt.c1.IsMet(), tt.name)
			assert.True(t, tt.c2.IsMet(), tt.name)
		}
		shared, _ := tt.c1.Shared(tt.c2)
		if shared == nil || tt.state == NonConvergent {
			continue
		}
		shared.AsLine().Normalize()
		assert.Equal(t, tt.desired.GetID(), shared.GetID(), tt.name)
		assert.InDelta(t, tt.desired.GetA(), shared.AsLine().GetA(), utils.StandardCompare, tt.name)
		assert.InDelta(t, tt.desired.GetB(), shared.AsLine().GetB(), utils.StandardCompare, tt.name)
		assert.InDelta(t, tt.desired.GetC(), shared.AsLine().GetC(), utils.StandardCompare, tt.name)
	}
}
