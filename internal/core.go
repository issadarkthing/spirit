package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	seqStr       = reflect.TypeOf((*Seq)(nil)).Elem().Name()
	invokableStr = reflect.TypeOf((*Seq)(nil)).Elem().Name()
)

// Case implements the switch case construct.
func caseForm(scope Scope, args []Value) (Value, error) {

	argc := len(args)
	if argc < 2 {
		return nil, ArgumentError{
			Got: argc,
			Fn:  "case",
		}
	}

	res, err := Eval(scope, args[0])
	if err != nil {
		return nil, err
	}

	if len(args) == 2 {
		return Eval(scope, args[1])
	}

	start := 1
	for ; start < len(args); start += 2 {
		val := args[start]
		if start+1 >= len(args) {
			return val, nil
		}

		if Compare(res, val) {
			return Eval(scope, args[start+1])
		}
	}

	return nil, fmt.Errorf("no matching clause for '%s'", res)
}

// MacroExpand is a wrapper around the internal MacroExpand function that
// ignores the expanded bool flag.
func macroExpand(scope Scope, f Value) (Value, error) {
	f, _, err := MacroExpand(scope, f)
	return f, err
}

// Throw converts args to strings and returns an error with all the strings
// joined.
func throw(scope Scope, args ...Value) error {
	return errors.New(strings.Trim(makeString(args...).String(), "\""))
}

// Realize realizes a sequence by continuously calling First() and Next()
// until the sequence becomes nil.
func realize(seq Seq) *List {
	var vals []Value

	for seq != nil {
		v := seq.First()
		if v == nil {
			break
		}
		vals = append(vals, v)
		seq = seq.Next()
	}

	return &List{Values: vals}
}

// TypeOf returns the type information object for the given argument.
func typeOf(v interface{}) Value {
	return ValueOf(reflect.TypeOf(v))
}

