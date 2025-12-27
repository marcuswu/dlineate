package constraint

import (
	"fmt"
	"math"
	"math/big"

	el "github.com/marcuswu/dlineate/internal/element"
	"github.com/marcuswu/dlineate/utils"
	"github.com/rs/zerolog"
)

// Type of a Constraint(Distance or Angle)
type Type uint

// ConstraintType constants
const (
	Distance Type = iota
	Angle
)

func (t Type) String() string {
	switch t {
	case Distance:
		return "Distance"
	case Angle:
		return "Angle"
	default:
		return fmt.Sprintf("%d", int(t))
	}
}

// Constraint interface representing a constraint
/*type Constraint interface {
	SetID(uint)
	GetID() uint
	GetType() ConstraintType
	GetValue() float64
	UpdateValue(float64)
	HasElementID(uint)
	First() SketchElement
	Second() SketchElement
	EquationCount() uint
	ValueCount() uint
	FillValues([]float64)
	CheckSolution([]float64, float64)
	Equals(Constraint) bool

	Calculate()
	Check()
}*/

// Constraint Represents a 2D constraint
type Constraint struct {
	id       uint
	Type     Type
	Value    big.Float
	Element1 uint
	Element2 uint
	Solved   bool
}

// GetID returns the constraint identifier
func (c *Constraint) GetID() uint {
	return c.id
}

// GetValue returns the constraint's value
func (c *Constraint) GetValue() *big.Float {
	var ret big.Float
	return ret.Copy(&c.Value)
}

// UpdateValue updates the constraint's value
func (c *Constraint) UpdateValue(v *big.Float) {
	c.Value.Set(v)
}

// HasElementID returns whether an element with the passed ID
// exists in this constraint
func (c *Constraint) HasElementID(eID uint) bool {
	return c.Element1 == eID || c.Element2 == eID
}

func (c *Constraint) HasElements(ids ...uint) bool {
	for _, id := range ids {
		if id != c.Element1 && id != c.Element2 {
			return false
		}
	}

	return true
}

// First returns the first element in the constraint
func (c *Constraint) First() uint {
	return c.Element1
}

// Second returns the second element in the constraint
func (c *Constraint) Second() uint {
	return c.Element2
}

func (c *Constraint) ElementIDs() []uint {
	return []uint{c.Element1, c.Element2}
}

/*func (c *Constraint) Element(this uint, ea accessors.ElementAccessor) (el.SketchElement, bool) {
	if this == c.Element1 {
		return c.First(-1, ea), true
	}
	return c.Second(-1, ea), this == c.Element2
}*/

func (c *Constraint) Other(this uint) (uint, bool) {
	if this == c.Element1 {
		return c.Element2, true
	}
	return c.Element1, this == c.Element2
}

func (c *Constraint) Shared(o *Constraint) (uint, bool) {
	if o.HasElementID(c.Element1) {
		return c.Element1, true
	}
	if o.HasElementID(c.Element2) {
		return c.Element2, true
	}

	return 0, false
}

func (c *Constraint) IsMet(e1 el.SketchElement, e2 el.SketchElement) bool {
	var temp big.Float
	current := e1.DistanceTo(e2)
	if c.Type == Angle {
		current = e1.AsLine().AngleToLine(e2.AsLine())
	}

	comparison := utils.StandardBigFloatCompare(temp.Abs(current), temp.Abs(&c.Value))
	if comparison != 0 {
		c.Solved = false
	} else {
		c.Solved = true
	}

	return c.Solved
}

