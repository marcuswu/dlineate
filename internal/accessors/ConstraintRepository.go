package accessors

import (
	"github.com/marcuswu/dlineate/internal/constraint"
	"github.com/marcuswu/dlineate/utils"
	"github.com/rs/zerolog"
)

type ConstraintRepository struct {
	constraints map[uint]*constraint.Constraint
	eToC        map[uint][]*constraint.Constraint
}

func NewConstraintRepository() *ConstraintRepository {
	r := &ConstraintRepository{}
	r.Clear()
	return r
}

func (r *ConstraintRepository) Clear() {
	r.constraints = make(map[uint]*constraint.Constraint, 0)
	r.eToC = make(map[uint][]*constraint.Constraint, 0)
}

func (r *ConstraintRepository) GetConstraint(cId uint) (*constraint.Constraint, bool) {
	c, ok := r.constraints[cId]
	return c, ok
}

func (r *ConstraintRepository) AddConstraint(c *constraint.Constraint) *constraint.Constraint {
	r.constraints[c.GetID()] = c
	if _, ok := r.eToC[c.Element1]; !ok {
		r.eToC[c.Element1] = make([]*constraint.Constraint, 0)
	}
	if _, ok := r.eToC[c.Element2]; !ok {
		r.eToC[c.Element2] = make([]*constraint.Constraint, 0)
	}
	r.eToC[c.Element1] = append(r.eToC[c.Element1], c)
	r.eToC[c.Element2] = append(r.eToC[c.Element2], c)
	return c
}

func (r *ConstraintRepository) RemoveConstraint(cId uint) {
	constraint := r.constraints[cId]
	delete(r.constraints, cId)
	removeConstraintFromElement := func(eId uint, cId uint) {
		constraints := r.eToC[eId]
		for i := len(constraints) - 1; i >= 0; i-- {
			c := constraints[i]
			if c.GetID() != cId {
				continue
			}
			constraints[i] = constraints[len(constraints)-1]
			constraints = constraints[:len(constraints)-1]
		}
		r.eToC[eId] = constraints
	}
	removeConstraintFromElement(constraint.Element1, cId)
	removeConstraintFromElement(constraint.Element2, cId)
}

func (r *ConstraintRepository) ConstraintsForElement(eId uint) []*constraint.Constraint {
	constraints, ok := r.eToC[eId]
	if !ok {
		return []*constraint.Constraint{}
	}
	return constraints
}

func (r *ConstraintRepository) SetConstraintElement(oldId uint, newElement uint) {
	for _, constraint := range r.constraints {
		if constraint.Element1 == oldId {
			constraint.Element1 = newElement
		}
		if constraint.Element2 == oldId {
			constraint.Element2 = newElement
		}
	}
	r.eToC[newElement] = append(r.eToC[newElement], r.eToC[oldId]...)
}

func (r *ConstraintRepository) NextId() uint {
	return uint(len(r.constraints))
}

func (r *ConstraintRepository) IdSet() *utils.Set {
	ids := utils.NewSet()
	for id, _ := range r.constraints {
		ids.Add(id)
	}
	return ids
}

func (r *ConstraintRepository) Count() int {
	return len(r.constraints)
}

func (r *ConstraintRepository) LogConstraints(level zerolog.Level) {
	utils.Logger.WithLevel(level).Msg("Constraints: ")
	for _, c := range r.constraints {
		utils.Logger.WithLevel(level).Msgf("%v", c)
	}
	utils.Logger.WithLevel(level).Msg("")
}

func (r *ConstraintRepository) IsMet(constr uint, cluster int, ea ElementAccessor) bool {
	c := r.constraints[constr]
	e1, ok := ea.GetElement(cluster, c.Element1)
	if !ok {
		return false
	}
	e2, ok := ea.GetElement(cluster, c.Element2)
	if !ok {
		return false
	}
	return c.IsMet(e1, e2)
}

func (r *ConstraintRepository) ReplaceElement(original, new uint) {
	for _, c := range r.constraints {
		if c.Element1 == original {
			c.Element1 = new
		}
		if c.Element2 == original {
			c.Element2 = new
		}
	}
}
