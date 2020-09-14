package xlisp

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/spy16/sabre"
)

// Case implements the switch case construct.
func Case(scope sabre.Scope, args []sabre.Value) (sabre.Value, error) {

	err := checkArityAtLeast(2, len(args))
	if err != nil {
		return nil, err
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
func ToType(to sabre.Type, val sabre.Value) (sabre.Value, error) {
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

	err := checkArityAtLeast(1, len(args))
	if err != nil {
		return nil, err
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

func slangRange(args ...int) (Any, error) {
	var result []sabre.Value

	switch len(args) {
	case 1:
		result = createRange(0, args[0], 1)
	case 2:
		result = createRange(args[0], args[1], 1)
	case 3:
		result = createRange(args[0], args[1], args[2])
	}

	return &sabre.List{Values: result}, nil
}


func createRange(min, max, step int) []sabre.Value {

	result := make([]sabre.Value, 0, max-min)
	for i := min; i < max; i += step {
		result = append(result, sabre.Int64(i))
	}
	return result
}

func doSeq(scope sabre.Scope, args []sabre.Value) (sabre.Value, error) {

	arg1 := args[0]
	vecs, ok := arg1.(sabre.Vector)
	if !ok {
		return nil, fmt.Errorf("Invalid type")
	}

	coll, err := vecs.Values[1].Eval(scope)
	if err != nil {
		return nil, err
	}

	l, ok := coll.(sabre.Seq)
	if !ok {
		return nil, fmt.Errorf("Invalid type")
	}

	list := Realize(l)

	symbol, ok := vecs.Values[0].(sabre.Symbol)
	if !ok {
		return nil, fmt.Errorf("invalid type; expected symbol")
	}

	for _, v := range list.Values {
		scope.Bind(symbol.Value, v)
		for _, body := range args[1:] {
			_, err := body.Eval(scope)
			if err != nil {
				return nil, err
			}
		}
	}

	return sabre.Nil{}, nil
}

func swap(scope sabre.Scope, args []sabre.Value) (sabre.Value, error) {

	err := checkArity(2, len(args))
	if err != nil {
		return nil, err
	}

	symbol, ok := args[0].(sabre.Symbol)
	if !ok {
		return nil, fmt.Errorf("Expected symbol")
	}

	value, err := args[1].Eval(scope)
	if err != nil {
		return nil, err
	}

	scope.Bind(symbol.Value, value)
	return value, nil
}

func recur(scope sabre.Scope, args []sabre.Value) (sabre.Value, error) {

	symbol := sabre.Symbol{
		Value: "recur",
	}

	results, err := evalValueList(scope, args)
	if err != nil {
		return nil, err
	}

	results = append([]sabre.Value{symbol}, results...)
	return &sabre.List{Values: results}, nil
}

// Evaluate the expressions in another goroutine; returns chan
func future(scope sabre.Scope, args []sabre.Value) (sabre.Value, error) {

	ch := make(chan sabre.Value)
	
	go func() {
		
		for i, v := range args {

			list, ok := v.(*sabre.List)
			if !ok {
				return
			}

			symbol, ok := list.First().(sabre.Symbol)
			if !ok {
				continue
			}

			fn, err := scope.Resolve(symbol.Value)
			if err != nil {
				continue
			}

			res, err := fn.(sabre.Invokable).Invoke(scope, list.Values[1:]...)
			if err != nil {
				continue
			}

			if i == len(args)-1 {
				ch <- res
				close(ch)
			}
		}
	}()

	return sabre.ValueOf(ch), nil
}

// Deref chan from future to get the value. This call is blocking until future is resolved.
// The result will be cached.
func deref(scope sabre.Scope) (func(sabre.Value, <-chan sabre.Value) (sabre.Value, error)) {

	return func(symbol sabre.Value, ch <-chan sabre.Value) (sabre.Value, error) {

		derefSymbol := fmt.Sprintf("__deref__%s__result__", symbol.String())

		value :=<-ch	
		if value != nil {
			scope.Bind(derefSymbol, value)
			return value, nil
		}

		value, err := scope.Resolve(derefSymbol)
		if err != nil {
			return nil, err
		}

		return value, nil
	}
}
