package spirit

import (
	"github.com/issadarkthing/spirit/internal"
)

type any interface{}

// Add adds given floating point numbers and returns the sum.
func add(args ...any) any {
	switch args[0].(type) {
	case internal.Int64:
		var sum internal.Int64
		for _, a := range args {
			sum += a.(internal.Int64)
		}
		return sum
	case internal.Float64:
		var sum internal.Float64
		for _, a := range args {
			sum += a.(internal.Float64)
		}
		return sum
	default:
		return nil
	}
}

// Sub subtracts args from 'x' and returns the final result.
func sub(x any, args ...any) any {
	switch x.(type) {
	case internal.Float64:
		var result internal.Float64 = x.(internal.Float64)
		if len(args) == 0 {
			return -1 * x.(internal.Float64)
		}

		for _, a := range args {
			result -= a.(internal.Float64)
		}
		return result
	case internal.Int64:
		var result internal.Int64 = x.(internal.Int64)
		if len(args) == 0 {
			return -1 * x.(internal.Int64)
		}

		for _, a := range args {
			result -= a.(internal.Int64)
		}
		return result
	default:
		return nil
	}
}

// Multiply multiplies the given args to 1 and returns the result.
func multiply(first any, args ...any) any {
	switch args[0].(type) {
	case internal.Int64:
		result := first.(internal.Int64)
		for _, a := range args {
			result *= a.(internal.Int64)
		}
		return result
	case internal.Float64:
		result := first.(internal.Float64)
		for _, a := range args {
			result *= a.(internal.Float64)
		}
		return result
	default:
		return nil
	}
}

// Divide returns the product of given numbers.
func divide(first any, args ...any) any {

	switch first.(type) {
	case internal.Float64:
		var result internal.Float64 = first.(internal.Float64)

		if len(args) == 0 {
			return 1 / first.(internal.Float64)
		}

		for _, a := range args {
			result /= a.(internal.Float64)
		}
		return result

	case internal.Int64:
		var result internal.Int64 = first.(internal.Int64)

		if len(args) == 0 {
			return 1 / internal.Float64(first.(internal.Int64))
		}

		for _, a := range args {
			result /= a.(internal.Int64)
		}
		return result
	default:
		return nil
	}

}

// Lt returns true if the given args are monotonically increasing.
func lt(base float64, args ...float64) bool {
	inc := true
	for _, arg := range args {
		inc = inc && (arg > base)
	}
	return inc
}

// LtE returns true if the given args are monotonically increasing or
// are all equal.
func ltE(base float64, args ...float64) bool {
	inc := true
	for _, arg := range args {
		inc = inc && (arg >= base)
	}
	return inc
}

// Gt returns true if the given args are monotonically decreasing.
func gt(base float64, args ...float64) bool {
	inc := true
	for _, arg := range args {
		inc = inc && (arg < base)
	}
	return inc
}

// GtE returns true if the given args are monotonically decreasing or
// all equal.
func gtE(base float64, args ...float64) bool {
	inc := true
	for _, arg := range args {
		inc = inc && (arg <= base)
	}
	return inc
}
