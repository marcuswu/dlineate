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

	// p1, p2, and p3 should remain the same
	if p1.GetX() != 1 || p1.GetY() != 1 {
		t.Error("Expected p1 to remain at 1, 1, got: ", p1)
	}
	if p2.GetX() != 3 || p2.GetY() != 5 {
		t.Error("Expected p1 to remain at 3, 5, got: ", p2)
	}
	if p3.GetX() != 0 || p3.GetY() != 2 {
		t.Error("Expected p1 to remain at 0, 2, got: ", p3)
	}

	if utils.StandardFloatCompare(p1.DistanceTo(newP3), 1) != 0 {
		t.Error("Expected newP3 to have distance of 1 to p1, got ", p1.DistanceTo(newP3))
	}

	if utils.StandardFloatCompare(p2.DistanceTo(newP3), 5) != 0 {
		t.Error("Expected newP3 to have distance of 5 to p2, got ", p2.DistanceTo(newP3))
	}

	p3 = NewSketchPoint(2, 2, 1)
	var newP32 SketchElement

	newP32, state = pointFromPoints(p1, p2, p3, 1, 5)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if newP32.GetID() != 2 {
		t.Error("Expected newP3 to have id 2, got ", newP3.GetID())
	}

	if utils.StandardFloatCompare(p1.DistanceTo(newP32), 1) != 0 {
		t.Error("Expected newP3 to have distance of 1 to p1, got ", p1.DistanceTo(newP3))
	}

	if utils.StandardFloatCompare(p2.DistanceTo(newP32), 5) != 0 {
		t.Error("Expected newP3 to have distance of 5 to p2, got ", p2.DistanceTo(newP3))
	}

	if utils.StandardFloatCompare(newP3.GetX(), newP32.GetX()) == 0 {
		t.Error("Expected newP3 and newP32 to be different points, got ", newP3, newP32)
	}

	if utils.StandardFloatCompare(newP3.GetY(), newP32.GetY()) == 0 {
		t.Error("Expected newP3 and newP32 to be different points, got ", newP3, newP32)
	}
}
