package core

import (
	"fmt"
	"math"
	"sort"
	"testing"

	"github.com/marcuswu/dlineate/internal/constraint"
	el "github.com/marcuswu/dlineate/internal/element"
	"github.com/marcuswu/dlineate/internal/solver"
	"github.com/marcuswu/dlineate/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddElement(t *testing.T) {
	e1 := el.NewSketchPoint(0, 0, 1)
	e2 := el.NewSketchPoint(1, 2, 1)

	g := NewGraphCluster(1)
	g.AddElement(e1)

	if g.elements.Count() != 1 {
		t.Error("expected one element to be added to the cluster, found", g.elements.Count())
	}

	g.AddElement(e1)

	if g.elements.Count() != 1 {
		t.Error("expected no change to the cluster element length, found", g.elements.Count())
	}

	g.AddElement(e2)

	if g.elements.Count() != 2 {
		t.Error("expected two elements to be added to the cluster, found", g.elements.Count())
	}
}

func TestAddConstraint(t *testing.T) {
	e1 := el.NewSketchPoint(0, 0, 1)
	e2 := el.NewSketchPoint(1, 2, 1)
	e3 := el.NewSketchLine(2, 2, 1, -1)
	c1 := constraint.NewConstraint(0, constraint.Distance, e1, e2, 5, false)
	c2 := constraint.NewConstraint(1, constraint.Distance, e2, e3, 7, false)

	g := NewGraphCluster(0)
	g.AddConstraint(c1)

	if g.GetID() != 0 {
		t.Error("expected cluster id to be 0")
	}

	if g.constraints.Count() != 1 {
		t.Error("expected graph cluster to have one constraint, found", g.constraints.Count())
	}
	if g.elements.Count() != 2 {
		t.Error("expected graph cluster to have 2 elements, found", g.elements.Count())
	}

	c1.Solved = true
	g.AddConstraint(c1)
	if g.constraints.Count() != 1 {
		t.Error("expected no change to cluster constraints after adding the same constraint twice")
	}
	if g.elements.Count() != 2 {
		t.Error("expected no change to elements after adding the same constraint twice")
	}

	g.AddConstraint(c2)

	if g.constraints.Count() != 2 {
		t.Error("expected graph cluster to have two constraint, found", g.constraints.Count())
	}
	if g.elements.Count() != 3 {
		t.Error("expected graph cluster to have 3 elements, found", g.elements.Count())
	}
}

func TestHasElementID(t *testing.T) {
	e1 := el.NewSketchPoint(0, 0, 1)
	e2 := el.NewSketchPoint(1, 2, 1)
	e3 := el.NewSketchLine(2, 2, 1, -1)
	e4 := el.NewSketchLine(3, 2, 2, -0)
	c1 := constraint.NewConstraint(0, constraint.Distance, e1, e2, 5, false)
	c2 := constraint.NewConstraint(1, constraint.Distance, e2, e3, 7, false)
	c3 := constraint.NewConstraint(2, constraint.Distance, e3, e4, 2, false)

	g := NewGraphCluster(0)
	g.AddConstraint(c1)
	g.AddConstraint(c2)

	o := NewGraphCluster(1)
	o.AddConstraint(c3)

	if !g.elements.Contains(0) {
		t.Error("expected graph cluster to have element 0, but element was not found")
	}
	if !g.elements.Contains(1) {
		t.Error("expected graph cluster to have element 1, but element was not found")
	}
	if !g.elements.Contains(2) {
		t.Error("expected graph cluster to have element 2, but element was not found")
	}
	if g.elements.Contains(3) {
		t.Error("expected graph cluster to have element 3, but element was not found")
	}
	if g.elements.Contains(4) {
		t.Error("expected graph cluster to not have element 4, but element was found")
	}
}

func TestHasElement(t *testing.T) {
	e1 := el.NewSketchPoint(0, 0, 1)
	e2 := el.NewSketchPoint(1, 2, 1)
	e3 := el.NewSketchLine(2, 2, 1, -1)
	e4 := el.NewSketchLine(3, 2, 2, -0)
	c1 := constraint.NewConstraint(0, constraint.Distance, e1, e2, 5, false)
	c2 := constraint.NewConstraint(1, constraint.Distance, e2, e3, 7, false)
	c3 := constraint.NewConstraint(2, constraint.Distance, e3, e4, 2, false)

	g := NewGraphCluster(0)
	g.AddConstraint(c1)
	g.AddConstraint(c2)

	o := NewGraphCluster(1)
	o.AddConstraint(c3)

	elements1 := []el.SketchElement{e1, e2, e3}
	elements2 := []el.SketchElement{e3, e4}
	for _, e := range elements1 {
		assert.True(t, g.HasElement(e.GetID()), fmt.Sprintf("cluster 1 has expected element %d", e.GetID()))
	}
	for _, e := range elements2 {
		assert.True(t, o.HasElement(e.GetID()), fmt.Sprintf("cluster 2 has expected element %d", e.GetID()))
	}
	falseElement := el.NewSketchPoint(100, 0, 0)
	assert.False(t, g.HasElement(falseElement.GetID()))
	assert.False(t, o.HasElement(falseElement.GetID()))
}

func TestSharedElements(t *testing.T) {
	e1 := el.NewSketchPoint(0, 0, 1)
	e2 := el.NewSketchPoint(1, 2, 1)
	e3 := el.NewSketchLine(2, 2, 1, -1)
	e4 := el.NewSketchLine(3, 2, 2, -0)
	c1 := constraint.NewConstraint(0, constraint.Distance, e1, e2, 5, false)
	c2 := constraint.NewConstraint(1, constraint.Distance, e2, e3, 7, false)
	c3 := constraint.NewConstraint(2, constraint.Distance, e3, e4, 2, false)

	g := NewGraphCluster(0) // 0, 1, 2
	g.AddConstraint(c1)
	g.AddConstraint(c2)

	o := NewGraphCluster(1) // 2, 3
	o.AddConstraint(c3)

	g2 := NewGraphCluster(2)
	g3 := NewGraphCluster(3)
	e5 := el.NewSketchPoint(4, 0, 1)
	e6 := el.NewSketchPoint(5, 1, 2)
	c4 := constraint.NewConstraint(3, constraint.Distance, e4, e5, 12, false)
	c5 := constraint.NewConstraint(3, constraint.Distance, e5, e6, 12, false)
	g2.AddConstraint(c4) // 3, 4
	g3.AddConstraint(c5) // 4, 5

	shared := g.SharedElements(g3)

	if shared.Count() != 0 {
		t.Error("There should be no shared element between g and g2, found", shared.Count())
	}

	shared = g.SharedElements(g2)
	if shared.Count() != 0 {
		t.Error("There should be no shared elements between g and g2, found", shared.Count())
	}
	shared = o.SharedElements(g2)
	if shared.Count() != 1 {
		t.Error("There should be no shared elements between g and g2, found", shared.Count())
	}
	if !shared.Contains(3) {
		t.Error("Expected the shared element between g and g2 to have ID 3, got", shared.Contents()[0])
	}

	shared = g2.SharedElements(g3)
	if shared.Count() != 1 {
		t.Error("There should be one shared element between g2 and g3, found", shared.Count())
	}
	if !shared.Contains(4) {
		t.Error("Expected the shared element between g2 and g3 to have ID 4, got", shared.Contents()[0])
	}
}

