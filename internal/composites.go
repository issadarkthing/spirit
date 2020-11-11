package internal

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/xiaq/persistent/hash"
	"github.com/xiaq/persistent/hashmap"
	"github.com/xiaq/persistent/vector"
)

// List represents an list of forms/vals. Evaluating a list leads to a
// function invocation.
type List struct {
	Values
	Position

	special *Fn
}

// Eval performs an invocation.
func (lf *List) Eval(scope Scope) (Value, error) {
	if lf.Size() == 0 {
		return lf, nil
	}

	err := lf.parse(scope)
	if err != nil {
		return nil, err
	}


	fnCall := Call{
		Name:     lf.Values[0].String(),
		Position: lf.Position,
	}

	if lf.special != nil {
		scope.Push(fnCall)
		val, err := lf.special.Invoke(scope, lf.Values[1:]...)
		if err != nil {
			return nil, addStackTrace(scope, err)
		}
		scope.Pop()
		return val, nil
	}

	target, err := Eval(scope, lf.Values[0])
	if err != nil {
		return nil, err
	}

	invokable, ok := target.(Invokable)
	if !ok {
		err = fmt.Errorf(
			"cannot invoke value of type '%s'", reflect.TypeOf(target),
		)

		return nil, err
	}

	scope.Push(fnCall)
	val, err := invokable.Invoke(scope, lf.Values[1:]...)
	if err != nil {
		return nil, addStackTrace(scope, err)
	}
	scope.Pop()

	return val, nil
}

func (lf List) String() string {
	return containerString(lf.Values, "(", ")", " ")
}

func (lf *List) parse(scope Scope) error {
	if lf.Size() == 0 {
		return nil
	}

	form, expanded, err := MacroExpand(scope, lf)
	if err != nil {
		return err
	}

	if expanded {
		lf.Values = Values{
			Symbol{Value: "do"},
			form,
		}
	}

	special, err := resolveSpecial(scope, lf.First())
	if err != nil {
		return err
	} else if special == nil {
		return analyzeSeq(scope, lf.Values)
	}

	fn, err := special.Parse(scope, lf.Values[1:])
	if err != nil {
		return fmt.Errorf("%s: %v", special.Name, err)
	}
	lf.special = fn
	return nil
}

// Vector represents a list of values. Unlike List type, evaluation of
// vector does not lead to function invoke.
type Vector struct {
	Values
	Position
}

func (vf Vector) Conj(vals ...Value) Seq {
	return *&Vector{Values: append(vf.Values, vals...)}
}

// Eval evaluates each value in the vector form and returns the resultant
// values as new vector.
func (vf Vector) Eval(scope Scope) (Value, error) {
	vals, err := EvalValueList(scope, vf.Values)
	if err != nil {
		return nil, err
	}

	return Vector{Values: vals}, nil
}

// Invoke of a vector performs a index lookup. Only arity 1 is allowed
// and should be an integer value to be used as index.
func (vf Vector) Invoke(scope Scope, args ...Value) (Value, error) {
	vals, err := EvalValueList(scope, args)
	if err != nil {
		return nil, err
	}

	if len(vals) != 1 {
		return nil, fmt.Errorf("call requires exactly 1 argument, got %d", len(vals))
	}

	index, isInt := vals[0].(Number)
	if !isInt {
		return nil, fmt.Errorf("key must be integer")
	}

	if int(index) >= len(vf.Values) {
		return nil, fmt.Errorf("index out of bounds")
	}

	return vf.Values[int(index)], nil
}

func (vf Vector) String() string {
	return containerString(vf.Values, "[", "]", " ")
}

// Set represents a list of unique values. (Experimental)
type Set struct {
	Values
	Position
}

// Eval evaluates each value in the set form and returns the resultant
// values as new set.
func (set Set) Eval(scope Scope) (Value, error) {
	vals, err := EvalValueList(scope, set.Uniq())
	if err != nil {
		return nil, err
	}

	return Set{Values: Values(vals).Uniq()}, nil
}

func (set Set) String() string {
	return containerString(set.Values, "#{", "}", " ")
}

// TODO: Remove this naive solution
func (set Set) valid() bool {
	s := map[string]struct{}{}

	for _, v := range set.Values {
		str := v.String()
		if _, found := s[str]; found {
			return false
		}
		s[v.String()] = struct{}{}
	}

	return true
}

