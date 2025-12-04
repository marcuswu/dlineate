package numeric

import (
	"sort"
	"strconv"

	"github.com/marcuswu/dlineate/internal/accessors"
	"github.com/marcuswu/dlineate/internal/constraint"
	el "github.com/marcuswu/dlineate/internal/element"
	"github.com/marcuswu/dlineate/utils"
	"gonum.org/v1/gonum/optimize"
)

type Solver struct {
	elements      accessors.ElementAccessor
	constraints   accessors.ConstraintAccessor
	fixedElements *utils.Set
}

func NewSolver() *Solver {
	s := new(Solver)
	s.elements = accessors.NewElementRepository()
	s.constraints = accessors.NewConstraintRepository()
	s.fixedElements = utils.NewSet()
	return s
}

func (s *Solver) AddElement(e el.SketchElement) {
	s.elements.AddElement(e)
	if e.IsFixed() {
		s.fixedElements.Add(e.GetID())
	}
}

func (s *Solver) GetElement(id uint) (el.SketchElement, bool) {
	return s.elements.GetElement(-1, id)
}

func (s *Solver) AddConstraint(c *constraint.Constraint) {
	s.constraints.AddConstraint(c)
}

func (s *Solver) FreeParams() utils.Set {
	freeParams := utils.NewSet()
	for _, eId := range s.elements.IdSet().Contents() {
		if !s.fixedElements.Contains(eId) {
			freeParams.Add(eId)
		}
	}
	return *freeParams
}

func (s *Solver) FreeValues() []float64 {
	freeValues := make([]float64, 0, s.elements.Count()*3)
	sortedElements := s.elements.IdSet().Contents()
	sort.Slice(sortedElements, func(i, j int) bool {
		return sortedElements[i] < sortedElements[j]
	})
	for _, eId := range sortedElements {
		if !s.fixedElements.Contains(eId) {
			e, _ := s.elements.GetElement(-1, eId)
			freeValues = append(freeValues, el.ElementValues(e)...)
		}
	}
	return freeValues
}

func (s *Solver) Update(values []float64) {
	sortedElements := s.elements.IdSet().Contents()
	sort.Slice(sortedElements, func(i, j int) bool {
		return sortedElements[i] < sortedElements[j]
	})
	for _, eId := range sortedElements {
		if s.fixedElements.Contains(eId) {
			continue
		}
		e, _ := s.elements.GetElement(-1, eId)
		numParams := 3
		if e.GetType() == el.Point {
			numParams = 2
		}
		paramValues := values[:numParams]
		values = values[numParams:]
		el.SetElementValues(e, paramValues)
	}
}

func (s *Solver) Error() float64 {
	totalError := 0.0
	for _, cId := range s.constraints.IdSet().Contents() {
		constraint, _ := s.constraints.GetConstraint(cId)
		e1, _ := s.elements.GetElement(-1, constraint.Element1)
		e2, _ := s.elements.GetElement(-1, constraint.Element2)
		totalError += constraint.Error(e1, e2)
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
	strValues := make([]string, 0, len(initialValues))
	for _, value := range initialValues {
		strValues = append(strValues, strconv.FormatFloat(value, 'f', -1, 64))
	}
	utils.Logger.Debug().
		Strs("initial values", strValues).
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
