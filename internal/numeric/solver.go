package numeric

import (
	"sort"

	"github.com/marcuswu/dlineate/internal/accessors"
	"github.com/marcuswu/dlineate/internal/constraint"
	el "github.com/marcuswu/dlineate/internal/element"
	"github.com/marcuswu/dlineate/utils"
	"github.com/rs/zerolog"
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
	utils.Logger.Debug().
		Str("element", e.String()).
		Msg("Adding Element")
	if segment, ok := e.(*Segment); ok {
		el, ok := s.Elements.GetElement(-1, segment.start.GetID())
		if !ok {
			s.Elements.AddElement(segment.start)
		} else {
			segment.start = el.AsPoint()
		}
		if segment.start.IsFixed() {
			s.fixedElements.Add(segment.start.GetID())
		}
		el, ok = s.Elements.GetElement(-1, segment.end.GetID())
		if !ok {
			s.Elements.AddElement(segment.end)
		} else {
			segment.end = el.AsPoint()
		}
		if segment.end.IsFixed() {
			s.fixedElements.Add(segment.end.GetID())
		}
	}
	if _, ok := s.Elements.GetElement(-1, e.GetID()); !ok {
		s.Elements.AddElement(e)
	}
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
		}
		return
	}
	// For a segment, add its points -- only ever add points to the numerical solver data
	if !s.fixedElements.Contains(segment.start.GetID()) {
		values.Add(segment.start.GetID())
	}
	if !s.fixedElements.Contains(segment.end.GetID()) {
		values.Add(segment.end.GetID())
	}
	s.valueOrder = values.Contents()
}

func (s *Solver) AddConstraint(c *constraint.Constraint) {
	s.Constraints.AddConstraint(c)
	s.addValueOrder(c.Element1)
	s.addValueOrder(c.Element2)
}

func (s *Solver) FreeValues() []float64 {
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
}

func (s *Solver) Error() float64 {
	totalError := 0.0
	for _, cId := range s.Constraints.IdSet().Contents() {
		constraint, _ := s.Constraints.GetConstraint(cId)
		e1, _ := s.Elements.GetElement(-1, constraint.Element1)
		e2, _ := s.Elements.GetElement(-1, constraint.Element2)
		constraintError := constraint.Error(e1, e2)
		totalError += constraintError
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

	initialValues := s.FreeValues()
	if len(initialValues) == 0 {
		utils.Logger.Debug().
			Msg("Numeric solver: no free values to solve")
		return false
	}
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

	if utils.LogLevel() <= zerolog.DebugLevel {
		elements := s.Elements.IdSet().Contents()
		sort.Slice(elements, func(i, j int) bool {
			return elements[i] < elements[j]
		})
		for _, eId := range elements {
			e, _ := s.Elements.GetElement(-1, eId)
			if e.GetType() != el.Point {
				continue
			}
			utils.Logger.Info().
				Uint("element id", e.GetID()).
				Str("element", e.String()).
				Msg("Final element position")
		}

		constraints := s.Constraints.IdSet().Contents()
		sort.Slice(constraints, func(i, j int) bool {
			return constraints[i] < constraints[j]
		})
		for _, cId := range constraints {
			constraint, _ := s.Constraints.GetConstraint(cId)
			e1, _ := s.Elements.GetElement(-1, constraint.Element1)
			e2, _ := s.Elements.GetElement(-1, constraint.Element2)
			constraintError := constraint.Error(e1, e2)
			utils.Logger.Info().
				Uint("constraint id", constraint.GetID()).
				Str("element 1", e1.String()).
				Str("element 2", e2.String()).
				Str("desired", constraint.Value.Text('f', 4)).
				Float64("error", constraintError).
				Msg("Final constraint error")
		}
	}

	solved := solution.F <= tolerance
	return solved
}
