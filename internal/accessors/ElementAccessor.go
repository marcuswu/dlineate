package accessors

import (
	el "github.com/marcuswu/dlineate/internal/element"
	"github.com/marcuswu/dlineate/utils"
	"github.com/rs/zerolog"
)

type ElementAccessor interface {
	GetElement(cId int, eId uint) (el.SketchElement, bool)
	IsFixed(eId uint) bool
	IsShared(eId uint) bool
	AddElement(el.SketchElement) el.SketchElement
	AddElementToCluster(uint, int)
	SetConstraintLevel(uint, el.ConstraintLevel)
	ConstraintLevel(uint) el.ConstraintLevel
	RemoveElement(uint)
	SharedElements(int, int) *utils.Set
	MergeElements(int, int)
	MergeToRoot(int)
	CopyToCluster(int, int, uint)
	NextId() uint
	IdSet() *utils.Set
	Count() int
	Clear()
	ClearClusters()
	LogElements(zerolog.Level)
}
