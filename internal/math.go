package internal

import (
	"math"
)

// Add adds given floating point numbers and returns the sum.
func add(args ...Number) Number {
	var result Number
	for _, v := range args {
		result += v
	}
	return result
}

// Sub subtracts args from 'x' and returns the final result.
func sub(x Number, args ...Number) Number {
	result := x
	if len(args) == 0 {
		return -1 * result
	}

	for _, v := range args {
		result -= v
	}
	return result
}

// Multiply multiplies the given args to 1 and returns the result.
func multiply(first Number, args ...Number) Number {
	result := first
	for _, v := range args {
		result *= v
	}
	return result
}

// Divide returns the product of given numbers.
func divide(first Number, args ...Number) Number {

	result := first

	if len(args) == 0 {
		return 1 / first
	}

	for _, v := range args {
		result /= v
	}
	return result

}

// Lt returns true if the given args are monotonically increasing.
func lt(base Number, args ...Number) bool {
	inc := true
	for _, arg := range args {
		inc = inc && (arg > base)
	}
	return inc
}

// LtE returns true if the given args are monotonically increasing or
// are all equal.
func ltE(base Number, args ...Number) bool {
	inc := true
	for _, arg := range args {
		inc = inc && (arg >= base)
	}
	return inc
}

// Gt returns true if the given args are monotonically decreasing.
func gt(base Number, args ...Number) bool {
	inc := true
	for _, arg := range args {
		inc = inc && (arg < base)
	}
	return inc
}

// GtE returns true if the given args are monotonically decreasing or
// all equal.
func gtE(base Number, args ...Number) bool {
	inc := true
	for _, arg := range args {
		inc = inc && (arg <= base)
	}
	return inc
}

func isPrime(value Number) bool {
	for i := 2; i <= int(math.Floor(float64(value)/2)); i++ {
		if int(value)%i == 0 {
			return false
		}
	}
	return value > 1
}