func TestTranslate(t *testing.T) {
	ea := NewElementRepository()
	ca := NewConstraintRepository()
	e1 := el.NewSketchPoint(0, 0, 1)
	e2 := el.NewSketchPoint(1, 2, 1)
	e3 := el.NewSketchLine(2, 2, 1, -1)
	ea.AddElement(e1)
	ea.AddElement(e2)
	ea.AddElement(e3)
	ea.AddElementToCluster(e1.GetID(), 0)
	ea.AddElementToCluster(e2.GetID(), 0)
	ea.AddElementToCluster(e3.GetID(), 0)
	originalPointNearest := e3.PointNearestOrigin()
	c1 := constraint.NewConstraint(0, constraint.Distance, e1, e2, 5, false)
	c2 := constraint.NewConstraint(1, constraint.Distance, e2, e3, 7, false)
	ca.AddConstraint(c1)
	ca.AddConstraint(c2)

	g := NewGraphCluster(0)
	g.AddConstraint(c1)
	g.AddConstraint(c2)

	g.TranslateCluster(ea, 1, 1)
	ge1, _ := ea.GetElement(g.GetID(), 0)
	e1 = ge1.AsPoint()
	ge2, _ := ea.GetElement(g.GetID(), 1)
	e2 = ge2.AsPoint()
	ge3, _ := ea.GetElement(g.GetID(), 2)
	e3 = ge3.AsLine()

	if e1.GetX() != 1 && e1.GetY() != 2 {
		t.Error("Expected the e1 to be 1, 2, got", e1.GetX(), ",", e1.GetY())
	}
	if e2.GetX() != 3 && e2.GetY() != 2 {
		t.Error("Expected the e1 to be 3, 2, got", e2)
	}
	e3Point := el.NewSketchPoint(0, originalPointNearest.GetX()+1, ((e3.GetA()*(originalPointNearest.GetX()+1) + e3.GetC()) / -e3.GetB()))
	if utils.StandardFloatCompare(e3.DistanceTo(e3Point), 0) != 0 {
		t.Error("Expected e3Point to be on e3. Distance is", e3.DistanceTo(e3Point))
	}
	if utils.StandardFloatCompare(e3Point.GetX(), originalPointNearest.GetX()+1) != 0 {
		t.Error("Expected the X difference between e3 and its original point nearest origin to be 1. Original X", originalPointNearest.GetX(), ", new X", e3Point.GetY())
	}
	if utils.StandardFloatCompare(e3Point.GetY(), originalPointNearest.GetY()+1) != 0 {
		t.Error("Expected the Y difference between e3 and its original point nearest origin to be 1. Original Y", originalPointNearest.GetY(), ", new Y", e3Point.GetY())
	}
}

func TestRotate(t *testing.T) {
	ea := NewElementRepository()
	ca := NewConstraintRepository()
	e1 := el.NewSketchPoint(0, 0, 1)
	e2 := el.NewSketchPoint(1, 2, 1)
	e3 := el.NewSketchLine(2, 2, 1, -1)
	o := el.NewSketchPoint(3, 0, 0)
	ea.AddElement(e1)
	ea.AddElement(e2)
	ea.AddElement(e3)
	ea.AddElement(o)
	ea.AddElementToCluster(e1.GetID(), 0)
	ea.AddElementToCluster(e2.GetID(), 0)
	ea.AddElementToCluster(e3.GetID(), 0)
	ea.AddElementToCluster(o.GetID(), 0)
	c1 := constraint.NewConstraint(0, constraint.Distance, e1, e2, 5, false)
	c2 := constraint.NewConstraint(1, constraint.Distance, e2, e3, 7, false)
	ca.AddConstraint(c1)
	ca.AddConstraint(c2)

	g := NewGraphCluster(0)
	g.AddConstraint(c1)
	g.AddConstraint(c2)

	g.RotateCluster(ea, o, math.Pi/2.0)
	ge1, _ := ea.GetElement(g.GetID(), 0)
	e1 = ge1.AsPoint()
	ge2, _ := ea.GetElement(g.GetID(), 1)
	e2 = ge2.AsPoint()
	ge3, _ := ea.GetElement(g.GetID(), 2)
	e3 = ge3.AsLine()

	mag := math.Sqrt((0.4 * 0.4) + (0.8 * 0.8))
	A := -0.4 / mag
	B := 0.8 / mag
	C := -0.4 / mag
	if utils.StandardFloatCompare(e3.GetA(), A) != 0 ||
		utils.StandardFloatCompare(e3.GetB(), B) != 0 ||
		utils.StandardFloatCompare(e3.GetC(), C) != 0 {
		t.Error("Expected e3 to be", A, ",", B, ",", C, ". Got", e3.GetA(), ",", e3.GetB(), ",", e3.GetC())
	}

	if utils.StandardFloatCompare(e1.GetX(), -1) != 0 ||
		utils.StandardFloatCompare(e1.GetY(), 0.0) != 0 {
		t.Error("Expected -1, 0 got", e1.GetX(), ",", e1.GetY())
	}

	if utils.StandardFloatCompare(e2.GetX(), -1) != 0 ||
		utils.StandardFloatCompare(e2.GetY(), 2.0) != 0 {
		t.Error("Expected -1, 2 got", e2.GetX(), ",", e2.GetY())
	}
}

