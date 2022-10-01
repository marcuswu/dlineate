package element

import (
	"math"
	"testing"

	"github.com/marcuswu/dlineation/utils"
)

func TestAngleToLine(t *testing.T) {
	l1 := NewSketchLine(0, -0.611735, -0.791063, 6.155367)
	l2 := NewSketchLine(1, -0.563309, 0.826247, -3.804226)

	a := l1.AngleToLine(l2)
	b := l2.AngleToLine(l1)

	var angle = 108 * math.Pi / 180
	if utils.StandardFloatCompare(a, -angle) != 0 {
		t.Errorf("Expected angle to be -108º (%f), got %f\n", angle, a)
	}
	if utils.StandardFloatCompare(b, angle) != 0 {
		t.Errorf("Expected angle to be -108º (%f), got %f\n", angle, b)
	}
}
