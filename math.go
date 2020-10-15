package spirit

import (
	"github.com/issadarkthing/spirit/internal"
)


// Add adds given floating point numbers and returns the sum.
func add(args ...internal.Number) internal.Number {
	var result internal.Number
	for _, v := range args {
		result += v
	}
	return result
}

// Sub subtracts args from 'x' and returns the final result.
func sub(x internal.Number, args ...internal.Number) internal.Number {
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
func multiply(first internal.Number, args ...internal.Number) internal.Number {
	result := first
	for _, v := range args {
		result *= v
	}
	return result
}

// Divide returns the product of given numbers.
func divide(first internal.Number, args ...internal.Number) internal.Number {

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
func lt(base internal.Number, args ...internal.Number) bool {
	inc := true
	for _, arg := range args {
		inc = inc && (arg > base)
	}
	return inc
}

// LtE returns true if the given args are monotonically increasing or
// are all equal.
func ltE(base internal.Number, args ...internal.Number) bool {
	inc := true
	for _, arg := range args {
		inc = inc && (arg >= base)
	}
	return inc
}

// Gt returns true if the given args are monotonically decreasing.
func gt(base internal.Number, args ...internal.Number) bool {
	inc := true
	for _, arg := range args {
		inc = inc && (arg < base)
	}
	return inc
}

// GtE returns true if the given args are monotonically decreasing or
// all equal.
func gtE(base internal.Number, args ...internal.Number) bool {
	inc := true
	for _, arg := range args {
		inc = inc && (arg <= base)
	}
	return inc
}
