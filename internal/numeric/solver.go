package numeric

import (
	"github.com/marcuswu/dlineate/internal/accessors"
	"github.com/marcuswu/dlineate/internal/constraint"
	el "github.com/marcuswu/dlineate/internal/element"
	"github.com/marcuswu/dlineate/utils"
	"gonum.org/v1/gonum/optimize"
)

type Solver struct {
	Elements      accessors.ElementAccessor
	Constraints   accessors.ConstraintAccessor
	fixedElements *utils.Set
	valueOrder    []uint
}

func NewSolver() *Solver {
	s := new(Solver)
	s.Elements = accessors.NewElementRepository()
	s.Constraints = accessors.NewConstraintRepository()
	s.fixedElements = utils.NewSet()
	return s
}

func (s *Solver) AddElement(e el.SketchElement) {
	if segment, ok := e.(*Segment); ok {
		s.Elements.AddElement(segment.start)
		if segment.start.IsFixed() {
			s.fixedElements.Add(segment.start.GetID())
		}
		s.Elements.AddElement(segment.end)
		if segment.end.IsFixed() {
			s.fixedElements.Add(segment.end.GetID())
		}
	}
	s.Elements.AddElement(e)
	if e.IsFixed() {
		s.fixedElements.Add(e.GetID())
	}
}

func (s *Solver) GetElement(id uint) (el.SketchElement, bool) {
	return s.Elements.GetElement(-1, id)
}

func (s *Solver) addValueOrder(eId uint) {
	e, ok := s.Elements.GetElement(-1, eId)
	if !ok {
		return
	}
	values := utils.NewSetFromList(s.valueOrder)
	segment, ok := e.(*Segment)
	if !ok {
		if !s.fixedElements.Contains(eId) {
			values.Add(eId)
			s.valueOrder = values.Contents()
			// utils.Logger.Debug().
			// 	Uint("id", eId).
			// 	Msg("Adding value order id")
			// utils.Logger.Debug().
			// 	Uints("ids", s.valueOrder).
			// 	Msg("ValueOrder")
		}
		return
	}
	// For a segment, add its points -- only ever add points to the numerical solver data
	if !s.fixedElements.Contains(segment.start.GetID()) {
		values.Add(segment.start.GetID())
		// utils.Logger.Debug().
		// 	Uint("id", segment.start.GetID()).
		// 	Msg("Adding value order id")
	}
	if !s.fixedElements.Contains(segment.end.GetID()) {
		values.Add(segment.end.GetID())
		// utils.Logger.Debug().
		// 	Uint("id", segment.end.GetID()).
		// 	Msg("Adding value order id")
	}
	s.valueOrder = values.Contents()
	// utils.Logger.Debug().
	// 	Uints("ids", s.valueOrder).
	// 	Msg("ValueOrder")
}

func (s *Solver) AddConstraint(c *constraint.Constraint) {
	s.Constraints.AddConstraint(c)
	s.addValueOrder(c.Element1)
	s.addValueOrder(c.Element2)
}

/*func (s *Solver) FreeParams() utils.Set {
	freeParams := utils.NewSet()
	for _, eId := range s.Elements.IdSet().Contents() {
		if s.fixedElements.Contains(eId) {
			continue
		}
		if e, ok := s.Elements.GetElement(-1, eId); ok && e.GetType() == el.Line {
			segment := e.(*Segment)
			freeParams.Add(segment.start.GetID())
			freeParams.Add(segment.end.GetID())
		}
		freeParams.Add(eId)
	}
	return *freeParams
}*/

func (s *Solver) FreeValues() []float64 {
	// freeValues := make([]float64, 0, s.Elements.Count()*2)
	freeValues := make([]float64, 0, len(s.valueOrder)*2)
	for _, eId := range s.valueOrder {
		if s.fixedElements.Contains(eId) {
			continue
		}
		e, ok := s.Elements.GetElement(-1, eId)
		if !ok {
			utils.Logger.Error().
				Uint("Element id", eId).
				Msg("Failed to find element while building FreeValues")
			continue
		}
		if e.GetType() != el.Point {
			continue
		}
		freeValues = append(freeValues, el.ElementValues(e)...)
	}
	// utils.Logger.Debug().
	// 	Uints("ids", s.valueOrder).
	// 	Floats64("values", freeValues).
	// 	Msg("Numeric solver: retrieved free values for elements")
	return freeValues
}

func (s *Solver) Update(values []float64) {
	valuesLeft := values
	for _, eId := range s.valueOrder {
		if s.fixedElements.Contains(eId) {
			continue
		}
		e, ok := s.Elements.GetElement(-1, eId)
		if !ok {
			continue
		}
		if e.GetType() != el.Point {
			continue
		}
		paramValues := valuesLeft[:2]
		valuesLeft = valuesLeft[2:]
		el.SetElementValues(e, paramValues)
	}
	// utils.Logger.Debug().
	// 	Uints("ids", s.valueOrder).
	// 	Floats64("values", values).
	// 	Msg("Numeric solver: updated elements with values")
}

func (s *Solver) Error() float64 {
	totalError := 0.0
	for _, cId := range s.Constraints.IdSet().Contents() {
		constraint, _ := s.Constraints.GetConstraint(cId)
		e1, _ := s.Elements.GetElement(-1, constraint.Element1)
		e2, _ := s.Elements.GetElement(-1, constraint.Element2)
		constraintError := constraint.Error(e1, e2)
		totalError += constraintError
		// utils.Logger.Debug().
		// 	Uint("constraint id", constraint.GetID()).
		// 	Str("element 1", e1.String()).
		// 	Str("element 2", e2.String()).
		// 	Str("desired", constraint.Value.Text('f', 4)).
		// 	Float64("error", constraintError).
		// 	Msg("Constraint error")
	}
	return totalError
}

func (s *Solver) Solve(tolerance float64, maxIterations int) bool {
	problem := optimize.Problem{
		Func: func(x []float64) float64 {
			s.Update(x)
			return s.Error()
		},
	}

	settings := optimize.Settings{
		MajorIterations: maxIterations,
	}

	// initialError := s.Error()
	// utils.Logger.Debug().
	// 	Float64("initial error", initialError).
	// 	Msg("Numeric solver: initial error before optimization")
	// iterations := 20

	// for i := 0; i < iterations && s.Error() > tolerance; i++ {
	initialValues := s.FreeValues()
	// utils.Logger.Debug().
	// 	Floats64("initial values", initialValues).
	// 	Msg("Numeric solver: starting optimization")
	utils.Logger.Debug().
		Uints("value", s.valueOrder).
		Msg("value order")
	solution, err := optimize.Minimize(problem, initialValues, &settings, nil)
	if err != nil {
		utils.Logger.Debug().Err(err).
			Msg("Numeric solver: optimization error")
		return false
	}
	utils.Logger.Debug().
		Float64("final error", solution.F).
		Int("iterations", solution.Stats.MajorIterations).
		Int("max iterations", maxIterations).
		Str("status", solution.Status.String()).
		Float64("tolerance", tolerance).
		Msg("Numeric solver: optimization completed")
	s.Update(solution.X)
	// }
	// Might need to check each constraint individually
	solved := solution.F <= tolerance
	return solved
}