// Implements checks if given value implements the interface represented
// by 't'. Returns error if 't' does not represent an interface type.
func implements(v interface{}, t Type) (bool, error) {
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
func toType(to Type, val Value) (Value, error) {

	// convert string to number
	if to == TypeOf(Number(0)) {
		if str, ok := val.(String); ok {
			fl, err := strconv.ParseFloat(string(str), 64)
			if err != nil {
				return nil, err
			}

			return Number(fl), nil
		}
	}

	rv := reflect.ValueOf(val)
	if rv.Type().ConvertibleTo(to.T) || rv.Type().AssignableTo(to.T) {
		return ValueOf(rv.Convert(to.T).Interface()), nil
	}

	return nil, fmt.Errorf("cannot convert '%s' to '%s'", rv.Type(), to.T)
}

// MakeString returns stringified version of all args.
func makeString(vals ...Value) Value {
	argc := len(vals)
	switch argc {
	case 0:
		return String("")

	case 1:
		nilVal := Nil{}
		if vals[0] == nilVal || vals[0] == nil {
			return String("")
		}

		return String(strings.Trim(vals[0].String(), "\""))

	default:
		var sb strings.Builder
		for _, v := range vals {
			sb.WriteString(strings.Trim(v.String(), "\""))
		}
		return String(sb.String())
	}
}

func doSeq(scope Scope, args []Value) (Value, error) {

	argc := len(args)
	if argc < 1 {
		return nil, ArgumentError{
			Got: argc,
			Fn:  "doseq",
		}
	}

	arg1 := args[0]
	// function arguments binding
	vecs, ok := arg1.(*Vector)
	if !ok {
		return nil, TypeError{
			Expected: &Vector{},
			Got:      arg1,
		}
	}

	argc = vecs.Size()
	if argc < 2 {
		return nil, ArgumentError{
			Got: argc,
			Fn:  "binding",
		}
	}

	coll, err := vecs.Index(1).Eval(scope)
	if err != nil {
		return nil, err
	}

	l, ok := coll.(Seq)
	if !ok {
		return nil, ImplementError{
			Name: seqStr,
			Val:  coll,
		}
	}

	symbol, ok := vecs.Index(0).(Symbol)
	var result Value

	for curr := l; curr != nil && curr.First() != nil; curr = curr.Next() {
		scope.Bind(symbol.Value, curr.First())
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
func swap(scope Scope, args []Value) (Value, error) {

	argc := len(args)
	if argc != 2 {
		return nil, ArgumentError{
			Got: argc,
			Fn:  "swap",
		}
	}

	symbol, ok := args[0].(Symbol)
	if !ok {
		return nil, TypeError{
			Expected: Symbol{},
			Got:      args[0],
		}
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
func recur(scope Scope, args []Value) (Value, error) {

	symbol := Symbol{
		Value: "recur",
	}

	results, err := EvalValueList(scope, args)
	if err != nil {
		return nil, err
	}

	results = append([]Value{symbol}, results...)
	return &List{Values: Values(results)}, nil
}

// Returns string representation of type
func stringTypeOf(v interface{}) string {
	return reflect.TypeOf(v).String()
}

// Evaluate the expressions in another goroutine; returns chan
func future(scope Scope, args []Value) (Value, error) {

	ch := &Future{
		Channel: make(chan Value),
		Value:   Nil{},
	}

	ch.Submit(scope, args[0])

	return ch, nil
}

type chanWrapper func(*Future) (Value, error)

// Deref chan from future to get the value. This call is blocking until future is resolved.
// The result will be cached.
func deref(scope Scope) chanWrapper {
	return func(ch *Future) (Value, error) {
		for {
			if ch.Realized {
				return ch.Value, nil
			}
		}
	}
}

func sleep(s int) {
	time.Sleep(time.Millisecond * time.Duration(s))
}

func futureRealize(ch *Future) bool {
	return ch.Realized
}

func xlispTime(scope Scope, args []Value) (Value, error) {

	var lastVal Value
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

func parseLoop(scope Scope, args []Value) (*Fn, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("call requires at-least bindings argument")
	}

	vec, isVector := args[0].(*Vector)
	if !isVector {
		return nil, fmt.Errorf(
			"first argument to let must be bindings vector, not %v",
			reflect.TypeOf(args[0]),
		)
	}

	if vec.Size()%2 != 0 {
		return nil, fmt.Errorf("bindings must contain even forms")
	}

	var bindings []binding
	for i := 0; i < vec.Size(); i += 2 {
		sym, isSymbol := vec.Index(i).(Symbol)
		if !isSymbol {
			return nil, fmt.Errorf(
				"item at %d must be symbol, not %s",
				i, vec.Index(i),
			)
		}

		bindings = append(bindings, binding{
			Name: sym.Value,
			Expr: vec.Index(i + 1),
		})
	}

	return &Fn{
		Func: func(scope Scope, _ []Value) (Value, error) {
			letScope := NewScope(scope)
			for _, b := range bindings {
				v, err := b.Expr.Eval(letScope)
				if err != nil {
					return nil, err
				}
				_ = letScope.Bind(b.Name, v)
			}

			result, err := Module(args[1:]).Eval(letScope)
			if err != nil {
				return nil, err
			}

			for isRecur(result) {

				newBindings := result.(*List).Values[1:]
				for i, b := range bindings {
					letScope.Bind(b.Name, newBindings[i])
				}

				result, err = Module(args[1:]).Eval(letScope)
				if err != nil {
					return nil, err
				}
			}

			return result, err
		},
	}, nil
}

func and(x Value, y Value) bool {
	return isTruthy(x) && isTruthy(y)
}

func or(x Value, y Value) bool {
	return isTruthy(x) || isTruthy(y)
}

func safeSwap(scope Scope, args []Value) (Value, error) {

	argc := len(args)
	if argc != 2 {
		return nil, ArgumentError{
			Got: argc,
			Fn:  "swap",
		}
	}

	args, err := EvalValueList(scope, args[:2])
	if err != nil {
		return nil, err
	}

	sym, ok := args[0].(*Atom)
	if !ok {
		return nil, TypeError{
			Expected: &Atom{},
			Got:      args[0],
		}
	}

	fn, ok := args[1].(Invokable)
	if !ok {
		return nil, TypeError{
			Expected: Invokable(nil),
			Got:      args[1],
		}
	}

	return sym.UpdateState(scope, fn.(Invokable))
}

func bound(scope Scope) func(Symbol) bool {
	return func(sym Symbol) bool {
		_, err := scope.Resolve(sym.Value)
		return err == nil
	}
}

func resolve(scope Scope) func(Symbol) Value {
	return func(sym Symbol) Value {
		val, _ := scope.Resolve(sym.Value)
		return val
	}
}

func keyword(str string) Keyword {
	return Keyword(str)
}

func assoc(hm Assoc, args ...Value) (Assoc, error) {

	if len(args)%2 != 0 {
		return nil, ArgumentError{
			Got: len(args),
			Fn:  "assoc",
		}
	}

	h := hm
	for i := 0; i < len(args); i += 2 {

		key := args[i]
		value := args[i+1]

		if object, ok := hm.(Object); ok {
			keyword, ok := key.(Keyword)
			if !ok {
				return nil, fmt.Errorf("object requires Keyword as key")
			}

			if !object.InstanceOf.Exists(keyword) {
				return nil, fmt.Errorf("cannot find member or method %s", string(keyword))
			}
		}

		if vec, ok := h.(*Vector); ok {

			index, ok := key.(Number)
			if !ok {
				return nil, TypeError{
					Expected: Number(0),
					Got:      key,
				}
			}

			if int(index) < 0 || int(index) > vec.Size()-1 {
				return nil, fmt.Errorf("vector out of bound")
			}
		}

		h = h.Set(key, value).(Assoc)
	}

	return h, nil
}

func parsejson(rawJson string) (*HashMap, error) {

	var data map[string]interface{}

	err := json.Unmarshal([]byte(rawJson), &data)
	if err != nil {
		return nil, err
	}

	return convert(data), nil
}

func convert(data map[string]interface{}) *HashMap {
	pm := NewHashMap()
	for k, v := range data {

		// nested object
		if nest, ok := v.(map[string]interface{}); ok {
			pm = pm.Set(Keyword(k), convert(nest)).(*HashMap)

			// key with array value
		} else if nestArr, ok := v.([]interface{}); ok {
			vals := make([]Value, 0, len(nestArr))

			for _, n := range nestArr {
				// nested object
				if nested, ok := n.(map[string]interface{}); ok {
					vals = append(vals, convert(nested))
				} else {
					vals = append(vals, ValueOf(n))
				}
			}
			pm = pm.Set(Keyword(k), ValueOf(vals)).(*HashMap)

			// others can simply use ValueOf
		} else {
			pm = pm.Set(Keyword(k), ValueOf(v)).(*HashMap)
		}
	}
	return pm
}

func apply(scope Scope, args []Value) (Value, error) {

	argc := len(args)
	if argc < 2 {
		return nil, ArgumentError{
			Got: argc,
			Fn:  "<>",
		}
	}

	evaledArgs, err := EvalValueList(scope, args)
	if err != nil {
		return nil, err
	}

	fn, ok := evaledArgs[0].(Invokable)
	if !ok {
		return nil, ImplementError{
			Name: invokableStr,
			Val:  evaledArgs[0],
		}
	}

	fnArgs := evaledArgs[1 : len(evaledArgs)-1]

	lastArg := evaledArgs[len(evaledArgs)-1]
	coll, ok := lastArg.(Seq)
	if !ok {
		return nil, ImplementError{
			Name: seqStr,
			Val: lastArg,
		}
	}

	if coll.Size() != 0 {
		for it := coll; it != nil; it = it.Next() {
			fnArgs = append(fnArgs, it.First())
		}
	}

	val, err := fn.Invoke(scope, fnArgs...)
	if err != nil {
		return nil, err
	}

	return val, nil
}

func eval(scope Scope, args []Value) (Value, error) {
	form, err := args[0].Eval(scope)
	if err != nil {
		return nil, err
	}
	return form.Eval(scope)
}

func evalStr(scope Scope, args []Value) (Value, error) {
	form, err := args[0].Eval(scope)
	if err != nil {
		return nil, err
	}

	fromStr, ok := form.(String)
	if !ok {
		return nil, TypeError{
			Expected: String(""),
			Got:      form,
		}
	}

	return ReadEvalStr(scope, string(fromStr))
}

func lazyRange(min, max, step int) LazySeq {
	return LazySeq{
		Min:  min,
		Max:  max,
		Step: step,
	}
}

func source(scope Scope) func(string) (Value, error) {
	return func(file string) (Value, error) {
		content, err := readFile(file)
		if err != nil {
			return nil, err
		}

		value, err := ReadEvalStr(scope, content)
		if err != nil {
			return nil, err
		}

		return value, nil
	}
}

func defClass(scope Scope, args []Value) (Value, error) {

	argc := len(args)
	if argc < 2 {
		return nil, ArgumentError{
			Fn:  "defclass",
			Got: argc,
		}
	}

	name, ok := args[0].(Symbol)
	if !ok {
		return nil, TypeError{
			Expected: Symbol{},
			Got:      args[0],
		}
	}

	class := Class{
		Name: name.String(),
	}

	// identifies the position of the member declaration
	hashMapIndex := 1

	// handles inheritence
	if symbol, ok := args[1].(Symbol); ok {
		if symbol.Value != "<-" {
			return nil, fmt.Errorf("expecting hashMap or <- symbol")
		}

		arg2, err := args[2].Eval(scope)
		if err != nil {
			return nil, err
		}

		if parent, ok := arg2.(Class); ok {
			class.Parent = &parent
		}

		hashMapIndex = 3
	}

	evaledArgs, err := EvalValueList(scope, args[hashMapIndex:])
	if err != nil {
		return nil, err
	}

	// evaluate passed hash map
	hashMap, ok := evaledArgs[0].(*HashMap)
	if !ok {
		return nil, TypeError{
			Expected: &HashMap{},
			Got:      evaledArgs[0],
		}
	}

	// ensure all the keys in hashmap are keywords
	for it := hashMap.Data.Iterator(); it.HasElem(); it.Next() {
		key, _ := it.Elem()

		_, ok := key.(Keyword)
		if !ok {
			return nil, TypeError{
				Expected: Keyword(""),
				Got:      key.(Value),
			}
		}
	}

	class.Members = hashMap

	methods := NewHashMap()
	staticMethods := NewHashMap()

	for _, m := range evaledArgs[1:] {

		method, ok := m.(*List)
		if !ok {
			return nil, fmt.Errorf("expected defmethod")
		}

		methodType := method.First().String()
		name := Keyword(method.Next().First().String())
		var body Value = method.Next().Next().First()

		body, err = body.Eval(scope)
		if err != nil {
			return nil, err
		}

		fn, ok := body.(Invokable)
		if !ok {
			return nil, ImplementError{
				Name: invokableStr,
				Val: body,
			}
		}

		if methodType == "method" {
			methods = methods.Set(name, fn).(*HashMap)
		} else if methodType == "static" {
			staticMethods = staticMethods.Set(name, fn).(*HashMap)
		}
	}

	class.Methods = methods
	class.StaticsMethod = staticMethods

	// define in global variable
	if scope.Parent() != nil {
		scope = scope.Parent()
	}

	scope.Bind(class.Name, class)

	return class, nil
}

func forceGC() {
	runtime.GC()
}

func mem(scope Scope, args []Value) (Value, error) {

	var mem runtime.MemStats

	runtime.ReadMemStats(&mem)
	initialByte := mem.TotalAlloc

	val, err := EvalValueLast(scope, args)
	if err != nil {
		return nil, err
	}

	runtime.ReadMemStats(&mem)
	alloc := mem.TotalAlloc - initialByte
	fmt.Println("Total memory used: ", byteCount(alloc))

	return val, nil
}

func memory() {
	var memStat runtime.MemStats
	runtime.ReadMemStats(&memStat)
	fmt.Println(byteCount(memStat.HeapAlloc))
}

// converts byte to human readable format
func byteCount(b uint64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := unit, 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}
