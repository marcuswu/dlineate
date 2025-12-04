package utils

import (
	"math"
	"math/big"
)

// StandardCompare is the tolerance for a standard float64 comparison
const FloatPrecision = 200
const StandardBigCompare = 0.0000001
const StandardCompare = 0.00001
const MaxNumericIterations = 10000

// BigFloatCompare returns 0 if floats are equal, -1 if a < b, 1 if a > b using a tolerance
func BigFloatCompare(a *big.Float, b *big.Float, tolerance float64) int {
	var val, posVal, tol, negTol big.Float
	if a.IsInf() || b.IsInf() {
		return a.Cmp(b)
	}
	val.Sub(a, b)
	posVal.Abs(&val)
	tol.SetFloat64(tolerance)
	negTol.Neg(&tol)
	if posVal.Cmp(&tol) < 0 {
		return 0
	}

	if val.Cmp(&negTol) < 0 {
		return -1
	}
	return 1
}

// StandardBigFloatCompare returns 0 if floats are equal, -1 if a < b, 1 if a > b using a tolerance
func StandardBigFloatCompare(a *big.Float, b *big.Float) int {
	return BigFloatCompare(a, b, StandardBigCompare)
}

// FloatCompare returns 0 if floats are equal, -1 if a < b, 1 if a > b using a tolerance
func FloatCompare(a float64, b float64, tol float64) int {
	if math.Abs(a-b) < tol {
		return 0
	}
	if a-b < 0-tol {
		return -1
	}
	return 1
}

// StandardFloatCompare returns 0 if floats are equal, -1 if a < b, 1 if a > b using a tolerance
func StandardFloatCompare(a float64, b float64) int {
	return FloatCompare(a, b, StandardCompare)
}
