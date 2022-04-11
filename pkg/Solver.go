package dlineate

import (
	"errors"

	core "github.com/marcuswu/dlineate/internal"
)

type Sketch struct {
	plane       *Workplane
	sketch      *core.SketchGraph
	Elements    []*Element
	constraints []*Constraint
	passes      int
}

func NewSketch() *Sketch {
	s := new(Sketch)
	s.sketch = core.NewSketch()
	s.passes = 0

	return s
}

func (s *Sketch) SetWorkplane(plane *Workplane) {
	s.plane = plane
}

func (s *Sketch) findConstraints(e *Element) []*Constraint {
	found := make([]*Constraint, 0, 2)
	for _, c := range s.constraints {
		for _, el := range c.elements {
			if el == e {
				found = append(found, c)
			}
		}
	}
	return found
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
	p.element = s.sketch.AddPoint(p.values[0], p.values[1])
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

	l.element = s.sketch.AddLine(a, b, c) // AddLine normalizes a, b, c
	s.Elements = append(s.Elements, l)

	start := s.AddPoint(l.values[0], l.values[1])
	start.isChild = true
	end := s.AddPoint(l.values[2], l.values[3])
	end.isChild = true
	l.children = append(l.children, start)
	l.children = append(l.children, end)
	s.addDistanceConstraint(l, start, 0.0)
	s.addDistanceConstraint(l, end, 0.0)
	return l
}

func (s *Sketch) NewCircle(x float64, y float64, r float64) *Element {
	c := emptyElement()
	c.elementType = Circle
	c.values = append(c.values, x)
	c.values = append(c.values, y)
	c.values = append(c.values, r)

	s.Elements = append(s.Elements, c)

	center := s.AddPoint(c.values[0], c.values[1])
	center.isChild = true
	c.children = append(c.children, center)
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

	s.Elements = append(s.Elements, a)

	center := s.AddPoint(a.values[0], a.values[1])
	center.isChild = true
	a.children = append(a.children, center)

	start := s.AddPoint(a.values[2], a.values[3])
	start.isChild = true
	end := s.AddPoint(a.values[4], a.values[5])
	end.isChild = true
	a.children = append(a.children, start)
	a.children = append(a.children, end)
	s.addDistanceConstraint(a, start, 0.0)
	s.addDistanceConstraint(a, end, 0.0)
	return a
}

func (s *Sketch) resolveConstraint(c *Constraint) bool {
	switch c.constraintType {
	case Coincident:
		fallthrough
	case Distance:
		fallthrough
	case Angle:
		fallthrough
	case Perpendicular:
		fallthrough
	case Parallel:
		c.state = Resolved
		return true
	case Ratio:
		return s.resolveRatioConstraint(c)
	case Midpoint:
		return s.resolveMidpointConstraint(c)
	case Tangent:
		return s.resolveTangentConstraint(c)
	}

	return c.state == Resolved
}

func (s *Sketch) isElementSolved(e *Element) bool {
	// Need any internal constraint related to this element
	constraints := s.findConstraints(e)
	// If there are 2, this element is fully constrained (more is over constrained)
	if len(constraints) < 2 {
		return false
	}

	// If those have been solved, then the element is solved
	numSolved := 0
	for _, c := range constraints {
		if c.state == Solved {
			numSolved++
		}
	}

	if s.passes == 0 {
		return false
	}

	return numSolved > 1
}

func (s *Sketch) getDistanceConstraint(e *Element) (*Constraint, bool) {
	dc, err := s.findConstraint(Distance, e)
	if err == nil {
		return dc, true
	}

	if e.elementType != Line {
		return nil, false
	}

	constraints := s.findConstraints(e.children[0])
	for _, c := range constraints {
		if c.elements[0] == e.children[1] || c.elements[1] == e.children[2] {
			return c, true
		}
	}

	return nil, false
}

func (s *Sketch) resolveLineLength(e *Element) (float64, bool) {
	if e.elementType != Line {
		return 0, false
	}

	dc, ok := s.getDistanceConstraint(e)
	if ok {
		v := dc.constraints[0].Value
		return v, ok
	}

	startConstrained := s.isElementSolved(e.children[0])
	endConstrained := s.isElementSolved(e.children[1])
	if startConstrained && endConstrained {
		// resolve constraint setting p2's distance to the distance from p1 start to p1 end
		v := e.children[0].element.AsPoint().DistanceTo(e.children[1].element.AsPoint())

		return v, true
	}

	return 0, false
}

func (s *Sketch) resolveCurveRadius(e *Element) (float64, bool) {
	if e.elementType != Arc && e.elementType != Circle {
		return 0, false
	}

	dc, ok := s.getDistanceConstraint(e)
	if ok {
		v := dc.constraints[0].Value
		return v, ok
	}

	// Circles and Arcs with solved center and solved elements coincident or distance to the circle / arc
	if centerSolved := s.isElementSolved(e.children[0]); centerSolved {
		eAll := s.findConstraints(e)
		var other *Element = nil
		for _, ec := range eAll {
			if ec.constraintType != Distance && ec.constraintType != Coincident {
				continue
			}
			other = ec.elements[0]
			if other == e {
				other = ec.elements[1]
			}
			if !s.isElementSolved(other) {
				continue
			}
			// Other & e have a distance constraint between them. dist(other, e.center) - c.value is radius
			distFromCurve := other.element.AsPoint().DistanceTo(e.children[0].element.AsPoint())
			radius := distFromCurve - ec.constraints[0].Value
			return radius, true
		}
	}

	return 0, false
}

func (s *Sketch) Solve() error {

	return nil
}
