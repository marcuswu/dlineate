package core

import (
	"github.com/marcuswu/dlineate/internal/constraint"
	el "github.com/marcuswu/dlineate/internal/element"
	"github.com/marcuswu/dlineate/utils"
	"github.com/rs/zerolog"
)

type ConstraintAccessor interface {
	GetConstraint(id uint) (*constraint.Constraint, bool)
	ConstraintsForElement(eId uint) []*constraint.Constraint
	SetConstraintElement(oldId uint, newElement el.SketchElement)
	AddConstraint(*constraint.Constraint)
	RemoveConstraint(uint)
	Count() int
	NextId() uint
	IdSet() *utils.Set
	logConstraints(logger *zerolog.Event)
}
