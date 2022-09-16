package element

import "fmt"

// Type of a SketchElement (Point or Line)
type Type uint

// SolveState constants
const (
	Point Type = iota
	Line
)

type ConstraintLevel uint

const (
	OverConstrained ConstraintLevel = iota
	UnderConstrained
	FullyConstrained
)

func (cl ConstraintLevel) String() string {
	switch cl {
	case OverConstrained:
		return "over constrained"
	case UnderConstrained:
		return "under constrained"
	case FullyConstrained:
		return "fully constrained"
	default:
		return fmt.Sprintf("%d", int(cl))
	}
}

// SketchElement A 2D element within a Sketch
type SketchElement interface {
	SetID(uint)
	GetID() uint
	GetType() Type
	AngleTo(*Vector) float64
	Translate(tx float64, ty float64)
	TranslateByElement(SketchElement)
	ReverseTranslateByElement(SketchElement)
	Rotate(tx float64)
	Is(SketchElement) bool
	SquareDistanceTo(SketchElement) float64
	DistanceTo(SketchElement) float64
	VectorTo(SketchElement) *Vector
	AsPoint() *SketchPoint
	AsLine() *SketchLine
	ConstraintLevel() ConstraintLevel
	SetConstraintLevel(ConstraintLevel)
}

// List is a list of SketchElements
type List []SketchElement

func (e List) Len() int           { return len(e) }
func (e List) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }
func (e List) Less(i, j int) bool { return e[i].GetID() < e[j].GetID() }

// CopySketchElement creates a deep copy of a SketchElement
func CopySketchElement(e SketchElement) SketchElement {
	var n SketchElement
	if e.GetType() == Point {
		p := e.(*SketchPoint)
		n = NewSketchPoint(e.GetID(), p.GetX(), p.GetY())
		n.SetConstraintLevel(e.ConstraintLevel())
		return n
	}
	l := e.(*SketchLine)
	n = NewSketchLine(l.GetID(), l.GetA(), l.GetB(), l.GetC())
	n.SetConstraintLevel(e.ConstraintLevel())
	return n
}

// IdentityMap is a map of id to SketchElement
type IdentityMap = map[uint]SketchElement
