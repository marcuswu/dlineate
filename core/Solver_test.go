package core

import (
	"testing"

	"github.com/marcuswu/dlineation/utils"
)

func TestPointFromPoints(t *testing.T) {
	p1 := NewSketchPoint(0, 1, 1)
	p2 := NewSketchPoint(1, 3, 5)
	p3 := NewSketchPoint(2, 0, 2)

	newP3, state := pointFromPoints(p1, p2, p3, 1, 3)

	if state != NonConvergent {
		t.Error("Expected non-convergent state got ", state)
	}

	newP3, state = pointFromPoints(p1, p2, p3, 1, 5)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if newP3.GetID() != 2 {
		t.Error("Expected newP3 to have id 2, got ", newP3.GetID())
	}

	if utils.StandardFloatCompare(p1.DistanceTo(newP3), 1) != 0 {
	// if utils.StandardFloatCompare(newP3.GetX(), 0.447213595499958) != 0 && utils.StandardFloatCompare(newP3.GetY(), 1.8944271909999164) != 0 {
		t.Error("Expected newP3 to have distance of 1 to p1, got ", p1.DistanceTo(newP3))
		// t.Error("Expected newP3 to have location 0.447213595499958, 1.8944271909999164, got ", newP3)
	}

	if utils.StandardFloatCompare(p2.DistanceTo(newP3), 5) != 0 {
	// if utils.StandardFloatCompare(newP3.GetX(), 0.447213595499958) != 0 && utils.StandardFloatCompare(newP3.GetY(), 1.8944271909999164) != 0 {
		t.Error("Expected newP3 to have distance of 5 to p2, got ", p2.DistanceTo(newP3))
		// t.Error("Expected newP3 to have location 0.447213595499958, 1.8944271909999164, got ", newP3)
	}

	p3 = NewSketchPoint(2, 2, 1)
	newP3, state = pointFromPoints(p1, p2, p3, 1, 5)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if newP3.GetID() != 2 {
		t.Error("Expected newP3 to have id 2, got ", newP3.GetID())
	}

	if utils.StandardFloatCompare(newP3.GetX(), 2.0472135954999584) != 0 && utils.StandardFloatCompare(newP3.GetY(), 1.094427190999916) != 0 {
		t.Error("Expected newP3 to have location 2, 1.2, got ", newP3)
	}
}
