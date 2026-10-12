package dlineate

import (
	"fmt"

	el "github.com/marcuswu/dlineate/internal/element"
)

// SolveStatus describes the outcome of a sketch solve. Solved means the
// current solver accepted its constraints; it does not imply that the sketch
// has zero remaining degrees of freedom. The current graph solver may classify
// sketches with remaining freedom as NonConvergent instead of
// UnderConstrained.
type SolveStatus uint8

const (
	SolveStatusUnknown SolveStatus = iota
	SolveStatusSolved
	SolveStatusUnderConstrained
	SolveStatusOverConstrained
	SolveStatusNonConvergent
)

func (s SolveStatus) String() string {
	switch s {
	case SolveStatusSolved:
		return "solved"
	case SolveStatusUnderConstrained:
		return "under constrained"
	case SolveStatusOverConstrained:
		return "over constrained"
	case SolveStatusNonConvergent:
		return "non-convergent"
	default:
		return "unknown"
	}
}

// ElementSolveStatus is the public, package-independent constraint status of
// an element in a solve report.
type ElementSolveStatus uint8

const (
	ElementSolveStatusUnknown ElementSolveStatus = iota
	ElementSolveStatusUnderConstrained
	ElementSolveStatusFullyConstrained
	ElementSolveStatusOverConstrained
)

func (s ElementSolveStatus) String() string {
	switch s {
	case ElementSolveStatusUnderConstrained:
		return "under constrained"
	case ElementSolveStatusFullyConstrained:
		return "fully constrained"
	case ElementSolveStatusOverConstrained:
		return "over constrained"
	default:
		return "unknown"
	}
}

// ElementSolveResult is an immutable snapshot of one sketch element at the
// end of a solve. Values follows Element.Values() ordering for its type.
type ElementSolveResult struct {
	ID                    uint
	Type                  ElementType
	Status                ElementSolveStatus
	DegreesOfFreedom      int
	DegreesOfFreedomKnown bool
	Values                []float64
}

// ConstraintSolveResult is an immutable snapshot of one public constraint.
// CoreIDs identify the lower-level constraints used internally; callers
// should use ID as the stable public identifier.
type ConstraintSolveResult struct {
	ID          uint64
	Type        ConstraintType
	State       ConstraintState
	ElementIDs  []uint
	Value       float64
	ValueKnown  bool
	CoreIDs     []uint
	Conflicting bool
}

// SolveDiagnostic provides a machine-readable summary of a solve outcome.
// ConstraintIDs and ElementIDs refer to public sketch-level identifiers.
type SolveDiagnostic struct {
	Code          string
	Message       string
	ConstraintIDs []uint64
	ElementIDs    []uint
}

// SolveReport is a snapshot of the sketch after a solve attempt. The current
// solver reports solved, over-constrained, or non-convergent. Under-constrained
// is included in the public model but is not yet reliably distinguished from
// non-convergence by the underlying algorithms.
type SolveReport struct {
	Status      SolveStatus
	Passes      int
	Error       string
	Elements    []ElementSolveResult
	Constraints []ConstraintSolveResult
	Conflicts   []ConstraintSolveResult
	Diagnostics []SolveDiagnostic
}

// Err returns the error represented by the report. It returns nil when the
// legacy Sketch.Solve() call would return nil; otherwise it returns the same
// error text, or the generic solve failure for an empty/unknown report.
func (r SolveReport) Err() error {
	if r.Error != "" {
		return fmt.Errorf("%s", r.Error)
	}
	if r.Status != SolveStatusUnknown {
		return nil
	}
	return fmt.Errorf("failed to solve completely")
}

func publicElementStatus(level el.ConstraintLevel) ElementSolveStatus {
	switch level {
	case el.OverConstrained:
		return ElementSolveStatusOverConstrained
	case el.UnderConstrained:
		return ElementSolveStatusUnderConstrained
	case el.FullyConstrained:
		return ElementSolveStatusFullyConstrained
	default:
		return ElementSolveStatusUnknown
	}
}
