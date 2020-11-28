package spirit

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/issadarkthing/spirit/internal"
)

// Case implements the switch case construct.
func caseForm(scope internal.Scope, args []internal.Value) (internal.Value, error) {

	argc := len(args)
	if argc < 2 {
		return nil, internal.ErrWrongArgumentCount(argc, "case")
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

	// convert string to number
	if to == internal.TypeOf(internal.Number(0)) {
		if str, ok := val.(internal.String); ok {
			fl, err := strconv.ParseFloat(string(str), 64)
			if err != nil {
				return nil, err
			}

			return internal.Number(fl), nil
		}
	}

	rv := reflect.ValueOf(val)
	if rv.Type().ConvertibleTo(to.T) || rv.Type().AssignableTo(to.T) {
		return internal.ValueOf(rv.Convert(to.T).Interface()), nil
	}

	return nil, fmt.Errorf("cannot convert '%s' to '%s'", rv.Type(), to.T)
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



func isTruthy(v internal.Value) bool {
	if v == nil || v == (internal.Nil{}) {
		return false
	}

	if b, ok := v.(internal.Bool); ok {
		return bool(b)
	}

	return true
}


func doSeq(scope internal.Scope, args []internal.Value) (internal.Value, error) {

	argc := len(args)
	if argc < 1 {
		return nil, internal.ErrWrongArgumentCount(argc, "doseq")
	}

	arg1 := args[0]
	// function arguments binding
	vecs, ok := arg1.(*internal.PersistentVector)
	if !ok {
		return nil, invalidType(internal.NewPersistentVector(), arg1)
	}

	argc = vecs.Size()
	if argc < 2 {
		return nil, internal.ErrWrongArgumentCount(argc, "doseq")
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
func swap(scope internal.Scope, args []internal.Value) (internal.Value, error) {

	argc := len(args)
	if argc != 2 {
		return nil, internal.ErrWrongArgumentCount(argc, "swap")
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

	results, err := internal.EvalValueList(scope, args)
	if err != nil {
		return nil, err
	}

	results = append([]internal.Value{symbol}, results...)
	return &internal.List{Values: internal.Values(results)}, nil
}

// Returns string representation of type
func stringTypeOf(v interface{}) string {
	return reflect.TypeOf(v).String()
}

// Evaluate the expressions in another goroutine; returns chan
func future(scope internal.Scope, args []internal.Value) (internal.Value, error) {

	ch := &internal.Future{
		Channel: make(chan internal.Value),
		Value: internal.Nil{},
	}

	ch.Submit(scope, args[0])

	return ch, nil
}

type chanWrapper func(*internal.Future) (internal.Value, error)

// Deref chan from future to get the value. This call is blocking until future is resolved.
// The result will be cached.
func deref(scope internal.Scope) chanWrapper {
	return func(ch *internal.Future) (internal.Value, error) {
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

func futureRealize(ch *internal.Future) bool {
	return ch.Realized
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
			Expr: vec.Index(i + 1),
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

	argc := len(args)
	if argc != 2 {
		return nil, internal.ErrWrongArgumentCount(argc, "swap")
	}

	atom := args[0]
	atom, err := atom.Eval(scope)
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

func assoc(hm internal.Assoc, args ...internal.Value) (internal.Assoc, error) {

	if len(args)%2 != 0 {
		return nil, fmt.Errorf("invalid number of arguments passed")
	}

	h := hm
	for i := 0; i < len(args); i += 2 {

		key := args[i]
		value := args[i+1]

		if object, ok := hm.(internal.Object); ok {
			keyword, ok := key.(internal.Keyword); 
			if !ok {
				return nil, fmt.Errorf("object requires Keyword as key")
			}

			if !object.InstanceOf.Exists(keyword) {
				return nil, fmt.Errorf("cannot find member or method %s", string(keyword))
			}
		}

		h = h.Set(key, value).(internal.Assoc)
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

		// nested object
		if nest, ok := v.(map[string]interface{}); ok {
			pm = pm.Set(internal.Keyword(k), convert(nest)).(*internal.PersistentMap)

			// key with array value
		} else if nestArr, ok := v.([]interface{}); ok {
			vals := make([]internal.Value, 0, len(nestArr))

			for _, n := range nestArr {
				// nested object
				if nested, ok := n.(map[string]interface{}); ok {
					vals = append(vals, convert(nested))
				} else {
					vals = append(vals, internal.ValueOf(n))
				}
			}
			pm = pm.Set(internal.Keyword(k), internal.ValueOf(vals)).(*internal.PersistentMap)

			// others can simply use ValueOf
		} else {
			pm = pm.Set(internal.Keyword(k), internal.ValueOf(v)).(*internal.PersistentMap)
		}
	}
	return pm
}

func apply(scope internal.Scope, args []internal.Value) (internal.Value, error) {

	argc := len(args)
	if argc < 2 {
		return nil, internal.ErrWrongArgumentCount(argc, "<>")
	}

	evaledArgs, err := internal.EvalValueList(scope, args)
	if err != nil {
		return nil, err
	}

	fn, ok := evaledArgs[0].(internal.Invokable)
	if !ok {
		return nil, doesNotImplementInvokable(evaledArgs[0])
	}

	fnArgs := evaledArgs[1 : len(evaledArgs)-1]

	lastArg := evaledArgs[len(evaledArgs)-1]
	coll, ok := lastArg.(internal.Seq)
	if !ok {
		return nil, doesNotImplementSeq(lastArg)
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

func eval(scope internal.Scope, args []internal.Value) (internal.Value, error) {
	form, err := args[0].Eval(scope)
	if err != nil {
		return nil, err
	}
	return form.Eval(scope)
}

func evalStr(scope internal.Scope, args []internal.Value) (internal.Value, error) {
	form, err := args[0].Eval(scope)
	if err != nil {
		return nil, err
	}

	fromStr, ok := form.(internal.String)
	if !ok {
		return nil, invalidType(internal.String(""), form)
	}

	return internal.ReadEvalStr(scope, string(fromStr))
}


func lazyRange(min, max, step int) internal.LazySeq {
	return internal.LazySeq{
		Min: min,
		Max: max,
		Step: step,
	}
}


func source(scope internal.Scope) func(string) (internal.Value, error) {
	return func(file string) (internal.Value, error) {
		content, err := readFile(file)
		if err != nil {
			return nil, err
		}

		value, err := internal.ReadEvalStr(scope, content)
		if err != nil {
			return nil, err
		}

		return value, nil
	}
}

func defClass(scope internal.Scope, args []internal.Value) (internal.Value, error) {
	
	argc := len(args)
	if argc < 2 {
		return nil, internal.ErrWrongArgumentCount(argc, "defclass")
	}

	name, ok := args[0].(internal.Symbol)
	if !ok {
		return nil, invalidType(internal.Symbol{}, args[0])
	}

	class := internal.Class{
		Name: name.String(),
	}

	// identifies the position of the member declaration
	hashMapIndex := 1

	// handles inheritence
	if symbol, ok := args[1].(internal.Symbol); ok {
		if symbol.Value != "<-" {
			return nil, fmt.Errorf("expecting hashMap or <- symbol")
		}

		arg2, err := args[2].Eval(scope)
		if err != nil {
			return nil, err
		}

		if parent, ok := arg2.(internal.Class); ok {
			class.Parent = &parent	
		}

		hashMapIndex = 3
	}

	evaledArgs, err := internal.EvalValueList(scope, args[hashMapIndex:])
	if err != nil {
		return nil, err
	}

	
	// evaluate passed hash map
	hashMap, ok := evaledArgs[0].(*internal.PersistentMap)
	if !ok {
		return nil, invalidType(&internal.PersistentMap{}, evaledArgs[0])
	}

	// ensure all the keys in hashmap are keywords
	for it := hashMap.Data.Iterator(); it.HasElem(); it.Next() {
		key, _ := it.Elem()

		_, ok := key.(internal.Keyword)
		if !ok {
			return nil, invalidType(internal.Keyword(""), key.(internal.Value))
		}
	}

	class.Members = hashMap

	methods := internal.NewPersistentMap()
	for _, m := range evaledArgs[1:] {
		
		method, ok := m.(*internal.List)
		if !ok {
			return nil, fmt.Errorf("expected defmethod")
		}

		name := internal.Keyword(method.First().String())
		var body internal.Value = method.Next().First()
		
		body, err = body.Eval(scope)
		if err != nil {
			return nil, err
		}
		
		fn, ok := body.(internal.Invokable)
		if !ok {
			return nil, fmt.Errorf("expecting invokable")
		}

		methods = methods.Set(name, fn).(*internal.PersistentMap)
	}

	class.Methods = methods

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

func mem(scope internal.Scope, args []internal.Value) (internal.Value, error) {

	var mem runtime.MemStats

	runtime.ReadMemStats(&mem)
	initialByte := mem.TotalAlloc

	val, err := internal.EvalValueLast(scope, args)
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
