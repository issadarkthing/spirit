package spirit

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/issadarkthing/spirit/internal"
)

// Case implements the switch case construct.
func caseForm(scope internal.Scope, args []internal.Value) (internal.Value, error) {

	err := checkArityAtLeast(2, len(args))
	if err != nil {
		return nil, err
	}

	res, err := internal.Eval(scope, args[0])
	if err != nil {
		return nil, err
	}

	if len(args) == 2 {
		return internal.Eval(scope, args[1])
	}

	start := 1
	for ; start < len(args); start += 2 {
		val := args[start]
		if start+1 >= len(args) {
			return val, nil
		}

		if internal.Compare(res, val) {
			return internal.Eval(scope, args[start+1])
		}
	}

	return nil, fmt.Errorf("no matching clause for '%s'", res)
}

// MacroExpand is a wrapper around the internal MacroExpand function that
// ignores the expanded bool flag.
func macroExpand(scope internal.Scope, f internal.Value) (internal.Value, error) {
	f, _, err := internal.MacroExpand(scope, f)
	return f, err
}

// Throw converts args to strings and returns an error with all the strings
// joined.
func throw(scope internal.Scope, args ...internal.Value) error {
	return errors.New(strings.Trim(makeString(args...).String(), "\""))
}

// Realize realizes a sequence by continuously calling First() and Next()
// until the sequence becomes nil.
func realize(seq internal.Seq) *internal.List {
	var vals []internal.Value

	for seq != nil {
		v := seq.First()
		if v == nil {
			break
		}
		vals = append(vals, v)
		seq = seq.Next()
	}

	return &internal.List{Values: vals}
}

// TypeOf returns the type information object for the given argument.
func typeOf(v interface{}) internal.Value {
	return internal.ValueOf(reflect.TypeOf(v))
}

// Implements checks if given value implements the interface represented
// by 't'. Returns error if 't' does not represent an interface type.
func implements(v interface{}, t internal.Type) (bool, error) {
	if t.T.Kind() == reflect.Ptr {
		t.T = t.T.Elem()
	}

	if t.T.Kind() != reflect.Interface {
		return false, fmt.Errorf("type '%s' is not an interface type", t)
	}

	return reflect.TypeOf(v).Implements(t.T), nil
}

// ToType attempts to convert given internal value to target type. Returns
// error if conversion not possible.
func toType(to internal.Type, val internal.Value) (internal.Value, error) {
	rv := reflect.ValueOf(val)
	if rv.Type().ConvertibleTo(to.T) || rv.Type().AssignableTo(to.T) {
		return internal.ValueOf(rv.Convert(to.T).Interface()), nil
	}

	return nil, fmt.Errorf("cannot convert '%s' to '%s'", rv.Type(), to.T)
}

// ThreadFirst threads the expressions through forms by inserting result of
// eval as first argument to next expr.
func threadFirst(scope internal.Scope, args []internal.Value) (internal.Value, error) {
	return threadCall(scope, args, false)
}

// ThreadLast threads the expressions through forms by inserting result of
// eval as last argument to next expr.
func threadLast(scope internal.Scope, args []internal.Value) (internal.Value, error) {
	return threadCall(scope, args, true)
}

// MakeString returns stringified version of all args.
func makeString(vals ...internal.Value) internal.Value {
	argc := len(vals)
	switch argc {
	case 0:
		return internal.String("")

	case 1:
		nilVal := internal.Nil{}
		if vals[0] == nilVal || vals[0] == nil {
			return internal.String("")
		}

		return internal.String(strings.Trim(vals[0].String(), "\""))

	default:
		var sb strings.Builder
		for _, v := range vals {
			sb.WriteString(strings.Trim(v.String(), "\""))
		}
		return internal.String(sb.String())
	}
}

func threadCall(scope internal.Scope, args []internal.Value, last bool) (internal.Value, error) {

	err := checkArityAtLeast(1, len(args))
	if err != nil {
		return nil, err
	}

	res := args[0]
	// res, err := internal.Eval(scope, args[0])
	// if err != nil {
	// 	return nil, err
	// }

	for args = args[1:]; len(args) > 0; args = args[1:] {
		form := args[0]

		switch f := form.(type) {
		case *internal.List:
			if last {
				f.Values = append(f.Values, res)
			} else {
				f.Values = append([]internal.Value{f.Values[0], res}, f.Values[1:]...)
			}
			res, err = internal.Eval(scope, f)
			if v, ok := res.(*internal.List); ok {
				res = v.Cons(internal.Symbol{Value: "list"})
			}

		case internal.Invokable:
			res, err = f.Invoke(scope, res)

		default:
			return nil, fmt.Errorf("%s is not invokable", reflect.TypeOf(res))
		}

		if err != nil {
			return nil, err
		}
	}

	if res, ok := res.(*internal.List); ok {
		return res.Eval(scope)
	}

	return res, nil
}