func TestSolve0(t *testing.T) {
	ea := NewElementRepository()
	ca := NewConstraintRepository()
	g := NewGraphCluster(0)
	/*
		GraphCluster 0 (from test)
			l1: 0.000000x + 1.000000y + 0.000000 = 0
			p1: (0.000000, 0.000000)
			p2: (4.000000, 0.000000)
	*/

	l1 := el.NewSketchLine(0, 0, 1, 0)
	p1 := el.NewSketchPoint(1, 0, 0)
	p2 := el.NewSketchPoint(2, 3.13, 0)
	ea.AddElement(l1)
	ea.AddElement(p1)
	ea.AddElement(p2)
	g.AddElement(l1)
	g.AddElement(p1)
	g.AddElement(p2)
	c1 := constraint.NewConstraint(0, constraint.Distance, p2, p1, 4, false)
	g.AddConstraint(c1)
	ca.AddConstraint(c1)
	c2 := constraint.NewConstraint(1, constraint.Distance, p1, l1, 0, false)
	g.AddConstraint(c2)
	ca.AddConstraint(c2)
	c3 := constraint.NewConstraint(2, constraint.Distance, p2, l1, 0, false)
	g.AddConstraint(c3)
	ca.AddConstraint(c3)

	state := g.Solve(ea, ca)

	c1 = ca.constraints[0]
	c2 = ca.constraints[1]
	c3 = ca.constraints[2]

	if state != solver.Solved {
		t.Error("Expected solved state(4), got", state)
	}

	t.Logf(`elements after solve: 
	l1: %fx + %fy + %f = 0
	p1: (%f, %f)
	p2: (%f, %f)
	`,
		l1.GetA(), l1.GetB(), l1.GetC(),
		p1.GetX(), p1.GetY(),
		p2.GetX(), p2.GetY(),
	)

	cValue := c1.Element1.DistanceTo(c1.Element2)
	if utils.StandardFloatCompare(cValue, c1.Value) != 0 {
		t.Error("Expected point p1 to be distance", c1.Value, "from point p2, distance is", cValue)
	}

	cValue = c2.Element1.DistanceTo(c2.Element2)
	if utils.StandardFloatCompare(cValue, c2.Value) != 0 {
		t.Error("Expected point p1 to be on line l1, distance is", cValue)
	}

	cValue = c3.Element1.DistanceTo(c3.Element2)
	if utils.StandardFloatCompare(cValue, c3.Value) != 0 {
		t.Error("Expected point p2 to be on line l1, distance is", cValue)
	}
}

func TestSolve1(t *testing.T) {
	ea := NewElementRepository()
	ca := NewConstraintRepository()
	g := NewGraphCluster(0)

	/*
		GraphCluster 1 (from test)
			l2: -0.748682x + 0.662930y + 2.341692 = 0
			l3: 0.861839x + 0.507182y + -5.071811 = 0
			p2: (2.132349, -1.124164)
			p3: (4.784067, 1.870563)
	*/

	l2 := el.NewSketchLine(3, -2.27, 2.01, 7.1)
	l3 := el.NewSketchLine(4, 2.45, 2.86, -19.1)
	p2 := el.NewSketchPoint(2, 3.13, 0)
	p3 := el.NewSketchPoint(5, 5.14, 2.27)
	ea.AddElement(p2)
	ea.AddElement(l2)
	ea.AddElement(p3)
	ea.AddElement(l3)
	g.AddElement(p2)
	g.AddElement(l2)
	g.AddElement(p3)
	g.AddElement(l3)
	c1 := constraint.NewConstraint(0, constraint.Distance, p2, p3, 4, false)
	ca.AddConstraint(c1)
	g.AddConstraint(c1)
	c2 := constraint.NewConstraint(1, constraint.Distance, p2, l2, 0, false)
	ca.AddConstraint(c2)
	g.AddConstraint(c2)
	c3 := constraint.NewConstraint(2, constraint.Distance, p3, l2, 0, false)
	ca.AddConstraint(c3)
	g.AddConstraint(c3)
	c4 := constraint.NewConstraint(3, constraint.Distance, p3, l3, 0, false)
	ca.AddConstraint(c4)
	g.AddConstraint(c4)
	c5 := constraint.NewConstraint(4, constraint.Angle, l2, l3, -(108.0/180.0)*math.Pi, false)
	ca.AddConstraint(c5)
	g.AddConstraint(c5)

	// Solves:
	// 0. l1 to l2 angle first
	// 1. Then p2 to l1 and l2
	// 2. Finally p1 to p2 and l1
	state := g.Solve(ea, ca)

	c1 = ca.constraints[0]
	c2 = ca.constraints[1]
	c3 = ca.constraints[2]
	c4 = ca.constraints[3]
	c5 = ca.constraints[4]

	if state != solver.Solved {
		t.Error("Expected solved state(4), got", state)
	}

	t.Logf(`elements after solve: 
	l2: %fx + %fy + %f = 0
	l3: %fx + %fy + %f = 0
	p2: (%f, %f)
	p3: (%f, %f)
	`,
		l2.GetA(), l2.GetB(), l2.GetC(),
		l3.GetA(), l3.GetB(), l3.GetC(),
		p2.GetX(), p2.GetY(),
		p3.GetX(), p3.GetY(),
	)

	cValue := c1.Element1.DistanceTo(c1.Element2)
	if utils.StandardFloatCompare(cValue, c1.Value) != 0 {
		t.Error("Expected point p1 to distance", c1.Value, "from point p2, distance is", cValue)
	}

	cValue = c2.Element1.DistanceTo(c2.Element2)
	if utils.StandardFloatCompare(cValue, c2.Value) != 0 {
		t.Error("Expected point p1 to be on line l1, distance is", cValue)
	}

	cValue = c3.Element1.DistanceTo(c3.Element2)
	if utils.StandardFloatCompare(cValue, c3.Value) != 0 {
		t.Error("Expected point p2 to be on line l1, distance is", cValue)
	}

	cValue = c4.Element1.DistanceTo(c4.Element2)
	if utils.StandardFloatCompare(cValue, c4.Value) != 0 {
		t.Error("Expected point p2 to be on line l2, distance is", cValue)
	}

	angle := c5.Element1.(*el.SketchLine).AngleToLine(c5.Element2.(*el.SketchLine))
	if utils.StandardFloatCompare(angle, c5.Value) != 0 {
		t.Error("Expected line l2 to be", c5.Value, "radians from line l2, angle is", angle)
	}
}

