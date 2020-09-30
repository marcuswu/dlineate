package core

// ConstraintType of a Constraint(Distance or Angle)
type ConstraintType uint

// ConstraintType constants
const (
	Distance ConstraintType = iota
	Angle
)

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
	id             uint
	constraintType ConstraintType
	value          float64
	element1       SketchElement
	element2       SketchElement
}

// GetID returns the constraint identifier
func (c *Constraint) GetID() uint {
	return c.id
}

// GetValue returns the constraint's value
func (c *Constraint) GetValue() float64 {
	return c.value
}

// UpdateValue updates the constraint's value
func (c *Constraint) UpdateValue(v float64) {
	c.value = v
}

// HasElementID returns whether an element with the passed ID
// exists in this constraint
func (c *Constraint) HasElementID(eID uint) bool {
	return c.element1.GetID() == eID || c.element2.GetID() == eID
}

// First returns the first element in the constraint
func (c *Constraint) First() SketchElement {
	return c.element1
}

// Second returns the second element in the constraint
func (c *Constraint) Second() SketchElement {
	return c.element2
}

// Equals returns whether two constraints are equal
func (c *Constraint) Equals(o Constraint) bool {
	return c.id == o.GetID()
}

// NewConstraint creates a new constraint
func NewConstraint(id uint, constraintType ConstraintType, a SketchElement, b SketchElement, v float64) Constraint {
	// TODO: construct and return distance or angle constraint
	return Constraint{
		id:             id,
		constraintType: constraintType,
		value:          v,
		element1:       a,
		element2:       b,
	}
}
