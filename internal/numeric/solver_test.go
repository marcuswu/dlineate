package numeric

import (
	"math"
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

	solveElement := &Segment{id: s.Elements.NextId(), start: p1.AsPoint(), end: p2.AsPoint(), fixed: p1.IsFixed() && p2.IsFixed(), elementType: el.Line}

	s.AddElement(solveElement)

	return solveElement
}

func TestSolveTriangle(t *testing.T) {
	utils.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, NoColor: true})
	solver := NewSolver()

	// Add elements and constraints to the solver here
	origin := addPoint(solver, 0, 0, true)
	xap := addPoint(solver, 1, 0, true)
	xa := addLine(solver, origin, xap)

	p1 := addPoint(solver, 0, 0, false)
	p2 := addPoint(solver, 0.8, 0.2, false)
	p3 := addPoint(solver, 1, 2, false)
	l1 := addLine(solver, p1, p2)
	l2 := addLine(solver, p2, p3)
	l3 := addLine(solver, p3, p1)

	c1 := constraint.NewConstraint(solver.Constraints.NextId(), constraint.Distance, p1.GetID(), p2.GetID(), big.NewFloat(1), false)
	solver.AddConstraint(c1)
	// c2 := constraint.NewConstraint(solver.Constraints.NextId(), constraint.Distance, l1.GetID(), xa.GetID(), big.NewFloat(0), false)
	// solver.AddConstraint(c2)

	// p2-p3 and p3-p1 distance constraints
	// Uncomment here and comment angles below to solve by distances
	/*
		c2 := constraint.NewConstraint(solver.Constraints.NextId(), constraint.Distance, p2.GetID(), p3.GetID(), big.NewFloat(1), false)
		solver.AddConstraint(c2)
		c3 := constraint.NewConstraint(solver.Constraints.NextId(), constraint.Distance, p3.GetID(), p1.GetID(), big.NewFloat(1), false)
		solver.AddConstraint(c3)
	*/

	// l1-l2 and l2-l3 angle constraints
	// outsideAngle := (-120. / 180.) * math.Pi
	outsideAngle := (120. / 180.) * math.Pi
	// sixtyDeg := big.NewFloat(sixtyDegFloat)
	c4 := constraint.NewConstraint(solver.Constraints.NextId(), constraint.Angle, l1.GetID(), l2.GetID(), big.NewFloat(outsideAngle), false)
	solver.AddConstraint(c4)
	c5 := constraint.NewConstraint(solver.Constraints.NextId(), constraint.Angle, l2.GetID(), l3.GetID(), big.NewFloat(outsideAngle), false)
	solver.AddConstraint(c5)
	c6 := constraint.NewConstraint(solver.Constraints.NextId(), constraint.Angle, xa.GetID(), l1.GetID(), big.NewFloat(0), false)
	solver.AddConstraint(c6)

	solver.Solve(utils.StandardCompare, utils.MaxNumericIterations)

	// Output final positions of points
	for _, eId := range solver.Elements.IdSet().Contents() {
		e, _ := solver.Elements.GetElement(-1, eId)
		if e.IsFixed() || e.GetType() != el.Point {
			continue
		}
		utils.Logger.Info().
			Uint("element id", e.GetID()).
			Str("element", e.String()).
			Bool("fixed", e.IsFixed()).
			Msg("Final element position")
	}
	p1, _ = solver.Elements.GetElement(-1, p1.GetID())
	p2, _ = solver.Elements.GetElement(-1, p2.GetID())
	p3, _ = solver.Elements.GetElement(-1, p3.GetID())
	d1, _ := p1.AsPoint().DistanceTo(p2).Float64()
	d2, _ := p2.AsPoint().DistanceTo(p3).Float64()
	d3, _ := p3.AsPoint().DistanceTo(p1).Float64()
	a1, _ := l1.AsLine().AngleToLine(l2.AsLine()).Float64()
	a2, _ := l2.AsLine().AngleToLine(l3.AsLine()).Float64()
	utils.Logger.Info().
		Float64("d1", d1).
		Float64("d2", d2).
		Float64("d3", d3).
		Float64("a1", a1).
		Float64("a2", a2).
		Msg("Final distances and angles:")
	if utils.StandardFloatCompare(d1, 1.0) != 0 ||
		utils.StandardFloatCompare(a1, outsideAngle) != 0 ||
		utils.StandardFloatCompare(a2, outsideAngle) != 0 {
		t.Errorf("Final distances and angles do not match constraints: d1=%f, a1=%f, a2=%f", d1, (a1/math.Pi)*180., (a2/math.Pi)*180.)
	}
}

