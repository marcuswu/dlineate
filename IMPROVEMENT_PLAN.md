# Dlineate Improvement Plan

This plan captures the work needed to improve Dlineate's maintainability, public API, solve diagnostics, and constraint coverage. It is intended to guide implementation in small, testable stages while preserving the current graph-based solver and its numeric fallback.

## Current baseline

- `go test ./...` passes for the module.
- `go vet ./...` passes for the module.
- The module already includes a graph/cluster solve path, a Gonum-based numeric fallback, sketch elements and constraints, SVG/GraphViz output, and APIs for element constraint levels and conflicting constraints.
- The current public solve result is a generic error; element status and conflict reporting need stronger, verified semantics before they can safely drive UI behavior.

## Goals

1. Provide reliable overall and per-element constraint status for UI coloring.
2. Return actionable details for inconsistent constraints and distinguish those from under-constrained or non-convergent solves.
3. Make the public Go API consistent, safe, discoverable, and usable without exposing internal implementation types.
4. Improve solver determinism, diagnosability, and numeric fallback behavior.
5. Fill documented constraint/functionality gaps and keep documentation aligned with the supported API.

## Phase 1 — Characterize current behavior

### Work

- Add end-to-end sketch fixtures for under-constrained, fully constrained, over-constrained, and non-convergent cases.
- Cover isolated and disconnected geometry, points, lines, circles, and arcs, plus both ordinary graph merges and numeric fallback merges.
- Include redundant-but-consistent constraints as well as genuinely conflicting constraints.
- Assert solved geometry, solve status, per-element status, and conflict membership; record expected behavior where current output is wrong or ambiguous.
- Preserve representative existing examples as regression cases.

### Exit criteria

- Tests express the intended public behavior independently of internal cluster state.
- Each known status/reporting defect has a focused failing test before its implementation change.
- The full existing test suite remains a baseline for compatibility.

### Characterized baseline

The initial regression fixtures are in `solver_characterization_test.go` and `internal/graph/numeric_merge_characterization_test.go`. They capture the current behavior, including known defects to be corrected in later phases:

- Isolated point, line, circle, and arc geometry returns the generic `failed to solve completely` error, preserves its initial values, reports no conflicts, and currently labels every element fully constrained.
- Two disconnected components, each consisting of a fixed anchor and a point constrained only by distance, also fail with the generic error; their initial geometry is preserved. The solver does not currently expose the remaining freedom as an under-constrained result.
- The fixed-distance pentagon fixture solves successfully with expected bounds and fully constrained line statuses. This preserves a representative graph-cluster/merge regression.
- The modified-cylinder sketch solves successfully and preserves representative line and arc geometry as an integration regression.
- Two fixed points with duplicate, identical distance constraints fail and mark one duplicate as conflicting. This is a false positive for inconsistency; the geometry and constraints are mutually consistent.
- A deliberately inconsistent circle/line tangent sketch fails but returns no conflicting constraints. This is a false negative for actionable conflict reporting.
- A fixed four-triangle ring fixture in the graph package verifies a successful numeric fallback merge through both the solved graph state and fallback diagnostics.

These are baseline observations, not desired final semantics. Update the relevant assertions as the result/status and conflict behavior is implemented, while retaining the geometry and solver-path regressions.

## Phase 2 — Define public solve and diagnostic results

### Work

- Add public status/result types for overall solve state, per-element state, degrees of freedom when available, and diagnostics.
- Distinguish fully constrained, under-constrained, over-constrained/inconsistent, and non-convergent outcomes.
- Include stable element and constraint IDs, constraint kind/value, and referenced element IDs in diagnostics.
- Add a reporting solve method while retaining `Solve() error` as a compatibility wrapper during migration.
- Expose public status and constraint types rather than requiring users to depend on types from `internal` packages.
- Provide stable query methods for element state, constraint details, and conflict details. Keep legacy APIs as wrappers where practical.

### Exit criteria

- Callers can obtain structured solve results without parsing errors or strings.
- Existing users of `Solve() error` continue to compile during the transition.
- Public documentation identifies which diagnostic fields are guaranteed and which are estimates.

### Initial implementation

The first reporting API is now available through `Sketch.SolveWithReport()` and `Sketch.LastSolveReport()`. It provides public solve and element status enums, element/value snapshots, stable sketch-level constraint IDs, constraint snapshots, conflict membership, diagnostic codes, and an `Err()` compatibility conversion. Existing `Sketch.Solve() error` remains available.

This is the API foundation, not the completion of status analysis: degrees-of-freedom fields are present but marked unknown, and the current solver's under-constrained classification remains incomplete. `Solved` means the current solver accepted its constraints, not that the sketch is proven fully constrained. Status/conflict correctness remains Phase 4 work.

## Phase 3 — Standardize and validate constraint construction

### Work

