package main

import (
	"fmt"

	dlineation "github.com/marcuswu/dlineation/pkg"
)

func main() {
	sketch := dlineation.NewSketch()

	// Add elements
	offset := sketch.AddLine(0, 0, 6, 0)
	line1 := sketch.AddLine(6, 0, 9.4, 0)
	line2 := sketch.AddLine(9.4, 0, 9.4, -1)
	line3 := sketch.AddLine(9.4, -1, 8, -2)
	arc1 := sketch.AddArc(8.5, -2.5, 8, -2, 8, -3)
	line4 := sketch.AddLine(8, -3, 9.4, -3.8)
	line5 := sketch.AddLine(9.4, -3.8, 9.4, -8)
	line6 := sketch.AddLine(9.4, -8, 6, -8)
	line7 := sketch.AddLine(6, -8, 6, 0)
	arc2 := sketch.AddArc(6, 0, 6, 3, 8, -2)
	arc3 := sketch.AddArc(6, -5.21, 6, -8.21, 8, -3.0)

	// Add constraints
	// Bottom of pentagon starts at origin and aligns with x axis
	sketch.AddCoincidentConstraint(sketch.Origin, offset.Start())
	sketch.AddParallelConstraint(sketch.XAxis, offset)
	sketch.AddDistanceConstraint(offset, nil, 6)

	// line points are coincident
	sketch.AddCoincidentConstraint(offset.End(), line1.Start())
	sketch.AddCoincidentConstraint(line1.End(), line2.Start())
	sketch.AddCoincidentConstraint(line2.End(), line3.Start())
	sketch.AddCoincidentConstraint(line3.End(), arc1.Start())
	sketch.AddCoincidentConstraint(arc1.End(), line4.Start())
	sketch.AddCoincidentConstraint(line4.End(), line5.Start())
	sketch.AddCoincidentConstraint(line5.End(), line6.Start())
	sketch.AddCoincidentConstraint(line6.End(), line7.Start())
	sketch.AddCoincidentConstraint(line7.End(), line1.Start())

	// line1 constraints
	sketch.AddParallelConstraint(sketch.XAxis, line1)
	sketch.AddDistanceConstraint(line1, nil, 3.4)

	// line2 constraints
	sketch.AddParallelConstraint(sketch.YAxis, line2)
	sketch.AddAngleConstraint(line2, line3, 135)

	// line3 constraints
	sketch.AddAngleConstraint(line3, line4, 90)

	// arc1 constraints
	sketch.AddDistanceConstraint(arc1, nil, 0.5)
	sketch.AddDistanceConstraint(arc1.Center(), line7, 2.5)
	sketch.AddTangentConstraint(arc1, line3)
	sketch.AddTangentConstraint(arc1, line4)

	// line5 constraints
	sketch.AddParallelConstraint(sketch.YAxis, line5)

	// line6 constraints
	sketch.AddParallelConstraint(sketch.XAxis, line6)
	sketch.AddDistanceConstraint(line6, nil, 3.4)

	// line7 constraints
	sketch.AddParallelConstraint(sketch.YAxis, line7)
	sketch.AddDistanceConstraint(line7, nil, 8)

	// arc2 constraints
	sketch.AddCoincidentConstraint(arc2.Center(), line1.Start())
	sketch.AddCoincidentConstraint(arc2.Start(), line7)
	sketch.AddCoincidentConstraint(arc2.End(), line3)
	sketch.AddTangentConstraint(arc2, line3)

	// arc3 constraints
	sketch.AddCoincidentConstraint(arc3.Center(), line7)
	sketch.AddCoincidentConstraint(arc3.End(), line7)
	sketch.AddCoincidentConstraint(arc3.Start(), line4)
	sketch.AddTangentConstraint(arc3, line4)

	// Solve
	err := sketch.Solve()

	// Output results
	if err != nil {
		fmt.Printf("Solve error %s\n", err)
	}

	fmt.Printf("offset start constraint level %v\n", offset.Start().ConstraintLevel())
	fmt.Printf("offset end constraint level %v\n", offset.End().ConstraintLevel())
	fmt.Printf("l1 start constraint level %v\n", line1.Start().ConstraintLevel())
	fmt.Printf("l1 end constraint level %v\n", line1.End().ConstraintLevel())
	fmt.Printf("l2 start constraint level %v\n", line2.Start().ConstraintLevel())
	fmt.Printf("l2 end constraint level %v\n", line2.End().ConstraintLevel())
	fmt.Printf("l3 start constraint level %v\n", line3.Start().ConstraintLevel())
	fmt.Printf("l3 end constraint level %v\n", line3.End().ConstraintLevel())
	fmt.Printf("arc1 center constraint level %v\n", arc1.Center().ConstraintLevel())
	fmt.Printf("arc1 start constraint level %v\n", arc1.Start().ConstraintLevel())
	fmt.Printf("arc1 end constraint level %v\n", arc1.End().ConstraintLevel())
	fmt.Printf("l4 start constraint level %v\n", line4.Start().ConstraintLevel())
	fmt.Printf("l4 end constraint level %v\n", line4.End().ConstraintLevel())
	fmt.Printf("l5 start constraint level %v\n", line5.Start().ConstraintLevel())
	fmt.Printf("l5 end constraint level %v\n", line5.End().ConstraintLevel())
	fmt.Printf("l6 start constraint level %v\n", line6.Start().ConstraintLevel())
	fmt.Printf("l6 end constraint level %v\n", line6.End().ConstraintLevel())
	fmt.Printf("l7 start constraint level %v\n", line7.Start().ConstraintLevel())
	fmt.Printf("l7 end constraint level %v\n", line7.End().ConstraintLevel())
	fmt.Printf("l1 constraint level %v\n", line1.ConstraintLevel())
	fmt.Printf("l2 constraint level %v\n", line2.ConstraintLevel())
	fmt.Printf("l3 constraint level %v\n", line3.ConstraintLevel())
	fmt.Printf("l4 constraint level %v\n", line4.ConstraintLevel())
	fmt.Printf("l5 constraint level %v\n", line5.ConstraintLevel())
	fmt.Printf("l5 constraint level %v\n", line6.ConstraintLevel())
	fmt.Printf("l5 constraint level %v\n", line7.ConstraintLevel())

	// Export Image
	sketch.ExportImage("modifiedCylinderExample.svg")

	fmt.Println("Solved sketch: ")
	values := line1.Values(sketch)
	fmt.Printf("l1: %f, %f to %f, %f\n", values[0], values[1], values[2], values[3])
	values = line2.Values(sketch)
	fmt.Printf("l2: %f, %f to %f, %f\n", values[0], values[1], values[2], values[3])
	values = line3.Values(sketch)
	fmt.Printf("l3: %f, %f to %f, %f\n", values[0], values[1], values[2], values[3])
	values = arc1.Values(sketch)
	fmt.Printf("arc1: center %f, %f from %f, %f to %f, %f\n", values[0], values[1], values[2], values[3], values[4], values[5])
	values = line4.Values(sketch)
	fmt.Printf("l4: %f, %f to %f, %f\n", values[0], values[1], values[2], values[3])
	values = line5.Values(sketch)
	fmt.Printf("l5: %f, %f to %f, %f\n", values[0], values[1], values[2], values[3])
	values = line6.Values(sketch)
	fmt.Printf("l6: %f, %f to %f, %f\n", values[0], values[1], values[2], values[3])
	values = line7.Values(sketch)
	fmt.Printf("l7: %f, %f to %f, %f\n", values[0], values[1], values[2], values[3])
}
