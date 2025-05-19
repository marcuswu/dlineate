package accessors

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
	topE, topOk := r.elements[eId]
	if cId < 0 {
		return topE, topOk
	}
	clusterElements, ok := r.clusterElements[cId]
	if !ok {
		return topE, topOk
	}
	e, ok := clusterElements[eId]
	if !ok {
		return topE, topOk
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

func (r *ElementRepository) IsShared(eId uint) bool {
	set := r.clustersContaining(eId)
	return set.Count() > 1
}

func (r *ElementRepository) AddElement(e el.SketchElement) el.SketchElement {
	r.elements[e.GetID()] = e
	return e
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
	for _, cMap := range r.clusterElements {
		delete(cMap, rem)
	}
	delete(r.elements, rem)
}

func (r *ElementRepository) ReplaceElement(cId int, eId uint, e el.SketchElement) {
	replaceAll := cId < 0
	noMatch := true
	for id, cMap := range r.clusterElements {
		clusterMatch := replaceAll || cId == id
		if _, ok := cMap[eId]; !clusterMatch || !ok {
			continue
		}
		noMatch = false
		cMap[eId] = e
	}
	if replaceAll || noMatch {
		r.elements[eId] = e
	}
}

func (r *ElementRepository) SharedElements(c1 int, c2 int) *utils.Set {
	shared := utils.NewSet()
	c1Shared := r.clusterElements[c1]
	c2Shared := r.clusterElements[c2]
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
	c1Elements := r.clusterElements[c1]
	c2Elements := r.clusterElements[c2]
	for eId, e := range c2Elements {
		if _, ok := c1Elements[eId]; ok {
			continue
		}
		c1Elements[eId] = e
	}
	delete(r.clusterElements, c2)
}

func (r *ElementRepository) MergeToRoot(cluster int) {
	clusterElements := r.clusterElements[cluster]
	for eId, e := range clusterElements {
		r.elements[eId] = e
	}
	delete(r.clusterElements, cluster)
}

func (r *ElementRepository) CopyToCluster(to int, from int, eId uint) {
	fromCluster := r.clusterElements[from]
	toCluster := r.clusterElements[to]
	toCluster[eId] = el.CopySketchElement(fromCluster[eId])
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

func (r *ElementRepository) LogElements(level zerolog.Level) {
	utils.Logger.WithLevel(level).Msg("Elements: ")
	for _, e := range r.elements {
		utils.Logger.WithLevel(level).Msgf("%v", e)
	}
	utils.Logger.WithLevel(level).Msg("")
}

func (r *ElementRepository) IsFixed(eId uint) bool {
	e, ok := r.elements[eId]
	if !ok {
		return false
	}
	return e.IsFixed()
}
