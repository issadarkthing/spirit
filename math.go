package xlisp

import (
	"github.com/spy16/sabre"
)

type Any interface{}

// Add adds given floating point numbers and returns the sum.
func Add(args ...Any) Any {
	switch args[0].(type) {
	case sabre.Int64:
		var sum sabre.Int64
		for _, a := range args {
			sum += a.(sabre.Int64)
		}
		return sum
	case sabre.Float64:
		var sum sabre.Float64
		for _, a := range args {
			sum += a.(sabre.Float64)
		}
		return sum
	default:
		return nil
	}
}

// Sub subtracts args from 'x' and returns the final result.
func Sub(x Any, args ...Any) Any {
	switch x.(type) {
	case sabre.Float64:
		var result sabre.Float64 = x.(sabre.Float64)
		if len(args) == 0 {
			return -1 * x.(sabre.Float64)
		}

		for _, a := range args {
			result -= a.(sabre.Float64)
		}
		return result
	case sabre.Int64:
		var result sabre.Int64 = x.(sabre.Int64)
		if len(args) == 0 {
			return -1 * x.(sabre.Int64)
		}

		for _, a := range args {
			result -= a.(sabre.Int64)
		}
		return result
	default:
		return nil
	}
}

// Multiply multiplies the given args to 1 and returns the result.
func Multiply(first Any, args ...Any) Any {
	switch args[0].(type) {
	case sabre.Int64:
		result := first.(sabre.Int64)
		for _, a := range args {
			result *= a.(sabre.Int64)
		}
		return result
	case sabre.Float64:
		result := first.(sabre.Float64)
		for _, a := range args {
			result *= a.(sabre.Float64)
		}
		return result
	default:
		return nil
	}
}

// Divide returns the product of given numbers.
func Divide(first Any, args ...Any) Any {

	switch first.(type) {
	case sabre.Float64:
		var result sabre.Float64 = first.(sabre.Float64)

		if len(args) == 0 {
			return 1 / first.(sabre.Float64)
		}

		for _, a := range args {
			result /= a.(sabre.Float64)
		}
		return result

	case sabre.Int64:
		var result sabre.Int64 = first.(sabre.Int64)

		if len(args) == 0 {
			return 1 / sabre.Float64(first.(sabre.Int64))
		}

		for _, a := range args {
			result /= a.(sabre.Int64)
		}
		return result
	default:
		return nil
	}

}

// Lt returns true if the given args are monotonically increasing.
func Lt(base float64, args ...float64) bool {
	inc := true
	for _, arg := range args {
		inc = inc && (arg > base)
	}
	return inc
}

// LtE returns true if the given args are monotonically increasing or
// are all equal.
func LtE(base float64, args ...float64) bool {
	inc := true
	for _, arg := range args {
		inc = inc && (arg >= base)
	}
	return inc
}

// Gt returns true if the given args are monotonically decreasing.
func Gt(base float64, args ...float64) bool {
	inc := true
	for _, arg := range args {
		inc = inc && (arg < base)
	}
	return inc
}

// GtE returns true if the given args are monotonically decreasing or
// all equal.
func GtE(base float64, args ...float64) bool {
	inc := true
	for _, arg := range args {
		inc = inc && (arg <= base)
	}
	return inc
}
