package solver

import (
	"math"
	"math/big"
	"testing"

	"github.com/marcuswu/dlineate/internal/accessors"
	"github.com/marcuswu/dlineate/internal/constraint"
	el "github.com/marcuswu/dlineate/internal/element"
	"github.com/marcuswu/dlineate/utils"
	"github.com/stretchr/testify/assert"
)

func TestSolveAngleConstraint(t *testing.T) {
	ea := accessors.NewElementRepository()
	ea.AddElement(el.NewSketchLine(0, big.NewFloat(0), big.NewFloat(1), big.NewFloat(0)))
	ea.AddElement(el.NewSketchLine(1, big.NewFloat(-0.951057), big.NewFloat(0.309017), big.NewFloat(0)))
	ea.AddElement(el.NewSketchLine(2, big.NewFloat(-0.506732), big.NewFloat(-0.862104), big.NewFloat(0)))
	ea.AddElement(el.NewSketchLine(3, big.NewFloat(-0.506732), big.NewFloat(-0.862104), big.NewFloat(0)))
	ea.AddElement(el.NewSketchLine(4, big.NewFloat(1.5), big.NewFloat(0.3), big.NewFloat(0.1)))
	ea.AddElement(el.NewSketchLine(5, big.NewFloat(0.3), big.NewFloat(1.5), big.NewFloat(-0.1)))
	ea.AddElement(el.NewSketchPoint(6, big.NewFloat(0), big.NewFloat(0)))
	ea.AddElement(el.NewSketchPoint(7, big.NewFloat(1), big.NewFloat(1)))
	tests := []struct {
		name       string
		constraint *constraint.Constraint
		solveFor   uint
		solveState SolveState
	}{
		{
			"Test -108º constraint",
			constraint.NewConstraint(0, constraint.Angle, 0, 1, big.NewFloat(-(108.0/180.0)*math.Pi), false),
			0,
			Solved,
		},
		{
			"Test -108º constraint 2",
			constraint.NewConstraint(0, constraint.Angle, 2, 3, big.NewFloat((108.0/180.0)*math.Pi), false),
			3,
			Solved,
		},
		{
			"Test reverse rotation",
			constraint.NewConstraint(0, constraint.Angle, 4, 5, big.NewFloat((70.0/180.0)*math.Pi), false),
			5,
			Solved,
		},
		{
			"Test incorrect Constraint",
			constraint.NewConstraint(0, constraint.Distance, 6, 7, big.NewFloat(2), false),
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
			var v1, v2 big.Float
			v1.Abs(&tt.constraint.Value)
			v2.Abs(first.AngleToLine(second))
			assert.Equal(t, 0, utils.StandardBigFloatCompare(&v1, &v2), tt.name)
		} else {
			assert.Nil(t, newLine, tt.name)
		}
		assert.Equal(t, tt.solveState, status, tt.name)
	}
}

func TestLineFromPointLine(t *testing.T) {
	ea := accessors.NewElementRepository()
	ea.AddElement(el.NewSketchLine(0, big.NewFloat(1.5), big.NewFloat(0.3), big.NewFloat(0.1)))
	ea.AddElement(el.NewSketchLine(1, big.NewFloat(0.3), big.NewFloat(1.5), big.NewFloat(-0.1)))
	ea.AddElement(el.NewSketchPoint(2, big.NewFloat(1), big.NewFloat(1)))
	ea.AddElement(el.NewSketchPoint(3, big.NewFloat(-1), big.NewFloat(-1)))
	ca := accessors.NewConstraintRepository()
	c0 := constraint.NewConstraint(0, constraint.Angle, 0, 1, big.NewFloat((70.0/180.0)*math.Pi), false)
	c1 := constraint.NewConstraint(1, constraint.Distance, 2, 1, big.NewFloat(1), false)
	c2 := constraint.NewConstraint(2, constraint.Angle, 0, 1, big.NewFloat((70.0/180.0)*math.Pi), false)
	c3 := constraint.NewConstraint(3, constraint.Distance, 3, 1, big.NewFloat(1), false)
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
			el.NewSketchLine(1, big.NewFloat(0.1510894582), big.NewFloat(0.9885200937), big.NewFloat(-0.1396095519)),
			Solved,
		},
		{
			"test 2",
			c2,
			c3,
			el.NewSketchLine(1, big.NewFloat(0.1510894582), big.NewFloat(0.9885200937), big.NewFloat(0.1396095519)),
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
			t.Logf("LineFromPointLine expected line A: %s, found %s\n", tt.desired.GetA().String(), newLine.GetA().String())
			t.Logf("LineFromPointLine expected line B: %s, found %s\n", tt.desired.GetB().String(), newLine.GetB().String())
			t.Logf("LineFromPointLine expected line C: %s, found %s\n", tt.desired.GetC().String(), newLine.GetC().String())
			assert.Equal(t, 0, utils.StandardBigFloatCompare(tt.desired.GetA(), newLine.GetA()), tt.name)
			assert.Equal(t, 0, utils.StandardBigFloatCompare(tt.desired.GetB(), newLine.GetB()), tt.name)
			assert.Equal(t, 0, utils.StandardBigFloatCompare(tt.desired.GetC(), newLine.GetC()), tt.name)
		}
	}
}