// HashMap represents a container for key-value pairs.
type HashMap struct {
	Position
	Data map[Value]Value
}

// Eval evaluates all keys and values and returns a new HashMap containing
// the evaluated values.
func (hm *HashMap) Eval(scope Scope) (Value, error) {
	res := &HashMap{Data: map[Value]Value{}}
	for k, v := range hm.Data {
		key, err := k.Eval(scope)
		if err != nil {
			return nil, err
		}

		val, err := v.Eval(scope)
		if err != nil {
			return nil, err
		}

		res.Data[key] = val
	}

	return res, nil
}

func (hm *HashMap) String() string {
	var fields []Value
	for k, v := range hm.Data {
		fields = append(fields, k, v)
	}
	return containerString(fields, "{", "}", " ")
}

// Get returns the value associated with the given key if found.
// Returns def otherwise.
func (hm *HashMap) Get(key Value, def Value) Value {
	if !isHashable(key) {
		return def
	}

	v, found := hm.Data[key]
	if !found {
		return def
	}

	return v
}

// Set changes the value associated with the given key.
// destructive update
func (hm *HashMap) Set(key, val Value) error {
	if !isHashable(key) {
		return fmt.Errorf("value of type '%s' is not hashable", key)
	}

	hm.Data[key] = val
	return nil
}

// Keys returns all the keys in the hashmap.
func (hm *HashMap) Keys() Values {
	var res []Value
	for k := range hm.Data {
		res = append(res, k)
	}
	return res
}

// Values returns all the values in the hashmap.
func (hm *HashMap) Values() Values {
	var res []Value
	for _, v := range hm.Data {
		res = append(res, v)
	}
	return res
}

// Module represents a group of forms. Evaluating a module leads to evaluation
// of each form in order and result will be the result of last evaluation.
type Module []Value

// Eval evaluates all the vals in the module body and returns the result of the
// last evaluation.
func (mod Module) Eval(scope Scope) (Value, error) {
	res, err := EvalValueLast(scope, mod)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// Compare returns true if the 'v' is also a module and all forms in the
// module are equivalent.
func (mod Module) Compare(v Value) bool {
	other, ok := v.(Module)
	if !ok {
		return false
	}

	if len(mod) != len(other) {
		return false
	}

	for i := range mod {
		if !Compare(mod[i], other[i]) {
			return false
		}
	}

	return true
}

func (mod Module) String() string { return containerString(mod, "", "\n", "\n") }

// PersistentMap is persistant and does not mutate on change
// under the hood, it uses structural sharing to reduce the cost of copying
type PersistentMap struct {
	Position
	Data hashmap.Map
}

func NewPersistentMap() *PersistentMap {
	return &PersistentMap{Data: hashmap.New(compare, hasher)}
}

func (p *PersistentMap) Set(k, v Value) *PersistentMap {
	return &PersistentMap{
		Data: p.Data.Assoc(k, v),
	}
}

func (p *PersistentMap) Get(key, defValue Value) Value {
	val, ok := p.Data.Index(key)
	if !ok {
		return defValue
	}
	return val.(Value)
}

func (p *PersistentMap) Invoke(scope Scope, args ...Value) (Value, error) {
	
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("invoking hash map requires 1 or 2 arguments")
	}

	key := args[0]
	var defaultVal Value = Nil{}

	if len(args) == 2 {
		defaultVal = args[1]
	}

	return p.Get(key, defaultVal), nil
}

func (p *PersistentMap) Delete(k Value) *PersistentMap {
	return &PersistentMap{Data: p.Data.Dissoc(k)}
}

func (p *PersistentMap) Compare(other Value) bool {

	otherMap, ok := other.(*PersistentMap)
	if !ok {
		return false
	}

	if otherMap.Data.Len() != p.Data.Len() {
		return false
	}

	for it := p.Data.Iterator(); it.HasElem(); it.Next() {

		k1, v1 := it.Elem()

		v2, ok := otherMap.Data.Index(k1)
		if !ok {
			return false
		}

		if v1 != v2 {
			return false
		}
	}

	return true
}

func (p *PersistentMap) Eval(scope Scope) (Value, error) {
	res := &PersistentMap{Data: hashmap.New(compare, hasher)}

	for it := p.Data.Iterator(); it.HasElem(); it.Next() {
		k, v := it.Elem()

		key, err := k.(Value).Eval(scope)
		if err != nil {
			return nil, err
		}

		value, err := v.(Value).Eval(scope)
		if err != nil {
			return nil, err
		}

		res.Data = res.Data.Assoc(key, value)
	}

	return res, nil
}

