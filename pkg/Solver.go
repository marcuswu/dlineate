package dlineate

import (
	"errors"

	core "github.com/marcuswu/dlineate/internal"
	el "github.com/marcuswu/dlineate/internal/element"
	"github.com/marcuswu/dlineate/internal/solver"
)

type Sketch struct {
	plane       *Workplane
	sketch      *core.SketchGraph
	Elements    []*Element
	constraints []*Constraint
	eToC        map[uint][]*Constraint
	passes      int
	Origin      *Element
	XAxis       *Element
	YAxis       *Element
}

func NewSketch() *Sketch {
	s := new(Sketch)
	s.sketch = core.NewSketch()
	s.passes = 0
	s.eToC = make(map[uint][]*Constraint)
	s.Origin = s.addOrigin()
	s.XAxis = s.addAxis(0, -1, 0)
	s.YAxis = s.addAxis(1, 0, 0)
	s.AddAngleConstraint(s.XAxis, s.YAxis, 90)
	s.AddCoincidentConstraint(s.Origin, s.XAxis)
	s.AddCoincidentConstraint(s.Origin, s.YAxis)

	return s
}

func (s *Sketch) SetWorkplane(plane *Workplane) {
	s.plane = plane
}

func (s *Sketch) findConstraints(e *Element) []*Constraint {
	return s.eToC[e.element.GetID()]
}

func (s *Sketch) findConstraint(ctype ConstraintType, e *Element) (*Constraint, error) {
	for _, c := range s.eToC[e.element.GetID()] {
		if c.constraintType != ctype {
			continue
		}
		return c, nil
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
	s.eToC[p.element.GetID()] = make([]*Constraint, 2)
	return p
}

func (s *Sketch) addOrigin() *Element {
	o := emptyElement()
	o.elementType = Point
	o.values = append(o.values, 0)
	o.values = append(o.values, 0)

	o.element = s.sketch.AddPoint(0, 0) // AddLine normalizes a, b, c
	return o
}

func (s *Sketch) addAxis(a float64, b float64, c float64) *Element {
	ax := emptyElement()
	ax.elementType = Axis
	ax.values = append(ax.values, a)
	ax.values = append(ax.values, b)
	ax.values = append(ax.values, c)

	ax.element = s.sketch.AddLine(a, b, c) // AddLine normalizes a, b, c
	return ax
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
	s.eToC[start.element.GetID()] = make([]*Constraint, 2)
	l.children = append(l.children, end)
	s.eToC[end.element.GetID()] = make([]*Constraint, 2)
	s.addDistanceConstraint(l, start, 0.0)
	s.addDistanceConstraint(l, end, 0.0)
	s.eToC[l.element.GetID()] = make([]*Constraint, 2)
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
	s.eToC[center.element.GetID()] = make([]*Constraint, 2)
	s.eToC[c.element.GetID()] = make([]*Constraint, 2)
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
	s.eToC[center.element.GetID()] = make([]*Constraint, 2)

	start := s.AddPoint(a.values[2], a.values[3])
	start.isChild = true
	s.eToC[start.element.GetID()] = make([]*Constraint, 2)
	end := s.AddPoint(a.values[4], a.values[5])
	end.isChild = true
	s.eToC[end.element.GetID()] = make([]*Constraint, 2)
	a.children = append(a.children, start)
	a.children = append(a.children, end)
	s.addDistanceConstraint(a, start, 0.0)
	s.addDistanceConstraint(a, end, 0.0)
	s.eToC[a.element.GetID()] = make([]*Constraint, 2)
	return a
}

func (s *Sketch) resolveConstraint(c *Constraint) bool {
	if c.state == Resolved {
		return true
	}

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

func (s *Sketch) resolveConstraints() int {
	unresolved := 0

	for _, c := range s.constraints {
		if !s.resolveConstraint(c) {
			unresolved++
		}
	}

	return unresolved
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
	solveState := solver.None
	passes := 0

	// This isn't correct -- should run until everything is solved
	for numUnresolved := s.resolveConstraints(); numUnresolved > 0 && passes < 2; numUnresolved = s.resolveConstraints() {
		solveState = s.sketch.Solve()
		passes++
	}
	s.passes += passes

	// Handle if origin is translated
	s.Origin.valuesFromSketch(s)
	if s.Origin.element.AsPoint().X != 0 || s.Origin.element.AsPoint().Y != 0 {
		s.sketch.Translate(-s.Origin.element.AsPoint().X, -s.Origin.element.AsPoint().Y)
	}

	// Andle if x/y axes are rotated
	s.XAxis.valuesFromSketch(s)
	s.YAxis.valuesFromSketch(s)
	yaxis := el.NewSketchLine(0, 1, 0, 0)
	angle := s.YAxis.element.AsLine().AngleToLine(yaxis)
	if angle != 0 {
		s.sketch.Rotate(s.Origin.element.AsPoint(), angle)
	}

	// Load solved values back into our elements
	for _, e := range s.Elements {
		e.valuesFromSketch(s)
	}

	switch solveState {
	case solver.None:
		return errors.New("unknown solver error")
	case solver.UnderConstrained:
		// TODO: return information about which elements are underconstrained
		return errors.New("the sketch is under constrained")
	case solver.OverConstrained:
		// TODO: return information about which constraints are overconstrained
		return errors.New("the sketch is over constrained")
	case solver.NonConvergent:
		// TODO: return information about which constraints did not converge on a solved state
		return errors.New("the solver did not converge on a solution")
	default:
		return nil
	}
}
