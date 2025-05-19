package main

import (
	"fmt"
	"os"

	"github.com/marcuswu/dlineate"
	"github.com/marcuswu/dlineate/utils"
	"github.com/rs/zerolog"
)

func main() {
	utils.Logger = utils.Logger.Level(zerolog.DebugLevel).Output(zerolog.ConsoleWriter{Out: os.Stderr})
	sketch := dlineate.NewSketch()

	// Add elements
	l1 := sketch.AddLine(0.0, 0.0, 3.13, 0.0)
	l2 := sketch.AddLine(3.13, 0.0, 5.14, 2.27)
	l3 := sketch.AddLine(5.14, 2.27, 2.28, 4.72)
	l4 := sketch.AddLine(2.28, 4.72, -1.04, 3.56)
	l5 := sketch.AddLine(-1.04, 3.56, 0.0, 0.0)

	// Add constraints
	// Bottom of pentagon starts at origin and aligns with x axis
	sketch.AddCoincidentConstraint(sketch.Origin, l1.Start())
	sketch.AddParallelConstraint(sketch.XAxis, l1)

	// line points are coincident
	sketch.AddCoincidentConstraint(l1.End(), l2.Start())
	sketch.AddCoincidentConstraint(l2.End(), l3.Start())
	sketch.AddCoincidentConstraint(l3.End(), l4.Start())
	sketch.AddCoincidentConstraint(l4.End(), l5.Start())
	sketch.AddCoincidentConstraint(l5.End(), l1.Start())

	// 108 degrees between lines (skip 2 to not over constrain)
	sketch.AddAngleConstraint(l2, l3, 108, true)
	sketch.AddAngleConstraint(l3, l4, 108, true)
	sketch.AddAngleConstraint(l4, l5, 108, true)

	// 4 unit length on lines (skip 1 to not over constrain)
	sketch.AddDistanceConstraint(l1, nil, 4.0)
	sketch.AddDistanceConstraint(l2, nil, 4.0)
	sketch.AddDistanceConstraint(l4, nil, 4.0)
	sketch.AddDistanceConstraint(l5, nil, 4.0)

	// Solve
	err := sketch.Solve()

	sketch.ExportGraphViz("pentagon.dot")

	// Output results
	if err != nil {
		fmt.Printf("Solve error %s\n", err)
	}

	fmt.Printf("l1 start constraint level %v\n", l1.Start().ConstraintLevel())
	fmt.Printf("l1 end constraint level %v\n", l1.End().ConstraintLevel())
	fmt.Printf("l2 start constraint level %v\n", l2.Start().ConstraintLevel())
	fmt.Printf("l2 end constraint level %v\n", l2.End().ConstraintLevel())
	fmt.Printf("l3 start constraint level %v\n", l3.Start().ConstraintLevel())
	fmt.Printf("l3 end constraint level %v\n", l3.End().ConstraintLevel())
	fmt.Printf("l4 start constraint level %v\n", l4.Start().ConstraintLevel())
	fmt.Printf("l4 end constraint level %v\n", l4.End().ConstraintLevel())
	fmt.Printf("l5 start constraint level %v\n", l5.Start().ConstraintLevel())
	fmt.Printf("l5 end constraint level %v\n", l5.End().ConstraintLevel())
	fmt.Printf("l1 constraint level %v\n", l1.ConstraintLevel())
	fmt.Printf("l2 constraint level %v\n", l2.ConstraintLevel())
	fmt.Printf("l3 constraint level %v\n", l3.ConstraintLevel())
	fmt.Printf("l4 constraint level %v\n", l4.ConstraintLevel())
	fmt.Printf("l5 constraint level %v\n", l5.ConstraintLevel())

	// Export Image
	sketch.ExportImage("pentagonExample.svg")

	fmt.Println("Solved sketch: ")
	values := l1.Values(sketch)
	fmt.Printf("l1: %f, %f to %f, %f\n", values[0], values[1], values[2], values[3])
	values = l2.Values(sketch)
	fmt.Printf("l2: %f, %f to %f, %f\n", values[0], values[1], values[2], values[3])
	values = l3.Values(sketch)
	fmt.Printf("l3: %f, %f to %f, %f\n", values[0], values[1], values[2], values[3])
	values = l4.Values(sketch)
	fmt.Printf("l4: %f, %f to %f, %f\n", values[0], values[1], values[2], values[3])
	values = l5.Values(sketch)
	fmt.Printf("l5: %f, %f to %f, %f\n", values[0], values[1], values[2], values[3])
}
