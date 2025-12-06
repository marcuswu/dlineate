package numeric

import (
	"math/big"
	"os"
	"testing"

	"github.com/marcuswu/dlineate/internal/constraint"
	el "github.com/marcuswu/dlineate/internal/element"
	"github.com/marcuswu/dlineate/utils"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func addPoint(s *Solver, x, y float64, fixed bool) el.SketchElement {
	e := el.NewSketchPoint(s.Elements.NextId(), big.NewFloat(x), big.NewFloat(y))
	e.SetFixed(fixed)
	s.AddElement(e)
	return e
}

func addLine(s *Solver, p1, p2 el.SketchElement) el.SketchElement {
	if p1 == nil || p2 == nil {
		return nil
	}
	if p1.AsPoint() == nil || p2.AsPoint() == nil {
		return nil
	}
	x1, _ := p1.AsPoint().X.Float64()
	y1, _ := p1.AsPoint().Y.Float64()
	x2, _ := p2.AsPoint().X.Float64()
	y2, _ := p2.AsPoint().Y.Float64()

	var a, b, c, t big.Float
	a.SetPrec(utils.FloatPrecision).SetFloat64(y2 - y1) // y' - y
	b.SetPrec(utils.FloatPrecision).SetFloat64(x1 - x2) // x - x'
	c.SetPrec(utils.FloatPrecision).Neg(&a)
	t.SetPrec(utils.FloatPrecision).SetFloat64(x1)
	// c = -ax - by from ax + by + c = 0
	c.Mul(&c, &t)
	t.SetFloat64(y1)
	t.Mul(&t, &b)
	c.Sub(&c, &t)
	l := el.NewSketchLine(s.Elements.NextId(), &a, &b, &c)

	s.AddElement(l)

	// Automatically constrain the line to pass through the two points
	c1 := constraint.NewConstraint(s.Constraints.NextId(), constraint.Distance, p1.GetID(), l.GetID(), big.NewFloat(0), false)
	s.AddConstraint(c1)
	c2 := constraint.NewConstraint(s.Constraints.NextId(), constraint.Distance, p2.GetID(), l.GetID(), big.NewFloat(0), false)
	s.AddConstraint(c2)

	return l
}

func TestSolve(t *testing.T) {
	utils.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	solver := NewSolver()

	// Add elements and constraints to the solver here
	p1 := addPoint(solver, 0, 0, true)
	p2 := addPoint(solver, 1, 0, false)
	p3 := addPoint(solver, 1, -2, false)
	_ = addLine(solver, p1, p2)
	_ = addLine(solver, p2, p3)
	_ = addLine(solver, p3, p1)

	c1 := constraint.NewConstraint(solver.Constraints.NextId(), constraint.Distance, p1.GetID(), p2.GetID(), big.NewFloat(1), false)
	solver.AddConstraint(c1)
	c2 := constraint.NewConstraint(solver.Constraints.NextId(), constraint.Distance, p2.GetID(), p3.GetID(), big.NewFloat(1), false)
	solver.AddConstraint(c2)
	c3 := constraint.NewConstraint(solver.Constraints.NextId(), constraint.Distance, p3.GetID(), p1.GetID(), big.NewFloat(1), false)
	solver.AddConstraint(c3)

	solver.Solve(utils.FloatPrecision, utils.MaxNumericIterations)

	// Output final positions of points
	for _, eId := range solver.Elements.IdSet().Contents() {
		e, _ := solver.Elements.GetElement(-1, eId)
		utils.Logger.Info().
			Uint("element id", e.GetID()).
			Str("element", e.String()).
			Msg("Final element position")
	}
	p1, _ = solver.Elements.GetElement(-1, p1.GetID())
	p2, _ = solver.Elements.GetElement(-1, p2.GetID())
	p3, _ = solver.Elements.GetElement(-1, p3.GetID())
	d1, _ := p1.AsPoint().DistanceTo(p2).Float64()
	d2, _ := p2.AsPoint().DistanceTo(p3).Float64()
	d3, _ := p3.AsPoint().DistanceTo(p1).Float64()
	utils.Logger.Info().
		Float64("d1", d1).
		Float64("d2", d2).
		Float64("d3", d3).
		Msg("Final distances:")
	if utils.StandardFloatCompare(d1, 1.0) != 0 ||
		utils.StandardFloatCompare(d2, 1.0) != 0 ||
		utils.StandardFloatCompare(d3, 1.0) != 0 {
		t.Errorf("Final distances do not match constraints: d1=%f, d2=%f, d3=%f", d1, d2, d3)
	}
}
