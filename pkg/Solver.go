package dlineate

import (
	"errors"

	core "github.com/marcuswu/dlineate/internal"
	constraint "github.com/marcuswu/dlineate/internal/constraint"
)

type Sketch struct {
	plane       *Workplane
	sketch      *core.SketchGraph
	Elements    []*Element
	constraints []*Constraint
}

func NewSketch() *Sketch {
	s := new(Sketch)
	s.sketch = core.NewSketch()

	return s
}

func (s *Sketch) SetWorkplane(plane *Workplane) {
	s.plane = plane
}

func (s *Sketch) findConstraint(ctype ConstraintType, e *Element) (*Constraint, error) {
	for _, c := range s.constraints {
		if c.constraintType != ctype {
			continue
		}
		for _, el := range c.elements {
			if el == e {
				return c, nil
			}
		}
	}

	return nil, errors.New("no such constraint")
}

func (s *Sketch) AddPoint(x float64, y float64) *Element {
	p := emptyElement()
	p.elementType = Point
	p.values = append(p.values, x)
	p.values = append(p.values, y)
	p.elements = append(p.elements, s.sketch.AddPoint(p.values[0], p.values[1]))
	s.Elements = append(s.Elements, p)
	return p
}

func (s *Sketch) AddLine(x1 float64, y1 float64, x2 float64, y2 float64) *Element {
	l := emptyElement()
	l.elementType = Line

	a := y2 - y1              // y' - y
	b := x1 - x2              // x - x'
	c := (-a * x1) - (b * y1) // c = -ax - by from ax + by + c = 0
	l.values = append(l.values, x1)
	l.values = append(l.values, y1)
	l.values = append(l.values, x2)
	l.values = append(l.values, y2)

	l.elements = append(l.elements, s.sketch.AddLine(a, b, c)) // AddLine normalizes a, b, c
	l.elements = append(l.elements, s.sketch.AddPoint(l.values[0], l.values[1]))
	l.elements = append(l.elements, s.sketch.AddPoint(l.values[2], l.values[3]))
	l.constraints = append(l.constraints, s.sketch.AddConstraint(constraint.Distance, l.elements[0], l.elements[1], 0.0))
	l.constraints = append(l.constraints, s.sketch.AddConstraint(constraint.Distance, l.elements[0], l.elements[2], 0.0))
	s.Elements = append(s.Elements, l)
	return l
}

func (s *Sketch) NewCircle(x float64, y float64, r float64) *Element {
	c := emptyElement()
	c.elementType = Circle
	c.values = append(c.values, x)
	c.values = append(c.values, y)
	c.values = append(c.values, r)
	c.elements = append(c.elements, s.sketch.AddPoint(c.values[0], c.values[1]))
	s.Elements = append(s.Elements, c)
	return c
}

func (s *Sketch) NewArc(x1 float64, y1 float64, x2 float64, y2 float64, x3 float64, y3 float64) *Element {
	a := emptyElement()
	a.elementType = Arc
	a.values = append(a.values, x1)
	a.values = append(a.values, y1)
	a.values = append(a.values, x2)
	a.values = append(a.values, y2)
	a.values = append(a.values, x3)
	a.values = append(a.values, y3)

	a.elements = append(a.elements, s.sketch.AddPoint(a.values[0], a.values[1]))
	a.elements = append(a.elements, s.sketch.AddPoint(a.values[2], a.values[3]))
	a.elements = append(a.elements, s.sketch.AddPoint(a.values[4], a.values[5]))
	s.Elements = append(s.Elements, a)
	return a
}

func (s *Sketch) resolveConstraintDependencies() {
	
}

func (s *Sketch) Solve() error {




	return nil
}
