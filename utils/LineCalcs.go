package utils

import "math/big"

func LineFromPoints(x1, y1, x2, y2 float64) (a, b, c float64) {
	a = y2 - y1 // y' - y
	b = x1 - x2 // x - x'
	// c = -ax - by from ax + by + c = 0
	c = (-a * x1) - (b * y1)
	return
}

func BigFloatLineFromPoints(x1, y1, x2, y2 float64) (a, b, c *big.Float) {
	af, bf, cf := LineFromPoints(x1, y1, x2, y2)

	a = big.NewFloat(af)
	b = big.NewFloat(bf)
	c = big.NewFloat(cf)
	return
}

func BigFloatLineFromBigPoints(x1, y1, x2, y2 *big.Float) (a, b, c *big.Float) {
	x1f, _ := x1.Float64()
	y1f, _ := y1.Float64()
	x2f, _ := x2.Float64()
	y2f, _ := y2.Float64()
	af, bf, cf := LineFromPoints(x1f, y1f, x2f, y2f)

	a = big.NewFloat(af)
	b = big.NewFloat(bf)
	c = big.NewFloat(cf)
	return
}