func isTruthy(v internal.Value) bool {
	if v == nil || v == (internal.Nil{}) {
		return false
	}

	if b, ok := v.(internal.Bool); ok {
		return bool(b)
	}

	return true
}

func slangRange(args ...int) (*internal.List, error) {
	var result []internal.Value

	switch len(args) {
	case 1:
		result = createRange(0, args[0], 1)
	case 2:
		result = createRange(args[0], args[1], 1)
	case 3:
		result = createRange(args[0], args[1], args[2])
	}

	return &internal.List{Values: result}, nil
}

func createRange(min, max, step int) []internal.Value {

	result := make([]internal.Value, 0, max-min)
	for i := min; i < max; i += step {
		result = append(result, internal.Number(i))
	}
	return result
}

func doSeq(scope internal.Scope, args []internal.Value) (internal.Value, error) {

	arg1 := args[0]
	vecs, ok := arg1.(*internal.PersistentVector)
	if !ok {
		return nil, invalidType(internal.NewPersistentVector(), arg1)
	}

	coll, err := vecs.Index(1).Eval(scope)
	if err != nil {
		return nil, err
	}

	l, ok := coll.(internal.Seq)
	if !ok {
		return nil, doesNotImplementSeq(l)
	}

	symbol, ok := vecs.Index(0).(internal.Symbol)
	var result internal.Value

	list := realize(l)
	for _, v := range list.Values {
		scope.Bind(symbol.Value, v)
		for _, body := range args[1:] {
			result, err = body.Eval(scope)
			if err != nil {
				return nil, err
			}
		}
	}

	return result, nil
}