func TestLineFromPoints(t *testing.T) {
	ea := accessors.NewElementRepository()
	ea.AddElement(el.NewSketchPoint(0, big.NewFloat(1.5), big.NewFloat(0.3)))
	ea.AddElement(el.NewSketchPoint(1, big.NewFloat(0.3), big.NewFloat(1.5)))
	ea.AddElement(el.NewSketchPoint(2, big.NewFloat(1), big.NewFloat(1)))
	ea.AddElement(el.NewSketchLine(3, big.NewFloat(1.1), big.NewFloat(0.1), big.NewFloat(0.1)))
	ea.AddElement(el.NewSketchLine(4, big.NewFloat(0.3), big.NewFloat(1.5), big.NewFloat(-0.1)))
	ea.AddElement(el.NewSketchLine(5, big.NewFloat(-1), big.NewFloat(-1), big.NewFloat(0.0)))
	ea.AddElement(el.NewSketchPoint(6, big.NewFloat(1.5), big.NewFloat(0.3)))
	ea.AddElement(el.NewSketchLine(7, big.NewFloat(0.3), big.NewFloat(1.5), big.NewFloat(-0.1)))
	ea.AddElement(el.NewSketchPoint(8, big.NewFloat(1), big.NewFloat(1)))
	ea.AddElement(el.NewSketchPoint(9, big.NewFloat(1), big.NewFloat(1)))
	ea.AddElement(el.NewSketchLine(10, big.NewFloat(0.3), big.NewFloat(1.5), big.NewFloat(-0.1)))
	ea.AddElement(el.NewSketchPoint(11, big.NewFloat(1.5), big.NewFloat(0.3)))
	ea.AddElement(el.NewSketchPoint(12, big.NewFloat(1.5), big.NewFloat(0.3)))
	ea.AddElement(el.NewSketchLine(13, big.NewFloat(0.3), big.NewFloat(1.5), big.NewFloat(-0.1)))
	ea.AddElement(el.NewSketchPoint(14, big.NewFloat(1), big.NewFloat(1)))
	ea.AddElement(el.NewSketchPoint(15, big.NewFloat(1.5), big.NewFloat(0.3)))                    // (1.5, 0.3)
	ea.AddElement(el.NewSketchLine(16, big.NewFloat(0.3), big.NewFloat(1.5), big.NewFloat(-0.1))) // 0.196116x + 0.980580y - 0.065372 = 0 normalized
	ea.AddElement(el.NewSketchPoint(17, big.NewFloat(-1), big.NewFloat(1)))                       // (-1, 1)
	ea.AddElement(el.NewSketchPoint(18, big.NewFloat(1.5), big.NewFloat(0.3)))
	ea.AddElement(el.NewSketchLine(19, big.NewFloat(0.3), big.NewFloat(1.5), big.NewFloat(-0.1)))
	ea.AddElement(el.NewSketchPoint(20, big.NewFloat(-1), big.NewFloat(0)))
	ca := accessors.NewConstraintRepository()
	c0 := constraint.NewConstraint(0, constraint.Distance, 0, 1, big.NewFloat(1), false)
	c1 := constraint.NewConstraint(1, constraint.Distance, 2, 1, big.NewFloat(1), false)
	c2 := constraint.NewConstraint(2, constraint.Distance, 3, 4, big.NewFloat((70.0/180.0)*math.Pi), false)
	c3 := constraint.NewConstraint(3, constraint.Distance, 5, 4, big.NewFloat(1), false)
	c4 := constraint.NewConstraint(4, constraint.Distance, 6, 7, big.NewFloat(0), false)
	c5 := constraint.NewConstraint(5, constraint.Distance, 8, 7, big.NewFloat(0), false)
	c6 := constraint.NewConstraint(6, constraint.Distance, 9, 10, big.NewFloat(0), false)
	c7 := constraint.NewConstraint(7, constraint.Distance, 11, 10, big.NewFloat(0), false)
	c8 := constraint.NewConstraint(8, constraint.Distance, 12, 13, big.NewFloat(1), false)
	c9 := constraint.NewConstraint(9, constraint.Distance, 14, 13, big.NewFloat(1), false)
	c10 := constraint.NewConstraint(10, constraint.Distance, 15, 16, big.NewFloat(0.25), false)
	c11 := constraint.NewConstraint(11, constraint.Distance, 17, 16, big.NewFloat(0.25), false)
	c12 := constraint.NewConstraint(12, constraint.Distance, 18, 19, big.NewFloat(0.25), false)
	c13 := constraint.NewConstraint(13, constraint.Distance, 20, 19, big.NewFloat(0.25), false)
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
		{"Can't find points", c2, c3, el.NewSketchLine(4, big.NewFloat(0.3), big.NewFloat(1.5), big.NewFloat(-0.1)), NonConvergent},
		{"Coincident to both points", c4, c5, el.NewSketchLine(7, big.NewFloat(0.7), big.NewFloat(0.5), big.NewFloat(-1.2)), Solved},
		{"Coincident to both points alt slope", c6, c7, el.NewSketchLine(10, big.NewFloat(0.7), big.NewFloat(0.5), big.NewFloat(-1.2)), Solved},
		{"Distance between points is large", c8, c9, el.NewSketchLine(13, big.NewFloat(0.8137334712), big.NewFloat(0.5812381937), big.NewFloat(-0.3949716649)), Solved},
		{"Solve by tangent external", c10, c11, el.NewSketchLine(16, big.NewFloat(0.2696299255), big.NewFloat(0.9629640197), big.NewFloat(-0.4433340942)), Solved},
		{"Solve by tangent internal", c12, c13, el.NewSketchLine(19, big.NewFloat(-0.08038836581), big.NewFloat(-0.9967636182), big.NewFloat(0.1696116342)), Solved},
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
			t.Logf("LineFromPoints expected line A: %s, found %s\n", tt.desired.GetA().String(), newLine.GetA().String())
			t.Logf("LineFromPoints expected line B: %s, found %s\n", tt.desired.GetB().String(), newLine.GetB().String())
			t.Logf("LineFromPoints expected line C: %s, found %s\n", tt.desired.GetC().String(), newLine.GetC().String())
			assert.Equal(t, 0, utils.StandardBigFloatCompare(tt.desired.GetA(), newLine.GetA()), tt.name)
			assert.Equal(t, 0, utils.StandardBigFloatCompare(tt.desired.GetB(), newLine.GetB()), tt.name)
			assert.Equal(t, 0, utils.StandardBigFloatCompare(tt.desired.GetC(), newLine.GetC()), tt.name)
		}

		if tt.state == Solved {
			assert.True(t, ca.IsMet(tt.c1.GetID(), -1, ea), tt.name)
			assert.True(t, ca.IsMet(tt.c2.GetID(), -1, ea), tt.name)
		}
	}
}

