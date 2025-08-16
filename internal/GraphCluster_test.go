package core

import (
	"fmt"
	"math"
	"math/big"
	"testing"

	"github.com/marcuswu/dlineate/internal/accessors"
	"github.com/marcuswu/dlineate/internal/constraint"
	el "github.com/marcuswu/dlineate/internal/element"
	"github.com/marcuswu/dlineate/internal/solver"
	"github.com/marcuswu/dlineate/utils"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestAddElement(t *testing.T) {
	e1 := el.NewSketchPoint(0, big.NewFloat(0), big.NewFloat(1))
	e2 := el.NewSketchPoint(1, big.NewFloat(2), big.NewFloat(1))

	g := NewGraphCluster(1)
	g.AddElement(e1.GetID())

	if g.elements.Count() != 1 {
		t.Error("expected one element to be added to the cluster, found", g.elements.Count())
	}

	g.AddElement(e1.GetID())

	if g.elements.Count() != 1 {
		t.Error("expected no change to the cluster element length, found", g.elements.Count())
	}

	g.AddElement(e2.GetID())

	if g.elements.Count() != 2 {
		t.Error("expected two elements to be added to the cluster, found", g.elements.Count())
	}
}

func TestAddConstraint(t *testing.T) {
	e1 := el.NewSketchPoint(0, big.NewFloat(0), big.NewFloat(1))
	e2 := el.NewSketchPoint(1, big.NewFloat(2), big.NewFloat(1))
	e3 := el.NewSketchLine(2, big.NewFloat(2), big.NewFloat(1), big.NewFloat(-1))
	c1 := constraint.NewConstraint(0, constraint.Distance, e1.GetID(), e2.GetID(), big.NewFloat(5), false)
	c2 := constraint.NewConstraint(1, constraint.Distance, e2.GetID(), e3.GetID(), big.NewFloat(7), false)

	g := NewGraphCluster(0)
	g.AddConstraint(c1)

	if g.GetID() != 0 {
		t.Error("expected cluster id to be 0")
	}

	if len(g.constraints) != 1 {
		t.Error("expected graph cluster to have one constraint, found", len(g.constraints))
	}
	if g.elements.Count() != 2 {
		t.Error("expected graph cluster to have 2 elements, found", g.elements.Count())
	}

	c1.Solved = true
	g.AddConstraint(c1)
	if len(g.constraints) != 1 {
		t.Error("expected no change to cluster constraints after adding the same constraint twice")
	}
	if g.elements.Count() != 2 {
		t.Error("expected no change to elements after adding the same constraint twice")
	}

	g.AddConstraint(c2)

	if len(g.constraints) != 2 {
		t.Error("expected graph cluster to have two constraint, found", len(g.constraints))
	}
	if g.elements.Count() != 3 {
		t.Error("expected graph cluster to have 3 elements, found", g.elements.Count())
	}
}

func TestHasElementID(t *testing.T) {
	e1 := el.NewSketchPoint(0, big.NewFloat(0), big.NewFloat(1))
	e2 := el.NewSketchPoint(1, big.NewFloat(2), big.NewFloat(1))
	e3 := el.NewSketchLine(2, big.NewFloat(2), big.NewFloat(1), big.NewFloat(-1))
	e4 := el.NewSketchLine(3, big.NewFloat(2), big.NewFloat(2), big.NewFloat(-0))
	c1 := constraint.NewConstraint(0, constraint.Distance, e1.GetID(), e2.GetID(), big.NewFloat(5), false)
	c2 := constraint.NewConstraint(1, constraint.Distance, e2.GetID(), e3.GetID(), big.NewFloat(7), false)
	c3 := constraint.NewConstraint(2, constraint.Distance, e3.GetID(), e4.GetID(), big.NewFloat(2), false)

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
	e1 := el.NewSketchPoint(0, big.NewFloat(0), big.NewFloat(1))
	e2 := el.NewSketchPoint(1, big.NewFloat(2), big.NewFloat(1))
	e3 := el.NewSketchLine(2, big.NewFloat(2), big.NewFloat(1), big.NewFloat(-1))
	e4 := el.NewSketchLine(3, big.NewFloat(2), big.NewFloat(2), big.NewFloat(-0))
	c1 := constraint.NewConstraint(0, constraint.Distance, e1.GetID(), e2.GetID(), big.NewFloat(5), false)
	c2 := constraint.NewConstraint(1, constraint.Distance, e2.GetID(), e3.GetID(), big.NewFloat(7), false)
	c3 := constraint.NewConstraint(2, constraint.Distance, e3.GetID(), e4.GetID(), big.NewFloat(2), false)

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
	falseElement := el.NewSketchPoint(100, big.NewFloat(0), big.NewFloat(0))
	assert.False(t, g.HasElement(falseElement.GetID()))
	assert.False(t, o.HasElement(falseElement.GetID()))
}

func TestSharedElements(t *testing.T) {
	e1 := el.NewSketchPoint(0, big.NewFloat(0), big.NewFloat(1))
	e2 := el.NewSketchPoint(1, big.NewFloat(2), big.NewFloat(1))
	e3 := el.NewSketchLine(2, big.NewFloat(2), big.NewFloat(1), big.NewFloat(-1))
	e4 := el.NewSketchLine(3, big.NewFloat(2), big.NewFloat(2), big.NewFloat(-0))
	c1 := constraint.NewConstraint(0, constraint.Distance, e1.GetID(), e2.GetID(), big.NewFloat(5), false)
	c2 := constraint.NewConstraint(1, constraint.Distance, e2.GetID(), e3.GetID(), big.NewFloat(7), false)
	c3 := constraint.NewConstraint(2, constraint.Distance, e3.GetID(), e4.GetID(), big.NewFloat(2), false)

	g := NewGraphCluster(0) // 0, 1, 2
	g.AddConstraint(c1)
	g.AddConstraint(c2)

	o := NewGraphCluster(1) // 2, 3
	o.AddConstraint(c3)

	g2 := NewGraphCluster(2)
	g3 := NewGraphCluster(3)
	e5 := el.NewSketchPoint(4, big.NewFloat(0), big.NewFloat(1))
	e6 := el.NewSketchPoint(5, big.NewFloat(1), big.NewFloat(2))
	c4 := constraint.NewConstraint(3, constraint.Distance, e4.GetID(), e5.GetID(), big.NewFloat(12), false)
	c5 := constraint.NewConstraint(3, constraint.Distance, e5.GetID(), e6.GetID(), big.NewFloat(12), false)
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
	ea := accessors.NewElementRepository()
	ca := accessors.NewConstraintRepository()
	e1 := el.NewSketchPoint(0, big.NewFloat(0), big.NewFloat(1))
	e2 := el.NewSketchPoint(1, big.NewFloat(2), big.NewFloat(1))
	e3 := el.NewSketchLine(2, big.NewFloat(2), big.NewFloat(1), big.NewFloat(-1))
	ea.AddElement(e1)
	ea.AddElement(e2)
	ea.AddElement(e3)
	ea.AddElementToCluster(e1.GetID(), 0)
	ea.AddElementToCluster(e2.GetID(), 0)
	ea.AddElementToCluster(e3.GetID(), 0)
	originalPointNearest := e3.PointNearestOrigin()
	c1 := constraint.NewConstraint(0, constraint.Distance, e1.GetID(), e2.GetID(), big.NewFloat(5), false)
	c2 := constraint.NewConstraint(1, constraint.Distance, e2.GetID(), e3.GetID(), big.NewFloat(7), false)
	ca.AddConstraint(c1)
	ca.AddConstraint(c2)

	g := NewGraphCluster(0)
	g.AddConstraint(c1)
	g.AddConstraint(c2)

	g.TranslateCluster(ea, big.NewFloat(1), big.NewFloat(1))
	ge1, _ := ea.GetElement(g.GetID(), 0)
	e1 = ge1.AsPoint()
	ge2, _ := ea.GetElement(g.GetID(), 1)
	e2 = ge2.AsPoint()
	ge3, _ := ea.GetElement(g.GetID(), 2)
	e3 = ge3.AsLine()

	if e1.GetX().Cmp(big.NewFloat(1)) != 0 && e1.GetY().Cmp(big.NewFloat(2)) != 0 {
		t.Error("Expected the e1 to be 1, 2, got", e1.GetX(), ",", e1.GetY())
	}
	if e2.GetX().Cmp(big.NewFloat(3)) != 0 && e2.GetY().Cmp(big.NewFloat(2)) != 0 {
		t.Error("Expected the e1 to be 3, 2, got", e2)
	}
	var x, y, temp big.Float
	x.Add(originalPointNearest.GetX(), big.NewFloat(1))
	y.Add(originalPointNearest.GetX(), big.NewFloat(1))
	y.Mul(&y, e3.GetA())
	y.Add(&y, e3.GetC())
	temp.Neg(e3.GetB())
	y.Quo(&y, &temp)
	e3Point := el.NewSketchPoint(0, &x, &y)
	if utils.StandardBigFloatCompare(e3.DistanceTo(e3Point), big.NewFloat(0)) != 0 {
		t.Error("Expected e3Point to be on e3. Distance is", e3.DistanceTo(e3Point))
	}
	x.Add(originalPointNearest.GetX(), big.NewFloat(1))
	y.Add(originalPointNearest.GetY(), big.NewFloat(1))
	if utils.StandardBigFloatCompare(e3Point.GetX(), &x) != 0 {
		t.Error("Expected the X difference between e3 and its original point nearest origin to be 1. Original X", originalPointNearest.GetX(), ", new X", e3Point.GetY())
	}
	if utils.StandardBigFloatCompare(e3Point.GetY(), &y) != 0 {
		t.Error("Expected the Y difference between e3 and its original point nearest origin to be 1. Original Y", originalPointNearest.GetY(), ", new Y", e3Point.GetY())
	}
}

func TestRotate(t *testing.T) {
	ea := accessors.NewElementRepository()
	ca := accessors.NewConstraintRepository()
	e1 := el.NewSketchPoint(0, big.NewFloat(0), big.NewFloat(1))
	e2 := el.NewSketchPoint(1, big.NewFloat(2), big.NewFloat(1))
	e3 := el.NewSketchLine(2, big.NewFloat(2), big.NewFloat(1), big.NewFloat(-1))
	o := el.NewSketchPoint(3, big.NewFloat(0), big.NewFloat(0))
	ea.AddElement(e1)
	ea.AddElement(e2)
	ea.AddElement(e3)
	ea.AddElement(o)
	ea.AddElementToCluster(e1.GetID(), 0)
	ea.AddElementToCluster(e2.GetID(), 0)
	ea.AddElementToCluster(e3.GetID(), 0)
	ea.AddElementToCluster(o.GetID(), 0)
	c1 := constraint.NewConstraint(0, constraint.Distance, e1.GetID(), e2.GetID(), big.NewFloat(5), false)
	c2 := constraint.NewConstraint(1, constraint.Distance, e2.GetID(), e3.GetID(), big.NewFloat(7), false)
	ca.AddConstraint(c1)
	ca.AddConstraint(c2)

	g := NewGraphCluster(0)
	g.AddConstraint(c1)
	g.AddConstraint(c2)

	g.RotateCluster(ea, o, big.NewFloat(math.Pi/2.0))
	ge1, _ := ea.GetElement(g.GetID(), 0)
	e1 = ge1.AsPoint()
	ge2, _ := ea.GetElement(g.GetID(), 1)
	e2 = ge2.AsPoint()
	ge3, _ := ea.GetElement(g.GetID(), 2)
	e3 = ge3.AsLine()

	mag := math.Sqrt((0.4 * 0.4) + (0.8 * 0.8))
	var A, B, C big.Float
	A.Quo(big.NewFloat(-0.4), big.NewFloat(mag))
	B.Quo(big.NewFloat(0.8), big.NewFloat(mag))
	C.Quo(big.NewFloat(-0.4), big.NewFloat(mag))
	if utils.StandardBigFloatCompare(e3.GetA(), &A) != 0 ||
		utils.StandardBigFloatCompare(e3.GetB(), &B) != 0 ||
		utils.StandardBigFloatCompare(e3.GetC(), &C) != 0 {
		t.Error("Expected e3 to be", A, ",", B, ",", C, ". Got", e3.GetA(), ",", e3.GetB(), ",", e3.GetC())
	}

	if utils.StandardBigFloatCompare(e1.GetX(), big.NewFloat(-1)) != 0 ||
		utils.StandardBigFloatCompare(e1.GetY(), big.NewFloat(0.0)) != 0 {
		t.Error("Expected -1, 0 got", e1.GetX(), ",", e1.GetY())
	}

	if utils.StandardBigFloatCompare(e2.GetX(), big.NewFloat(-1)) != 0 ||
		utils.StandardBigFloatCompare(e2.GetY(), big.NewFloat(2.0)) != 0 {
		t.Error("Expected -1, 2 got", e2.GetX(), ",", e2.GetY())
	}
}

func TestSolve0(t *testing.T) {
	ea := accessors.NewElementRepository()
	ca := accessors.NewConstraintRepository()
	g := NewGraphCluster(0)
	/*
		GraphCluster 0 (from test)
			l1: 0.000000x + 1.000000y + 0.000000 = 0
			p1: (0.000000, 0.000000)
			p2: (4.000000, 0.000000)
	*/

	l1 := el.NewSketchLine(0, big.NewFloat(0), big.NewFloat(1), big.NewFloat(0))
	p1 := el.NewSketchPoint(1, big.NewFloat(0), big.NewFloat(0))
	p2 := el.NewSketchPoint(2, big.NewFloat(3.13), big.NewFloat(0))
	ea.AddElement(l1)
	ea.AddElement(p1)
	ea.AddElement(p2)
	g.AddElement(l1.GetID())
	g.AddElement(p1.GetID())
	g.AddElement(p2.GetID())
	c1 := constraint.NewConstraint(0, constraint.Distance, p2.GetID(), p1.GetID(), big.NewFloat(4), false)
	g.AddConstraint(c1)
	ca.AddConstraint(c1)
	c2 := constraint.NewConstraint(1, constraint.Distance, p1.GetID(), l1.GetID(), big.NewFloat(0), false)
	g.AddConstraint(c2)
	ca.AddConstraint(c2)
	c3 := constraint.NewConstraint(2, constraint.Distance, p2.GetID(), l1.GetID(), big.NewFloat(0), false)
	g.AddConstraint(c3)
	ca.AddConstraint(c3)

	state := g.Solve(ea, ca)

	c1, _ = ca.GetConstraint(0)
	c2, _ = ca.GetConstraint(1)
	c3, _ = ca.GetConstraint(2)

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

	e1, _ := ea.GetElement(-1, c1.Element1)
	e2, _ := ea.GetElement(-1, c1.Element2)
	cValue := e1.DistanceTo(e2)
	if utils.StandardBigFloatCompare(cValue, &c1.Value) != 0 {
		t.Error("Expected point p1 to be distance", c1.Value, "from point p2, distance is", cValue)
	}

	e1, _ = ea.GetElement(-1, c2.Element1)
	e2, _ = ea.GetElement(-1, c2.Element2)
	cValue = e1.DistanceTo(e2)
	if utils.StandardBigFloatCompare(cValue, &c2.Value) != 0 {
		t.Error("Expected point p1 to be on line l1, distance is", cValue)
	}

	e1, _ = ea.GetElement(-1, c3.Element1)
	e2, _ = ea.GetElement(-1, c3.Element2)
	cValue = e1.DistanceTo(e2)
	if utils.StandardBigFloatCompare(cValue, &c3.Value) != 0 {
		t.Error("Expected point p2 to be on line l1, distance is", cValue)
	}
}

func TestSolve1(t *testing.T) {
	ea := accessors.NewElementRepository()
	ca := accessors.NewConstraintRepository()
	g := NewGraphCluster(0)

	/*
		GraphCluster 1 (from test)
			l2: -0.748682x + 0.662930y + 2.341692 = 0
			l3: 0.861839x + 0.507182y + -5.071811 = 0
			p2: (2.132349, -1.124164)
			p3: (4.784067, 1.870563)
	*/

	l2 := el.NewSketchLine(3, big.NewFloat(-2.27), big.NewFloat(2.01), big.NewFloat(7.1))
	l3 := el.NewSketchLine(4, big.NewFloat(2.45), big.NewFloat(2.86), big.NewFloat(-19.1))
	p2 := el.NewSketchPoint(2, big.NewFloat(3.13), big.NewFloat(0))
	p3 := el.NewSketchPoint(5, big.NewFloat(5.14), big.NewFloat(2.27))
	ea.AddElement(p2)
	ea.AddElement(l2)
	ea.AddElement(p3)
	ea.AddElement(l3)
	g.AddElement(p2.GetID())
	g.AddElement(l2.GetID())
	g.AddElement(p3.GetID())
	g.AddElement(l3.GetID())
	c1 := constraint.NewConstraint(0, constraint.Distance, p2.GetID(), p3.GetID(), big.NewFloat(4), false)
	ca.AddConstraint(c1)
	g.AddConstraint(c1)
	c2 := constraint.NewConstraint(1, constraint.Distance, p2.GetID(), l2.GetID(), big.NewFloat(0), false)
	ca.AddConstraint(c2)
	g.AddConstraint(c2)
	c3 := constraint.NewConstraint(2, constraint.Distance, p3.GetID(), l2.GetID(), big.NewFloat(0), false)
	ca.AddConstraint(c3)
	g.AddConstraint(c3)
	c4 := constraint.NewConstraint(3, constraint.Distance, p3.GetID(), l3.GetID(), big.NewFloat(0), false)
	ca.AddConstraint(c4)
	g.AddConstraint(c4)
	c5 := constraint.NewConstraint(4, constraint.Angle, l2.GetID(), l3.GetID(), big.NewFloat(-(108.0/180.0)*math.Pi), false)
	ca.AddConstraint(c5)
	g.AddConstraint(c5)

	// Solves:
	// 0. l1 to l2 angle first
	// 1. Then p2 to l1 and l2
	// 2. Finally p1 to p2 and l1
	state := g.Solve(ea, ca)

	c1, _ = ca.GetConstraint(0)
	c2, _ = ca.GetConstraint(1)
	c3, _ = ca.GetConstraint(2)
	c4, _ = ca.GetConstraint(3)
	c5, _ = ca.GetConstraint(4)

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

	e1, _ := ea.GetElement(-1, c1.Element1)
	e2, _ := ea.GetElement(-1, c1.Element2)
	cValue := e1.DistanceTo(e2)
	if utils.StandardBigFloatCompare(cValue, &c1.Value) != 0 {
		t.Error("Expected point p1 to distance", c1.Value, "from point p2, distance is", cValue)
	}

	e1, _ = ea.GetElement(-1, c2.Element1)
	e2, _ = ea.GetElement(-1, c2.Element2)
	cValue = e1.DistanceTo(e2)
	if utils.StandardBigFloatCompare(cValue, &c2.Value) != 0 {
		t.Error("Expected point p1 to be on line l1, distance is", cValue)
	}

	e1, _ = ea.GetElement(-1, c3.Element1)
	e2, _ = ea.GetElement(-1, c3.Element2)
	cValue = e1.DistanceTo(e2)
	if utils.StandardBigFloatCompare(cValue, &c3.Value) != 0 {
		t.Error("Expected point p2 to be on line l1, distance is", cValue)
	}

	e1, _ = ea.GetElement(-1, c4.Element1)
	e2, _ = ea.GetElement(-1, c4.Element2)
	cValue = e1.DistanceTo(e2)
	if utils.StandardBigFloatCompare(cValue, &c4.Value) != 0 {
		t.Error("Expected point p2 to be on line l2, distance is", cValue)
	}

	e1, _ = ea.GetElement(-1, c5.Element1)
	e2, _ = ea.GetElement(-1, c5.Element2)
	angle := e1.(*el.SketchLine).AngleToLine(e2.(*el.SketchLine))
	if utils.StandardBigFloatCompare(angle, &c5.Value) != 0 {
		t.Error("Expected line l2 to be", c5.Value, "radians from line l2, angle is", angle)
	}
}

func TestSolve2(t *testing.T) {
	ea := accessors.NewElementRepository()
	ca := accessors.NewConstraintRepository()
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
	l3 := el.NewSketchLine(4, big.NewFloat(2.45), big.NewFloat(2.86), big.NewFloat(-19.1))
	p4 := el.NewSketchPoint(8, big.NewFloat(2.28), big.NewFloat(4.72))
	l4 := el.NewSketchLine(6, big.NewFloat(1.16), big.NewFloat(-3.32), big.NewFloat(13))
	p5 := el.NewSketchPoint(9, big.NewFloat(-1.04), big.NewFloat(3.56))
	l5 := el.NewSketchLine(7, big.NewFloat(3.56), big.NewFloat(1.04), big.NewFloat(0))
	p1 := el.NewSketchPoint(1, big.NewFloat(0), big.NewFloat(0))
	ea.AddElement(l3)
	ea.AddElement(p4)
	ea.AddElement(l4)
	ea.AddElement(p5)
	ea.AddElement(l5)
	ea.AddElement(p1)
	g.AddElement(l3.GetID())
	g.AddElement(p4.GetID())
	g.AddElement(l4.GetID())
	g.AddElement(p5.GetID())
	g.AddElement(l5.GetID())
	g.AddElement(p1.GetID())
	ea.AddElementToCluster(l3.GetID(), g.GetID())
	ea.AddElementToCluster(p4.GetID(), g.GetID())
	ea.AddElementToCluster(l4.GetID(), g.GetID())
	ea.AddElementToCluster(p5.GetID(), g.GetID())
	ea.AddElementToCluster(l5.GetID(), g.GetID())
	ea.AddElementToCluster(p1.GetID(), g.GetID())
	c1 := constraint.NewConstraint(0, constraint.Distance, p4.GetID(), l3.GetID(), big.NewFloat(0), false)
	ca.AddConstraint(c1)
	g.AddConstraint(c1)
	c2 := constraint.NewConstraint(1, constraint.Angle, l3.GetID(), l4.GetID(), big.NewFloat(-(108.0/180.0)*math.Pi), false)
	ca.AddConstraint(c2)
	g.AddConstraint(c2)
	c3 := constraint.NewConstraint(2, constraint.Distance, p4.GetID(), l4.GetID(), big.NewFloat(0), false)
	ca.AddConstraint(c3)
	g.AddConstraint(c3)
	c4 := constraint.NewConstraint(3, constraint.Distance, p5.GetID(), l4.GetID(), big.NewFloat(0), false)
	ca.AddConstraint(c4)
	g.AddConstraint(c4)
	c5 := constraint.NewConstraint(4, constraint.Distance, p4.GetID(), p5.GetID(), big.NewFloat(4), false)
	ca.AddConstraint(c5)
	g.AddConstraint(c5)
	c6 := constraint.NewConstraint(5, constraint.Angle, l4.GetID(), l5.GetID(), big.NewFloat(-(108.0/180.0)*math.Pi), false)
	ca.AddConstraint(c6)
	g.AddConstraint(c6)
	c7 := constraint.NewConstraint(6, constraint.Distance, p5.GetID(), l5.GetID(), big.NewFloat(0), false)
	ca.AddConstraint(c7)
	g.AddConstraint(c7)
	c8 := constraint.NewConstraint(7, constraint.Distance, p1.GetID(), p5.GetID(), big.NewFloat(4), false)
	ca.AddConstraint(c8)
	g.AddConstraint(c8)
	c9 := constraint.NewConstraint(8, constraint.Distance, p1.GetID(), l5.GetID(), big.NewFloat(0), false)
	ca.AddConstraint(c9)
	g.AddConstraint(c9)

	state := g.Solve(ea, ca)

	c1, _ = ca.GetConstraint(0)
	c2, _ = ca.GetConstraint(1)
	c3, _ = ca.GetConstraint(2)
	c4, _ = ca.GetConstraint(3)
	c5, _ = ca.GetConstraint(4)
	c6, _ = ca.GetConstraint(5)
	c7, _ = ca.GetConstraint(6)
	c8, _ = ca.GetConstraint(7)
	c9, _ = ca.GetConstraint(8)

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

	e1, _ := ea.GetElement(-1, c1.Element1)
	e2, _ := ea.GetElement(-1, c1.Element2)
	cValue := e1.DistanceTo(e2)
	if utils.StandardBigFloatCompare(cValue, &c1.Value) != 0 {
		t.Error("Expected point p1 to distance", c1.Value, "from point p5, distance is", cValue)
	}

	e1, _ = ea.GetElement(-1, c7.Element1)
	e2, _ = ea.GetElement(-1, c7.Element2)
	cValue = e1.DistanceTo(e2)
	if utils.StandardBigFloatCompare(cValue, &c7.Value) != 0 {
		t.Error("Expected point p4 to distance", c7.Value, "from point p5, distance is", cValue)
	}

	e1, _ = ea.GetElement(-1, c9.Element1)
	e2, _ = ea.GetElement(-1, c9.Element2)
	cValue = e1.DistanceTo(e2)
	if utils.StandardBigFloatCompare(cValue, &c9.Value) != 0 {
		t.Error("Expected point p1 to be on line l5, distance is", cValue)
	}

	e1, _ = ea.GetElement(-1, c4.Element1)
	e2, _ = ea.GetElement(-1, c4.Element2)
	cValue = e1.DistanceTo(e2)
	if utils.StandardBigFloatCompare(cValue, &c4.Value) != 0 {
		t.Error("Expected point p5 to be on line l4, distance is", cValue)
	}

	e1, _ = ea.GetElement(-1, c5.Element1)
	e2, _ = ea.GetElement(-1, c5.Element2)
	cValue = e1.DistanceTo(e2)
	if utils.StandardBigFloatCompare(cValue, &c5.Value) != 0 {
		t.Error("Expected point p5 to be on line l5, distance is", cValue)
	}

	e1, _ = ea.GetElement(-1, c8.Element1)
	e2, _ = ea.GetElement(-1, c8.Element2)
	cValue = e1.DistanceTo(e2)
	if utils.StandardBigFloatCompare(cValue, &c8.Value) != 0 {
		t.Error("Expected point p4 to be on line l3, distance is", cValue)
	}

	e1, _ = ea.GetElement(-1, c3.Element1)
	e2, _ = ea.GetElement(-1, c3.Element2)
	cValue = e1.DistanceTo(e2)
	if utils.StandardBigFloatCompare(cValue, &c3.Value) != 0 {
		t.Error("Expected point p4 to be on line l4, distance is", cValue)
	}

	e1, _ = ea.GetElement(-1, c2.Element1)
	e2, _ = ea.GetElement(-1, c2.Element2)
	angle := e1.AsLine().AngleToLine(e2.AsLine())
	var v1, v2 big.Float
	v1.Abs(angle)
	v2.Abs(c2.GetValue())
	assert.Equal(t, 0, utils.StandardBigFloatCompare(&v1, &v2), "Expected line l5 angle to be correct")

	e1, _ = ea.GetElement(-1, c6.Element1)
	e2, _ = ea.GetElement(-1, c6.Element2)
	angle = e1.AsLine().AngleToLine(e2.AsLine())
	v1.Abs(angle)
	v2.Abs(c6.GetValue())
	assert.Equal(t, 0, utils.StandardBigFloatCompare(&v1, &v2), "Expected line l3 angle to be correct")
}

func TestSolveMerge(t *testing.T) {
	utils.Logger.Level(zerolog.DebugLevel)
	/*
		GraphCluster 0 (from test)
			p0: (0.000000, 0.000000)
			l1: 0.000000x + -1.000000y + 0.000000 = 0
			l2: 1.000000x + 0.000000y + 0.000000 = 0
			l3: 0.000000x + -1.000000y + 0.000000 = 0
			p5: (4.000000, 0.000000)

		GraphCluster 1 (from test)
			p0: (0.000000, 0.000000)
			l15: -3.839516x + -1.121656y + 0.000000 = 0
			l12: -0.563309x + 0.826247y + -3.804226 = 0
			p11: (2.183330, 6.092751)
			l9: 0.611735x + 0.791063y + -6.155367 = 0
			p14: (-1.121656, 3.839516)

		GraphCluster 2 (from test)
			p5: (2.488281, -0.724727)
			p8: (5.140000, 2.270000)
			l6: 2.994727x + -2.651719y + -9.373495 = 0
			l9: 0.861839x + 0.507182y + -5.581155 = 0

		Each cluster shares one element with another:
			GraphCluster 0 and 1 share p0
			GraphCluster 0 and 2 share p5
			GraphCluster 1 and 2 share l9

		solveMerge should merge the three clusters into a single solved graph
	*/
	ea := accessors.NewElementRepository()
	ca := accessors.NewConstraintRepository()
	g0 := NewGraphCluster(0)

	var prec uint = 200

	p0 := el.NewSketchPoint(0, new(big.Float).SetPrec(prec).SetFloat64(0.0), new(big.Float).SetPrec(prec).SetFloat64(0.0))
	ea.AddElement(p0)
	g0.AddElement(p0.GetID())
	ea.AddElementToCluster(p0.GetID(), g0.GetID())
	l1 := el.NewSketchLine(1, new(big.Float).SetPrec(prec).SetFloat64(0), new(big.Float).SetPrec(prec).SetFloat64(-1), new(big.Float).SetPrec(prec).SetFloat64(0))
	ea.AddElement(l1)
	g0.AddElement(l1.GetID())
	ea.AddElementToCluster(l1.GetID(), g0.GetID())
	l2 := el.NewSketchLine(2, new(big.Float).SetPrec(prec).SetFloat64(1), new(big.Float).SetPrec(prec).SetFloat64(0), new(big.Float).SetPrec(prec).SetFloat64(0))
	ea.AddElement(l2)
	g0.AddElement(l2.GetID())
	ea.AddElementToCluster(l2.GetID(), g0.GetID())
	l3 := el.NewSketchLine(3, new(big.Float).SetPrec(prec).SetFloat64(0), new(big.Float).SetPrec(prec).SetFloat64(-1), new(big.Float).SetPrec(prec).SetFloat64(0))
	ea.AddElement(l3)
	g0.AddElement(l3.GetID())
	ea.AddElementToCluster(l3.GetID(), g0.GetID())
	p5 := el.NewSketchPoint(5, new(big.Float).SetPrec(prec).SetFloat64(4.0), new(big.Float).SetPrec(prec).SetFloat64(0.0))
	ea.AddElement(p5)
	g0.AddElement(p5.GetID())
	ea.AddElementToCluster(p5.GetID(), g0.GetID())

	g1 := NewGraphCluster(1)

	ea.AddElementToCluster(p0.GetID(), g1.GetID())
	g1.AddElement(p0.GetID())
	l15 := el.NewSketchLine(15, new(big.Float).SetPrec(prec).SetFloat64(-3.839516468), new(big.Float).SetPrec(prec).SetFloat64(-1.121656496), new(big.Float).SetPrec(prec).SetFloat64(0.000000))
	ea.AddElement(l15)
	g1.AddElement(l15.GetID())
	ea.AddElementToCluster(l15.GetID(), g1.GetID())
	l12 := el.NewSketchLine(12, new(big.Float).SetPrec(prec).SetFloat64(-0.5633086396), new(big.Float).SetPrec(prec).SetFloat64(0.8262465592), new(big.Float).SetPrec(prec).SetFloat64(-3.804226065))
	ea.AddElement(l12)
	g1.AddElement(l12.GetID())
	ea.AddElementToCluster(l12.GetID(), g1.GetID())
	p11 := el.NewSketchPoint(11, new(big.Float).SetPrec(prec).SetFloat64(2.183329741), new(big.Float).SetPrec(prec).SetFloat64(6.092751026))
	ea.AddElement(p11)
	g1.AddElement(p11.GetID())
	ea.AddElementToCluster(p11.GetID(), g1.GetID())
	l9 := el.NewSketchLine(9, new(big.Float).SetPrec(prec).SetFloat64(0.6117352315), new(big.Float).SetPrec(prec).SetFloat64(0.7910625807), new(big.Float).SetPrec(prec).SetFloat64(-6.155367074))
	ea.AddElement(l9)
	g1.AddElement(l9.GetID())
	ea.AddElementToCluster(l9.GetID(), g1.GetID())
	p14 := el.NewSketchPoint(14, new(big.Float).SetPrec(prec).SetFloat64(-1.121656496), new(big.Float).SetPrec(prec).SetFloat64(3.839516468))
	ea.AddElement(p14)
	g1.AddElement(p14.GetID())
	ea.AddElementToCluster(p14.GetID(), g1.GetID())

	g2 := NewGraphCluster(2)

	ea.AddElementToCluster(p5.GetID(), g2.GetID())
	g2.AddElement(p5.GetID())
	e, _ := ea.GetElement(g2.GetID(), p5.GetID())
	p5 = e.AsPoint()
	p5.X.SetFloat64(2.488281499).SetPrec(prec)
	p5.Y.SetFloat64(-0.7247268643).SetPrec(prec)
	p8 := el.NewSketchPoint(8, new(big.Float).SetPrec(prec).SetFloat64(5.140000), new(big.Float).SetPrec(prec).SetFloat64(2.270000))
	ea.AddElement(p8)
	g2.AddElement(p8.GetID())
	ea.AddElementToCluster(p8.GetID(), g2.GetID())
	l6 := el.NewSketchLine(6, new(big.Float).SetPrec(prec).SetFloat64(2.994726864), new(big.Float).SetPrec(prec).SetFloat64(-2.651718501), new(big.Float).SetPrec(prec).SetFloat64(-9.373495085))
	ea.AddElement(l6)
	g2.AddElement(l6.GetID())
	ea.AddElementToCluster(l6.GetID(), g2.GetID())
	ea.AddElementToCluster(l9.GetID(), g2.GetID())
	g2.AddElement(l9.GetID())
	e, _ = ea.GetElement(g2.GetID(), l9.GetID())
	l9 = e.AsLine()
	l9.SetA(new(big.Float).SetPrec(prec).SetFloat64(0.8618389136))
	l9.SetB(new(big.Float).SetPrec(prec).SetFloat64(0.5071821044))
	l9.SetC(new(big.Float).SetPrec(prec).SetFloat64(-5.581155393))

	state := g0.solveMerge(ea, ca, g1, g2)

	if state != solver.Solved {
		t.Error("Expected solved state(4), got", state)
	}

	t.Logf("g0 elements length %d\n", g0.elements.Count())
	t.Logf("all elements count %d\n", len(ea.IdSet().Contents()))

	e, _ = ea.GetElement(g0.GetID(), 0)
	p0 = e.AsPoint()
	e, _ = ea.GetElement(g0.GetID(), 3)
	l3 = e.AsLine()
	e, _ = ea.GetElement(g0.GetID(), 5)
	p5 = e.AsPoint()
	e, _ = ea.GetElement(g0.GetID(), 6)
	l6 = e.AsLine()
	e, _ = ea.GetElement(g0.GetID(), 8)
	p8 = e.AsPoint()
	e, _ = ea.GetElement(g0.GetID(), 9)
	l9 = e.AsLine()
	e, _ = ea.GetElement(g0.GetID(), 11)
	p11 = e.AsPoint()
	e, _ = ea.GetElement(g0.GetID(), 12)
	l12 = e.AsLine()
	e, _ = ea.GetElement(g0.GetID(), 14)
	p14 = e.AsPoint()
	e, _ = ea.GetElement(g0.GetID(), 15)
	l15 = e.AsLine()

	// rad2Deg := func(rad float64) float64 { return (rad / math.Pi) * 180 }
	deg2Rad := func(deg float64) *big.Float {
		var pi, piDeg big.Float
		pi.SetPrec(utils.FloatPrecision).SetFloat64(math.Pi)
		piDeg.SetPrec(utils.FloatPrecision).SetFloat64(180)
		d := new(big.Float).SetPrec(utils.FloatPrecision).SetFloat64(deg)
		d.Quo(d, &piDeg)
		d.Mul(d, &pi)
		return d
	}
	desired := deg2Rad(108)
	desiredAlt := deg2Rad(72)
	// All angles should be 108 or 72 degrees
	var angle big.Float
	angle.Abs(l9.AngleToLine(l6))
	if utils.StandardBigFloatCompare(&angle, desired) != 0 && utils.StandardBigFloatCompare(&angle, desiredAlt) != 0 {
		t.Error("Expected l9 to l6 to be", desired.String(), "or", desiredAlt.String(), "degrees, got", angle.String())
	}
	angle.Abs(l6.AngleToLine(l3))
	if utils.StandardBigFloatCompare(&angle, desired) != 0 && utils.StandardBigFloatCompare(&angle, desiredAlt) != 0 {
		t.Error("Expected l6 to l3 to be", desired.String(), "or", desiredAlt.String(), "degrees, got", angle.String())
	}
	angle.Abs(l3.AngleToLine(l15))
	if utils.StandardBigFloatCompare(&angle, desired) != 0 && utils.StandardBigFloatCompare(&angle, desiredAlt) != 0 {
		t.Error("Expected l3 to l15 to be", desired.String(), "or", desiredAlt.String(), "degrees, got", angle.String())
	}
	angle.Abs(l15.AngleToLine(l12))
	if utils.StandardBigFloatCompare(&angle, desired) != 0 && utils.StandardBigFloatCompare(&angle, desiredAlt) != 0 {
		t.Error("Expected l12 to l15 to be", desired.String(), "or", desiredAlt.String(), "degrees, got", angle.String())
	}
	angle.Abs(l12.AngleToLine(l9))
	if utils.StandardBigFloatCompare(&angle, desired) != 0 && utils.StandardBigFloatCompare(&angle, desiredAlt) != 0 {
		t.Error("Expected l12 to l9 to be", desired.String(), "or", desiredAlt.String(), "degrees, got", angle.String())
	}

	desired = new(big.Float).SetPrec(prec).SetFloat64(4.0)
	measured := p0.DistanceTo(p5)
	if utils.StandardBigFloatCompare(measured, desired) != 0 {
		t.Error("Expected p0 to p5 to be", desired, ", got", measured)
	}
	measured = p5.DistanceTo(p8)
	if utils.StandardBigFloatCompare(measured, desired) != 0 {
		t.Error("Expected p5 to p8 to be", desired, ", got", measured)
	}
	measured = p8.DistanceTo(p11)
	if utils.StandardBigFloatCompare(measured, desired) != 0 {
		t.Error("Expected p8 to p11 to be", desired, ", got", measured)
	}
	measured = p11.DistanceTo(p14)
	if utils.StandardBigFloatCompare(measured, desired) != 0 {
		t.Error("Expected p11 to p14 to be", desired, ", got", measured)
	}
	measured = p14.DistanceTo(p0)
	if utils.StandardBigFloatCompare(measured, desired) != 0 {
		t.Error("Expected p14 to p0 to be", desired, ", got", measured)
	}

	t.Logf(`elements after solve:
	l3: %fx + %fy + %f = 0
	l6: %fx + %fy + %f = 0
	l9: %fx + %fy + %f = 0
	l12: %fx + %fy + %f = 0
	l15: %fx + %fy + %f = 0
	p0: (%f, %f)
	p5: (%f, %f)
	p8: (%f, %f)
	p11: (%f, %f)
	p14: (%f, %f)
	`,
		l3.GetA(), l3.GetB(), l3.GetC(),
		l6.GetA(), l6.GetB(), l6.GetC(),
		l9.GetA(), l9.GetB(), l9.GetC(),
		l12.GetA(), l12.GetB(), l12.GetC(),
		l15.GetA(), l15.GetB(), l15.GetC(),
		p0.GetX(), p0.GetY(),
		p5.GetX(), p5.GetY(),
		p8.GetX(), p8.GetY(),
		p11.GetX(), p11.GetY(),
		p14.GetX(), p14.GetY(),
	)
}

func TestSolveEdgeCases(t *testing.T) {
	ea := accessors.NewElementRepository()
	ca := accessors.NewConstraintRepository()
	e3 := el.NewSketchLine(2, big.NewFloat(2), big.NewFloat(1), big.NewFloat(-1))
	ea.AddElement(e3)
	ea.AddElementToCluster(2, 1)
	e4 := el.NewSketchLine(3, big.NewFloat(2), big.NewFloat(2), big.NewFloat(-0))
	ea.AddElement(e4)
	ea.AddElementToCluster(3, 1)
	c3 := constraint.NewConstraint(2, constraint.Distance, e3.GetID(), e4.GetID(), big.NewFloat(2), false)
	ca.AddConstraint(c3)

	o := NewGraphCluster(1)
	o.AddConstraint(c3)

	state := o.Solve(ea, ca)
	assert.Equal(t, solver.Solved, state, "Test local solve with solveorder < 2")

	o.solved = utils.NewSet()
	e1 := el.NewSketchPoint(0, big.NewFloat(0), big.NewFloat(1))
	ea.AddElement(e1)
	ea.AddElementToCluster(0, 1)
	e2 := el.NewSketchPoint(1, big.NewFloat(2), big.NewFloat(1))
	ea.AddElement(e2)
	ea.AddElementToCluster(1, 1)
	c1 := constraint.NewConstraint(0, constraint.Distance, e1.GetID(), e2.GetID(), big.NewFloat(5), false)
	ca.AddConstraint(c1)
	o.AddConstraint(c1)

	t.Log(o.ToGraphViz(ea, ca))
	state = o.Solve(ea, ca)
	assert.Equal(t, solver.NonConvergent, state, "Test local solve without enough constraints to solve desired element")

	c2 := constraint.NewConstraint(1, constraint.Distance, e2.GetID(), e3.GetID(), big.NewFloat(7), false)
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
	ea := accessors.NewElementRepository()
	ca := accessors.NewConstraintRepository()
	// Create fixed element cluster
	// Create cluster w/ square
	// merge the two -- use solveMerge instead of mergeOne
	g := NewGraphCluster(0)
	var e el.SketchElement = el.NewSketchPoint(0, big.NewFloat(0), big.NewFloat(0))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), g.GetID())
	g.AddElement(e.GetID())
	e = el.NewSketchLine(1, big.NewFloat(0), big.NewFloat(1), big.NewFloat(0))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), g.GetID())
	g.AddElement(e.GetID())
	e = el.NewSketchLine(2, big.NewFloat(1), big.NewFloat(0), big.NewFloat(0))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), g.GetID())
	g.AddElement(e.GetID())

	o := NewGraphCluster(1)
	e = el.NewSketchLine(1, big.NewFloat(-0.029929), big.NewFloat(-0.999552), big.NewFloat(0))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o.GetID())
	o.AddElement(e.GetID())

	state := g.solveMerge(ea, ca, o, nil)
	assert.Equal(t, solver.NonConvergent, state, "Merge containing only one shared element should fail to solve")

	e = el.NewSketchLine(2, big.NewFloat(0.999552), big.NewFloat(-0.029929), big.NewFloat(0))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o.GetID())
	o.AddElement(e.GetID())

	state = g.solveMerge(ea, ca, o, nil)
	assert.Equal(t, solver.NonConvergent, state, "Merge where shared elements are both lines should fail to solve")

	o = NewGraphCluster(1)
	e = el.NewSketchPoint(0, big.NewFloat(0), big.NewFloat(0))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o.GetID())
	o.AddElement(e.GetID())
	e = el.NewSketchPoint(5, big.NewFloat(3.998208), big.NewFloat(-0.119717))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o.GetID())
	o.AddElement(e.GetID())
	e = el.NewSketchLine(12, big.NewFloat(-0.563309), big.NewFloat(0.826247), big.NewFloat(-3.804226))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o.GetID())
	o.AddElement(e.GetID())
	e = el.NewSketchPoint(11, big.NewFloat(2.183330), big.NewFloat(6.092751))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o.GetID())
	o.AddElement(e.GetID())
	e = el.NewSketchLine(9, big.NewFloat(0.611735), big.NewFloat(0.791063), big.NewFloat(-6.155367))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o.GetID())
	o.AddElement(e.GetID())
	e = el.NewSketchPoint(8, big.NewFloat(5.347580), big.NewFloat(3.645810))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o.GetID())
	o.AddElement(e.GetID())
	e = el.NewSketchLine(1, big.NewFloat(-0.029929), big.NewFloat(-0.999552), big.NewFloat(0))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o.GetID())
	o.AddElement(e.GetID())
	e = el.NewSketchLine(2, big.NewFloat(0.999552), big.NewFloat(-0.029929), big.NewFloat(0))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o.GetID())
	o.AddElement(e.GetID())
	e = el.NewSketchLine(15, big.NewFloat(-0.959879), big.NewFloat(-0.280414), big.NewFloat(0))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o.GetID())
	o.AddElement(e.GetID())
	e = el.NewSketchPoint(14, big.NewFloat(-1.121656), big.NewFloat(3.839516))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o.GetID())
	o.AddElement(e.GetID())
	e = el.NewSketchLine(6, big.NewFloat(0.941382), big.NewFloat(-0.337343), big.NewFloat(-3.804226))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o.GetID())
	o.AddElement(e.GetID())
	e = el.NewSketchLine(3, big.NewFloat(-0.029929), big.NewFloat(-0.999552), big.NewFloat(0))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o.GetID())
	o.AddElement(e.GetID())
	state = g.solveMerge(ea, ca, o, nil)
	assert.Equal(t, solver.Solved, state, "Merge should solve successfully")

	g = NewGraphCluster(0)
	e = el.NewSketchPoint(0, big.NewFloat(0), big.NewFloat(0))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), g.GetID())
	g.AddElement(e.GetID())
	e = el.NewSketchLine(100, big.NewFloat(0), big.NewFloat(1), big.NewFloat(0))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), g.GetID())
	g.AddElement(e.GetID())
	e = el.NewSketchPoint(5, big.NewFloat(4), big.NewFloat(0))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), g.GetID())
	g.AddElement(e.GetID())

	state = g.solveMerge(ea, ca, o, nil)
	assert.Equal(t, solver.Solved, state, "Merge with two shared points should solve successfully")
}