- Standardize constraint-creation signatures and error handling; consistently return errors for invalid inputs.
- Validate nil elements, supported element combinations, element ownership by the target sketch, finite numeric values, and invalid/degenerate geometry before mutating sketch state.
- Make failure atomic: invalid constraint creation must not partially change element maps, graph constraints, or public constraint lists.
- Use sketch-owned builders or equivalent APIs to support ergonomic constraint creation without making elements depend on their owning sketch.
- Define the public constraint-handle/ID behavior and query APIs.
- Preserve current methods as deprecated or forwarding wrappers as appropriate.

### Exit criteria

- Invalid requests fail predictably and do not alter solver state.
- API behavior and units are documented and consistent across constraint methods.
- Tests cover cross-sketch references, nils, invalid types, NaN/Inf, and degeneracy.

## Phase 4 — Correct constraint status and conflict analysis

### Work

- Specify status semantics for compound elements and their child points, including how status aggregates for lines, circles, and arcs.
- Stop relying on optimistic element defaults; calculate under-constrained status from the completed solve state.
- Track degrees of freedom where the graph solver can establish them; for numeric cases, evaluate an appropriate rank/Jacobian-based estimate rather than treating optimizer convergence as proof of full constraint.
- Separate redundant constraint count from actual inconsistency.
- Replace heuristic conflict marking with verified conflict detection; return a useful reduced conflict set where feasible, and document if it is not guaranteed minimal.
- Ensure statuses and diagnostics remain meaningful when a solve fails partway or partially solves a sketch.

### Exit criteria

- UI-facing status is correct for the Phase 1 fixtures.
- Redundant-but-consistent constraints are not incorrectly reported as inconsistent.
- Conflict output refers only to constraints in the sketch and is stable across repeated runs.

## Phase 5 — Stabilize graph and numeric solving

### Work

- Make graph construction, cluster ordering, merge candidate selection, set iteration where order affects decisions, and diagnostic output deterministic by ordering IDs explicitly.
- Review solve-pass reset/build behavior and ensure repeated solves or added constraints do not retain stale cluster or solved state.
- Add configurable numeric solver options for tolerance and iteration limits, with documented defaults.
- Return numeric termination details, including final residual/error and optimizer termination reason, through the solve report.
- Test difficult initial positions, near-degenerate geometry, boundary cases, and fallback behavior for under-constrained versus inconsistent systems.
- Profile before optimizing; add performance tests/benchmarks for representative larger sketches if needed.

### Exit criteria

- Repeated identical inputs produce stable statuses, conflicts, and solve results within documented numeric tolerances.
- Re-solving after supported sketch mutations does not reuse invalid state.
- Numeric fallback behavior and failure modes are observable to callers.

## Phase 6 — Fill documented functionality gaps

### Work

- Reassess the planned equal-angle and symmetric constraints listed in `plans.md` and confirm scope/priorities.
- For each selected constraint, specify supported element kinds, units/sign conventions, graph-solver behavior, numeric fallback behavior, and diagnostic/reporting behavior.
- Add implementation and tests before advertising a constraint as supported.
- Audit README feature claims against actual constraint behavior and known limitations.

### Exit criteria

- Each implemented constraint has API documentation and positive, invalid-input, conflict, under-constrained, and integration coverage as applicable.
- Unsupported constraints are clearly marked as unsupported rather than implied by old planning notes.

## Phase 7 — Documentation and downstream integration

### Work

- Update the README's installation/import path and getting-started code to match the actual package API.
- Document units, angle conventions, arc direction, fixed geometry, solve lifecycle, status semantics, result fields, numeric tolerance, and conflict-report limitations.
- Add a MakerCAD-oriented example showing creation, solve, and retrieval of per-element state and conflicting constraints.
- Reconcile `TODO.md`, `plans.md`, and the test script with completed work and the verified test workflow; remove or label stale historical failure/hang notes.
- Document compatibility/deprecation policy for any API migration.

### Exit criteria

- A new user can build a sketch, solve it, and inspect UI-relevant status from documentation alone.
- Examples compile and run as part of a documented verification step.
- Planning and maintenance documents agree with the shipped functionality.

## Suggested implementation order

1. Phase 1 characterization tests.
2. Phase 2 result types and compatibility API.
3. Phase 3 input validation and consistent construction.
4. Phase 4 status and conflict correctness.
5. Phase 5 deterministic graph/numeric behavior.
6. Phase 6 selected constraint gaps.
7. Phase 7 documentation and MakerCAD integration.

The order is intentional: reliable tests and explicit result contracts should precede the solver changes that implement them. Avoid combining broad solver algorithm rewrites with public API changes in a single step.

## Ongoing verification

For each implementation change:

- Run focused tests for the changed behavior.
- Run `go test ./...` and `go vet ./...` from the Dlineate module before merging.
- Add a regression test for each corrected defect.
- Check representative MakerCAD integration callers when changing exported API behavior.
- Confirm generated/debug output is deterministic when the change affects graph ordering or diagnostics.
