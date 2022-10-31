package solver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSolveStateToString(t *testing.T) {
	tests := []struct {
		name    string
		state   SolveState
		desired string
	}{
		{"Test Over Constraint to string", OverConstrained, "over constrained"},
		{"Test Non Convergent to string", NonConvergent, "non-convergent"},
		{"Test Solved to string", Solved, "solved"},
		{"Test unknown solve state to string", 7, "7"},
	}
	for _, tt := range tests {
		result := tt.state.String()
		assert.Equal(t, tt.desired, result)
	}
}