func TestSolveMergeEdgeCases(t *testing.T) {
	ea := accessors.NewElementRepository()
	ca := accessors.NewConstraintRepository()
	g := NewGraphCluster(0)
	var e el.SketchElement = el.NewSketchPoint(0, big.NewFloat(0), big.NewFloat(0))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), g.GetID())
	g.AddElement(e.GetID())
	e = el.NewSketchLine(1, big.NewFloat(0), big.NewFloat(1), big.NewFloat(0))
	g.AddElement(e.GetID())
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), g.GetID())
	e = el.NewSketchLine(2, big.NewFloat(1), big.NewFloat(0), big.NewFloat(0))
	g.AddElement(e.GetID())
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), g.GetID())

	o1 := NewGraphCluster(1)
	e = el.NewSketchPoint(0, big.NewFloat(0), big.NewFloat(0))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o1.GetID())
	o1.AddElement(e.GetID())
	e = el.NewSketchPoint(5, big.NewFloat(3.998208), big.NewFloat(-0.119717))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o1.GetID())
	o1.AddElement(e.GetID())
	e = el.NewSketchLine(12, big.NewFloat(-0.563309), big.NewFloat(0.826247), big.NewFloat(-3.804226))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o1.GetID())
	o1.AddElement(e.GetID())
	e = el.NewSketchPoint(11, big.NewFloat(2.183330), big.NewFloat(6.092751))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o1.GetID())
	o1.AddElement(e.GetID())
	e = el.NewSketchLine(9, big.NewFloat(0.611735), big.NewFloat(0.791063), big.NewFloat(-6.155367))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o1.GetID())
	o1.AddElement(e.GetID())

	o2 := NewGraphCluster(2)
	e = el.NewSketchLine(15, big.NewFloat(-0.959879), big.NewFloat(-0.280414), big.NewFloat(0))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o2.GetID())
	o2.AddElement(e.GetID())
	e = el.NewSketchPoint(5, big.NewFloat(-1.121656), big.NewFloat(3.839516))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o2.GetID())
	o2.AddElement(e.GetID())
	e = el.NewSketchLine(6, big.NewFloat(0.941382), big.NewFloat(-0.337343), big.NewFloat(-3.804226))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o2.GetID())
	o2.AddElement(e.GetID())
	e = el.NewSketchLine(3, big.NewFloat(-0.029929), big.NewFloat(-0.999552), big.NewFloat(0))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o2.GetID())
	o2.AddElement(e.GetID())

	state := g.solveMerge(ea, ca, o1, o2)
	assert.Equal(t, solver.NonConvergent, state, "Three cluster solve with only two shared elements should fail")

	ea.Clear()
	// Solve merge with three lines lines
	g = NewGraphCluster(0)
	e = el.NewSketchPoint(0, big.NewFloat(0), big.NewFloat(0))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), g.GetID())
	g.AddElement(e.GetID())
	e = el.NewSketchLine(9, big.NewFloat(0.611735), big.NewFloat(0.791063), big.NewFloat(-6.155367))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), g.GetID())
	g.AddElement(e.GetID())
	e = el.NewSketchLine(1, big.NewFloat(0), big.NewFloat(1), big.NewFloat(0))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), g.GetID())
	g.AddElement(e.GetID())

	o1 = NewGraphCluster(1)
	e = el.NewSketchPoint(7, big.NewFloat(0), big.NewFloat(1))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o1.GetID())
	o1.AddElement(e.GetID())
	e = el.NewSketchPoint(5, big.NewFloat(3.998208), big.NewFloat(-0.119717))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o1.GetID())
	o1.AddElement(e.GetID())
	e = el.NewSketchPoint(11, big.NewFloat(2.183330), big.NewFloat(6.092751))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o1.GetID())
	o1.AddElement(e.GetID())
	e = el.NewSketchLine(9, big.NewFloat(0.611735), big.NewFloat(0.791063), big.NewFloat(-6.155367))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o1.GetID())
	o1.AddElement(e.GetID())
	e = el.NewSketchLine(2, big.NewFloat(1), big.NewFloat(0), big.NewFloat(0))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o1.GetID())
	o1.AddElement(e.GetID())

	o2 = NewGraphCluster(2)
	e = el.NewSketchLine(15, big.NewFloat(-0.959879), big.NewFloat(-0.280414), big.NewFloat(0))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o2.GetID())
	o2.AddElement(e.GetID())
	e = el.NewSketchPoint(8, big.NewFloat(3.998208), big.NewFloat(-0.119717))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o2.GetID())
	o2.AddElement(e.GetID())
	e = el.NewSketchLine(1, big.NewFloat(0), big.NewFloat(1), big.NewFloat(0))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o2.GetID())
	o2.AddElement(e.GetID())
	e = el.NewSketchLine(2, big.NewFloat(1), big.NewFloat(0.0), big.NewFloat(0.0))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o2.GetID())
	o2.AddElement(e.GetID())

	state = g.solveMerge(ea, ca, o1, o2)
	assert.Equal(t, solver.Solved, state, "Three cluster solve with three lines")

	ea.Clear()
	// Solve merge with one point and two lines where lines are in clusters 0 and 1
	// I don't know where I got these values... they may be incorrect
	g = NewGraphCluster(0)
	e = el.NewSketchPoint(0, big.NewFloat(0), big.NewFloat(0))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), g.GetID())
	g.AddElement(e.GetID())
	e = el.NewSketchLine(3, big.NewFloat(0), big.NewFloat(-1), big.NewFloat(0))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), g.GetID())
	g.AddElement(e.GetID())
	e = el.NewSketchPoint(5, big.NewFloat(4), big.NewFloat(0))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), g.GetID())
	g.AddElement(e.GetID())
	e = el.NewSketchLine(1, big.NewFloat(0), big.NewFloat(-1), big.NewFloat(0))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), g.GetID())
	g.AddElement(e.GetID())

	o1 = NewGraphCluster(1)
	e = el.NewSketchPoint(11, big.NewFloat(2.183330), big.NewFloat(6.092751))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o1.GetID())
	o1.AddElement(e.GetID())
	e = el.NewSketchLine(12, big.NewFloat(-0.563309), big.NewFloat(0.826247), big.NewFloat(-3.804226))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o1.GetID())
	o1.AddElement(e.GetID())
	e = el.NewSketchLine(15, big.NewFloat(-3.839516), big.NewFloat(-1.121656), big.NewFloat(0))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o1.GetID())
	o1.AddElement(e.GetID())
	e = el.NewSketchLine(9, big.NewFloat(0.611735), big.NewFloat(0.791063), big.NewFloat(-6.155367))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o1.GetID())
	o1.AddElement(e.GetID())
	e = el.NewSketchPoint(14, big.NewFloat(-1.121656), big.NewFloat(3.839516))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o1.GetID())
	o1.AddElement(e.GetID())
	e = el.NewSketchPoint(0, big.NewFloat(0), big.NewFloat(0))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o1.GetID())
	o1.AddElement(e.GetID())

	o2 = NewGraphCluster(2)
	e = el.NewSketchPoint(8, big.NewFloat(5.14), big.NewFloat(2.27))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o2.GetID())
	o2.AddElement(e.GetID())
	e = el.NewSketchLine(6, big.NewFloat(2.029929), big.NewFloat(-2.651719), big.NewFloat(-9.373495))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o2.GetID())
	o2.AddElement(e.GetID())
	e = el.NewSketchPoint(5, big.NewFloat(2.488281), big.NewFloat(-0.724727))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o2.GetID())
	o2.AddElement(e.GetID())
	e = el.NewSketchLine(9, big.NewFloat(0.861839), big.NewFloat(0.507182), big.NewFloat(-5.581155))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o2.GetID())
	o2.AddElement(e.GetID())

	state = g.solveMerge(ea, ca, o1, o2)
	assert.Equal(t, solver.Solved, state, "Three cluster solve with three shared elements should solve")

	ea.ReplaceElement(g.GetID(), 0, el.NewSketchLine(0, big.NewFloat(1), big.NewFloat(1), big.NewFloat(1)))
	ea.ReplaceElement(o1.GetID(), 0, el.NewSketchLine(0, big.NewFloat(3), big.NewFloat(2), big.NewFloat(1)))
	line, _ := ea.GetElement(-1, 9)
	line.AsLine().SetB(big.NewFloat(0.235))
	line.AsLine().SetC(big.NewFloat(2))
	state = g.solveMerge(ea, ca, o1, o2)
	assert.Equal(t, solver.NonConvergent, state, "Three cluster solve with non-convergent lines")

	ea.ReplaceElement(o1.GetID(), 2, el.NewSketchLine(2, big.NewFloat(1), big.NewFloat(0), big.NewFloat(0)))
	ea.ReplaceElement(o2.GetID(), 2, el.NewSketchLine(2, big.NewFloat(1), big.NewFloat(0), big.NewFloat(0)))
	ea.ReplaceElement(g.GetID(), 1, el.NewSketchLine(1, big.NewFloat(1), big.NewFloat(0), big.NewFloat(6)))
	ea.ReplaceElement(o2.GetID(), 1, el.NewSketchLine(1, big.NewFloat(1), big.NewFloat(0), big.NewFloat(6)))
	ea.ReplaceElement(g.GetID(), 0, el.NewSketchPoint(0, big.NewFloat(-10), big.NewFloat(1)))
	ea.ReplaceElement(o1.GetID(), 0, el.NewSketchPoint(0, big.NewFloat(3), big.NewFloat(2)))
	state = g.solveMerge(ea, ca, o1, o2)
	assert.Equal(t, solver.NonConvergent, state, "Fail to solve final element")

	ea.Clear()

	g = NewGraphCluster(0)
	e = el.NewSketchPoint(0, big.NewFloat(0), big.NewFloat(0))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), g.GetID())
	g.AddElement(e.GetID())
	e = el.NewSketchLine(9, big.NewFloat(0.611735), big.NewFloat(0.791063), big.NewFloat(-6.155367))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), g.GetID())
	g.AddElement(e.GetID())
	e = el.NewSketchLine(1, big.NewFloat(0), big.NewFloat(1), big.NewFloat(0))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), g.GetID())
	g.AddElement(e.GetID())

	o1 = NewGraphCluster(1)
	e = el.NewSketchPoint(0, big.NewFloat(0), big.NewFloat(0))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o1.GetID())
	o1.AddElement(e.GetID())
	e = el.NewSketchPoint(5, big.NewFloat(3.998208), big.NewFloat(-0.119717))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o1.GetID())
	o1.AddElement(e.GetID())
	e = el.NewSketchPoint(11, big.NewFloat(2.183330), big.NewFloat(6.092751))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o1.GetID())
	o1.AddElement(e.GetID())
	e = el.NewSketchLine(6, big.NewFloat(0.611735), big.NewFloat(0.791063), big.NewFloat(-6.155367))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o1.GetID())
	o1.AddElement(e.GetID())
	e = el.NewSketchLine(2, big.NewFloat(1), big.NewFloat(0), big.NewFloat(0))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o1.GetID())
	o1.AddElement(e.GetID())

	o2 = NewGraphCluster(2)
	e = el.NewSketchLine(15, big.NewFloat(-0.959879), big.NewFloat(-0.280414), big.NewFloat(0))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o2.GetID())
	o2.AddElement(e.GetID())
	e = el.NewSketchPoint(8, big.NewFloat(3.998208), big.NewFloat(-0.119717))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o2.GetID())
	o2.AddElement(e.GetID())
	e = el.NewSketchLine(1, big.NewFloat(0), big.NewFloat(1), big.NewFloat(0))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o2.GetID())
	o2.AddElement(e.GetID())
	e = el.NewSketchLine(2, big.NewFloat(1), big.NewFloat(0.0), big.NewFloat(0.0))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o2.GetID())
	o2.AddElement(e.GetID())
	e = el.NewSketchPoint(0, big.NewFloat(0), big.NewFloat(0))
	ea.AddElement(e)
	ea.AddElementToCluster(e.GetID(), o2.GetID())
	o2.AddElement(e.GetID())

	state = g.solveMerge(ea, ca, o1, o2)
	assert.Equal(t, solver.NonConvergent, state, "Nonconvergent where an element has too many parents")
}

