package dlineate

import (
	"reflect"

	core "github.com/marcuswu/dlineate/internal"
	"github.com/marcuswu/dlineate/internal/constraint"
	el "github.com/marcuswu/dlineate/internal/element"
)

// Solver a sketch solver instance
type Solver struct {
	graph *core.SketchGraph
}

/*------------ Everything below here needs rewrite -------------*/

// Geometry any geometry in a sketch
type Geometry interface {
	constraintOther() el.SketchElement
}

// Point a point in a sketch.
type Point struct {
	point *el.SketchPoint
}

func (p *Point) constraintOther() el.SketchElement { return p.point }

// Line a line in a sketch. Has a start, an end, and a line formula
type Line struct {
	start, end *el.SketchPoint
	line       *el.SketchLine
}

func (l *Line) constraintOther() el.SketchElement { return l.line }

// Circle a circle in a sketch. Has a center and a radius
type Circle struct {
	center *el.SketchPoint
	radius float64
}

func (c *Circle) constraintOther() el.SketchElement { return c.center }

// Arc an arc in a sketch. Has a center and a start and end
type Arc struct {
	Circle
	start *el.SketchPoint
	end   *el.SketchPoint
}

func (a *Arc) constraintOther() el.SketchElement { return a.center }

// New creates a new solver
func New() *Solver { return &Solver{graph: core.NewSketch()} }

// NewPoint creates a new point
func (s *Solver) NewPoint(x, y float64) *Point {
	return &Point{point: s.graph.AddPoint(x, y).(*el.SketchPoint)}
}

// NewLine creates a new line
func (s *Solver) NewLine(xStart, yStart, xEnd, yEnd float64) *Line {
	start := s.graph.AddPoint(xStart, yStart).(*el.SketchPoint)
	end := s.graph.AddPoint(yEnd, yEnd).(*el.SketchPoint)
	a := -1 * (yEnd - yStart)
	b := xEnd - xStart
	c := -1 * ((a * xStart) + (b * yStart))
	line := s.graph.AddLine(a, b, c).(*el.SketchLine)
	s.graph.AddConstraint(constraint.Distance, start, line, 0)
	s.graph.AddConstraint(constraint.Distance, end, line, 0)
	return &Line{start: start, end: end, line: line}
}

// NewCircle creates a new circle
func (s *Solver) NewCircle(xCenter, yCenter, radius float64) *Circle {
	center := s.graph.AddPoint(xCenter, yCenter).(*el.SketchPoint)
	return &Circle{center: center, radius: radius}
}

// NewArc creates a new arc
func (s *Solver) NewArc(xCenter, yCenter, xStart, yStart, xEnd, yEnd float64) *Arc {
	start := s.graph.AddPoint(xStart, yStart).(*el.SketchPoint)
	end := s.graph.AddPoint(yEnd, yEnd).(*el.SketchPoint)
	center := s.graph.AddPoint(xCenter, yCenter).(*el.SketchPoint)

	return &Arc{Circle: Circle{center: center}, start: start, end: end}
}

// Native constraints

func (s *Solver) distancePoints(p1, p2 *Point, value float64) *constraint.Constraint {
	return s.graph.AddConstraint(constraint.Distance, p1.point, p2.point, value)
}

// Angle -- between lines
func (s *Solver) Angle(l1, l2 *Line, angle float64) *constraint.Constraint {
	return s.graph.AddConstraint(constraint.Distance, l1.line, l2.line, angle)
}

// Derived constraints

func (s *Solver) radiusCircle(c *Circle, value float64) {
	c.radius = value
}

func (s *Solver) lengthLine(l *Line, value float64) *constraint.Constraint {
	return s.graph.AddConstraint(constraint.Distance, l.start, l.end, value)
}

func (s *Solver) distanceCircle(c *Circle, g Geometry, value float64) *constraint.Constraint {
	// TODO: doing this means radius must be set first and alterations to radius need constraint updates
	return s.graph.AddConstraint(constraint.Distance, c.center, g.constraintOther(), c.radius+value)
}

// Distance -- combine the above distance methods
func (s *Solver) Distance(g1, g2 Geometry, value float64) (*constraint.Constraint, error) {
	switch v := g1.(type) {
	case *Point:
		if _, ok := g2.(*Point); !ok {
			return s.Distance(g2, g1, value)
		}
		return s.distancePoints(v, g2.(*Point), value), nil
	case *Circle:
		if reflect.ValueOf(g2).IsNil() {
			s.radiusCircle(g1.(*Circle), value)
			return nil, nil
		}
		return s.distanceCircle(v, g2, value), nil
	case *Line:
		_, isCircle := g2.(*Circle)
		_, isArc := g2.(*Arc)
		if isArc || isCircle {
			return s.Distance(g2, g1, value)
		}
		if reflect.ValueOf(g2).IsNil() {
			return s.lengthLine(v, value), nil
		}
		return s.graph.AddConstraint(constraint.Distance, v.line, g2.constraintOther(), value), nil
	case *Arc:
		if reflect.ValueOf(g2).IsNil() {
			v.radius = value
			s.graph.AddConstraint(constraint.Distance, v.center, v.end, value)
			return s.graph.AddConstraint(constraint.Distance, v.center, v.start, value), nil
		}
		return s.distanceCircle(&v.Circle, g2, value), nil
	default:
		return s.graph.AddConstraint(constraint.Distance, g1.constraintOther(), g2.constraintOther(), value), nil
	}
}