func TestSolveModifiedRectangle(t *testing.T) {
	utils.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, NoColor: true})
	solver := NewSolver()

	/*
	   "Point(7) (2, 11)"
	   "Point(10) (2, 0)"
	   "Point(13) (-2, 11)"
	   "Point(14) (0, -4.7)"
	   "Point(16) (-2, 0)"
	   "Segment(1) Start: Point(0) (0, 0), End: Point(2) (0, -1)"
	   "Segment(3) Start: Point(0) (0, 0), End: Point(4) (1, 0)"
	   "Segment(5) Start: Point(13) (-2, 11), End: Point(7) (2, 11)"
	   "Segment(8) Start: Point(7) (2, 11), End: Point(10) (2, 0)"
	   "Segment(11) Start: Point(16) (-2, 0), End: Point(13) (-2, 11)"
	*/
	p0 := addPoint(solver, 0, 0, true)
	p2 := addPoint(solver, 0, 1, true)
	p4 := addPoint(solver, 1, 0, true)

	p7 := addPoint(solver, 2, 11, false)    // top right    3
	p10 := addPoint(solver, 2, 0, false)    // bottom right 4
	p13 := addPoint(solver, -2, 11, false)  // top left     5
	p14 := addPoint(solver, 0, -4.7, false) // arc center   6
	p16 := addPoint(solver, -2, 0, false)   // bottom left  7

	s1 := addLine(solver, p0, p4)    // x axis
	s3 := addLine(solver, p0, p2)    // y axis
	s5 := addLine(solver, p13, p7)   // top horizontal
	s8 := addLine(solver, p7, p10)   // right vertical
	s11 := addLine(solver, p16, p13) // left vertical

	/*
		Constraints in numeric solver:
		"Constraint(0) type: Distance, e1: 1, e2: 14, v: 4.7"
		"Constraint(9) type: Distance, e1: 13, e2: 7, v: 4"
		"Constraint(10) type: Angle, e1: 5, e2: 1, v: 0 rad"
		"Constraint(11) type: Distance, e1: 7, e2: 10, v: 11"
		"Constraint(12) type: Angle, e1: 8, e2: 3, v: 0 rad"
		"Constraint(13) type: Distance, e1: 16, e2: 13, v: 11"
		"Constraint(14) type: Angle, e1: 11, e2: 3, v: 0 rad"
		"Constraint(15) type: Distance, e1: 3, e2: 14, v: 0"
		"Constraint(17) type: Distance, e1: 14, e2: 10, v: 6.5"
		"Constraint(18) type: Distance, e1: 14, e2: 16, v: 6.5"
	*/

	// Add elements and constraints to the solver here
	// Arc center 4.7 from X axis
	c0 := constraint.NewConstraint(solver.Constraints.NextId(), constraint.Distance, p14.GetID(), s1.GetID(), big.NewFloat(4.7), false)
	solver.AddConstraint(c0)
	// top horizontal length is 4
	c9 := constraint.NewConstraint(solver.Constraints.NextId(), constraint.Distance, p13.GetID(), p7.GetID(), big.NewFloat(4), false)
	solver.AddConstraint(c9)
	// top horizontal is parallel with x axis
	c10 := constraint.NewConstraint(solver.Constraints.NextId(), constraint.Angle, s5.GetID(), s1.GetID(), big.NewFloat(0), false)
	solver.AddConstraint(c10)
	// right vertical length is 11
	c11 := constraint.NewConstraint(solver.Constraints.NextId(), constraint.Distance, p7.GetID(), p10.GetID(), big.NewFloat(11), false)
	solver.AddConstraint(c11)
	// right vertical is parallel with y axis
	c12 := constraint.NewConstraint(solver.Constraints.NextId(), constraint.Angle, s8.GetID(), s3.GetID(), big.NewFloat(math.Pi), false)
	solver.AddConstraint(c12)
	// left vertical length is 11
	c13 := constraint.NewConstraint(solver.Constraints.NextId(), constraint.Distance, p16.GetID(), p13.GetID(), big.NewFloat(11), false)
	solver.AddConstraint(c13)
	// left vertical is parallel with y axis
	c14 := constraint.NewConstraint(solver.Constraints.NextId(), constraint.Angle, s11.GetID(), s3.GetID(), big.NewFloat(0), false)
	solver.AddConstraint(c14)
	// arc center is on y axis
	c15 := constraint.NewConstraint(solver.Constraints.NextId(), constraint.Distance, s3.GetID(), p14.GetID(), big.NewFloat(0), false)
	solver.AddConstraint(c15)
	// arc center is 6.5 from right vertical end
	c17 := constraint.NewConstraint(solver.Constraints.NextId(), constraint.Distance, p14.GetID(), p10.GetID(), big.NewFloat(6.5), false)
	solver.AddConstraint(c17)
	// arc center is 6.5 from left vertical start
	c18 := constraint.NewConstraint(solver.Constraints.NextId(), constraint.Distance, p14.GetID(), p16.GetID(), big.NewFloat(6.5), false)
	solver.AddConstraint(c18)

	solved := solver.Solve(utils.StandardCompare, utils.MaxNumericIterations)

	offBy := c0.Error(p0, p14)
	utils.Logger.Info().
		Float64("error", offBy).
		Msg("Constraint 0")
	offBy = c9.Error(p13, p7)
	utils.Logger.Info().
		Float64("error", offBy).
		Msg("Constraint 9")
	offBy = c10.Error(s5, s1)
	utils.Logger.Info().
		Float64("error", offBy).
		Msg("Constraint 10")
	offBy = c11.Error(p7, p10)
	utils.Logger.Info().
		Float64("error", offBy).
		Msg("Constraint 11")
	offBy = c12.Error(s8, s3)
	utils.Logger.Info().
		Float64("error", offBy).
		Msg("Constraint 12")
	offBy = c13.Error(p16, p13)
	utils.Logger.Info().
		Float64("error", offBy).
		Msg("Constraint 13")
	offBy = c14.Error(s3, s11)
	utils.Logger.Info().
		Float64("error", offBy).
		Msg("Constraint 14")
	offBy = c15.Error(s3, p14)
	utils.Logger.Info().
		Float64("error", offBy).
		Msg("Constraint 15")

	offBy = c17.Error(p14, p10)
	utils.Logger.Info().
		Float64("error", offBy).
		Msg("Constraint 17")
	offBy = c18.Error(p14, p16)
	utils.Logger.Info().
		Float64("error", offBy).
		Msg("Constraint 18")

	if !solved {
		t.Errorf("Could not solve with a good error margin")
	}
}
