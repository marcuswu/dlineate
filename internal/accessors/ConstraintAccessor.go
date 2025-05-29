package accessors

import (
	"github.com/marcuswu/dlineate/internal/constraint"
	"github.com/marcuswu/dlineate/utils"
	"github.com/rs/zerolog"
)

type ConstraintAccessor interface {
	GetConstraint(id uint) (*constraint.Constraint, bool)
	ConstraintsForElement(eId uint) []*constraint.Constraint
	SetConstraintElement(oldId uint, newElement uint)
	AddConstraint(*constraint.Constraint) *constraint.Constraint
	RemoveConstraint(uint)
	ReplaceElement(uint, uint)
	Count() int
	NextId() uint
	IdSet() *utils.Set
	LogConstraints(level zerolog.Level)
	IsMet(constraint uint, cluster int, ea ElementAccessor) bool
}