func (p PersistentMap) String() string {
	m := p.Data
	var str strings.Builder
	str.WriteRune('{')
	length := m.Len()
	i := 0
	for it := m.Iterator(); it.HasElem(); it.Next() {
		k, v := it.Elem()
		if i != 0 {
			str.WriteRune(' ')
		}
		str.WriteString(fmt.Sprintf("%v %v", k, v))
		if i != length-1 {
			str.WriteRune(',')
		}
		i++
	}
	str.WriteRune('}')
	return str.String()
}

type PersistentVector struct {
	Position
	Vec vector.Vector
}

func NewPersistentVector() *PersistentVector {
	return &PersistentVector{Vec: vector.Empty}
}

func (p *PersistentVector) First() Value {
	v, ok := p.Vec.Index(0)
	if !ok {
		return nil
	}
	return v.(Value)
}

func (p *PersistentVector) Eval(scope Scope) (Value, error) {
	var pv Seq = NewPersistentVector()
	for it := p.Vec.Iterator(); it.HasElem(); it.Next() {
		v := it.Elem()
		val, err := v.(Value).Eval(scope)
		if err != nil {
			return nil, err
		}
		pv = pv.Conj(val)
	}
	return pv, nil
}

func (p *PersistentVector) String() string {
	vals := make([]Value, 0, p.Vec.Len())
	for it := p.Vec.Iterator(); it.HasElem(); it.Next() {
		vals = append(vals, it.Elem().(Value))
	}
	return containerString(vals, "[", "]", " ")
}

func (p *PersistentVector) Next() Seq {
	if p.Vec.Len() == 1 || p.Vec.Len() == 0 {
		return nil
	}
	return &PersistentVector{
		Vec:      p.Vec.SubVector(1, p.Vec.Len()),
		Position: p.Position,
	}
}

func (p *PersistentVector) Conj(vals ...Value) Seq {
	pv := &PersistentVector{
		Vec:      p.Vec,
		Position: p.Position,
	}
	for _, v := range vals {
		pv.Vec = pv.Vec.Cons(v)
	}
	return pv
}

func (p *PersistentVector) Cons(v Value) Seq {

	pv := NewPersistentVector()
	pv.Vec = pv.Vec.Cons(v)
	pv.SetPosition(p.Position)
	for it := p.Vec.Iterator(); it.HasElem(); it.Next() {
		val := it.Elem()
		pv.Vec = pv.Vec.Cons(val.(Value))
	}

	return pv
}

func (p *PersistentVector) Assoc(i int, v Value) Seq {
	return &PersistentVector{
		Vec:      p.Vec.Assoc(i, v),
		Position: p.Position,
	}
}

func (p *PersistentVector) SubVector(i, j int) Seq {
	return &PersistentVector{
		Vec:      p.Vec.SubVector(i, j),
		Position: p.Position,
	}
}

func (p *PersistentVector) Index(i int) Value {
	val, ok := p.Vec.Index(i)
	if !ok {
		panic("error out of bound")
	}
	return val.(Value)
}

func (p *PersistentVector) SetPosition(pos Position) *PersistentVector {
	p.Position = pos
	return p
}

func (p *PersistentVector) Compare(other Value) bool {

	pv2, ok := other.(*PersistentVector)
	if !ok {
		return false
	}

	if p.Size() != pv2.Size() {
		return false
	}

	i := 0
	for it := p.Vec.Iterator(); it.HasElem(); it.Next() {
		v1 := it.Elem()
		v2 := pv2.Index(i)

		if !Compare(v1.(Value), v2.(Value)) {
			return false
		}

		i++
	}

	return true
}

func (p *PersistentVector) Size() int {
	return p.Vec.Len()
}

func (p *PersistentVector) GetValues() []Value {
	vals := make([]Value, 0, p.Size())
	for it := p.Vec.Iterator(); it.HasElem(); it.Next() {
		vals = append(vals, it.Elem().(Value))
	}
	return vals
}

