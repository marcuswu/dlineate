package dlineate

import (
	core "github.com/marcuswu/dlineate/internal/core"
	el "github.com/marcuswu/dlineate/internal/element"
	c "github.com/marcuswu/dlineate/internal/constraint"
)


type Sketch struct {
	plane *core.Workplane
	sketch *core.SketchGraph
	elements []*Element
	constraints []*Constraint
}

func NewSketch() *Sketch {
	s := new(Sketch)
	s.sketch = core.NewSketch()
}

func (s *Sketch) AddElement(e *Element) {
	e.addToSketch(s.sketch)
}

func (s *Sketch) SetWorkplane(plane *Workplane) {
	s.plane = plane
}

func (s *Sketch) AddPoint(x float64, y float64) *Element {
	p := emptyElement()
	p.elementType = Point
	append(p.values, x)
	append(p.values, y)
	append(e.elements, s.sketch.AddPoint(e.values[0], e.values[1]))
	append(s.elements, p)
	return p
}

func (s *Sketch) AddLine(x1 double, y1 double, x2 double, y2 double) *Element {
	l := emptyElement()
	l.elementType = Line

	a := y2 - y1 // y' - y
	b := x1 - x2 // x - x'
	c := (-a * x1) - (b * y1) // c = -ax - by from ax + by + c = 0
	append(l.values, x1)
	append(l.values, y1)
	append(l.values, x2)
	append(l.values, y2)

	append(l.elements, s.sketch.AddLine(a, b, c)) // AddLine normalizes a, b, c
	append(l.elements, s.sketch.AddPoint(e.values[0], e.values[1]))
	append(l.elements, s.sketch.AddPoint(e.values[2], e.values[3]))
	append(l.constraints, s.sketch.AddConstraint(c.Distance, e.elements[0], e.elements[1], 0.0))
	append(l.constraints, s.sketch.AddConstraint(c.Distance, e.elements[0], e.elements[2], 0.0))
	append(s.elements, l)
	return l
}

func (s *Sketch) NewCircle(x double, y double, r double) *Element {
	c := emptyElement()
	c.elementType = Circle
	append(c.values, x)
	append(c.values, y)
	append(c.values, r)
	append(c.elements, s.sketch.AddPoint(e.values[0], e.values[1])) 
	append(s.elements, c)
	return c
}

func (s *Sketch) NewArc(x1 double, y1 double, x2 double, y2 double, x3 double, y3 double) *Element {
	a := emptyElement()
	a.elementType = Arc
	append(a.values, x1)
	append(a.values, y1)
	append(a.values, x2)
	append(a.values, y2)
	append(a.values, x3)
	append(a.values, y3)

	append(a.elements, s.sketch.AddPoint(e.values[0], e.values[1]))
	append(a.elements, s.sketch.AddPoint(e.values[2], e.values[3]))
	append(a.elements, s.sketch.AddPoint(e.values[4], e.values[5]))
	append(s.elements, a)
	return a
}