func TestSolve2(t *testing.T) {
	ea := NewElementRepository()
	ca := NewConstraintRepository()
	g := NewGraphCluster(0)

	/*
		A more complicated cluster to solve. The below is a diagram of the desired result

		                * p1
		\              /
		 \ l3         / l5
		  \          /
		p4 *--------* p5
			   l4

		Graph should look like:

		* p1           * l3
		| \ l5    l4 / |
		|  *-------*   |
		| /,------´  \ |
		*--------------*
		p5             p4

		GraphCluster 2 (from test)
			l3: 0.650573x + 0.759444y + -5.071811 = 0
			l4: 0.521236x + -0.853412y + 3.696525 = 0
			l5: -0.972714x + -0.232006y + -1.016993 = 0
			p1: (-0.886306, -0.667527)
			p4: (1.599320, 5.308275)
			p5: (-1.814330, 3.223330)
	*/
	l3 := el.NewSketchLine(4, 2.45, 2.86, -19.1)
	p4 := el.NewSketchPoint(8, 2.28, 4.72)
	l4 := el.NewSketchLine(6, 1.16, -3.32, 13)
	p5 := el.NewSketchPoint(9, -1.04, 3.56)
	l5 := el.NewSketchLine(7, 3.56, 1.04, 0)
	p1 := el.NewSketchPoint(1, 0, 0)
	ea.AddElement(l3)
	ea.AddElement(p4)
	ea.AddElement(l4)
	ea.AddElement(p5)
	ea.AddElement(l5)
	ea.AddElement(p1)
	g.AddElement(l3)
	g.AddElement(p4)
	g.AddElement(l4)
	g.AddElement(p5)
	g.AddElement(l5)
	g.AddElement(p1)
	c1 := constraint.NewConstraint(0, constraint.Distance, p4, l3, 0, false)
	ca.AddConstraint(c1)
	g.AddConstraint(c1)
	c2 := constraint.NewConstraint(1, constraint.Angle, l3, l4, -(108.0/180.0)*math.Pi, false)
	ca.AddConstraint(c2)
	g.AddConstraint(c2)
	c3 := constraint.NewConstraint(2, constraint.Distance, p4, l4, 0, false)
	ca.AddConstraint(c3)
	g.AddConstraint(c3)
	c4 := constraint.NewConstraint(3, constraint.Distance, p5, l4, 0, false)
	ca.AddConstraint(c4)
	g.AddConstraint(c4)
	c5 := constraint.NewConstraint(4, constraint.Distance, p4, p5, 4, false)
	ca.AddConstraint(c5)
	g.AddConstraint(c5)
	c6 := constraint.NewConstraint(5, constraint.Angle, l4, l5, -(108.0/180.0)*math.Pi, false)
	ca.AddConstraint(c6)
	g.AddConstraint(c6)
	c7 := constraint.NewConstraint(6, constraint.Distance, p5, l5, 0, false)
	ca.AddConstraint(c7)
	g.AddConstraint(c7)
	c8 := constraint.NewConstraint(7, constraint.Distance, p1, p5, 4, false)
	ca.AddConstraint(c8)
	g.AddConstraint(c8)
	c9 := constraint.NewConstraint(8, constraint.Distance, p1, l5, 0, false)
	ca.AddConstraint(c9)
	g.AddConstraint(c9)

	state := g.Solve(ea, ca)

	c1 = ca.constraints[0]
	c2 = ca.constraints[1]
	c3 = ca.constraints[2]
	c4 = ca.constraints[3]
	c5 = ca.constraints[4]
	c6 = ca.constraints[5]
	c7 = ca.constraints[6]
	c8 = ca.constraints[7]
	c9 = ca.constraints[8]

	if state != solver.Solved {
		t.Error("Expected solved state(4), got", state)
	}

	t.Logf(`elements after solve: 
	l3: %fx + %fy + %f = 0
	l4: %fx + %fy + %f = 0
	l5: %fx + %fy + %f = 0
	p1: (%f, %f)
	p4: (%f, %f)
	p5: (%f, %f)
	`,
		l3.GetA(), l3.GetB(), l3.GetC(),
		l4.GetA(), l4.GetB(), l4.GetC(),
		l5.GetA(), l5.GetB(), l5.GetC(),
		p1.GetX(), p1.GetY(),
		p4.GetX(), p4.GetY(),
		p5.GetX(), p5.GetY(),
	)

	cValue := c1.Element1.DistanceTo(c1.Element2)
	if utils.StandardFloatCompare(cValue, c1.Value) != 0 {
		t.Error("Expected point p1 to distance", c1.Value, "from point p5, distance is", cValue)
	}

	cValue = c7.Element1.DistanceTo(c7.Element2)
	if utils.StandardFloatCompare(cValue, c7.Value) != 0 {
		t.Error("Expected point p4 to distance", c7.Value, "from point p5, distance is", cValue)
	}

	cValue = c9.Element1.DistanceTo(c9.Element2)
	if utils.StandardFloatCompare(cValue, c9.Value) != 0 {
		t.Error("Expected point p1 to be on line l5, distance is", cValue)
	}

	cValue = c4.Element1.DistanceTo(c4.Element2)
	if utils.StandardFloatCompare(cValue, c4.Value) != 0 {
		t.Error("Expected point p5 to be on line l4, distance is", cValue)
	}

	cValue = c5.Element1.DistanceTo(c5.Element2)
	if utils.StandardFloatCompare(cValue, c5.Value) != 0 {
		t.Error("Expected point p5 to be on line l5, distance is", cValue)
	}

	cValue = c8.Element1.DistanceTo(c8.Element2)
	if utils.StandardFloatCompare(cValue, c8.Value) != 0 {
		t.Error("Expected point p4 to be on line l3, distance is", cValue)
	}

	cValue = c3.Element1.DistanceTo(c3.Element2)
	if utils.StandardFloatCompare(cValue, c3.Value) != 0 {
		t.Error("Expected point p4 to be on line l4, distance is", cValue)
	}

	angle := c2.Element1.AsLine().AngleToLine(c2.Element2.AsLine())
	assert.InDelta(t, math.Abs(angle), math.Abs(c2.Value), utils.StandardCompare, "Expected line l5 angle to be correct")

	angle = c6.Element1.AsLine().AngleToLine(c6.Element2.AsLine())
	assert.InDelta(t, math.Abs(angle), math.Abs(c6.Value), utils.StandardCompare, "Expected line l3 angle to be correct")
}

