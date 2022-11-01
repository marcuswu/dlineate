package core

import (
	"fmt"
	"math"
	"sort"
	"testing"

	"github.com/marcuswu/dlineation/internal/constraint"
	el "github.com/marcuswu/dlineation/internal/element"
	"github.com/marcuswu/dlineation/internal/solver"
	"github.com/marcuswu/dlineation/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddElement(t *testing.T) {
	e1 := el.NewSketchPoint(0, 0, 1)
	e2 := el.NewSketchPoint(1, 2, 1)

	g := NewGraphCluster(1)
	g.AddElement(e1)

	if len(g.elements) != 1 {
		t.Error("expected one element to be added to the cluster, found", len(g.elements))
	}

	if len(g.solveOrder) != 1 {
		t.Error("expected one element to be added to the cluster's solve order, found", len(g.solveOrder))
	}

	g.AddElement(e1)

	if len(g.elements) != 1 {
		t.Error("expected no change tothe cluster element length, found", len(g.elements))
	}

	if len(g.solveOrder) != 1 {
		t.Error("expected no change to the cluster's solve order, found", len(g.solveOrder))
	}

	g.AddElement(e2)

	if len(g.elements) != 2 {
		t.Error("expected two elements to be added to the cluster, found", len(g.elements))
	}

	if len(g.solveOrder) != 2 {
		t.Error("expected two elements to be added to the cluster's solve order, found", len(g.solveOrder))
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

	if len(g.constraints) != 1 {
		t.Error("expected graph cluster to have one constraint, found", len(g.constraints))
	}
	if len(g.elements) != 2 {
		t.Error("expected graph cluster to have 2 elements, found", len(g.elements))
	}

	c1.Solved = true
	g.AddConstraint(c1)
	if len(g.constraints) != 1 {
		t.Error("expected no change to cluster constraints after adding the same constraint twice")
	}
	if len(g.elements) != 2 {
		t.Error("expected no change to elements after adding the same constraint twice")
	}
	if !g.constraints[c1.GetID()].Solved {
		t.Error("expected constraint solve state to change after adding the solved constraint")
	}

	g.AddConstraint(c2)

	if len(g.constraints) != 2 {
		t.Error("expected graph cluster to have two constraint, found", len(g.constraints))
	}
	if len(g.elements) != 3 {
		t.Error("expected graph cluster to have 3 elements, found", len(g.elements))
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

	if !g.HasElementID(0) {
		t.Error("expected graph cluster to have element 0, but element was not found")
	}
	if !g.HasElementID(1) {
		t.Error("expected graph cluster to have element 1, but element was not found")
	}
	if !g.HasElementID(2) {
		t.Error("expected graph cluster to have element 2, but element was not found")
	}
	if g.HasElementID(3) {
		t.Error("expected graph cluster to have element 3, but element was not found")
	}
	if g.HasElementID(4) {
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
		assert.True(t, g.HasElement(e), fmt.Sprintf("cluster 1 has expected element %d", e.GetID()))
	}
	for _, e := range elements2 {
		assert.True(t, o.HasElement(e), fmt.Sprintf("cluster 2 has expected element %d", e.GetID()))
	}
	falseElement := el.NewSketchPoint(100, 0, 0)
	assert.False(t, g.HasElement(falseElement))
	assert.False(t, o.HasElement(falseElement))
	assert.True(t, g.HasElement(nil))
}

func TestGetElement(t *testing.T) {
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

	element3, ok := g.GetElement(2)
	if !ok {
		t.Error("Should find element with id 2 got", ok)
	}
	if element3.GetType() != el.Line || utils.StandardFloatCompare(e3.AngleToLine(element3.AsLine()), 0) != 0 {
		t.Error("Element with id 2 should be equal to ", e3, ", got", element3)
	}

	element3, ok = g.GetElement(4)
	if ok {
		t.Error("Should not find element with id 4, got", ok)
	}
	if element3 != nil {
		t.Error("Element with id 4 should be nil, got", element3)
	}
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
	e1 := el.NewSketchPoint(0, 0, 1)
	e2 := el.NewSketchPoint(1, 2, 1)
	e3 := el.NewSketchLine(2, 2, 1, -1)
	originalPointNearest := e3.PointNearestOrigin()
	c1 := constraint.NewConstraint(0, constraint.Distance, e1, e2, 5, false)
	c2 := constraint.NewConstraint(1, constraint.Distance, e2, e3, 7, false)

	g := NewGraphCluster(0)
	g.AddConstraint(c1)
	g.AddConstraint(c2)

	g.Translate(1, 1)
	e1 = g.elements[0].AsPoint()
	e2 = g.elements[1].AsPoint()
	e3 = g.elements[2].AsLine()

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
	e1 := el.NewSketchPoint(0, 0, 1)
	e2 := el.NewSketchPoint(1, 2, 1)
	e3 := el.NewSketchLine(2, 2, 1, -1)
	o := el.NewSketchPoint(3, 0, 0)
	c1 := constraint.NewConstraint(0, constraint.Distance, e1, e2, 5, false)
	c2 := constraint.NewConstraint(1, constraint.Distance, e2, e3, 7, false)

	g := NewGraphCluster(0)
	g.AddConstraint(c1)
	g.AddConstraint(c2)

	g.Rotate(o, math.Pi/2.0)
	e1 = g.elements[0].AsPoint()
	e2 = g.elements[1].AsPoint()
	e3 = g.elements[2].AsLine()

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

func TestLocalSolve0(t *testing.T) {
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
	g.AddElement(l1)
	g.AddElement(p1)
	g.AddElement(p2)
	c1 := constraint.NewConstraint(0, constraint.Distance, p2, p1, 4, false)
	g.AddConstraint(c1)
	c2 := constraint.NewConstraint(1, constraint.Distance, p1, l1, 0, false)
	g.AddConstraint(c2)
	c3 := constraint.NewConstraint(2, constraint.Distance, p2, l1, 0, false)
	g.AddConstraint(c3)

	state := g.localSolve()

	c1 = g.constraints[0]
	c2 = g.constraints[1]
	c3 = g.constraints[2]

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

func TestLocalSolve1(t *testing.T) {
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
	g.AddElement(p2)
	g.AddElement(l2)
	g.AddElement(p3)
	g.AddElement(l3)
	c1 := constraint.NewConstraint(0, constraint.Distance, p2, p3, 4, false)
	g.AddConstraint(c1)
	c2 := constraint.NewConstraint(1, constraint.Distance, p2, l2, 0, false)
	g.AddConstraint(c2)
	c3 := constraint.NewConstraint(2, constraint.Distance, p3, l2, 0, false)
	g.AddConstraint(c3)
	c4 := constraint.NewConstraint(3, constraint.Distance, p3, l3, 0, false)
	g.AddConstraint(c4)
	c5 := constraint.NewConstraint(4, constraint.Angle, l2, l3, -(108.0/180.0)*math.Pi, false)
	g.AddConstraint(c5)

	// Solves:
	// 0. l1 to l2 angle first
	// 1. Then p2 to l1 and l2
	// 2. Finally p1 to p2 and l1
	state := g.localSolve()

	c1 = g.constraints[0]
	c2 = g.constraints[1]
	c3 = g.constraints[2]
	c4 = g.constraints[3]
	c5 = g.constraints[4]

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

func TestLocalSolve2(t *testing.T) {
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
	g.AddElement(l3)
	g.AddElement(p4)
	g.AddElement(l4)
	g.AddElement(p5)
	g.AddElement(l5)
	g.AddElement(p1)
	c1 := constraint.NewConstraint(0, constraint.Distance, p4, l3, 0, false)
	g.AddConstraint(c1)
	c2 := constraint.NewConstraint(1, constraint.Angle, l3, l4, -(108.0/180.0)*math.Pi, false)
	g.AddConstraint(c2)
	c3 := constraint.NewConstraint(2, constraint.Distance, p4, l4, 0, false)
	g.AddConstraint(c3)
	c4 := constraint.NewConstraint(3, constraint.Distance, p5, l4, 0, false)
	g.AddConstraint(c4)
	c5 := constraint.NewConstraint(4, constraint.Distance, p4, p5, 4, false)
	g.AddConstraint(c5)
	c6 := constraint.NewConstraint(5, constraint.Angle, l4, l5, -(108.0/180.0)*math.Pi, false)
	g.AddConstraint(c6)
	c7 := constraint.NewConstraint(6, constraint.Distance, p5, l5, 0, false)
	g.AddConstraint(c7)
	c8 := constraint.NewConstraint(7, constraint.Distance, p1, p5, 4, false)
	g.AddConstraint(c8)
	c9 := constraint.NewConstraint(8, constraint.Distance, p1, l5, 0, false)
	g.AddConstraint(c9)

	state := g.localSolve()

	c1 = g.constraints[0]
	c2 = g.constraints[1]
	c3 = g.constraints[2]
	c4 = g.constraints[3]
	c5 = g.constraints[4]
	c6 = g.constraints[5]
	c7 = g.constraints[6]
	c8 = g.constraints[7]
	c9 = g.constraints[8]

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
	g0 := NewGraphCluster(0)

	l1 := el.NewSketchLine(0, 0, 1, 0)
	p1 := el.NewSketchPoint(1, 0.0, 0.0)
	p2 := el.NewSketchPoint(2, 4.0, 0)
	c1 := constraint.NewConstraint(0, constraint.Distance, p1, p2, 4, false)
	g0.AddConstraint(c1)
	c2 := constraint.NewConstraint(1, constraint.Distance, p1, l1, 0, false)
	g0.AddConstraint(c2)
	c3 := constraint.NewConstraint(2, constraint.Distance, p2, l1, 0, false)
	g0.AddConstraint(c3)

	g1 := NewGraphCluster(1)

	l2 := el.NewSketchLine(3, -0.748682, 0.662930, 2.341692)
	l3 := el.NewSketchLine(4, 0.861839, 0.507182, -5.071811)
	p2 = el.NewSketchPoint(2, 2.132349, -1.124164)
	p3 := el.NewSketchPoint(5, 4.784067, 1.870563)
	c4 := constraint.NewConstraint(3, constraint.Distance, p2, p3, 4, false)
	g1.AddConstraint(c4)
	c5 := constraint.NewConstraint(4, constraint.Distance, p2, l2, 0, false)
	g1.AddConstraint(c5)
	c6 := constraint.NewConstraint(5, constraint.Distance, p3, l2, 0, false)
	g1.AddConstraint(c6)
	c7 := constraint.NewConstraint(6, constraint.Distance, p3, l3, 0, false)
	g1.AddConstraint(c7)
	c8 := constraint.NewConstraint(7, constraint.Angle, l3, l2, (108.0/180.0)*math.Pi, false)
	g1.AddConstraint(c8)

	g2 := NewGraphCluster(2)

	l3 = el.NewSketchLine(4, 0.650573, 0.759444, -5.071811)
	l4 := el.NewSketchLine(6, 0.521236, -0.853412, 3.696525)
	l5 := el.NewSketchLine(7, -0.972714, -0.232006, -1.016993)
	p1 = el.NewSketchPoint(1, -0.886306, -0.667527)
	p4 := el.NewSketchPoint(8, 1.599320, 5.308275)
	p5 := el.NewSketchPoint(9, -1.814330, 3.223330)
	c9 := constraint.NewConstraint(8, constraint.Distance, p1, p5, 4, false)
	g2.AddConstraint(c9)
	c10 := constraint.NewConstraint(9, constraint.Distance, p4, p5, 4, false)
	g2.AddConstraint(c10)
	c11 := constraint.NewConstraint(10, constraint.Distance, p1, l5, 0, false)
	g2.AddConstraint(c11)
	c12 := constraint.NewConstraint(11, constraint.Distance, p5, l4, 0, false)
	g2.AddConstraint(c12)
	c13 := constraint.NewConstraint(12, constraint.Distance, p5, l5, 0, false)
	g2.AddConstraint(c13)
	c14 := constraint.NewConstraint(13, constraint.Distance, p4, l3, 0, false)
	g2.AddConstraint(c14)
	c15 := constraint.NewConstraint(14, constraint.Distance, p4, l4, 0, false)
	g2.AddConstraint(c15)
	c16 := constraint.NewConstraint(15, constraint.Angle, l5, l4, (108.0/180.0)*math.Pi, false)
	g2.AddConstraint(c16)
	c17 := constraint.NewConstraint(16, constraint.Angle, l3, l4, (108.0/180.0)*math.Pi, false)
	g2.AddConstraint(c17)

	state := g0.solveMerge(g1, g2)

	if state != solver.Solved {
		t.Error("Expected solved state(4), got", state)
	}

	t.Logf("g0 elements length %d\n", len(g0.elements))
	t.Logf("element %d type %v\n", 0, g0.elements[0].GetType())
	l1 = g0.elements[0].AsLine()
	t.Logf("element %d type %v\n", 1, g0.elements[1].GetType())
	p1 = g0.elements[1].AsPoint()
	t.Logf("element %d type %v\n", 2, g0.elements[2].GetType())
	p2 = g0.elements[2].AsPoint()
	t.Logf("element %d type %v\n", 3, g0.elements[3].GetType())
	l2 = g0.elements[3].AsLine()
	t.Logf("element %d type %v\n", 4, g0.elements[4].GetType())
	l3 = g0.elements[4].AsLine()
	t.Logf("element %d type %v\n", 5, g0.elements[5].GetType())
	p3 = g0.elements[5].AsPoint()
	t.Logf("element %d type %v\n", 6, g0.elements[6].GetType())
	l4 = g0.elements[6].AsLine()
	t.Logf("element %d type %v\n", 7, g0.elements[7].GetType())
	l5 = g0.elements[7].AsLine()
	t.Logf("element %d type %v\n", 8, g0.elements[8].GetType())
	p4 = g0.elements[8].AsPoint()
	t.Logf("element %d type %v\n", 9, g0.elements[9].GetType())
	p5 = g0.elements[9].AsPoint()

	rad2Deg := func(rad float64) float64 { return (rad / math.Pi) * 180 }
	deg2Rad := func(deg float64) float64 { return (deg / 180) * math.Pi }
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
		var solved constraint.ConstraintList = g.solvedConstraintsFor(tt.eId)
		var unsolved constraint.ConstraintList = g.unsolvedConstraintsFor(tt.eId)
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

func TestLocalSolveEdgeCases(t *testing.T) {
	e3 := el.NewSketchLine(2, 2, 1, -1)
	e4 := el.NewSketchLine(3, 2, 2, -0)
	c3 := constraint.NewConstraint(2, constraint.Distance, e3, e4, 2, false)

	o := NewGraphCluster(1)
	o.AddConstraint(c3)

	state := o.localSolve()
	assert.Equal(t, solver.NonConvergent, state, "Test local solve with solveorder < 2")

	e1 := el.NewSketchPoint(0, 0, 1)
	e2 := el.NewSketchPoint(1, 2, 1)
	c1 := constraint.NewConstraint(0, constraint.Distance, e1, e2, 5, false)
	o.AddConstraint(c1)
	o.solveOrder = append(o.solveOrder, e1.GetID())
	o.solveOrder = append(o.solveOrder, e2.GetID())
	o.solveOrder = append(o.solveOrder, e3.GetID())

	state = o.localSolve()
	assert.Equal(t, solver.NonConvergent, state, "Test local solve without enough constraints to solve desired element")

	c2 := constraint.NewConstraint(1, constraint.Distance, e2, e3, 7, false)
	g := NewGraphCluster(0)
	g.solveOrder = append(g.solveOrder, e1.GetID())
	g.solveOrder = append(g.solveOrder, e2.GetID())
	g.solveOrder = append(g.solveOrder, e3.GetID())
	g.AddConstraint(c1)
	g.AddConstraint(c2)
	g.AddConstraint(c3)
	g.solved.Add(c2.GetID())
	g.solved.Add(c3.GetID())

	state = g.Solve()
	assert.Equal(t, solver.Solved, state, "Test local solve with pre-solved elements")
}

func TestMergeOne(t *testing.T) {
	// Create fixed element cluster
	// Create cluster w/ square
	// merge the two -- use solveMerge instead of mergeOne
	g := NewGraphCluster(0)
	g.AddElement(el.NewSketchPoint(0, 0, 0))
	g.AddElement(el.NewSketchLine(1, 0, 1, 0))
	g.AddElement(el.NewSketchLine(2, 1, 0, 0))

	o := NewGraphCluster(1)
	o.AddElement(el.NewSketchLine(1, -0.029929, -0.999552, 0))

	state := g.solveMerge(o, nil)
	assert.Equal(t, solver.NonConvergent, state, "Merge containing only one shared element should fail to solve")

	o.AddElement(el.NewSketchLine(2, 0.999552, -0.029929, 0))

	state = g.solveMerge(o, nil)
	assert.Equal(t, solver.NonConvergent, state, "Merge where shared elements are both lines should fail to solve")

	o = NewGraphCluster(1)
	o.AddElement(el.NewSketchPoint(0, 0, 0))
	o.AddElement(el.NewSketchPoint(5, 3.998208, -0.119717))
	o.AddElement(el.NewSketchLine(12, -0.563309, 0.826247, -3.804226))
	o.AddElement(el.NewSketchPoint(11, 2.183330, 6.092751))
	o.AddElement(el.NewSketchLine(9, 0.611735, 0.791063, -6.155367))
	o.AddElement(el.NewSketchPoint(8, 5.347580, 3.645810))
	o.AddElement(el.NewSketchLine(1, -0.029929, -0.999552, 0))
	o.AddElement(el.NewSketchLine(2, 0.999552, -0.029929, 0))
	o.AddElement(el.NewSketchLine(15, -0.959879, -0.280414, 0))
	o.AddElement(el.NewSketchPoint(14, -1.121656, 3.839516))
	o.AddElement(el.NewSketchLine(6, 0.941382, -0.337343, -3.804226))
	o.AddElement(el.NewSketchLine(3, -0.029929, -0.999552, 0))
	state = g.solveMerge(o, nil)
	assert.Equal(t, solver.Solved, state, "Merge should solve successfully")

	g = NewGraphCluster(0)
	g.AddElement(el.NewSketchPoint(0, 0, 0))
	g.AddElement(el.NewSketchLine(100, 0, 1, 0))
	g.AddElement(el.NewSketchPoint(5, 4, 0))

	state = g.solveMerge(o, nil)
	assert.Equal(t, solver.Solved, state, "Merge with two shared points should solve successfully")
}

func TestSolveMergeEdgeCases(t *testing.T) {
	g := NewGraphCluster(0)
	g.AddElement(el.NewSketchPoint(0, 0, 0))
	g.AddElement(el.NewSketchLine(1, 0, 1, 0))
	g.AddElement(el.NewSketchLine(2, 1, 0, 0))

	o1 := NewGraphCluster(1)
	o1.AddElement(el.NewSketchPoint(0, 0, 0))
	o1.AddElement(el.NewSketchPoint(5, 3.998208, -0.119717))
	o1.AddElement(el.NewSketchLine(12, -0.563309, 0.826247, -3.804226))
	o1.AddElement(el.NewSketchPoint(11, 2.183330, 6.092751))
	o1.AddElement(el.NewSketchLine(9, 0.611735, 0.791063, -6.155367))

	o2 := NewGraphCluster(2)
	o2.AddElement(el.NewSketchLine(15, -0.959879, -0.280414, 0))
	o2.AddElement(el.NewSketchPoint(5, -1.121656, 3.839516))
	o2.AddElement(el.NewSketchLine(6, 0.941382, -0.337343, -3.804226))
	o2.AddElement(el.NewSketchLine(3, -0.029929, -0.999552, 0))

	state := g.solveMerge(o1, o2)
	assert.Equal(t, solver.NonConvergent, state, "Three cluster solve with only two shared elements should fail")

	// Solve merge with three lines lines
	g = NewGraphCluster(0)
	g.AddElement(el.NewSketchPoint(0, 0, 0))
	g.AddElement(el.NewSketchLine(9, 0.611735, 0.791063, -6.155367))
	g.AddElement(el.NewSketchLine(1, 0, 1, 0))

	o1 = NewGraphCluster(1)
	o1.AddElement(el.NewSketchPoint(7, 0, 1))
	o1.AddElement(el.NewSketchPoint(5, 3.998208, -0.119717))
	o1.AddElement(el.NewSketchPoint(11, 2.183330, 6.092751))
	o1.AddElement(el.NewSketchLine(9, 0.611735, 0.791063, -6.155367))
	o1.AddElement(el.NewSketchLine(2, 1, 0, 0))

	o2 = NewGraphCluster(2)
	o2.AddElement(el.NewSketchLine(15, -0.959879, -0.280414, 0))
	o2.AddElement(el.NewSketchPoint(8, 3.998208, -0.119717))
	o2.AddElement(el.NewSketchLine(1, 0, 1, 0))
	o2.AddElement(el.NewSketchLine(2, 1, 0.0, 0.0))

	state = g.solveMerge(o1, o2)
	assert.Equal(t, solver.Solved, state, "Three cluster solve with three lines")

	// Solve merge with one point and two lines where lines are in clusters 0 and 1
	g = NewGraphCluster(0)
	g.AddElement(el.NewSketchPoint(0, 0, 0))
	g.AddElement(el.NewSketchLine(9, 0.611735, 0.791063, -6.155367))
	g.AddElement(el.NewSketchLine(1, 0, 1, 0))

	o1 = NewGraphCluster(1)
	o1.AddElement(el.NewSketchPoint(0, 0, 0))
	o1.AddElement(el.NewSketchPoint(5, 3.998208, -0.119717))
	o1.AddElement(el.NewSketchPoint(11, 2.183330, 6.092751))
	o1.AddElement(el.NewSketchLine(6, 0.611735, 0.791063, -6.155367))
	o1.AddElement(el.NewSketchLine(2, 1, 0, 0))

	o2 = NewGraphCluster(2)
	o2.AddElement(el.NewSketchLine(15, -0.959879, -0.280414, 0))
	o2.AddElement(el.NewSketchPoint(8, 3.998208, -0.119717))
	o2.AddElement(el.NewSketchLine(1, 0, 1, 0))
	o2.AddElement(el.NewSketchLine(2, 1, 0.0, 0.0))

	state = g.solveMerge(o1, o2)
	assert.Equal(t, solver.Solved, state, "Three cluster solve with three shared elements should solve")

	g.elements[0] = el.NewSketchLine(0, 1, 1, 1)
	o1.elements[0] = el.NewSketchLine(0, 3, 2, 1)
	line := o2.elements[2]
	line.AsLine().SetB(0.235)
	line.AsLine().SetC(2)
	state = g.solveMerge(o1, o2)
	assert.Equal(t, solver.NonConvergent, state, "Three cluster solve with non-convergent lines")

	o1.elements[2] = el.NewSketchLine(2, 1, 0, 0)
	o2.elements[2] = el.NewSketchLine(2, 1, 0, 0)
	g.elements[1] = el.NewSketchLine(1, 1, 0, 6)
	o2.elements[1] = el.NewSketchLine(1, 1, 0, 6)
	g.elements[0] = el.NewSketchPoint(0, -10, 1)
	o1.elements[0] = el.NewSketchPoint(0, 3, 2)
	state = g.solveMerge(o1, o2)
	assert.Equal(t, solver.NonConvergent, state, "Fail to solve final element")

	g = NewGraphCluster(0)
	g.AddElement(el.NewSketchPoint(0, 0, 0))
	g.AddElement(el.NewSketchLine(9, 0.611735, 0.791063, -6.155367))
	g.AddElement(el.NewSketchLine(1, 0, 1, 0))

	o1 = NewGraphCluster(1)
	o1.AddElement(el.NewSketchPoint(0, 0, 0))
	o1.AddElement(el.NewSketchPoint(5, 3.998208, -0.119717))
	o1.AddElement(el.NewSketchPoint(11, 2.183330, 6.092751))
	o1.AddElement(el.NewSketchLine(6, 0.611735, 0.791063, -6.155367))
	o1.AddElement(el.NewSketchLine(2, 1, 0, 0))

	o2 = NewGraphCluster(2)
	o2.AddElement(el.NewSketchLine(15, -0.959879, -0.280414, 0))
	o2.AddElement(el.NewSketchPoint(8, 3.998208, -0.119717))
	o2.AddElement(el.NewSketchLine(1, 0, 1, 0))
	o2.AddElement(el.NewSketchLine(2, 1, 0.0, 0.0))
	o2.AddElement(el.NewSketchPoint(0, 0, 0))

	state = g.solveMerge(o1, o2)
	assert.Equal(t, solver.NonConvergent, state, "Nonconvergent where an element has too many parents")
}

func TestToGraphViz(t *testing.T) {
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
	g.AddConstraint(c3)

	gvString := g.ToGraphViz()
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
