package graph

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/marcuswu/dlineate/internal/constraint"
	el "github.com/marcuswu/dlineate/internal/element"
	"github.com/marcuswu/dlineate/internal/solver"
	"github.com/marcuswu/dlineate/utils"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

// This fixture records a case where four rigid triangle clusters form a ring.
// Adjacent triangles share only one point, so the ordinary pair/triple merge
// patterns cannot join the clusters and the numeric fallback is required.
func TestCharacterizeNumericMergeRing(t *testing.T) {
	previousLogger := utils.Logger
	defer func() { utils.Logger = previousLogger }()

	var logs bytes.Buffer
	utils.Logger = zerolog.New(&logs).Level(zerolog.DebugLevel)

	g := NewSketch()
	a := g.AddPoint(big.NewFloat(0), big.NewFloat(0))
	b := g.AddPoint(big.NewFloat(1), big.NewFloat(0))
	c := g.AddPoint(big.NewFloat(1), big.NewFloat(1))
	d := g.AddPoint(big.NewFloat(0), big.NewFloat(1))
	u0 := g.AddPoint(big.NewFloat(0.4), big.NewFloat(-0.5))
	u1 := g.AddPoint(big.NewFloat(1.5), big.NewFloat(0.4))
	u2 := g.AddPoint(big.NewFloat(0.6), big.NewFloat(1.5))
	u3 := g.AddPoint(big.NewFloat(-0.5), big.NewFloat(0.6))

	triangles := [][3]el.SketchElement{
		{a, b, u0},
		{b, c, u1},
		{c, d, u2},
		{d, a, u3},
	}
	for _, triangle := range triangles {
		for _, pair := range [][2]el.SketchElement{
			{triangle[0], triangle[1]},
			{triangle[1], triangle[2]},
			{triangle[2], triangle[0]},
		} {
			g.AddConstraint(constraint.Distance, pair[0], pair[1], pair[0].DistanceTo(pair[1]))
		}
	}

	g.ResetClusters()
	g.BuildClusters()
	require.Equal(t, solver.Solved, g.Solve())
	require.True(t, g.IsSolved())
	require.Contains(t, logs.String(), "Looking for numeric merge")
	require.Contains(t, logs.String(), "Numeric merge: solver completed")
	require.Contains(t, logs.String(), `"solved":true`)
}