func TestSolveMerge(t *testing.T) {
	/*
		GraphCluster 0 (from test)
			l1: 0.000000x + 1.000000y + 0.000000 = 0
			p1: (0.000000, 0.000000)
			p2: (4.000000, 0.000000)

		GraphCluster 1 (from test)
			l2: -0.748682x + 0.662930y + 2.341692 = 0
			l3: 0.861839x + 0.507182y + -5.071811 = 0
			p2: (2.132349, -1.124164)
			p3: (4.784067, 1.870563)

		GraphCluster 2 (fron test)
			l3: 0.650573x + 0.759444y + -5.071811 = 0
			l4: 0.521236x + -0.853412y + 3.696525 = 0
			l5: -0.972714x + -0.232006y + -1.016993 = 0
			p1: (-0.886306, -0.667527)
			p4: (1.599320, 5.308275)
			p5: (-1.814330, 3.223330)

		Each cluster shares one element with another:
			GraphCluster 0 and 1 share p2
			GraphCluster 0 and 2 share p1
			GraphCluster 1 and 2 share l3

		solveMerge should merge the three clusters into a single solved graph
	*/
	ea := NewElementRepository()
	ca := NewConstraintRepository()
	g0 := NewGraphCluster(0)

	l1 := el.NewSketchLine(0, 0, 1, 0)
	ea.AddElement(l1)
	ea.AddElementToCluster(0, 0)
	p1 := el.NewSketchPoint(1, 0.0, 0.0)
	ea.AddElement(p1)
	ea.AddElementToCluster(1, 0)
	p2 := el.NewSketchPoint(2, 4.0, 0)
	ea.AddElement(p2)
	ea.AddElementToCluster(2, 0)
	c1 := constraint.NewConstraint(0, constraint.Distance, p1, p2, 4, false)
	ca.AddConstraint(c1)
	g0.AddConstraint(c1)
	c2 := constraint.NewConstraint(1, constraint.Distance, p1, l1, 0, false)
	ca.AddConstraint(c2)
	g0.AddConstraint(c2)
	c3 := constraint.NewConstraint(2, constraint.Distance, p2, l1, 0, false)
	ca.AddConstraint(c3)
	g0.AddConstraint(c3)

	g1 := NewGraphCluster(1)

	l2 := el.NewSketchLine(3, -0.748682, 0.662930, 2.341692)
	ea.AddElement(l2)
	ea.AddElementToCluster(3, 1)
	l3 := el.NewSketchLine(4, 0.861839, 0.507182, -5.071811)
	ea.AddElement(l3)
	ea.AddElementToCluster(4, 1)
	p2 = el.NewSketchPoint(2, 2.132349, -1.124164)
	ea.AddElementToCluster(2, 1)
	p3 := el.NewSketchPoint(5, 4.784067, 1.870563)
	ea.AddElement(p3)
	ea.AddElementToCluster(5, 1)
	c4 := constraint.NewConstraint(3, constraint.Distance, p2, p3, 4, false)
	ca.AddConstraint(c4)
	g1.AddConstraint(c4)
	c5 := constraint.NewConstraint(4, constraint.Distance, p2, l2, 0, false)
	ca.AddConstraint(c5)
	g1.AddConstraint(c5)
	c6 := constraint.NewConstraint(5, constraint.Distance, p3, l2, 0, false)
	ca.AddConstraint(c6)
	g1.AddConstraint(c6)
	c7 := constraint.NewConstraint(6, constraint.Distance, p3, l3, 0, false)
	ca.AddConstraint(c7)
	g1.AddConstraint(c7)
	c8 := constraint.NewConstraint(7, constraint.Angle, l3, l2, (108.0/180.0)*math.Pi, false)
	ca.AddConstraint(c8)
	g1.AddConstraint(c8)

	g2 := NewGraphCluster(2)

	l3 = el.NewSketchLine(4, 0.650573, 0.759444, -5.071811)
	ea.AddElementToCluster(4, 2)
	l4 := el.NewSketchLine(6, 0.521236, -0.853412, 3.696525)
	ea.AddElement(l4)
	ea.AddElementToCluster(6, 1)
	l5 := el.NewSketchLine(7, -0.972714, -0.232006, -1.016993)
	ea.AddElement(l5)
	ea.AddElementToCluster(7, 1)
	p1 = el.NewSketchPoint(1, -0.886306, -0.667527)
	ea.AddElementToCluster(1, 2)
	p4 := el.NewSketchPoint(8, 1.599320, 5.308275)
	ea.AddElement(p4)
	ea.AddElementToCluster(8, 1)
	p5 := el.NewSketchPoint(9, -1.814330, 3.223330)
	ea.AddElement(p5)
	ea.AddElementToCluster(9, 1)
	c9 := constraint.NewConstraint(8, constraint.Distance, p1, p5, 4, false)
	ca.AddConstraint(c9)
	g2.AddConstraint(c9)
	c10 := constraint.NewConstraint(9, constraint.Distance, p4, p5, 4, false)
	ca.AddConstraint(c10)
	g2.AddConstraint(c10)
	c11 := constraint.NewConstraint(10, constraint.Distance, p1, l5, 0, false)
	ca.AddConstraint(c11)
	g2.AddConstraint(c11)
	c12 := constraint.NewConstraint(11, constraint.Distance, p5, l4, 0, false)
	ca.AddConstraint(c12)
	g2.AddConstraint(c12)
	c13 := constraint.NewConstraint(12, constraint.Distance, p5, l5, 0, false)
	ca.AddConstraint(c13)
	g2.AddConstraint(c13)
	c14 := constraint.NewConstraint(13, constraint.Distance, p4, l3, 0, false)
	ca.AddConstraint(c14)
	g2.AddConstraint(c14)
	c15 := constraint.NewConstraint(14, constraint.Distance, p4, l4, 0, false)
	ca.AddConstraint(c15)
	g2.AddConstraint(c15)
	c16 := constraint.NewConstraint(15, constraint.Angle, l5, l4, (108.0/180.0)*math.Pi, false)
	ca.AddConstraint(c16)
	g2.AddConstraint(c16)
	c17 := constraint.NewConstraint(16, constraint.Angle, l3, l4, (108.0/180.0)*math.Pi, false)
	ca.AddConstraint(c17)
	g2.AddConstraint(c17)

	state := g0.solveMerge(ea, ca, g1, g2)

	if state != solver.Solved {
		t.Error("Expected solved state(4), got", state)
	}

	t.Logf("g0 elements length %d\n", g0.elements.Count())
	t.Logf("all elements count %d\n", len(ea.elements))
	t.Logf("element %d type %v\n", 0, ea.elements[0].GetType())
	l1 = ea.elements[0].AsLine()
	t.Logf("element %d type %v\n", 1, ea.elements[1].GetType())
	p1 = ea.elements[1].AsPoint()
	t.Logf("element %d type %v\n", 2, ea.elements[2].GetType())
	p2 = ea.elements[2].AsPoint()
	t.Logf("element %d type %v\n", 3, ea.elements[3].GetType())
	l2 = ea.elements[3].AsLine()
	t.Logf("element %d type %v\n", 4, ea.elements[4].GetType())
	l3 = ea.elements[4].AsLine()
	t.Logf("element %d type %v\n", 5, ea.elements[5].GetType())
	p3 = ea.elements[5].AsPoint()
	t.Logf("element %d type %v\n", 6, ea.elements[6].GetType())
	l4 = ea.elements[6].AsLine()
	t.Logf("element %d type %v\n", 7, ea.elements[7].GetType())
	l5 = ea.elements[7].AsLine()
	t.Logf("element %d type %v\n", 8, ea.elements[8].GetType())
	p4 = ea.elements[8].AsPoint()
	t.Logf("element %d type %v\n", 9, ea.elements[9].GetType())
	p5 = ea.elements[9].AsPoint()

	rad2Deg := func(rad float64) float64 { return (rad / math.Pi) * 180 }
	deg2Rad := func(deg float64) float64 { return (deg / 180.0) * math.Pi }
	desired := deg2Rad(72)
	angle := l1.AngleToLine(l2)
	if utils.StandardFloatCompare(angle, desired) != 0 {
		t.Error("Expected l1 to l2 to be", 72, "degrees, got", rad2Deg(angle))
	}
	desired = deg2Rad(-108)
	angle = l2.AngleToLine(l3)
	if utils.StandardFloatCompare(angle, desired) != 0 {
		t.Error("Expected l2 to l3 to be", -108, "degrees, got", rad2Deg(angle))
	}
	desired = deg2Rad(-108)
	angle = l3.AngleToLine(l4)
	if utils.StandardFloatCompare(angle, desired) != 0 {
		t.Error("Expected l3 to l4 to be", -108, "degrees, got", rad2Deg(angle))
	}
	desired = deg2Rad(-108)
	angle = l4.AngleToLine(l5)
	if utils.StandardFloatCompare(angle, desired) != 0 {
		t.Error("Expected l4 to l5 to be", -108, "degrees, got", rad2Deg(angle))
	}
	desired = deg2Rad(-108)
	angle = l5.AngleToLine(l1)
	if utils.StandardFloatCompare(angle, desired) != 0 {
		t.Error("Expected l5 to l1 to be", -108, "degrees, got", rad2Deg(angle))
	}

	desired = 4.0
	measured := p1.DistanceTo(p2)
	if utils.StandardFloatCompare(measured, desired) != 0 {
		t.Error("Expected p1 to p2 to be", desired, ", got", measured)
	}
	measured = p2.DistanceTo(p3)
	if utils.StandardFloatCompare(measured, desired) != 0 {
		t.Error("Expected p2 to p3 to be", desired, ", got", measured)
	}
	measured = p3.DistanceTo(p4)
	if utils.StandardFloatCompare(measured, desired) != 0 {
		t.Error("Expected p3 to p4 to be", desired, ", got", measured)
	}
	measured = p4.DistanceTo(p5)
	if utils.StandardFloatCompare(measured, desired) != 0 {
		t.Error("Expected p4 to p5 to be", desired, ", got", measured)
	}
	measured = p5.DistanceTo(p1)
	if utils.StandardFloatCompare(measured, desired) != 0 {
		t.Error("Expected p5 to p1 to be", desired, ", got", measured)
	}

	t.Logf(`elements after solve: 
	l1: %fx + %fy + %f = 0
	l2: %fx + %fy + %f = 0
	l3: %fx + %fy + %f = 0
	l4: %fx + %fy + %f = 0
	l5: %fx + %fy + %f = 0
	p1: (%f, %f)
	p2: (%f, %f)
	p3: (%f, %f)
	p4: (%f, %f)
	p5: (%f, %f)
	`,
		l1.GetA(), l1.GetB(), l1.GetC(),
		l2.GetA(), l2.GetB(), l2.GetC(),
		l3.GetA(), l3.GetB(), l3.GetC(),
		l4.GetA(), l4.GetB(), l4.GetC(),
		l5.GetA(), l5.GetB(), l5.GetC(),
		p1.GetX(), p1.GetY(),
		p2.GetX(), p2.GetY(),
		p3.GetX(), p3.GetY(),
		p4.GetX(), p4.GetY(),
		p5.GetX(), p5.GetY(),
	)
}

