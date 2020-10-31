package core

import (
	"math"
	"testing"

	"github.com/marcuswu/dlineate/internal/constraint"
	el "github.com/marcuswu/dlineate/internal/element"
	"github.com/marcuswu/dlineate/internal/solver"
	"github.com/marcuswu/dlineate/utils"
)

func TestAddConstraint(t *testing.T) {
	e1 := el.NewSketchPoint(0, 0, 1)
	e2 := el.NewSketchPoint(1, 2, 1)
	e3 := el.NewSketchLine(2, 2, 1, -1)
	c1 := constraint.NewConstraint(0, constraint.Distance, e1, e2, 5)
	c2 := constraint.NewConstraint(0, constraint.Distance, e2, e3, 7)

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
	c1 := constraint.NewConstraint(0, constraint.Distance, e1, e2, 5)
	c2 := constraint.NewConstraint(0, constraint.Distance, e2, e3, 7)
	c3 := constraint.NewConstraint(0, constraint.Distance, e3, e4, 2)

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
	c1 := constraint.NewConstraint(0, constraint.Distance, e1, e2, 5)
	c2 := constraint.NewConstraint(0, constraint.Distance, e2, e3, 7)
	c3 := constraint.NewConstraint(0, constraint.Distance, e3, e4, 2)

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
	if e3 != element3 {
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
	c1 := constraint.NewConstraint(0, constraint.Distance, e1, e2, 5)
	c2 := constraint.NewConstraint(1, constraint.Distance, e2, e3, 7)
	c3 := constraint.NewConstraint(2, constraint.Distance, e3, e4, 2)

	g := NewGraphCluster()
	g.AddConstraint(c1)
	g.AddConstraint(c2)

	o := NewGraphCluster()
	o.AddConstraint(c3)
	g.others = append(g.others, o)

	g2 := NewGraphCluster()
	g3 := NewGraphCluster()
	e5 := el.NewSketchPoint(4, 0, 1)
	e6 := el.NewSketchPoint(5, 1, 2)
	c4 := constraint.NewConstraint(3, constraint.Distance, e4, e5, 12)
	c5 := constraint.NewConstraint(3, constraint.Distance, e5, e6, 12)
	g2.AddConstraint(c4)
	g3.AddConstraint(c5)

	shared := g.SharedElements(g3)

	if shared.Count() != 0 {
		t.Error("There should be no shared element between g and g2, found", shared.Count())
	}

	shared = g.SharedElements(g2)
	if shared.Count() != 1 {
		t.Error("There should be one shared element between g and g2, found", shared.Count())
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
	c1 := constraint.NewConstraint(0, constraint.Distance, e1, e2, 5)
	c2 := constraint.NewConstraint(0, constraint.Distance, e2, e3, 7)

	g := NewGraphCluster()
	g.AddConstraint(c1)
	g.AddConstraint(c2)

	g.Translate(1, 1)

	if e1.GetX() != 1 && e1.GetY() != 2 {
		t.Error("Expected the e1 to be 1, 2, got", e1)
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
	c1 := constraint.NewConstraint(0, constraint.Distance, e1, e2, 5)
	c2 := constraint.NewConstraint(0, constraint.Distance, e2, e3, 7)

	g := NewGraphCluster()
	g.AddConstraint(c1)
	g.AddConstraint(c2)

	g.Rotate(o, math.Pi/2.0)

	if utils.StandardFloatCompare(e3.GetA(), 0.5) != 0 ||
		utils.StandardFloatCompare(e3.GetB(), -1.0) != 0 ||
		utils.StandardFloatCompare(e3.GetC(), 0.5) != 0 {
		t.Error("Expected e3 to be 0.5, -1, 0.5. Got", e3.GetA(), ",", e3.GetB(), ",", e3.GetC())
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

func TestLocalSolve(t *testing.T) {
	g := NewGraphCluster()

	// lines and points for a square -- intentionally off
	l1 := el.NewSketchLine(0, 1, 0.1, -1) // right line
	l2 := el.NewSketchLine(1, 0, 1, -1.1) // top line
	l3 := el.NewSketchLine(2, 1.1, 0, 0)  // left line
	l4 := el.NewSketchLine(3, 0, 1, 0.1)  // bottom line
	p1 := el.NewSketchPoint(4, 0.1, 1)    // top left
	p2 := el.NewSketchPoint(5, 1, 1.1)    // top right
	p3 := el.NewSketchPoint(6, 1.1, 0)    // botton right
	p4 := el.NewSketchPoint(7, 0, 0.1)    // bottom left
	c := constraint.NewConstraint(0, constraint.Distance, p1, p2, 1)
	g.AddConstraint(c)
	c = constraint.NewConstraint(1, constraint.Distance, p2, p3, 1)
	g.AddConstraint(c)
	c = constraint.NewConstraint(2, constraint.Distance, p3, p4, 1)
	g.AddConstraint(c)
	c = constraint.NewConstraint(3, constraint.Distance, p4, p1, 1)
	g.AddConstraint(c)
	c = constraint.NewConstraint(4, constraint.Distance, p2, l1, 0)
	g.AddConstraint(c)
	c = constraint.NewConstraint(5, constraint.Distance, p3, l1, 0)
	g.AddConstraint(c)
	c = constraint.NewConstraint(6, constraint.Distance, p1, l2, 0)
	g.AddConstraint(c)
	c = constraint.NewConstraint(7, constraint.Distance, p2, l2, 0)
	g.AddConstraint(c)
	c = constraint.NewConstraint(8, constraint.Distance, p1, l3, 0)
	g.AddConstraint(c)
	c = constraint.NewConstraint(9, constraint.Distance, p4, l3, 0)
	g.AddConstraint(c)
	c = constraint.NewConstraint(10, constraint.Distance, p4, l4, 0)
	g.AddConstraint(c)
	c = constraint.NewConstraint(11, constraint.Distance, p3, l4, 0)
	g.AddConstraint(c)
	c = constraint.NewConstraint(12, constraint.Angle, l1, l2, math.Pi/2)
	g.AddConstraint(c)
	c = constraint.NewConstraint(13, constraint.Angle, l3, l4, math.Pi/2)
	g.AddConstraint(c)

	state := g.localSolve()

	if state != solver.Solved {
		t.Error("Expected solved state(4), got", state)
	}
}
