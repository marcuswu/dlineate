package solver

import (
	"math"
	"testing"

	"github.com/marcuswu/dlineate/internal/accessors"
	"github.com/marcuswu/dlineate/internal/constraint"
	el "github.com/marcuswu/dlineate/internal/element"
	"github.com/marcuswu/dlineate/utils"
	"github.com/stretchr/testify/assert"
)

func TestSolveAngleConstraint(t *testing.T) {
	ea := accessors.NewElementRepository()
	ea.AddElement(el.NewSketchLine(0, 0, 1, 0))
	ea.AddElement(el.NewSketchLine(1, -0.951057, 0.309017, 0))
	ea.AddElement(el.NewSketchLine(2, -0.506732, -0.862104, 0))
	ea.AddElement(el.NewSketchLine(3, -0.506732, -0.862104, 0))
	ea.AddElement(el.NewSketchLine(4, 1.5, 0.3, 0.1))
	ea.AddElement(el.NewSketchLine(5, 0.3, 1.5, -0.1))
	ea.AddElement(el.NewSketchPoint(6, 0, 0))
	ea.AddElement(el.NewSketchPoint(7, 1, 1))
	tests := []struct {
		name       string
		constraint *constraint.Constraint
		solveFor   uint
		solveState SolveState
	}{
		{
			"Test -108º constraint",
			constraint.NewConstraint(0, constraint.Angle, 0, 1, -(108.0/180.0)*math.Pi, false),
			0,
			Solved,
		},
		{
			"Test -108º constraint 2",
			constraint.NewConstraint(0, constraint.Angle, 2, 3, (108.0/180.0)*math.Pi, false),
			3,
			Solved,
		},
		{
			"Test reverse rotation",
			constraint.NewConstraint(0, constraint.Angle, 4, 5, (70.0/180.0)*math.Pi, false),
			5,
			Solved,
		},
		{
			"Test incorrect Constraint",
			constraint.NewConstraint(0, constraint.Distance, 6, 7, 2, false),
			0,
			NonConvergent,
		},
	}
	for _, tt := range tests {
		newLine, status := SolveAngleConstraint(-1, ea, tt.constraint, tt.solveFor)
		if tt.solveState == Solved {
			e1, _ := ea.GetElement(-1, tt.constraint.Element1)
			e2, _ := ea.GetElement(-1, tt.constraint.Element2)
			first := e1.AsLine()
			second := e2.AsLine()
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
	ea := accessors.NewElementRepository()
	ea.AddElement(el.NewSketchLine(0, 1.5, 0.3, 0.1))
	ea.AddElement(el.NewSketchLine(1, 0.3, 1.5, -0.1))
	ea.AddElement(el.NewSketchPoint(2, 1, 1))
	ea.AddElement(el.NewSketchPoint(3, -1, -1))
	ca := accessors.NewConstraintRepository()
	c0 := constraint.NewConstraint(0, constraint.Angle, 0, 1, (70.0/180.0)*math.Pi, false)
	c1 := constraint.NewConstraint(1, constraint.Distance, 2, 1, 1, false)
	c2 := constraint.NewConstraint(2, constraint.Angle, 0, 1, (70.0/180.0)*math.Pi, false)
	c3 := constraint.NewConstraint(3, constraint.Distance, 3, 1, 1, false)
	ca.AddConstraint(c0)
	ca.AddConstraint(c1)
	ca.AddConstraint(c2)
	ca.AddConstraint(c3)
	tests := []struct {
		name    string
		c1      *constraint.Constraint
		c2      *constraint.Constraint
		desired *el.SketchLine
		state   SolveState
	}{
		{
			"test 1",
			c0,
			c1,
			el.NewSketchLine(1, 0.151089, 0.988520, -0.139610),
			Solved,
		},
		{
			"test 2",
			c2,
			c3,
			el.NewSketchLine(1, 0.151089, 0.988520, 0.139610),
			Solved,
		},
	}
	for _, tt := range tests {
		newLine, state := LineFromPointLine(-1, ea, tt.c1, tt.c2)
		assert.Equal(t, state, tt.state, tt.name)
		if tt.state == NonConvergent {
			assert.Nil(t, newLine)
		} else {
			c1Line, _ := ea.GetElement(-1, newLine.GetID())
			c2Line, _ := ea.GetElement(-1, newLine.GetID())
			c1Line.AsLine().SetA(newLine.GetA())
			c1Line.AsLine().SetB(newLine.GetB())
			c1Line.AsLine().SetC(newLine.GetC())
			c2Line.AsLine().SetA(newLine.GetA())
			c2Line.AsLine().SetB(newLine.GetB())
			c2Line.AsLine().SetC(newLine.GetC())
			assert.True(t, ca.IsMet(tt.c1.GetID(), -1, ea), tt.name)
			assert.True(t, ca.IsMet(tt.c2.GetID(), -1, ea), tt.name)
			assert.Equal(t, tt.desired.GetID(), newLine.GetID(), tt.name)
			assert.InDelta(t, tt.desired.GetA(), newLine.GetA(), utils.StandardCompare, tt.name)
			assert.InDelta(t, tt.desired.GetB(), newLine.GetB(), utils.StandardCompare, tt.name)
			assert.InDelta(t, tt.desired.GetC(), newLine.GetC(), utils.StandardCompare, tt.name)
		}
	}
}

func TestLineFromPoints(t *testing.T) {
	ea := accessors.NewElementRepository()
	ea.AddElement(el.NewSketchPoint(0, 1.5, 0.3))
	ea.AddElement(el.NewSketchPoint(1, 0.3, 1.5))
	ea.AddElement(el.NewSketchPoint(2, 1, 1))
	ea.AddElement(el.NewSketchLine(3, 1.1, 0.1, 0.1))
	ea.AddElement(el.NewSketchLine(4, 0.3, 1.5, -0.1))
	ea.AddElement(el.NewSketchLine(5, -1, -1, 0.0))
	ea.AddElement(el.NewSketchPoint(6, 1.5, 0.3))
	ea.AddElement(el.NewSketchLine(7, 0.3, 1.5, -0.1))
	ea.AddElement(el.NewSketchPoint(8, 1, 1))
	ea.AddElement(el.NewSketchPoint(9, 1, 1))
	ea.AddElement(el.NewSketchLine(10, 0.3, 1.5, -0.1))
	ea.AddElement(el.NewSketchPoint(11, 1.5, 0.3))
	ea.AddElement(el.NewSketchPoint(12, 1.5, 0.3))
	ea.AddElement(el.NewSketchLine(13, 0.3, 1.5, -0.1))
	ea.AddElement(el.NewSketchPoint(14, 1, 1))
	ea.AddElement(el.NewSketchPoint(15, 1.5, 0.3))      // (1.5, 0.3)
	ea.AddElement(el.NewSketchLine(16, 0.3, 1.5, -0.1)) // 0.196116x + 0.980580y - 0.065372 = 0 normalized
	ea.AddElement(el.NewSketchPoint(17, -1, 1))         // (-1, 1)
	ea.AddElement(el.NewSketchPoint(18, 1.5, 0.3))
	ea.AddElement(el.NewSketchLine(19, 0.3, 1.5, -0.1))
	ea.AddElement(el.NewSketchPoint(20, -1, 0))
	ca := accessors.NewConstraintRepository()
	c0 := constraint.NewConstraint(0, constraint.Distance, 0, 1, 1, false)
	c1 := constraint.NewConstraint(1, constraint.Distance, 2, 1, 1, false)
	c2 := constraint.NewConstraint(2, constraint.Distance, 3, 4, (70.0/180.0)*math.Pi, false)
	c3 := constraint.NewConstraint(3, constraint.Distance, 5, 4, 1, false)
	c4 := constraint.NewConstraint(4, constraint.Distance, 6, 7, 0, false)
	c5 := constraint.NewConstraint(5, constraint.Distance, 8, 7, 0, false)
	c6 := constraint.NewConstraint(6, constraint.Distance, 9, 10, 0, false)
	c7 := constraint.NewConstraint(7, constraint.Distance, 11, 10, 0, false)
	c8 := constraint.NewConstraint(8, constraint.Distance, 12, 13, 1, false)
	c9 := constraint.NewConstraint(9, constraint.Distance, 14, 13, 1, false)
	c10 := constraint.NewConstraint(10, constraint.Distance, 15, 16, 0.25, false)
	c11 := constraint.NewConstraint(11, constraint.Distance, 17, 16, 0.25, false)
	c12 := constraint.NewConstraint(12, constraint.Distance, 18, 19, 0.25, false)
	c13 := constraint.NewConstraint(13, constraint.Distance, 20, 19, 0.25, false)
	ca.AddConstraint(c0)
	ca.AddConstraint(c1)
	ca.AddConstraint(c2)
	ca.AddConstraint(c3)
	ca.AddConstraint(c4)
	ca.AddConstraint(c5)
	ca.AddConstraint(c6)
	ca.AddConstraint(c7)
	ca.AddConstraint(c8)
	ca.AddConstraint(c9)
	ca.AddConstraint(c10)
	ca.AddConstraint(c11)
	ca.AddConstraint(c12)
	ca.AddConstraint(c13)
	tests := []struct {
		name    string
		c1      *constraint.Constraint
		c2      *constraint.Constraint
		desired *el.SketchLine
		state   SolveState
	}{
		{"Can't find line", c0, c1, nil, NonConvergent},
		{"Can't find points", c2, c3, el.NewSketchLine(4, 0.3, 1.5, -0.1), NonConvergent},
		{"Coincident to both points", c4, c5, el.NewSketchLine(7, 0.7, 0.5, -1.2), Solved},
		{"Coincident to both points alt slope", c6, c7, el.NewSketchLine(10, 0.7, 0.5, -1.2), Solved},
		{"Distance between points is large", c8, c9, el.NewSketchLine(13, 0.813733, 0.581238, -0.394971), Solved},
		{"Solve by tangent external", c10, c11, el.NewSketchLine(16, 0.269630, 0.962964, -0.443334), Solved},
		{"Solve by tangent internal", c12, c13, el.NewSketchLine(19, -0.080388, -0.996764, 0.169612), Solved},
	}
	for _, tt := range tests {
		newLine, state := LineFromPoints(-1, ea, tt.c1, tt.c2)
		assert.Equal(t, tt.state, state, tt.name)
		if tt.desired == nil {
			assert.Nil(t, newLine, tt.name)
		} else {
			newLine.Normalize()
			c1Line, _ := ea.GetElement(-1, newLine.GetID())
			c2Line, _ := ea.GetElement(-1, newLine.GetID())
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
			assert.True(t, ca.IsMet(tt.c1.GetID(), -1, ea), tt.name)
			assert.True(t, ca.IsMet(tt.c2.GetID(), -1, ea), tt.name)
		}
	}
}

func TestMoveLineToPoint(t *testing.T) {
	ea := accessors.NewElementRepository()
	ea.AddElement(el.NewSketchPoint(0, 1.5, 0.3))
	ea.AddElement(el.NewSketchPoint(1, 0.3, 1.5))
	ea.AddElement(el.NewSketchPoint(2, 1.5, 0.3))
	ea.AddElement(el.NewSketchPoint(3, 0.3, 1.5))
	ea.AddElement(el.NewSketchPoint(4, 1.5, 0.3))
	ea.AddElement(el.NewSketchLine(5, 0.3, 1.5, 1))
	ea.AddElement(el.NewSketchLine(6, 0.3, 1.5, 1))
	ea.AddElement(el.NewSketchPoint(7, 1.5, -2))
	ca := accessors.NewConstraintRepository()
	c0 := constraint.NewConstraint(0, constraint.Angle, 0, 1, 1, false)
	c1 := constraint.NewConstraint(1, constraint.Distance, 2, 3, 1, false)
	c2 := constraint.NewConstraint(2, constraint.Distance, 4, 5, 1, false)
	c3 := constraint.NewConstraint(3, constraint.Distance, 6, 7, 1, false)
	ca.AddConstraint(c0)
	ca.AddConstraint(c1)
	ca.AddConstraint(c2)
	ca.AddConstraint(c3)
	tests := []struct {
		name    string
		c1      *constraint.Constraint
		desired *el.SketchLine
		state   SolveState
	}{
		{"The constraint should be a distance constraint", c0, nil, NonConvergent},
		{"The constraint should have a point and a line", c1, nil, NonConvergent},
		{"The constraint should be met 1", c2, el.NewSketchLine(5, 0.3, 1.5, 0.629706), Solved},
		{"The constraint should be met 2", c3, el.NewSketchLine(6, 0.196116, 0.980581, 2.666987), Solved},
	}
	for _, tt := range tests {
		state := MoveLineToPoint(ea, tt.c1)
		assert.Equal(t, tt.state, state, tt.name)
		if state != Solved {
			continue
		}
		assert.True(t, ca.IsMet(tt.c1.GetID(), -1, ea), tt.name)
		e, _ := ea.GetElement(-1, tt.c1.Element1)
		newLine := e.AsLine()
		if newLine == nil {
			e, _ := ea.GetElement(-1, tt.c1.Element2)
			newLine = e.AsLine()
		}
		assert.Equal(t, tt.desired.GetID(), newLine.GetID(), tt.name)
		assert.InDelta(t, tt.desired.GetA(), newLine.GetA(), utils.StandardCompare, tt.name)
		assert.InDelta(t, tt.desired.GetB(), newLine.GetB(), utils.StandardCompare, tt.name)
		assert.InDelta(t, tt.desired.GetC(), newLine.GetC(), utils.StandardCompare, tt.name)
	}
}

func TestLineResult(t *testing.T) {
	ea := accessors.NewElementRepository()
	ea.AddElement(el.NewSketchPoint(0, 1.5, 0.3))
	ea.AddElement(el.NewSketchLine(1, 0.3, 1.5, -0.1))
	ea.AddElement(el.NewSketchPoint(2, 1, 1))
	ea.AddElement(el.NewSketchLine(3, 1.5, 0.3, 0.1))
	ea.AddElement(el.NewSketchLine(4, 0.3, 1.5, -0.1))
	ea.AddElement(el.NewSketchPoint(5, 1, 1))
	ca := accessors.NewConstraintRepository()
	c0 := constraint.NewConstraint(0, constraint.Distance, 0, 1, 0, false)
	c1 := constraint.NewConstraint(1, constraint.Distance, 2, 1, 0, false)
	c2 := constraint.NewConstraint(2, constraint.Angle, 3, 4, (70.0/180.0)*math.Pi, false)
	c3 := constraint.NewConstraint(3, constraint.Distance, 5, 4, 1, false)
	ca.AddConstraint(c0)
	ca.AddConstraint(c1)
	ca.AddConstraint(c2)
	ca.AddConstraint(c3)
	tests := []struct {
		name    string
		c1      *constraint.Constraint
		c2      *constraint.Constraint
		desired *el.SketchLine
		state   SolveState
	}{
		{"Test Line From Points", c0, c1, el.NewSketchLine(1, 0.7, 0.5, -1.2), Solved},
		{"Test Line From Point Line", c2, c3, el.NewSketchLine(4, 0.151089, 0.988520, -0.139610), Solved},
	}
	for _, tt := range tests {
		newLine, state := LineResult(-1, ea, tt.c1, tt.c2)
		assert.Equal(t, state, tt.state, tt.name)
		if tt.desired == nil {
			assert.Nil(t, newLine, tt.name)
		} else {
			newLine.Normalize()
			c1Line, _ := ea.GetElement(-1, newLine.GetID())
			c2Line, _ := ea.GetElement(-1, newLine.GetID())
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
			assert.True(t, ca.IsMet(tt.c1.GetID(), -1, ea), tt.name)
			assert.True(t, ca.IsMet(tt.c2.GetID(), -1, ea), tt.name)
		}
	}
}

func TestSolveForLine(t *testing.T) {
	ea := accessors.NewElementRepository()
	ea.AddElement(el.NewSketchPoint(0, 1.5, 0.3))
	ea.AddElement(el.NewSketchPoint(1, 0.3, 1.5))
	ea.AddElement(el.NewSketchPoint(2, 1, 1))
	ea.AddElement(el.NewSketchPoint(3, 1.5, 0.3))
	ea.AddElement(el.NewSketchLine(4, 0.3, 1.5, -0.1))
	ea.AddElement(el.NewSketchPoint(5, 1, 1))
	ea.AddElement(el.NewSketchLine(6, 1.5, 0.3, 0.1))
	ea.AddElement(el.NewSketchLine(7, 0.3, 1.5, -0.1))
	ea.AddElement(el.NewSketchPoint(8, 1, 1))
	ca := accessors.NewConstraintRepository()
	c0 := constraint.NewConstraint(0, constraint.Distance, 0, 1, 0, false)
	c1 := constraint.NewConstraint(1, constraint.Distance, 2, 1, 0, false)
	c2 := constraint.NewConstraint(2, constraint.Distance, 3, 4, 0, false)
	c3 := constraint.NewConstraint(3, constraint.Distance, 5, 4, 0, false)
	c4 := constraint.NewConstraint(4, constraint.Angle, 6, 7, (70.0/180.0)*math.Pi, false)
	c5 := constraint.NewConstraint(5, constraint.Distance, 8, 7, 1, false)
	ca.AddConstraint(c0)
	ca.AddConstraint(c1)
	ca.AddConstraint(c2)
	ca.AddConstraint(c3)
	ca.AddConstraint(c4)
	ca.AddConstraint(c5)
	tests := []struct {
		name    string
		c1      *constraint.Constraint
		c2      *constraint.Constraint
		desired *el.SketchLine
		state   SolveState
	}{
		{"Test Nonconvergent", c0, c1, nil, NonConvergent},
		{"Test Line From Points", c2, c3, el.NewSketchLine(4, 0.7, 0.5, -1.2), Solved},
		{"Test Line From Point Line", c4, c5, el.NewSketchLine(7, 0.151089, 0.988520, -0.139610), Solved},
	}
	for _, tt := range tests {
		state := SolveForLine(-1, ea, tt.c1, tt.c2)
		assert.Equal(t, state, tt.state, tt.name)
		if tt.state == Solved {
			assert.True(t, ca.IsMet(tt.c1.GetID(), -1, ea), tt.name)
			assert.True(t, ca.IsMet(tt.c2.GetID(), -1, ea), tt.name)
		}
		e, ok := tt.c1.Shared(tt.c2)
		if !ok || tt.state == NonConvergent {
			continue
		}
		shared, _ := ea.GetElement(-1, e)
		shared.AsLine().Normalize()
		assert.Equal(t, tt.desired.GetID(), shared.GetID(), tt.name)
		assert.InDelta(t, tt.desired.GetA(), shared.AsLine().GetA(), utils.StandardCompare, tt.name)
		assert.InDelta(t, tt.desired.GetB(), shared.AsLine().GetB(), utils.StandardCompare, tt.name)
		assert.InDelta(t, tt.desired.GetC(), shared.AsLine().GetC(), utils.StandardCompare, tt.name)
	}
}