func TestSolvedUnsolvedConstraintsFor(t *testing.T) {
	ca := NewConstraintRepository()
	e1 := el.NewSketchPoint(0, 0, 1)
	e2 := el.NewSketchPoint(1, 2, 1)
	e3 := el.NewSketchLine(2, 2, 1, -1)
	e4 := el.NewSketchLine(3, 2, 2, -0)
	c1 := constraint.NewConstraint(0, constraint.Distance, e1, e2, 5, false)
	c2 := constraint.NewConstraint(1, constraint.Distance, e2, e3, 7, false)
	c3 := constraint.NewConstraint(2, constraint.Distance, e3, e4, 2, false)

	g := NewGraphCluster(0)
	ca.AddConstraint(c1)
	ca.AddConstraint(c2)
	ca.AddConstraint(c3)
	g.AddConstraint(c1)
	g.AddConstraint(c2)
	g.AddConstraint(c3)
	g.solved.Add(c1.GetID())
	g.solved.Add(c3.GetID())

	tests := []struct {
		name     string
		eId      uint
		solved   []uint
		unsolved []uint
	}{
		// Keep these lists sorted
		{"constraints for element 0", 0, []uint{0}, []uint{}},
		{"constraints for element 1", 1, []uint{0}, []uint{1}},
		{"constraints for element 2", 2, []uint{2}, []uint{1}},
		{"constraints for element 3", 3, []uint{2}, []uint{}},
	}
	for _, tt := range tests {
		var solved constraint.ConstraintList = g.solvedConstraintsFor(ca, tt.eId)
		var unsolved constraint.ConstraintList = g.unsolvedConstraintsFor(ca, tt.eId)
		sort.Sort(solved)
		sort.Sort(unsolved)
		assert.Equal(t, len(tt.solved), len(solved), tt.name)
		assert.Equal(t, len(tt.unsolved), len(unsolved), tt.name)
		for i, c := range solved {
			assert.Equal(t, tt.solved[i], c.GetID())
		}
		for i, c := range unsolved {
			assert.Equal(t, tt.unsolved[i], c.GetID())
		}
	}
}

func TestSolveEdgeCases(t *testing.T) {
	ea := NewElementRepository()
	ca := NewConstraintRepository()
	e3 := el.NewSketchLine(2, 2, 1, -1)
	ea.AddElement(e3)
	ea.AddElementToCluster(2, 1)
	e4 := el.NewSketchLine(3, 2, 2, -0)
	ea.AddElement(e4)
	ea.AddElementToCluster(3, 1)
	c3 := constraint.NewConstraint(2, constraint.Distance, e3, e4, 2, false)
	ca.AddConstraint(c3)

	o := NewGraphCluster(1)
	o.AddConstraint(c3)

	state := o.Solve(ea, ca)
	assert.Equal(t, solver.NonConvergent, state, "Test local solve with solveorder < 2")

	e1 := el.NewSketchPoint(0, 0, 1)
	ea.AddElement(e1)
	ea.AddElementToCluster(0, 1)
	e2 := el.NewSketchPoint(1, 2, 1)
	ea.AddElement(e2)
	ea.AddElementToCluster(1, 1)
	c1 := constraint.NewConstraint(0, constraint.Distance, e1, e2, 5, false)
	ca.AddConstraint(c1)
	o.AddConstraint(c1)

	state = o.Solve(ea, ca)
	assert.Equal(t, solver.NonConvergent, state, "Test local solve without enough constraints to solve desired element")

	c2 := constraint.NewConstraint(1, constraint.Distance, e2, e3, 7, false)
	ca.AddConstraint(c2)
	g := NewGraphCluster(0)
	g.AddConstraint(c1)
	g.AddConstraint(c2)
	g.AddConstraint(c3)
	g.solved.Add(c2.GetID())
	g.solved.Add(c3.GetID())

	state = g.Solve(ea, ca)
	assert.Equal(t, solver.Solved, state, "Test local solve with pre-solved elements")
}

