- [x] Return details about under / over constrained elements
- [x] Rework pkg/Solver to use maps to handle findConstraint(s)
- [x] Add SVG export functionality to visualize input and output
- [-] Build out some core examples / tests
  - [ ] Fix square example -- internal solver tests succeed, but using pkg/ fails... something wrong with interface
        Or something is wrong with internal that is exposed via how the interface is using it
- [x] Load element values from solver when solved
- [x] External interface should take two passes. One to solve independent constraints. One to solve dependent.
- [x] Ratio & equality lengths are dependent constraints