func (c *Constraint) Error(e1 el.SketchElement, e2 el.SketchElement) float64 {
	var result float64
	switch c.Type {
	case Angle:
		// Returns a value between -Pi and Pi
		current := e1.AsLine().AngleToLine(e2.AsLine())

		Lv, _ := current.Float64()
		Sv, _ := c.Value.Float64()
		if Sv > Lv {
			Lv, Sv = Sv, Lv
		}

		// Compare crossing -Pi / Pi boundary counting -Pi as equal to Pi
		// Always positive in the range [0, 2pi)
		dist1 := (Sv + math.Pi) + (math.Pi - Lv)

		// Direct compare -- always positive in the range [0, 2pi)
		dist2 := Lv - Sv

		// Take the smaller of the two values. This results in an error in the range [0, pi)
		result = dist1
		if dist2 < result {
			result = dist2
		}
		// utils.Logger.Debug().
		// 	Uint("constraint id", c.GetID()).
		// 	Float64("angle difference", result).
		// 	Str("current", current.String()).
		// 	Str("desired", c.Value.String()).
		// 	Msg("checking angle constraint error")
		result = result * result
		if math.IsInf(result, 0) {
			utils.Logger.Error().
				Uint("constraint id", c.GetID()).
				Uint("element 1", c.Element1).
				Uint("element 2", c.Element2).
				Str("current", current.String()).
				Str("value", c.Value.String()).
				Msg("Constraint error is infinite")
		}
	case Distance:
		first := e1
		other := e2
		// If using the numerical solver, e2 could be a segment so convert
		if e2.GetType() == el.Line {
			other = e2.AsLine()
		}
		if e1.GetType() == el.Line {
			first = e1.AsLine()
		}
		current := first.DistanceTo(other)
		Lv, _ := current.Float64()
		Sv, _ := c.Value.Float64()
		if Sv > Lv {
			Lv, Sv = Sv, Lv
		}
		result = Lv - Sv
		if math.IsInf(result, 0) {
			utils.Logger.Error().
				Uint("constraint id", c.GetID()).
				Str("element 1", e1.String()).
				Str("element 2", e2.String()).
				Str("current", current.Text('f', 4)).
				Str("value", c.Value.Text('f', 4)).
				Msg("Constraint error is infinite")
		}
		result = result * result
	}
	return result
}

func (c *Constraint) String() string {
	units := ""
	if c.Type == Angle {
		units = " rad"
	}
	return fmt.Sprintf("Constraint(%d) type: %v, e1: %d, e2: %d, v: %s%s", c.GetID(), c.Type, c.Element1, c.Element2, c.Value.String(), units)
}

func (c *Constraint) ToGraphViz(cId1, cId2 int) string {
	if cId1 < 0 && cId2 < 0 {
		return fmt.Sprintf("\t%d -- %d [label=\"%v (%d)\"]\n", c.Element1, c.Element2, c.Type, c.id)
	}
	return fmt.Sprintf("\t\"%d-%d\" -- \"%d-%d\" [label=\"%v (%d)\"]\n", cId1, c.Element1, cId2, c.Element2, c.Type, c.id)
}

// Equals returns whether two constraints are equal
func (c *Constraint) Equals(o Constraint) bool {
	return c.id == o.GetID()
}

// NewConstraint creates a new constraint
func NewConstraint(id uint, constraintType Type, a uint, b uint, v *big.Float, solved bool) *Constraint {
	var val big.Float
	val.Copy(v)
	return &Constraint{
		id:       id,
		Type:     constraintType,
		Value:    val,
		Element1: a,
		Element2: b,
		Solved:   false,
	}
}

// CopyConstraint creates a deep copy of a Constraint
func CopyConstraint(c *Constraint) *Constraint {
	var temp big.Float
	return NewConstraint(
		c.GetID(),
		c.Type,
		c.Element1,
		c.Element2,
		temp.Copy(&c.Value),
		c.Solved,
	)
}

type ConstraintList []*Constraint

func (cl ConstraintList) Len() int           { return len(cl) }
func (cl ConstraintList) Swap(i, j int)      { cl[i], cl[j] = cl[j], cl[i] }
func (cl ConstraintList) Less(i, j int) bool { return cl[i].id < cl[j].id }

func (l ConstraintList) MarshalZerologArray(a *zerolog.Array) {
	for _, c := range l {
		a.Uint(c.GetID())
	}
}