func TestMoveLineToPoint(t *testing.T) {
	ea := accessors.NewElementRepository()
	ea.AddElement(el.NewSketchPoint(0, big.NewFloat(1.5), big.NewFloat(0.3)))
	ea.AddElement(el.NewSketchPoint(1, big.NewFloat(0.3), big.NewFloat(1.5)))
	ea.AddElement(el.NewSketchPoint(2, big.NewFloat(1.5), big.NewFloat(0.3)))
	ea.AddElement(el.NewSketchPoint(3, big.NewFloat(0.3), big.NewFloat(1.5)))
	ea.AddElement(el.NewSketchPoint(4, big.NewFloat(1.5), big.NewFloat(0.3)))
	ea.AddElement(el.NewSketchLine(5, big.NewFloat(0.3), big.NewFloat(1.5), big.NewFloat(1)))
	ea.AddElement(el.NewSketchLine(6, big.NewFloat(0.3), big.NewFloat(1.5), big.NewFloat(1)))
	ea.AddElement(el.NewSketchPoint(7, big.NewFloat(1.5), big.NewFloat(-2)))
	ca := accessors.NewConstraintRepository()
	c0 := constraint.NewConstraint(0, constraint.Angle, 0, 1, big.NewFloat(1), false)
	c1 := constraint.NewConstraint(1, constraint.Distance, 2, 3, big.NewFloat(1), false)
	c2 := constraint.NewConstraint(2, constraint.Distance, 4, 5, big.NewFloat(1), false)
	c3 := constraint.NewConstraint(3, constraint.Distance, 6, 7, big.NewFloat(1), false)
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
		{"The constraint should be met 1", c2, el.NewSketchLine(5, big.NewFloat(0.1961161351), big.NewFloat(0.9805806757), big.NewFloat(0.4116515946)), Solved},
		{"The constraint should be met 2", c3, el.NewSketchLine(6, big.NewFloat(0.1961161351), big.NewFloat(0.9805806757), big.NewFloat(2.666987149)), Solved},
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
		t.Logf("MoveLineToPoint expected line A: %s, found %s\n", tt.desired.GetA().String(), newLine.GetA().String())
		t.Logf("MoveLineToPoint expected line B: %s, found %s\n", tt.desired.GetB().String(), newLine.GetB().String())
		t.Logf("MoveLineToPoint expected line C: %s, found %s\n", tt.desired.GetC().String(), newLine.GetC().String())
		assert.Equal(t, 0, utils.StandardBigFloatCompare(tt.desired.GetA(), newLine.GetA()), tt.name)
		assert.Equal(t, 0, utils.StandardBigFloatCompare(tt.desired.GetB(), newLine.GetB()), tt.name)
		assert.Equal(t, 0, utils.StandardBigFloatCompare(tt.desired.GetC(), newLine.GetC()), tt.name)
	}
}

