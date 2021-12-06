- [x] dlineate
  - [x] examples
  - [x] internal
    - [x] constraint
      - [x] Constraint.go
    - [x] element
      - [x] SketchElement.go
      - [x] SketchLine.go
      - [x] SketchPoint.go
    - [x] solver
      - [x] Solver.go
    - [x] GraphCluster.go
    - [x] SketchGraph.go
  - [x] utils
    - [x] FloatCompare.go
    - [x] Set.go
  - [x] pkg
    - [x] Solver.go -- Main interface for external access, describes the Sketch type. Sketch constructor creates internal sketch instance
      - [x] SetWorkplane(...)
      - [-] SetOrigin(...)
      - [x] AddPoint(x, y)
      - [x] AddLine(PointRef, PointRef)
      - [x] AddCircle(PointRef, Radius)
      - [x] AddArc(PointRef, PointRef, PointRef)
      - [-] AddElement(Element)
      - [ ] AddConstraint(Constraint)
      - [ ] Solve()
      - [ ] Elements()
    - [x] Constraint.go -- Base Constraint interface
      - [x] Also defines base constraint functionality
      - [x] Constraints will own internal constraints and internal elements
    - [ ] XConstraint.go -- Implementations of different constraint types
    - [x] Element.go -- Base Element interface
      - [x] Also defines base element functionality
      - [x] Elements will own internal constraints and internal elements
      - [x] Elements can export their values
      Implementations of different element types

libMakerCad will
 * instantiate DLineate, setting workplace and origin
 * proxy element and constraint creation
 * execute the solve -- return error / dof info. Should solve as much as possible regardless of errors, dof
 * retrieve solved element data
 * create a face using results

I would like to be able to add constraints onto an element -- line.perpendicular(other)
To do that, a line would need to know about the sketch it belongs to... creating a circular reference.