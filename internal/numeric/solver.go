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
	s.Elements.AddElement(e)
	if e.IsFixed() {
		s.fixedElements.Add(e.GetID())
	}
}

func (s *Solver) GetElement(id uint) (el.SketchElement, bool) {
	return s.Elements.GetElement(-1, id)
}

func (s *Solver) AddConstraint(c *constraint.Constraint) {
	s.Constraints.AddConstraint(c)
	values := utils.NewSetFromList(s.valueOrder)
	if !s.fixedElements.Contains(c.Element1) {
		values.Add(c.Element1)
	}
	if !s.fixedElements.Contains(c.Element2) {
		values.Add(c.Element2)
	}
	s.valueOrder = values.Contents()
}

func (s *Solver) FreeParams() utils.Set {
	freeParams := utils.NewSet()
	for _, eId := range s.Elements.IdSet().Contents() {
		if !s.fixedElements.Contains(eId) {
			freeParams.Add(eId)
		}
	}
	return *freeParams
}

func (s *Solver) FreeValues() []float64 {
	freeValues := make([]float64, 0, s.Elements.Count()*3)
	// sortedElements := s.Elements.IdSet().Contents()
	// sort.Slice(sortedElements, func(i, j int) bool {
	// 	return sortedElements[i] < sortedElements[j]
	// })
	// ids := make([]uint, 0, len(sortedElements))
	// for _, eId := range sortedElements {
	for _, eId := range s.valueOrder {
		if s.fixedElements.Contains(eId) {
			continue
		}
		e, _ := s.Elements.GetElement(-1, eId)
		freeValues = append(freeValues, el.ElementValues(e)...)
	}
	utils.Logger.Debug().
		Uints("ids", s.valueOrder).
		Floats64("values", freeValues).
		Msg("Numeric solver: retrieved free values for elements")
	return freeValues
}

func (s *Solver) Update(values []float64) {
	valuesLeft := values
	for _, eId := range s.valueOrder {
		if s.fixedElements.Contains(eId) {
			continue
		}
		e, _ := s.Elements.GetElement(-1, eId)
		numParams := 3
		if e.GetType() == el.Point {
			numParams = 2
		}
		paramValues := valuesLeft[:numParams]
		valuesLeft = valuesLeft[numParams:]
		el.SetElementValues(e, paramValues)
	}
	utils.Logger.Debug().
		Uints("ids", s.valueOrder).
		Floats64("values", values).
		Msg("Numeric solver: updated elements with values")
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
		// 	Float64("error", constraintError).
		// 	Msg("Constraint error")
		utils.Logger.Error().
			Uint("constraint id", constraint.GetID()).
			Str("element 1", e1.String()).
			Str("element 2", e2.String()).
			Str("desired", constraint.Value.Text('f', 4)).
			Float64("error", constraintError).
			Msg("Constraint error")
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

	// initialError := problem.Func(s.FreeValues())
	initialError := s.Error()
	utils.Logger.Debug().
		Float64("initial error", initialError).
		Msg("Numeric solver: initial error before optimization")
	initialValues := s.FreeValues()
	utils.Logger.Debug().
		Floats64("initial values", initialValues).
		Msg("Numeric solver: starting optimization")
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
		Msg("Numeric solver: optimization completed")
	s.Update(solution.X)
	// Might need to check each constraint individually
	solved := solution.Status == optimize.Success
	return solved
}