// unsafely swap the value. Does not mutate the value rather just swapping
func swap(scope internal.Scope, args []internal.Value) (internal.Value, error) {

	err := checkArity(2, len(args))
	if err != nil {
		return nil, err
	}

	symbol, ok := args[0].(internal.Symbol)
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

// Returns '(recur & expressions) so it'll be recognize by fn.Invoke method as
// tail recursive function
func recur(scope internal.Scope, args []internal.Value) (internal.Value, error) {

	symbol := internal.Symbol{
		Value: "recur",
	}

	results, err := evalValueList(scope, args)
	if err != nil {
		return nil, err
	}

	results = append([]internal.Value{symbol}, results...)
	return &internal.List{Values: results}, nil
}

// Returns string representation of type
func stringTypeOf(v interface{}) string {
	return reflect.TypeOf(v).String()
}

// Evaluate the expressions in another goroutine; returns chan
func future(scope internal.Scope, args []internal.Value) (internal.Value, error) {

	ch := make(chan internal.Value)

	go func() {

		val, err := args[0].Eval(scope)
		if err != nil {
			panic(err)
		}

		ch <- val
		close(ch)
	}()

	return internal.ValueOf(ch), nil
}

type chanWrapper func(internal.Symbol, <-chan internal.Value) (internal.Value, error)

// Deref chan from future to get the value. This call is blocking until future is resolved.
// The result will be cached.
func deref(scope internal.Scope) chanWrapper {

	return func(symbol internal.Symbol, ch <-chan internal.Value) (internal.Value, error) {

		derefSymbol := fmt.Sprintf("__deref__%s__result__", symbol.Value)

		value, ok := <-ch
		if ok {
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

func sleep(s int) {
	time.Sleep(time.Millisecond * time.Duration(s))
}

func futureRealize(ch <-chan internal.Value) bool {
	select {
	case _, ok := <-ch:
		return !ok
	default:
		return false
	}
}

func xlispTime(scope internal.Scope, args []internal.Value) (internal.Value, error) {

	var lastVal internal.Value
	var err error
	initial := time.Now()
	for _, v := range args {
		lastVal, err = v.Eval(scope)
		if err != nil {
			return nil, err
		}
	}
	final := time.Since(initial)
	fmt.Printf("Elapsed time: %s\n", final.String())

	return lastVal, nil
}

func parseLoop(scope internal.Scope, args []internal.Value) (*internal.Fn, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("call requires at-least bindings argument")
	}

	vec, isVector := args[0].(*internal.PersistentVector)
	if !isVector {
		return nil, fmt.Errorf(
			"first argument to let must be bindings vector, not %v",
			reflect.TypeOf(args[0]),
		)
	}

	if vec.Size()%2 != 0 {
		return nil, fmt.Errorf("bindings must contain event forms")
	}

	var bindings []binding
	for i := 0; i < vec.Size(); i += 2 {
		sym, isSymbol := vec.Index(i).(internal.Symbol)
		if !isSymbol {
			return nil, fmt.Errorf(
				"item at %d must be symbol, not %s",
				i, vec.Index(i),
			)
		}

		bindings = append(bindings, binding{
			Name: sym.Value,
			Expr: vec.Index(i+1),
		})
	}

	return &internal.Fn{
		Func: func(scope internal.Scope, _ []internal.Value) (internal.Value, error) {
			letScope := internal.NewScope(scope)
			for _, b := range bindings {
				v, err := b.Expr.Eval(letScope)
				if err != nil {
					return nil, err
				}
				_ = letScope.Bind(b.Name, v)
			}

			result, err := internal.Module(args[1:]).Eval(letScope)
			if err != nil {
				return nil, err
			}

			for isRecur(result) {

				newBindings := result.(*internal.List).Values[1:]
				for i, b := range bindings {
					letScope.Bind(b.Name, newBindings[i])
				}

				result, err = internal.Module(args[1:]).Eval(letScope)
				if err != nil {
					return nil, err
				}
			}

			return result, err
		},
	}, nil
}

func isRecur(value internal.Value) bool {

	list, ok := value.(*internal.List)
	if !ok {
		return false
	}

	sym, ok := list.First().(internal.Symbol)
	if !ok {
		return false
	}

	if sym.Value != "recur" {
		return false
	}

	return true
}

type binding struct {
	Name string
	Expr internal.Value
}

func and(x internal.Value, y internal.Value) bool {
	return isTruthy(x) && isTruthy(y)
}

func or(x internal.Value, y internal.Value) bool {
	return isTruthy(x) || isTruthy(y)
}

func safeSwap(scope internal.Scope, args []internal.Value) (internal.Value, error) {

	err := checkArity(2, len(args))
	if err != nil {
		return nil, err
	}

	atom := args[0]
	atom, err = atom.Eval(scope)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve symbol")
	}

	fn, err := args[1].Eval(scope)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve symbol")
	}

	return atom.(*internal.Atom).UpdateState(scope, fn.(internal.Invokable))
}

func bound(scope internal.Scope) func(internal.Symbol) bool {
	return func(sym internal.Symbol) bool {
		_, err := scope.Resolve(sym.Value)
		return err == nil
	}
}

func resolve(scope internal.Scope) func(internal.Symbol) internal.Value {
	return func(sym internal.Symbol) internal.Value {
		val, _ := scope.Resolve(sym.Value)
		return val
	}
}

func splitString(str, sep internal.String) *internal.List {
	result := strings.Split(string(str), string(sep))
	values := make([]internal.Value, 0, len(result))
	for _, v := range result {
		values = append(values, internal.String(v))
	}
	return &internal.List{Values: values}
}

func keyword(str string) internal.Keyword {
	return internal.Keyword(str)
}

func assoc(hm *internal.PersistentMap, args ...internal.Value) (*internal.PersistentMap, error) {

	if len(args)%2 != 0 {
		return nil, fmt.Errorf("invalid number of arguments passed")
	}

	h := hm
	for i := 0; i < len(args); i += 2 {
		h = h.Set(args[i], args[i+1])
	}

	return h, nil
}

func parsejson(rawJson string) (*internal.PersistentMap, error) {

	var data map[string]interface{}

	err := json.Unmarshal([]byte(rawJson), &data)
	if err != nil {
		return nil, err
	}

	return convert(data), nil
}

func convert(data map[string]interface{}) *internal.PersistentMap {
	pm := internal.NewPersistentMap()
	for k, v := range data {
		if nest, ok := v.(map[string]interface{}); ok {
			pm = pm.Set(internal.Keyword(k), convert(nest))
		} else {
			pm = pm.Set(internal.Keyword(k), internal.ValueOf(v))
		}
	}
	return pm
}
