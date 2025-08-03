package dlineate

import (
	"testing"

	"github.com/marcuswu/dlineate/internal/constraint"
	"github.com/marcuswu/dlineate/internal/element"
	"github.com/stretchr/testify/assert"
)

func TestElementTypeString(t *testing.T) {
	// Point
	assert.Equal(t, "Point", Point.String(), "point element")
	// Axis
	assert.Equal(t, "Axis", Axis.String(), "axis element")
	// Line
	assert.Equal(t, "Line", Line.String(), "line element")
	// Circle
	assert.Equal(t, "Circle", Circle.String(), "circle element")
	// Arc
	assert.Equal(t, "Arc", Arc.String(), "arcelement")
	// other
	var e ElementType = 20
	assert.Equal(t, "20", e.String(), "unknown element")
}

func TestValuesFromSketch(t *testing.T) {
	s := NewSketch()
	arc := s.AddArc(0, 0, 1, 1, 2, 2)
	arc.children[0].element.AsPoint().X = 1
	arc.children[0].element.AsPoint().Y = 2
	arc.children[1].element.AsPoint().X = 3
	arc.children[1].element.AsPoint().Y = 4
	arc.children[2].element.AsPoint().X = 5
	arc.children[2].element.AsPoint().Y = 6

	s.XAxis.valuesFromSketch(s)
	assert.Equal(t, 0.0, s.XAxis.values[0])
	assert.Equal(t, -1.0, s.XAxis.values[1])
	assert.Equal(t, 0.0, s.XAxis.values[2])

	arc.valuesFromSketch(s)
	assert.Equal(t, 1.0, arc.values[0])
	assert.Equal(t, 2.0, arc.values[1])
	assert.Equal(t, 3.0, arc.values[2])
	assert.Equal(t, 4.0, arc.values[3])
	assert.Equal(t, 5.0, arc.values[4])
	assert.Equal(t, 6.0, arc.values[5])
}

func TestGetCircleRadius(t *testing.T) {
	s := NewSketch()
	l := s.AddLine(0, 0, 1, 1)
	o := s.AddPoint(1, 1)
	c1 := s.AddDistanceConstraint(l, o, 1)
	_, err := l.getCircleRadius(s, c1)
	assert.NotNil(t, err, "Should get an error looking for circle radius on a line")

	c := s.AddCircle(0, 0, 2)
	c2 := s.AddDistanceConstraint(c, nil, 3)
	dist, err := c.getCircleRadius(s, c2)
	assert.Nil(t, err, "Should have no error looking for a circle radius on a circle")
	assert.Equal(t, 3.0, dist, "Should find circle distance")

	c = s.AddCircle(1, 0, 2)
	c3 := s.AddCoincidentConstraint(c, o)
	c3.constraints = append(c3.constraints, s.sketch.AddConstraint(constraint.Distance, c.Center().element, o.element, 1.0))
	dist, err = c.getCircleRadius(s, c3)
	assert.Nil(t, err, "Should have no error looking for a circle")
	assert.Equal(t, 1.0, dist, "Should find circle distance")
}

func TestGetValues(t *testing.T) {
	s := NewSketch()
	arc := s.AddArc(0, 0, 1, 1, 2, 2)
	arc.children[0].element.AsPoint().X = 1
	arc.children[0].element.AsPoint().Y = 2
	arc.children[1].element.AsPoint().X = 3
	arc.children[1].element.AsPoint().Y = 4
	arc.children[2].element.AsPoint().X = 5
	arc.children[2].element.AsPoint().Y = 6

	values := arc.Values()
	assert.Equal(t, 0.0, values[0])
	assert.Equal(t, 0.0, values[1])
	assert.Equal(t, 1.0, values[2])
	assert.Equal(t, 1.0, values[3])
	assert.Equal(t, 2.0, values[4])
	assert.Equal(t, 2.0, values[5])

	s.passes++
	arc.valuesFromSketch(s)
	values = arc.Values()
	assert.Equal(t, 1.0, values[0])
	assert.Equal(t, 2.0, values[1])
	assert.Equal(t, 3.0, values[2])
	assert.Equal(t, 4.0, values[3])
	assert.Equal(t, 5.0, values[4])
	assert.Equal(t, 6.0, values[5])
}

func TestConstraintLevel(t *testing.T) {
	s := NewSketch()
	arc := s.AddArc(0, 0, 1, 1, 2, 2)
	arc.children[0].element.SetConstraintLevel(element.FullyConstrained)
	arc.children[1].element.SetConstraintLevel(element.FullyConstrained)
	arc.children[2].element.SetConstraintLevel(element.FullyConstrained)

	level := arc.ConstraintLevel()
	assert.Equal(t, element.FullyConstrained, level, "Expect fully constrained")

	arc.children[2].element.SetConstraintLevel(element.UnderConstrained)
	level = arc.ConstraintLevel()
	assert.Equal(t, element.UnderConstrained, level, "Expect under constrained")
}

func TestMinMaxXY(t *testing.T) {
	s := NewSketch()
	arc := s.AddArc(1, 1, 0, 2, -1, 3)

	minx, miny, maxx, maxy := arc.minMaxXY()
	assert.Equal(t, -1.0, minx, "minx")
	assert.Equal(t, 1.0, miny, "miny")
	assert.Equal(t, 1.0, maxx, "maxx")
	assert.Equal(t, 3.0, maxy, "maxy")

	arc = s.AddArc(0, 0, 1, -0.9, 3, -1)
	minx, miny, maxx, maxy = arc.minMaxXY()
	assert.Equal(t, 0.0, minx, "minx")
	assert.Equal(t, -1.0, miny, "miny")
	assert.Equal(t, 3.0, maxx, "maxx")
	assert.Equal(t, 0.0, maxy, "maxy")

	cir := s.AddCircle(0, 0, 2)
	minx, miny, maxx, maxy = cir.minMaxXY()
	assert.Equal(t, -2.0, minx, "minx")
	assert.Equal(t, -2.0, miny, "miny")
	assert.Equal(t, 2.0, maxx, "maxx")
	assert.Equal(t, 2.0, maxy, "maxy")
}

func TestStartCenterEndEdgeCases(t *testing.T) {
	s := NewSketch()
	p := s.AddPoint(0, 0)
	e := p.Start()
	assert.Nil(t, e, "Point should not have a start")
	e = p.Center()
	assert.Nil(t, e, "Point should not have a center")
	e = p.End()
	assert.Nil(t, e, "Point should not have a end")
}
