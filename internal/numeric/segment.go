package numeric

import (
	"fmt"
	"math/big"

	el "github.com/marcuswu/dlineate/internal/element"
	"github.com/marcuswu/dlineate/utils"
)

// Segment is a pseudo element type for representing lines during numeric
// solving.
//
// The numerical solver handles points very well, but lines that need to
// fit to those points are secondary elements which are solved when the
// points are solved. We still need a line segment for angle constraints
// so we use this pseudo element.
type Segment struct {
	id          uint
	start       *el.SketchPoint
	end         *el.SketchPoint
	fixed       bool
	elementType el.Type
}

func NewSegmentFromLine(line *el.SketchLine) *Segment {
	isFixed := line.IsFixed() || (line.Start.IsFixed() && line.End.IsFixed())
	start := el.CopySketchElement(line.Start)
	end := el.CopySketchElement(line.End)
	return &Segment{id: line.GetID(), start: start.AsPoint(), end: end.AsPoint(), fixed: isFixed, elementType: el.Line}
}

// SetID sets the id of the element
func (s *Segment) SetID(id uint) {
	s.id = id
}

// GetID gets the id of the element
func (s *Segment) GetID() uint {
	return s.id
}

func (s *Segment) SetFixed(fixed bool) {
	s.fixed = fixed
}

func (s *Segment) IsFixed() bool {
	return s.fixed
}

func (s *Segment) GetType() el.Type { return s.elementType }

// Basis for most of the line segment operations
func (s *Segment) AsLine() *el.SketchLine {
	a, b, c := utils.BigFloatLineFromBigPoints(&s.start.X, &s.start.Y, &s.end.X, &s.end.Y)
	return el.NewSketchLine(s.id, a, b, c)
}

func (s *Segment) AngleTo(u *el.Vector) *big.Float {
	return s.AsLine().AngleTo(u)
}

func (s *Segment) Translate(tx *big.Float, ty *big.Float) {
	s.AsLine().Translate(tx, ty)
}

func (s *Segment) TranslateByElement(e el.SketchElement) {
	s.AsLine().TranslateByElement(e)
}

func (s *Segment) ReverseTranslateByElement(e el.SketchElement) {
	s.AsLine().ReverseTranslateByElement(e)
}

func (s *Segment) Rotate(radians *big.Float) {
	s.AsLine().Rotate(radians)
}

// Is returns true if the two elements have the same id
func (s *Segment) Is(o el.SketchElement) bool {
	return s.id == o.GetID()
}

// Is returns true if the two elements are equal
func (s *Segment) IsEqual(o el.SketchElement) bool {
	return s.AsLine().IsEqual(o)
}

func (s *Segment) SquareDistanceTo(o el.SketchElement) *big.Float {
	return s.AsLine().SquareDistanceTo(o)
}

func (s *Segment) DistanceTo(o el.SketchElement) *big.Float {
	return s.AsLine().DistanceTo(o)
}

func (s *Segment) VectorTo(o el.SketchElement) *el.Vector {
	return s.AsLine().VectorTo(o)
}

func (s *Segment) AsPoint() *el.SketchPoint { return nil }

func (s *Segment) ConstraintLevel() el.ConstraintLevel {
	return el.FullyConstrained
}

func (s *Segment) SetConstraintLevel(_ el.ConstraintLevel) {}

func (s *Segment) ToGraphViz(cId int) string {
	return s.AsLine().ToGraphViz(cId)
}

func (s *Segment) String() string {
	return fmt.Sprintf("Segment(%d) Start: %s, End: %s", s.id, s.start.String(), s.end.String())
}
