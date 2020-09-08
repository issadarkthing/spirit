package slang

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/spy16/sabre"
)

// Case implements the switch case construct.
func Case(scope sabre.Scope, args []sabre.Value) (sabre.Value, error) {
	if len(args) < 2 {
		return nil, errors.New("case requires at-least 2 args")
	}

	res, err := sabre.Eval(scope, args[0])
	if err != nil {
		return nil, err
	}

	if len(args) == 2 {
		return sabre.Eval(scope, args[1])
	}

	start := 1
	for ; start < len(args); start += 2 {
		val := args[start]
		if start+1 >= len(args) {
			return val, nil
		}

		if sabre.Compare(res, val) {
			return sabre.Eval(scope, args[start+1])
		}
	}

	return nil, fmt.Errorf("no matching clause for '%s'", res)
}

// MacroExpand is a wrapper around the sabre MacroExpand function that
// ignores the expanded bool flag.
func MacroExpand(scope sabre.Scope, f sabre.Value) (sabre.Value, error) {
	f, _, err := sabre.MacroExpand(scope, f)
	return f, err
}

// Throw converts args to strings and returns an error with all the strings
// joined.
func Throw(scope sabre.Scope, args ...sabre.Value) error {
	return errors.New(strings.Trim(MakeString(args...).String(), "\""))
}

// Realize realizes a sequence by continuously calling First() and Next()
// until the sequence becomes nil.
func Realize(seq sabre.Seq) *sabre.List {
	var vals []sabre.Value

	for seq != nil {
		v := seq.First()
		if v == nil {
			break
		}
		vals = append(vals, v)
		seq = seq.Next()
	}

	return &sabre.List{Values: vals}
}

// TypeOf returns the type information object for the given argument.
func TypeOf(v interface{}) sabre.Value {
	return sabre.ValueOf(reflect.TypeOf(v))
}

// Implements checks if given value implements the interface represented
// by 't'. Returns error if 't' does not represent an interface type.
func Implements(v interface{}, t sabre.Type) (bool, error) {
	if t.T.Kind() == reflect.Ptr {
		t.T = t.T.Elem()
	}

	if t.T.Kind() != reflect.Interface {
		return false, fmt.Errorf("type '%s' is not an interface type", t)
	}

	return reflect.TypeOf(v).Implements(t.T), nil
}

// ToType attempts to convert given sabre value to target type. Returns
// error if conversion not possible.
func ToType(val sabre.Value, to sabre.Type) (sabre.Value, error) {
	rv := reflect.ValueOf(val)
	if rv.Type().ConvertibleTo(to.T) || rv.Type().AssignableTo(to.T) {
		return sabre.ValueOf(rv.Convert(to.T).Interface()), nil
	}

	return nil, fmt.Errorf("cannot convert '%s' to '%s'", rv.Type(), to.T)
}

// ThreadFirst threads the expressions through forms by inserting result of
// eval as first argument to next expr.
func ThreadFirst(scope sabre.Scope, args []sabre.Value) (sabre.Value, error) {
	return threadCall(scope, args, false)
}

// ThreadLast threads the expressions through forms by inserting result of
// eval as last argument to next expr.
func ThreadLast(scope sabre.Scope, args []sabre.Value) (sabre.Value, error) {
	return threadCall(scope, args, true)
}

// MakeString returns stringified version of all args.
func MakeString(vals ...sabre.Value) sabre.Value {
	argc := len(vals)
	switch argc {
	case 0:
		return sabre.String("")

	case 1:
		nilVal := sabre.Nil{}
		if vals[0] == nilVal || vals[0] == nil {
			return sabre.String("")
		}

		return sabre.String(strings.Trim(vals[0].String(), "\""))

	default:
		var sb strings.Builder
		for _, v := range vals {
			sb.WriteString(strings.Trim(v.String(), "\""))
		}
		return sabre.String(sb.String())
	}
}

func threadCall(scope sabre.Scope, args []sabre.Value, last bool) (sabre.Value, error) {
	if len(args) == 0 {
		return nil, errors.New("at-least 1 argument required")
	}

	res, err := sabre.Eval(scope, args[0])
	if err != nil {
		return nil, err
	}

	for args = args[1:]; len(args) > 0; args = args[1:] {
		form := args[0]

		switch f := form.(type) {
		case *sabre.List:
			if last {
				f.Values = append(f.Values, res)
			} else {
				f.Values = append([]sabre.Value{f.Values[0], res}, f.Values[1:]...)
			}
			res, err = sabre.Eval(scope, f)

		case sabre.Invokable:
			res, err = f.Invoke(scope, res)

		default:
			return nil, fmt.Errorf("%s is not invokable", reflect.TypeOf(res))
		}

		if err != nil {
			return nil, err
		}
	}

	return res, nil
}

func isTruthy(v sabre.Value) bool {
	if v == nil || v == (sabre.Nil{}) {
		return false
	}

	if b, ok := v.(sabre.Bool); ok {
		return bool(b)
	}

	return true
}

func slangRange(min, max Any) (Any, error) {

	var imin, imax sabre.Int64
	result := make([]sabre.Value, 0, imax-imin)

	switch min.(type) {
	case sabre.Float64:
		imin = sabre.Int64(min.(sabre.Float64))
		imax = sabre.Int64(max.(sabre.Float64))
	case sabre.Int64:
		imin = min.(sabre.Int64)
		imax = max.(sabre.Int64)
	default:
		return nil, fmt.Errorf("Invalid type (%T, %T)", min, max)
	}

	for i := imin; i < imax; i++ {
		result = append(result, i)
	}

	return &sabre.List{Values: result}, nil
}

func slangMap(scope sabre.Scope, args []sabre.Value) (sabre.Value, error) {

	if len(args) < 2 {
		return nil, fmt.Errorf(
			"invalid number of argument; expected (%d) got (%d)",
			2, len(args),
		)
	}

	fn, err := sabre.Eval(scope, args[0])
	if err != nil {
		return nil, err
	}

	list, err := sabre.Eval(scope, args[1])
	if err != nil {
		return nil, err
	}

	switch list.(type) {
	case *sabre.List:

		result := make([]sabre.Value, 0, len(list.(*sabre.List).Values))
		for _, v := range list.(*sabre.List).Values {

			applied, err := fn.(sabre.MultiFn).Invoke(scope, v)
			if err != nil {
				return nil, err
			}

			result = append(result, applied)
		}
		return &sabre.List{Values: result}, nil

	case sabre.Vector:

		result := make([]sabre.Value, 0, len(list.(sabre.Vector).Values))
		for _, v := range list.(sabre.Vector).Values {

			applied, err := fn.(sabre.MultiFn).Invoke(scope, v)
			if err != nil {
				return nil, err
			}

			result = append(result, applied)
		}
		return &sabre.Vector{Values: result}, nil

	default:
		return nil, fmt.Errorf("Expected Seq instead got %T", list)
	}

}

func traverse(seq sabre.Seq, callBack func(val sabre.Value)) {

	curr := seq.First()
	if curr == nil {
		return
	}

	for curr != nil {
		callBack(curr)
		seq = seq.Next()
		curr = seq.First()
	}
}