func TestMergeOne(t *testing.T) {
	ea := NewElementRepository()
	ca := NewConstraintRepository()
	// Create fixed element cluster
	// Create cluster w/ square
	// merge the two -- use solveMerge instead of mergeOne
	g := NewGraphCluster(0)
	var e el.SketchElement = el.NewSketchPoint(0, 0, 0)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), g.GetID())
	g.AddElement(e)
	e = el.NewSketchLine(1, 0, 1, 0)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), g.GetID())
	g.AddElement(e)
	e = el.NewSketchLine(2, 1, 0, 0)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), g.GetID())
	g.AddElement(e)

	o := NewGraphCluster(1)
	e = el.NewSketchLine(1, -0.029929, -0.999552, 0)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o.GetID())
	o.AddElement(e)

	state := g.solveMerge(ea, ca, o, nil)
	assert.Equal(t, solver.NonConvergent, state, "Merge containing only one shared element should fail to solve")

	e = el.NewSketchLine(2, 0.999552, -0.029929, 0)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o.GetID())
	o.AddElement(e)

	state = g.solveMerge(ea, ca, o, nil)
	assert.Equal(t, solver.NonConvergent, state, "Merge where shared elements are both lines should fail to solve")

	o = NewGraphCluster(1)
	e = el.NewSketchPoint(0, 0, 0)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o.GetID())
	o.AddElement(e)
	e = el.NewSketchPoint(5, 3.998208, -0.119717)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o.GetID())
	o.AddElement(e)
	e = el.NewSketchLine(12, -0.563309, 0.826247, -3.804226)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o.GetID())
	o.AddElement(e)
	e = el.NewSketchPoint(11, 2.183330, 6.092751)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o.GetID())
	o.AddElement(e)
	e = el.NewSketchLine(9, 0.611735, 0.791063, -6.155367)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o.GetID())
	o.AddElement(e)
	e = el.NewSketchPoint(8, 5.347580, 3.645810)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o.GetID())
	o.AddElement(e)
	e = el.NewSketchLine(1, -0.029929, -0.999552, 0)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o.GetID())
	o.AddElement(e)
	e = el.NewSketchLine(2, 0.999552, -0.029929, 0)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o.GetID())
	o.AddElement(e)
	e = el.NewSketchLine(15, -0.959879, -0.280414, 0)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o.GetID())
	o.AddElement(e)
	e = el.NewSketchPoint(14, -1.121656, 3.839516)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o.GetID())
	o.AddElement(e)
	e = el.NewSketchLine(6, 0.941382, -0.337343, -3.804226)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o.GetID())
	o.AddElement(e)
	e = el.NewSketchLine(3, -0.029929, -0.999552, 0)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o.GetID())
	o.AddElement(e)
	state = g.solveMerge(ea, ca, o, nil)
	assert.Equal(t, solver.Solved, state, "Merge should solve successfully")

	g = NewGraphCluster(0)
	e = el.NewSketchPoint(0, 0, 0)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), g.GetID())
	g.AddElement(e)
	e = el.NewSketchLine(100, 0, 1, 0)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), g.GetID())
	g.AddElement(e)
	e = el.NewSketchPoint(5, 4, 0)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), g.GetID())
	g.AddElement(e)

	state = g.solveMerge(ea, ca, o, nil)
	assert.Equal(t, solver.Solved, state, "Merge with two shared points should solve successfully")
}

