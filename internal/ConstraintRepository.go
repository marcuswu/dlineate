package core

import (
	"github.com/marcuswu/dlineate/internal/constraint"
	el "github.com/marcuswu/dlineate/internal/element"
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

func (r *ConstraintRepository) AddConstraint(c *constraint.Constraint) {
	r.constraints[c.GetID()] = c
	if _, ok := r.eToC[c.Element1.GetID()]; !ok {
		r.eToC[c.Element1.GetID()] = make([]*constraint.Constraint, 0)
	}
	if _, ok := r.eToC[c.Element2.GetID()]; !ok {
		r.eToC[c.Element2.GetID()] = make([]*constraint.Constraint, 0)
	}
	r.eToC[c.Element1.GetID()] = append(r.eToC[c.Element1.GetID()], c)
	r.eToC[c.Element2.GetID()] = append(r.eToC[c.Element2.GetID()], c)
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
	removeConstraintFromElement(constraint.Element1.GetID(), cId)
	removeConstraintFromElement(constraint.Element2.GetID(), cId)
}

func (r *ConstraintRepository) ConstraintsForElement(eId uint) []*constraint.Constraint {
	constraints, ok := r.eToC[eId]
	if !ok {
		return []*constraint.Constraint{}
	}
	return constraints
}

func (r *ConstraintRepository) SetConstraintElement(oldId uint, newElement el.SketchElement) {
	for _, constraint := range r.constraints {
		if constraint.Element1.GetID() == oldId {
			constraint.Element1 = newElement
		}
		if constraint.Element2.GetID() == oldId {
			constraint.Element2 = newElement
		}
	}
	r.eToC[newElement.GetID()] = append(r.eToC[newElement.GetID()], r.eToC[oldId]...)
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

func (r *ConstraintRepository) logConstraints(logger *zerolog.Event) {
	logger.Msg("Constraints: ")
	for _, c := range r.constraints {
		logger.Msgf("%v", c)
	}
	logger.Msg("")
}