func TestLineResult(t *testing.T) {
	ea := accessors.NewElementRepository()
	ea.AddElement(el.NewSketchPoint(0, big.NewFloat(1.5), big.NewFloat(0.3)))
	ea.AddElement(el.NewSketchLine(1, big.NewFloat(0.3), big.NewFloat(1.5), big.NewFloat(-0.1)))
	ea.AddElement(el.NewSketchPoint(2, big.NewFloat(1), big.NewFloat(1)))
	ea.AddElement(el.NewSketchLine(3, big.NewFloat(1.5), big.NewFloat(0.3), big.NewFloat(0.1)))
	ea.AddElement(el.NewSketchLine(4, big.NewFloat(0.3), big.NewFloat(1.5), big.NewFloat(-0.1)))
	ea.AddElement(el.NewSketchPoint(5, big.NewFloat(1), big.NewFloat(1)))
	ca := accessors.NewConstraintRepository()
	c0 := constraint.NewConstraint(0, constraint.Distance, 0, 1, big.NewFloat(0), false)
	c1 := constraint.NewConstraint(1, constraint.Distance, 2, 1, big.NewFloat(0), false)
	c2 := constraint.NewConstraint(2, constraint.Angle, 3, 4, big.NewFloat((70.0/180.0)*math.Pi), false)
	c3 := constraint.NewConstraint(3, constraint.Distance, 5, 4, big.NewFloat(1), false)
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
		{"Test Line From Points", c0, c1, el.NewSketchLine(1, big.NewFloat(0.7), big.NewFloat(0.5), big.NewFloat(-1.2)), Solved},
		{"Test Line From Point Line", c2, c3, el.NewSketchLine(4, big.NewFloat(0.1510894582), big.NewFloat(0.9885200937), big.NewFloat(-0.1396095519)), Solved},
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
			t.Logf("LineResult expected line A: %s, found %s\n", tt.desired.GetA().String(), newLine.GetA().String())
			t.Logf("LineResult expected line B: %s, found %s\n", tt.desired.GetB().String(), newLine.GetB().String())
			t.Logf("LineResult expected line C: %s, found %s\n", tt.desired.GetC().String(), newLine.GetC().String())
			assert.Equal(t, 0, utils.StandardBigFloatCompare(tt.desired.GetA(), newLine.GetA()), tt.name)
			assert.Equal(t, 0, utils.StandardBigFloatCompare(tt.desired.GetB(), newLine.GetB()), tt.name)
			assert.Equal(t, 0, utils.StandardBigFloatCompare(tt.desired.GetC(), newLine.GetC()), tt.name)
		}

		if tt.state == Solved {
			assert.True(t, ca.IsMet(tt.c1.GetID(), -1, ea), tt.name)
			assert.True(t, ca.IsMet(tt.c2.GetID(), -1, ea), tt.name)
		}
	}
}