func TestToGraphViz(t *testing.T) {
	ea := accessors.NewElementRepository()
	ca := accessors.NewConstraintRepository()
	e1 := el.NewSketchPoint(0, big.NewFloat(0), big.NewFloat(1))
	e2 := el.NewSketchPoint(1, big.NewFloat(2), big.NewFloat(1))
	e3 := el.NewSketchLine(2, big.NewFloat(2), big.NewFloat(1), big.NewFloat(-1))
	e4 := el.NewSketchLine(3, big.NewFloat(2), big.NewFloat(2), big.NewFloat(-0))
	ea.AddElement(e1)
	ea.AddElement(e2)
	ea.AddElement(e3)
	ea.AddElement(e4)
	c1 := constraint.NewConstraint(0, constraint.Distance, e1.GetID(), e2.GetID(), big.NewFloat(5), false)
	c2 := constraint.NewConstraint(1, constraint.Distance, e2.GetID(), e3.GetID(), big.NewFloat(7), false)
	c3 := constraint.NewConstraint(2, constraint.Distance, e3.GetID(), e4.GetID(), big.NewFloat(2), false)

	g := NewGraphCluster(0)
	ca.AddConstraint(c1)
	g.AddConstraint(c1)
	ca.AddConstraint(c2)
	g.AddConstraint(c2)
	ca.AddConstraint(c3)
	g.AddConstraint(c3)

	gvString := g.ToGraphViz(ea, ca)
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
