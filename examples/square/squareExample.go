package main

import (
	"fmt"
	"os"

	dlineation "github.com/marcuswu/dlineation/pkg"
	"github.com/marcuswu/dlineation/utils"
	"github.com/rs/zerolog"
)

func main() {
	utils.Logger = utils.Logger.Level(zerolog.InfoLevel).Output(zerolog.ConsoleWriter{Out: os.Stderr})
	sketch := dlineation.NewSketch()
	// Add elements
	l1 := sketch.AddLine(0.1, -0.2, 1.1, 0.1)
	l2 := sketch.AddLine(1.01, 0.2, 1.1, 0.9)
	l3 := sketch.AddLine(1.1, 1.2, 0.1, 1.1)
	l4 := sketch.AddLine(-0.1, 1.2, 0.1, 0.1)

	// Add constraints
	sketch.AddCoincidentConstraint(sketch.Origin, l1.Start())
	sketch.AddParallelConstraint(sketch.XAxis, l1)
	sketch.AddCoincidentConstraint(l2.Start(), l1.End())
	sketch.AddCoincidentConstraint(l3.Start(), l2.End())
	sketch.AddCoincidentConstraint(l4.Start(), l3.End())
	sketch.AddCoincidentConstraint(l1.Start(), l4.End())
	sketch.AddPerpendicularConstraint(l1, l2)
	sketch.AddParallelConstraint(l1, l3)
	sketch.AddDistanceConstraint(l1, nil, 1.0)
	sketch.AddDistanceConstraint(l2, nil, 1.0)
	sketch.AddDistanceConstraint(l3, nil, 1.0)

	// Solve
	err := sketch.Solve()

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
	fmt.Printf("l1 constraint level %v\n", l1.ConstraintLevel())
	fmt.Printf("l2 constraint level %v\n", l2.ConstraintLevel())
	fmt.Printf("l3 constraint level %v\n", l3.ConstraintLevel())
	fmt.Printf("l4 constraint level %v\n", l4.ConstraintLevel())

	// Export Image
	sketch.ExportImage("squareExample.svg")

	fmt.Println("Solved sketch: ")
	values := l1.Values(sketch)
	fmt.Printf("l1: %f, %f to %f, %f\n", values[0], values[1], values[2], values[3])
	values = l2.Values(sketch)
	fmt.Printf("l2: %f, %f to %f, %f\n", values[0], values[1], values[2], values[3])
	values = l3.Values(sketch)
	fmt.Printf("l3: %f, %f to %f, %f\n", values[0], values[1], values[2], values[3])
	values = l4.Values(sketch)
	fmt.Printf("l4: %f, %f to %f, %f\n", values[0], values[1], values[2], values[3])
}
