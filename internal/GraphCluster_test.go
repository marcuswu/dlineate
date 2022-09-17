package core

import (
	"math"
	"testing"

	"github.com/marcuswu/dlineation/internal/constraint"
	el "github.com/marcuswu/dlineation/internal/element"
	"github.com/marcuswu/dlineation/internal/solver"
	"github.com/marcuswu/dlineation/utils"
)

func TestAddConstraint(t *testing.T) {
	e1 := el.NewSketchPoint(0, 0, 1)
	e2 := el.NewSketchPoint(1, 2, 1)
	e3 := el.NewSketchLine(2, 2, 1, -1)
	c1 := constraint.NewConstraint(0, constraint.Distance, e1, e2, 5, false)
	c2 := constraint.NewConstraint(0, constraint.Distance, e2, e3, 7, false)

	g := NewGraphCluster()
	g.AddConstraint(c1)

	if len(g.constraints) != 1 {
		t.Error("expected graph cluster to have one constraint, found", len(g.constraints))
	}
	if len(g.elements) != 2 {
		t.Error("expected graph cluster to have 2 elements, found", len(g.elements))
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
	c2 := constraint.NewConstraint(0, constraint.Distance, e2, e3, 7, false)
	c3 := constraint.NewConstraint(0, constraint.Distance, e3, e4, 2, false)

	g := NewGraphCluster()
	g.AddConstraint(c1)
	g.AddConstraint(c2)

	o := NewGraphCluster()
	o.AddConstraint(c3)
	g.others = append(g.others, o)

	if !g.HasElementID(0) {
		t.Error("expected graph cluster to have element 0, but element was not found")
	}
	if !g.HasElementID(1) {
		t.Error("expected graph cluster to have element 1, but element was not found")
	}
	if !g.HasElementID(2) {
		t.Error("expected graph cluster to have element 2, but element was not found")
	}
	if !g.HasElementID(3) {
		t.Error("expected graph cluster to have element 3, but element was not found")
	}
	if g.HasElementID(4) {
		t.Error("expected graph cluster to not have element 4, but element was found")
	}
}

func TestGetElement(t *testing.T) {
	e1 := el.NewSketchPoint(0, 0, 1)
	e2 := el.NewSketchPoint(1, 2, 1)
	e3 := el.NewSketchLine(2, 2, 1, -1)
	e4 := el.NewSketchLine(3, 2, 2, -0)
	c1 := constraint.NewConstraint(0, constraint.Distance, e1, e2, 5, false)
	c2 := constraint.NewConstraint(0, constraint.Distance, e2, e3, 7, false)
	c3 := constraint.NewConstraint(0, constraint.Distance, e3, e4, 2, false)

	g := NewGraphCluster()
	g.AddConstraint(c1)
	g.AddConstraint(c2)

	o := NewGraphCluster()
	o.AddConstraint(c3)
	g.others = append(g.others, o)

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

	g := NewGraphCluster() // 0, 1, 2
	g.AddConstraint(c1)
	g.AddConstraint(c2)

	o := NewGraphCluster() // 2, 3
	o.AddConstraint(c3)
	g.others = append(g.others, o)

	g2 := NewGraphCluster()
	g3 := NewGraphCluster()
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
	c2 := constraint.NewConstraint(0, constraint.Distance, e2, e3, 7, false)

	g := NewGraphCluster()
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
	c2 := constraint.NewConstraint(0, constraint.Distance, e2, e3, 7, false)

	g := NewGraphCluster()
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
	g := NewGraphCluster()
	/*
		GraphCluster 0 (from test)
			l1: 0.000000x + 1.000000y + 0.000000 = 0
			p1: (0.000000, 0.000000)
			p2: (4.000000, 0.000000)
	*/

	l1 := el.NewSketchLine(0, 0, 1, 0)
	p1 := el.NewSketchPoint(1, 0, 0)
	p2 := el.NewSketchPoint(2, 3.13, 0)
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
	g := NewGraphCluster()

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
	g := NewGraphCluster()

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
	l4 := el.NewSketchLine(6, 1.16, -3.32, 13)
	l5 := el.NewSketchLine(7, 3.56, 1.04, 0)
	p1 := el.NewSketchPoint(1, 0, 0)
	p4 := el.NewSketchPoint(8, 2.28, 4.72)
	p5 := el.NewSketchPoint(9, -1.04, 3.56)
	c1 := constraint.NewConstraint(0, constraint.Distance, p1, p5, 4, false)
	g.AddConstraint(c1)
	c2 := constraint.NewConstraint(1, constraint.Distance, p4, p5, 4, false)
	g.AddConstraint(c2)
	c3 := constraint.NewConstraint(2, constraint.Distance, p1, l5, 0, false)
	g.AddConstraint(c3)
	c4 := constraint.NewConstraint(3, constraint.Distance, p5, l4, 0, false)
	g.AddConstraint(c4)
	c5 := constraint.NewConstraint(4, constraint.Distance, p5, l5, 0, false)
	g.AddConstraint(c5)
	c6 := constraint.NewConstraint(5, constraint.Distance, p4, l3, 0, false)
	g.AddConstraint(c6)
	c7 := constraint.NewConstraint(6, constraint.Distance, p4, l4, 0, false)
	g.AddConstraint(c7)
	c8 := constraint.NewConstraint(7, constraint.Angle, l3, l4, -(108.0/180.0)*math.Pi, false)
	g.AddConstraint(c8)
	c9 := constraint.NewConstraint(8, constraint.Angle, l4, l5, -(108.0/180.0)*math.Pi, false)
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

	cValue = c2.Element1.DistanceTo(c2.Element2)
	if utils.StandardFloatCompare(cValue, c2.Value) != 0 {
		t.Error("Expected point p4 to distance", c2.Value, "from point p5, distance is", cValue)
	}

	cValue = c3.Element1.DistanceTo(c3.Element2)
	if utils.StandardFloatCompare(cValue, c3.Value) != 0 {
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

	cValue = c6.Element1.DistanceTo(c6.Element2)
	if utils.StandardFloatCompare(cValue, c6.Value) != 0 {
		t.Error("Expected point p4 to be on line l3, distance is", cValue)
	}

	cValue = c7.Element1.DistanceTo(c7.Element2)
	if utils.StandardFloatCompare(cValue, c7.Value) != 0 {
		t.Error("Expected point p4 to be on line l4, distance is", cValue)
	}

	angle := c8.Element1.AsLine().AngleToLine(c8.Element2.AsLine())
	if utils.StandardFloatCompare(angle, c8.Value) != 0 {
		t.Error("Expected line l5 to be", c8.Value, "radians from line l4, angle is", angle)
	}

	angle = c9.Element1.AsLine().AngleToLine(c9.Element2.AsLine())
	if utils.StandardFloatCompare(angle, c9.Value) != 0 {
		t.Error("Expected line l3 to be", c9.Value, "radians from line l4, angle is", angle)
	}
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
	g0 := NewGraphCluster()

	l1 := el.NewSketchLine(0, 0, 1, 0)
	p1 := el.NewSketchPoint(1, 0.0, 0.0)
	p2 := el.NewSketchPoint(2, 4.0, 0)
	c1 := constraint.NewConstraint(0, constraint.Distance, p1, p2, 4, false)
	g0.AddConstraint(c1)
	c2 := constraint.NewConstraint(1, constraint.Distance, p1, l1, 0, false)
	g0.AddConstraint(c2)
	c3 := constraint.NewConstraint(2, constraint.Distance, p2, l1, 0, false)
	g0.AddConstraint(c3)

	g1 := NewGraphCluster()

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
	c8 := constraint.NewConstraint(7, constraint.Angle, l3, l2, -(108.0/180.0)*math.Pi, false)
	g1.AddConstraint(c8)

	g2 := NewGraphCluster()

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
	c16 := constraint.NewConstraint(15, constraint.Angle, l5, l4, -(108.0/180.0)*math.Pi, false)
	g2.AddConstraint(c16)
	c17 := constraint.NewConstraint(16, constraint.Angle, l3, l4, -(108.0/180.0)*math.Pi, false)
	g2.AddConstraint(c17)

	g0.others = append(g0.others, g1)
	g0.others = append(g0.others, g2)

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

func TestSolve(t *testing.T) {

}
