package constraint

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func TestConstraintTypeString(t *testing.T) {
	tests := []struct {
		name           string
		constraintType Type
		expected       string
	}{
		{"Distance type to string", Distance, "Distance"},
		{"Angle type to string", Angle, "Angle"},
		{"Unknown type to string", 5, "5"},
	}

	for _, tt := range tests {
		str := tt.constraintType.String()
		assert.Equal(t, str, tt.expected)
	}
}

func TestConstraintIdAndValue(t *testing.T) {
	tests := []struct {
		name          string
		constraint    *Constraint
		expectedId    uint
		expectedValue float64
	}{
		{"Get constraint id, value", NewConstraint(0, Distance, 0, 0, 0.5, false), 0, 0.5},
	}

	updatedValue := 3.14159265358979
	for _, tt := range tests {
		id := tt.constraint.GetID()
		value := tt.constraint.GetValue()
		assert.Equal(t, id, tt.expectedId)
		assert.Equal(t, value, tt.expectedValue)
		tt.constraint.UpdateValue(updatedValue)
		assert.Equal(t, updatedValue, tt.constraint.GetValue())
	}
}

func TestConstraintElements(t *testing.T) {
	tests := []struct {
		name       string
		constraint *Constraint
		other      *Constraint
		ids        []uint
		hasIds     []bool
		otherIds   []uint
		hasAllIds  bool
		shared     uint
		hasShared  bool
		isMet      bool
		isEqual    bool
		isFirst    bool
	}{
		{
			"Two constraint elements shared and is met (not equal)",
			NewConstraint(0, Distance, 0, 1, 4, true),
			NewConstraint(1, Distance, 0, 1, 4, true),
			[]uint{0, 1, 2},
			[]bool{true, true, false},
			[]uint{1, 0, 0},
			false,
			0,
			true,
			true,
			false,
			true,
		},
		{
			"One constraint element shared and is not met (is equal)",
			NewConstraint(2, Distance, 2, 0, 5, false),
			NewConstraint(3, Distance, 0, 1, 4, true),
			[]uint{2, 0},
			[]bool{true, true},
			[]uint{0, 2},
			true,
			0,
			true,
			false,
			true,
			false,
		},
		{
			"Second constraint element shared",
			NewConstraint(4, Angle, 2, 3, math.Pi/2.0, true),
			NewConstraint(5, Distance, 4, 5, 0, true),
			[]uint{2, 3},
			[]bool{true, true},
			[]uint{3, 2},
			true,
			4,
			false,
			true,
			true,
			false,
		},
	}
	for _, tt := range tests {
		for i := range tt.ids {
			assert.Equal(t, tt.hasIds[i], tt.constraint.HasElementID(tt.ids[i]), tt.name)
			ok := tt.constraint.HasElementID(tt.ids[i])
			assert.Equal(t, tt.hasIds[i], ok, tt.name)
			other, ok := tt.constraint.Other(tt.ids[i])
			assert.Equal(t, tt.otherIds[i], other, tt.name)
			assert.Equal(t, tt.hasIds[i], ok, tt.name)
		}
		assert.Equal(t, tt.hasAllIds, tt.constraint.HasElements(tt.ids...), tt.name)
		shared, ok := tt.constraint.Shared(tt.other)
		assert.Equal(t, tt.hasShared, ok, tt.name)
		if tt.hasShared {
			assert.Equal(t, tt.shared, shared, tt.name)
		} else {
			assert.NotEqual(t, tt.shared, shared, tt.name)
		}
		// assert.Equal(t, tt.isMet, tt.constraint.IsMet(), tt.name)
		assert.Equal(t, tt.isEqual, tt.constraint.Equals(*tt.other), tt.name)
		assert.Equal(t, tt.isFirst, tt.constraint.First() == tt.shared, tt.name)
		assert.Equal(t, !tt.isFirst && tt.hasShared, tt.constraint.Second() == tt.shared, tt.name)
		if tt.hasAllIds {
			assert.Equal(t, tt.ids, tt.constraint.ElementIDs(), tt.name)
		} else {
			assert.NotEqual(t, tt.ids, tt.constraint.ElementIDs(), tt.name)
		}
	}
}

func TestConstraintStringGraphviz(t *testing.T) {
	tests := []struct {
		name       string
		constraint *Constraint
	}{
		{
			"Distance Constraint",
			NewConstraint(0, Distance, 0, 1, 4, true),
		},
		{
			"Angle Constraint",
			NewConstraint(0, Angle, 1, 2, math.Pi/2, true),
		},
	}
	for _, tt := range tests {
		str := tt.constraint.String()
		assert.True(t, strings.Contains(str, fmt.Sprintf("e1: %d", tt.constraint.Element1)))
		assert.True(t, strings.Contains(str, fmt.Sprintf("e2: %d", tt.constraint.Element2)))
		assert.True(t, strings.Contains(str, fmt.Sprintf("Constraint(%d)", tt.constraint.GetID())))
		assert.True(t, strings.Contains(str, fmt.Sprintf("v: %f", tt.constraint.Value)))

		str = tt.constraint.ToGraphViz(7)
		assert.True(t, strings.Contains(str, fmt.Sprintf("7-%d", tt.constraint.Element1)))
		assert.True(t, strings.Contains(str, fmt.Sprintf("7-%d", tt.constraint.Element2)))
		assert.True(t, strings.Contains(str, fmt.Sprintf("%v (%d)", tt.constraint.Type, tt.constraint.GetID())))

		str = tt.constraint.ToGraphViz(-1)
		assert.False(t, strings.Contains(str, fmt.Sprintf("-1-%d", tt.constraint.Element1)))
		assert.True(t, strings.Contains(str, fmt.Sprintf("%d", tt.constraint.Element1)))
		assert.False(t, strings.Contains(str, fmt.Sprintf("-1-%d", tt.constraint.Element2)))
		assert.True(t, strings.Contains(str, fmt.Sprintf("%d", tt.constraint.Element2)))
		assert.True(t, strings.Contains(str, fmt.Sprintf("%v (%d)", tt.constraint.Type, tt.constraint.GetID())))
	}
}

func TestCopyAndSort(t *testing.T) {
	constraintList := ConstraintList{
		NewConstraint(1, Distance, 0, 1, 4, true),
		NewConstraint(0, Angle, 1, 2, math.Pi/2, true),
	}

	constraintList = append(constraintList, CopyConstraint(constraintList[1]))
	sort.Sort(constraintList)

	sortOrder := []uint{0, 0, 1}

	for i, tt := range constraintList {
		assert.Equal(t, sortOrder[i], tt.GetID())
	}

	log.Logger.Trace().Array("test", constraintList)
}