func TestSolveMergeEdgeCases(t *testing.T) {
	ea := NewElementRepository()
	ca := NewConstraintRepository()
	g := NewGraphCluster(0)
	var e el.SketchElement = el.NewSketchPoint(0, 0, 0)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), g.GetID())
	g.AddElement(e)
	e = el.NewSketchLine(1, 0, 1, 0)
	g.AddElement(e)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), g.GetID())
	e = el.NewSketchLine(2, 1, 0, 0)
	g.AddElement(e)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), g.GetID())

	o1 := NewGraphCluster(1)
	e = el.NewSketchPoint(0, 0, 0)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o1.GetID())
	o1.AddElement(e)
	e = el.NewSketchPoint(5, 3.998208, -0.119717)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o1.GetID())
	o1.AddElement(e)
	e = el.NewSketchLine(12, -0.563309, 0.826247, -3.804226)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o1.GetID())
	o1.AddElement(e)
	e = el.NewSketchPoint(11, 2.183330, 6.092751)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o1.GetID())
	o1.AddElement(e)
	e = el.NewSketchLine(9, 0.611735, 0.791063, -6.155367)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o1.GetID())
	o1.AddElement(e)

	o2 := NewGraphCluster(2)
	e = el.NewSketchLine(15, -0.959879, -0.280414, 0)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o2.GetID())
	o2.AddElement(e)
	e = el.NewSketchPoint(5, -1.121656, 3.839516)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o2.GetID())
	o2.AddElement(e)
	e = el.NewSketchLine(6, 0.941382, -0.337343, -3.804226)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o2.GetID())
	o2.AddElement(e)
	e = el.NewSketchLine(3, -0.029929, -0.999552, 0)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o2.GetID())
	o2.AddElement(e)

	state := g.solveMerge(ea, ca, o1, o2)
	assert.Equal(t, solver.NonConvergent, state, "Three cluster solve with only two shared elements should fail")

	ea.Clear()
	// Solve merge with three lines lines
	g = NewGraphCluster(0)
	e = el.NewSketchPoint(0, 0, 0)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), g.GetID())
	g.AddElement(e)
	e = el.NewSketchLine(9, 0.611735, 0.791063, -6.155367)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), g.GetID())
	g.AddElement(e)
	e = el.NewSketchLine(1, 0, 1, 0)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), g.GetID())
	g.AddElement(e)

	o1 = NewGraphCluster(1)
	e = el.NewSketchPoint(7, 0, 1)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o1.GetID())
	o1.AddElement(e)
	e = el.NewSketchPoint(5, 3.998208, -0.119717)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o1.GetID())
	o1.AddElement(e)
	e = el.NewSketchPoint(11, 2.183330, 6.092751)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o1.GetID())
	o1.AddElement(e)
	e = el.NewSketchLine(9, 0.611735, 0.791063, -6.155367)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o1.GetID())
	o1.AddElement(e)
	e = el.NewSketchLine(2, 1, 0, 0)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o1.GetID())
	o1.AddElement(e)

	o2 = NewGraphCluster(2)
	e = el.NewSketchLine(15, -0.959879, -0.280414, 0)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o2.GetID())
	o2.AddElement(e)
	e = el.NewSketchPoint(8, 3.998208, -0.119717)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o2.GetID())
	o2.AddElement(e)
	e = el.NewSketchLine(1, 0, 1, 0)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o2.GetID())
	o2.AddElement(e)
	e = el.NewSketchLine(2, 1, 0.0, 0.0)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o2.GetID())
	o2.AddElement(e)

	state = g.solveMerge(ea, ca, o1, o2)
	assert.Equal(t, solver.Solved, state, "Three cluster solve with three lines")

	// Solve merge with one point and two lines where lines are in clusters 0 and 1
	// I don't know where I got these values... they may be incorrect
	g = NewGraphCluster(0)
	e = el.NewSketchPoint(0, 0, 0)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), g.GetID())
	g.AddElement(e)
	e = el.NewSketchLine(3, 0, -1, 0)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), g.GetID())
	g.AddElement(e)
	e = el.NewSketchPoint(5, 4, 0)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), g.GetID())
	g.AddElement(e)
	e = el.NewSketchLine(1, 0, -1, 0)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), g.GetID())
	g.AddElement(e)

	o1 = NewGraphCluster(1)
	e = el.NewSketchPoint(11, 2.183330, 6.092751)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o1.GetID())
	o1.AddElement(e)
	e = el.NewSketchLine(12, -0.563309, 0.826247, -3.804226)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o1.GetID())
	o1.AddElement(e)
	e = el.NewSketchLine(15, -3.839516, -1.121656, 0)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o1.GetID())
	o1.AddElement(e)
	e = el.NewSketchLine(9, 0.611735, 0.791063, -6.155367)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o1.GetID())
	o1.AddElement(e)
	e = el.NewSketchPoint(14, -1.121656, 3.839516)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o1.GetID())
	o1.AddElement(e)
	e = el.NewSketchPoint(0, 0, 0)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o1.GetID())
	o1.AddElement(e)

	o2 = NewGraphCluster(2)
	e = el.NewSketchPoint(8, 5.14, 2.27)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o2.GetID())
	o2.AddElement(e)
	e = el.NewSketchLine(6, 2.029929, -2.651719, -9.373495)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o2.GetID())
	o2.AddElement(e)
	e = el.NewSketchPoint(5, 2.488281, -0.724727)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o2.GetID())
	o2.AddElement(e)
	e = el.NewSketchLine(9, 0.861839, 0.507182, -5.581155)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o2.GetID())
	o2.AddElement(e)

	state = g.solveMerge(ea, ca, o1, o2)
	assert.Equal(t, solver.Solved, state, "Three cluster solve with three shared elements should solve")

	ea.clusterElements[g.GetID()][0] = el.NewSketchLine(0, 1, 1, 1)
	ea.clusterElements[o1.GetID()][0] = el.NewSketchLine(0, 3, 2, 1)
	line := ea.elements[9]
	line.AsLine().SetB(0.235)
	line.AsLine().SetC(2)
	state = g.solveMerge(ea, ca, o1, o2)
	assert.Equal(t, solver.NonConvergent, state, "Three cluster solve with non-convergent lines")

	ea.clusterElements[o1.GetID()][2] = el.NewSketchLine(2, 1, 0, 0)
	ea.clusterElements[o2.GetID()][2] = el.NewSketchLine(2, 1, 0, 0)
	ea.clusterElements[g.GetID()][1] = el.NewSketchLine(1, 1, 0, 6)
	ea.clusterElements[o2.GetID()][1] = el.NewSketchLine(1, 1, 0, 6)
	ea.clusterElements[g.GetID()][0] = el.NewSketchPoint(0, -10, 1)
	ea.clusterElements[o1.GetID()][0] = el.NewSketchPoint(0, 3, 2)
	state = g.solveMerge(ea, ca, o1, o2)
	assert.Equal(t, solver.NonConvergent, state, "Fail to solve final element")

	ea.Clear()

	g = NewGraphCluster(0)
	e = el.NewSketchPoint(0, 0, 0)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), g.GetID())
	g.AddElement(e)
	e = el.NewSketchLine(9, 0.611735, 0.791063, -6.155367)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), g.GetID())
	g.AddElement(e)
	e = el.NewSketchLine(1, 0, 1, 0)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), g.GetID())
	g.AddElement(e)

	o1 = NewGraphCluster(1)
	e = el.NewSketchPoint(0, 0, 0)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o1.GetID())
	o1.AddElement(e)
	e = el.NewSketchPoint(5, 3.998208, -0.119717)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o1.GetID())
	o1.AddElement(e)
	e = el.NewSketchPoint(11, 2.183330, 6.092751)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o1.GetID())
	o1.AddElement(e)
	e = el.NewSketchLine(6, 0.611735, 0.791063, -6.155367)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o1.GetID())
	o1.AddElement(e)
	e = el.NewSketchLine(2, 1, 0, 0)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o1.GetID())
	o1.AddElement(e)

	o2 = NewGraphCluster(2)
	e = el.NewSketchLine(15, -0.959879, -0.280414, 0)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o2.GetID())
	o2.AddElement(e)
	e = el.NewSketchPoint(8, 3.998208, -0.119717)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o2.GetID())
	o2.AddElement(e)
	e = el.NewSketchLine(1, 0, 1, 0)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o2.GetID())
	o2.AddElement(e)
	e = el.NewSketchLine(2, 1, 0.0, 0.0)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o2.GetID())
	o2.AddElement(e)
	e = el.NewSketchPoint(0, 0, 0)
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o2.GetID())
	o2.AddElement(e)

	state = g.solveMerge(ea, ca, o1, o2)
	assert.Equal(t, solver.NonConvergent, state, "Nonconvergent where an element has too many parents")
}

func TestToGraphViz(t *testing.T) {
	ca := NewConstraintRepository()
	e1 := el.NewSketchPoint(0, 0, 1)
	e2 := el.NewSketchPoint(1, 2, 1)
	e3 := el.NewSketchLine(2, 2, 1, -1)
	e4 := el.NewSketchLine(3, 2, 2, -0)
	c1 := constraint.NewConstraint(0, constraint.Distance, e1, e2, 5, false)
	c2 := constraint.NewConstraint(1, constraint.Distance, e2, e3, 7, false)
	c3 := constraint.NewConstraint(2, constraint.Distance, e3, e4, 2, false)

	g := NewGraphCluster(0)
	ca.AddConstraint(c1)
	g.AddConstraint(c1)
	ca.AddConstraint(c2)
	g.AddConstraint(c2)
	ca.AddConstraint(c3)
	g.AddConstraint(c3)

	gvString := g.ToGraphViz(ca)
	assert.Contains(t, gvString, "subgraph cluster_0")
	assert.Contains(t, gvString, "label = \"Cluster 0\"")
	assert.Contains(t, gvString, c1.ToGraphViz(0), "GraphViz output contains constraint 1")
	assert.Contains(t, gvString, c2.ToGraphViz(0), "GraphViz output contains constraint 2")
	assert.Contains(t, gvString, c3.ToGraphViz(0), "GraphViz output contains constraint 3")
	assert.Contains(t, gvString, e1.ToGraphViz(0), "GraphViz output contains element 1")
	assert.Contains(t, gvString, e2.ToGraphViz(0), "GraphViz output contains element 2")
	assert.Contains(t, gvString, e3.ToGraphViz(0), "GraphViz output contains element 3")
	assert.Contains(t, gvString, e4.ToGraphViz(0), "GraphViz output contains element 4")
}
