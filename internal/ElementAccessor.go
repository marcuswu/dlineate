package core

import (
	el "github.com/marcuswu/dlineate/internal/element"
	"github.com/marcuswu/dlineate/utils"
	"github.com/rs/zerolog"
)

type ElementAccessor interface {
	GetElement(cId int, eId uint) (el.SketchElement, bool)
	isShared(eId uint) bool
	AddElement(el.SketchElement)
	AddElementToCluster(uint, int)
	SetConstraintLevel(uint, el.ConstraintLevel)
	ConstraintLevel(uint) el.ConstraintLevel
	RemoveElement(uint)
	SharedElements(int, int) *utils.Set
	MergeElements(int, int)
	NextId() uint
	IdSet() *utils.Set
	Count() int
	Clear()
	ClearClusters()
	logElements(*zerolog.Event)
}
