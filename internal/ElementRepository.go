package core

import (
	el "github.com/marcuswu/dlineate/internal/element"
	"github.com/marcuswu/dlineate/utils"
	"github.com/rs/zerolog"
)

type ElementRepository struct {
	elements        map[uint]el.SketchElement
	clusterElements map[int]map[uint]el.SketchElement
}

func NewElementRepository() *ElementRepository {
	r := &ElementRepository{}
	r.Clear()
	return r
}

func (r *ElementRepository) Clear() {
	r.elements = make(map[uint]el.SketchElement, 0)
	r.clusterElements = make(map[int]map[uint]el.SketchElement, 0)
}

func (r *ElementRepository) ClearClusters() {
	r.clusterElements = make(map[int]map[uint]el.SketchElement, 0)
}

func (r *ElementRepository) GetElement(cId int, eId uint) (el.SketchElement, bool) {
	if cId < 0 {
		e, ok := r.elements[eId]
		return e, ok
	}
	clusterElements, ok := r.clusterElements[cId]
	if !ok {
		e, ok := r.elements[eId]
		return e, ok
	}
	e, ok := clusterElements[eId]
	if !ok {
		e, ok := r.elements[eId]
		return e, ok
	}
	return e, ok
}

func (r *ElementRepository) clustersContaining(eId uint) *utils.Set {
	set := utils.NewSet()
	for cId, cMap := range r.clusterElements {
		if _, ok := cMap[eId]; ok {
			set.Add(uint(cId))
		}
	}
	return set
}

func (r *ElementRepository) isShared(eId uint) bool {
	set := r.clustersContaining(eId)
	return set.Count() > 1
}

func (r *ElementRepository) AddElement(e el.SketchElement) {
	r.elements[e.GetID()] = e
}

func (r *ElementRepository) AddElementToCluster(eId uint, cId int) {
	if _, ok := r.elements[eId]; !ok {
		return
	}
	if _, ok := r.clusterElements[cId]; !ok {
		r.clusterElements[cId] = make(map[uint]el.SketchElement)
	}
	clusters := r.clustersContaining(eId)
	switch clusters.Count() {
	case 0:
		// regular element
		r.clusterElements[cId][eId] = r.elements[eId]
	case 1:
		// The previous element needs to be a copy
		shared := int(clusters.Contents()[0])
		r.clusterElements[shared][eId] = el.CopySketchElement(r.elements[eId])
		r.clusterElements[cId][eId] = el.CopySketchElement(r.elements[eId])
	default:
		// copies have already been made -- just create the current copy
		r.clusterElements[cId][eId] = el.CopySketchElement(r.elements[eId])
	}
}

func (r *ElementRepository) RemoveElement(rem uint) {
	delete(r.elements, rem)
}

func (r *ElementRepository) SharedElements(c1 int, c2 int) *utils.Set {
	shared := utils.NewSet()
	c1Shared, _ := r.clusterElements[c1]
	c2Shared, _ := r.clusterElements[c2]
	for eId, _ := range c1Shared {
		if _, ok := c2Shared[eId]; !ok {
			continue
		}
		shared.Add(eId)
	}
	return shared
}

// Merge elements between two clusters
// Move shared items from c2 to c1 then reevaluate shared items for c1
func (r *ElementRepository) MergeElements(c1 int, c2 int) {
	c1Shared, _ := r.clusterElements[c1]
	c2Shared, _ := r.clusterElements[c2]
	toDelete := make([]uint, 0, 2)
	for eId, e := range c1Shared {
		if _, ok := c2Shared[eId]; !ok {
			continue
		}
		toDelete = append(toDelete, eId)
		r.elements[eId] = e
	}
	for _, eId := range toDelete {
		delete(c2Shared, eId)
		if !r.isShared(eId) {
			r.elements[eId] = c1Shared[eId]
			delete(c1Shared, eId)
		}
	}
}

func (r *ElementRepository) IdSet() *utils.Set {
	ids := utils.NewSet()
	for id, _ := range r.elements {
		ids.Add(id)
	}
	return ids
}

func (r *ElementRepository) Count() int {
	return len(r.elements)
}

func (r *ElementRepository) NextId() uint {
	return uint(len(r.elements))
}

func (r *ElementRepository) SetConstraintLevel(eId uint, level el.ConstraintLevel) {
	r.elements[eId].SetConstraintLevel(level)
}

func (r *ElementRepository) ConstraintLevel(eId uint) el.ConstraintLevel {
	e, ok := r.elements[eId]
	if !ok {
		return el.UnderConstrained
	}
	return e.ConstraintLevel()
}

func (r *ElementRepository) logElements(logger *zerolog.Event) {
	logger.Msg("Elements: ")
	for _, e := range r.elements {
		logger.Msgf("%v", e)
	}
	logger.Msg("")
}