func (p *PersistentVector) Invoke(scope Scope, args ...Value) (Value, error) {
	vals, err := EvalValueList(scope, args)
	if err != nil {
		return nil, err
	}

	if len(vals) != 1 {
		return nil, fmt.Errorf("call requires exactly 1 argument, got %d", len(vals))
	}

	index, isInt := vals[0].(Number)
	if !isInt {
		return nil, fmt.Errorf("key must be integer")
	}

	i := int(index)

	if i >= p.Size() {
		return nil, fmt.Errorf("index out of bounds")
	}

	return p.Index(i), nil
}

type Call struct {
	Position
	Name string
}

// Stack contains function call. When fn is called, Call will be pushed in Stack,
// when the fn exits, the stack is popped
type Stack []Call

// Add function call to stack
func (s *Stack) Push(call Call) {
	*s = append(*s, call)
}

func (s Stack) Size() int {
	return len(s)
}

// Pops removes function call from Stack
func (s *Stack) Pop() Call {

	if s.Size() == 0 {
		return Call{}
	}

	last := (*s)[s.Size()-1]
	*s = (*s)[:s.Size()-1]
	return last
}

// StackTrace returns string representing current stack trace
func (s *Stack) StackTrace() string {

	var str strings.Builder
	// last index in slice
	last := s.Size() - 1
	for i := range *s {
		// iterate over slice in reverse
		call := (*s)[last-i]
		file, line, col := call.GetPos()
		str.WriteString(
			fmt.Sprintf("\nat %s (%s:%d:%d)", call.Name, file, line, col),
		)
	}

	return str.String()
}

type Future struct {
	Realized bool
	Value    Value
	Channel  chan Value
}

func (c *Future) Submit(scope Scope, form Value) {
	go func() {
		val, err := form.Eval(scope)
		if err != nil {
			panic(err)
		}
		c.Value = val
		c.Realized = true
	}()
}

func (c Future) String() string {
	return fmt.Sprintf("<Future(realized: %v value: %v)>", c.Realized, c.Value)
}

func (c *Future) Eval(_ Scope) (Value, error) {
	return c, nil
}

// Implements Seq interface but lazy
type LazySeq struct {
	Min    int
	Max    int
	Step   int
}

func (l LazySeq) First() Value {
	
	if l.Min >= l.Max {
		return ValueOf(nil)
	}

	return ValueOf(l.Min)
}

func (l LazySeq) Next() Seq {

	if l.Min+l.Step < l.Max {
		return LazySeq{
			Min: l.Min + l.Step,
			Max: l.Max,
			Step: l.Step,
		}
	}

	return nil
}

func (l LazySeq) values() []Value {

	if l.Size() == 0 {
		return []Value{}
	}
	result := make([]Value, 0, l.Size())
	var curr Seq
	for curr = l; curr != nil; curr = curr.Next() {
		result = append(result, curr.First())
	}
	return result
}

func (l LazySeq) Cons(v Value) Seq {
	return &List{
		Values: append([]Value{v}, l.values()...),
	}
}

func (l LazySeq) Conj(v ...Value) Seq {
	return &List{
		Values: append(l.values(), v...),
	}
}

func (l LazySeq) Size() int {
	return ((l.Max-l.Min) / l.Step)
}

func (l LazySeq) String() string {
	str := strings.Builder{}
	 
	str.WriteString("<LazySeq(")

	for i := l.Min; i < l.Max; i += l.Step {
		str.WriteString(fmt.Sprintf("%d ", i))
	}

	result := str.String()
	result = strings.TrimSuffix(result, " ")

	return result + ")>"
}

func (l LazySeq) Eval(_ Scope) (Value, error) {
	return l, nil
}


// ------------------ helper functions ---------------------------

func hasher(s interface{}) uint32 {
	return hash.String(s.(Value).String())
}

func compare(k1, k2 interface{}) bool {
	return Compare(k1.(Value), k2.(Value))
}

func containerString(vals []Value, begin, end, sep string) string {
	parts := make([]string, len(vals))
	for i, expr := range vals {
		parts[i] = fmt.Sprintf("%v", expr)
	}
	return begin + strings.Join(parts, sep) + end
}

// if an error occured when function is called, the error produced does not
// contain stackTrace. This helper function will add stackTrace to the EvalError
// type, and return as is if it has stackTrace
func addStackTrace(scope Scope, err error) error {
	if evalErr, ok := err.(EvalError); ok && evalErr.StackTrace == "" {
		evalErr.StackTrace = scope.StackTrace()
		return evalErr
	}
	return err
}