func TestSolveForLine(t *testing.T) {
	ea := accessors.NewElementRepository()
	ea.AddElement(el.NewSketchPoint(0, big.NewFloat(1.5), big.NewFloat(0.3)))
	ea.AddElement(el.NewSketchPoint(1, big.NewFloat(0.3), big.NewFloat(1.5)))
	ea.AddElement(el.NewSketchPoint(2, big.NewFloat(1), big.NewFloat(1)))
	ea.AddElement(el.NewSketchPoint(3, big.NewFloat(1.5), big.NewFloat(0.3)))
	ea.AddElement(el.NewSketchLine(4, big.NewFloat(0.3), big.NewFloat(1.5), big.NewFloat(-0.1)))
	ea.AddElement(el.NewSketchPoint(5, big.NewFloat(1), big.NewFloat(1)))
	ea.AddElement(el.NewSketchLine(6, big.NewFloat(1.5), big.NewFloat(0.3), big.NewFloat(0.1)))
	ea.AddElement(el.NewSketchLine(7, big.NewFloat(0.3), big.NewFloat(1.5), big.NewFloat(-0.1)))
	ea.AddElement(el.NewSketchPoint(8, big.NewFloat(1), big.NewFloat(1)))
	ca := accessors.NewConstraintRepository()
	c0 := constraint.NewConstraint(0, constraint.Distance, 0, 1, big.NewFloat(0), false)
	c1 := constraint.NewConstraint(1, constraint.Distance, 2, 1, big.NewFloat(0), false)
	c2 := constraint.NewConstraint(2, constraint.Distance, 3, 4, big.NewFloat(0), false)
	c3 := constraint.NewConstraint(3, constraint.Distance, 5, 4, big.NewFloat(0), false)
	c4 := constraint.NewConstraint(4, constraint.Angle, 6, 7, big.NewFloat((70.0/180.0)*math.Pi), false)
	c5 := constraint.NewConstraint(5, constraint.Distance, 8, 7, big.NewFloat(1), false)
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
		{"Test Line From Points", c2, c3, el.NewSketchLine(4, big.NewFloat(0.7), big.NewFloat(0.5), big.NewFloat(-1.2)), Solved},
		{"Test Line From Point Line", c4, c5, el.NewSketchLine(7, big.NewFloat(0.1510894582), big.NewFloat(0.9885200937), big.NewFloat(-0.1396095510)), Solved},
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
		t.Logf("SolveForLine expected line A: %s, found %s\n", tt.desired.GetA().String(), shared.AsLine().GetA().String())
		t.Logf("SolveForLine expected line B: %s, found %s\n", tt.desired.GetB().String(), shared.AsLine().GetB().String())
		t.Logf("SolveForLine expected line C: %s, found %s\n", tt.desired.GetC().String(), shared.AsLine().GetC().String())
		assert.Equal(t, 0, utils.StandardBigFloatCompare(tt.desired.GetA(), shared.AsLine().GetA()), tt.name)
		assert.Equal(t, 0, utils.StandardBigFloatCompare(tt.desired.GetB(), shared.AsLine().GetB()), tt.name)
		assert.Equal(t, 0, utils.StandardBigFloatCompare(tt.desired.GetC(), shared.AsLine().GetC()), tt.name)
	}
}
