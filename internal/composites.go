package internal

import (
	"fmt"
	"strings"

	"github.com/xiaq/persistent/hash"
	"github.com/xiaq/persistent/hashmap"
	"github.com/xiaq/persistent/vector"
)

const (
	IndentLevel = 4
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

	root := RootScope(scope)
	spirit, ok := root.(*Spirit)
	if !ok {
		return nil, fmt.Errorf("InternalError: cannot find root scope")
	}

	if lf.special != nil {
		spirit.Push(fnCall)
		val, err := lf.special.Invoke(scope, lf.Values[1:]...)
		if err != nil {
			err = newEvalErr(lf, err)
			return nil, addStackTrace(spirit.Stack, err)
		}
		spirit.Pop()
		return val, nil
	}

	target, err := Eval(scope, lf.Values[0])
	if err != nil {
		return nil, err
	}

	invokable, ok := target.(Invokable)
	if !ok {
		return nil, ImplementError{
			Name: invokableStr,
			Val:  target,
		}
	}

	spirit.Push(fnCall)
	val, err := invokable.Invoke(scope, lf.Values[1:]...)
	if err != nil {
		return nil, addStackTrace(spirit.Stack, err)
	}
	spirit.Pop()

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

// HashMap is persistant and does not mutate on change
// under the hood, it uses structural sharing to reduce the cost of copying
type HashMap struct {
	Position
	Data hashmap.Map
}

func NewHashMap() *HashMap {
	return &HashMap{Data: hashmap.New(compare, hasher)}
}

func (hm *HashMap) Set(k, v Value) Value {
	return &HashMap{
		Data: hm.Data.Assoc(k, v),
	}
}

func (hm *HashMap) Get(key Value) Value {
	val, ok := hm.Data.Index(key)
	if !ok {
		return nil
	}
	return val.(Value)
}

// HashMap implements Seq interface. Note that the order of items is unstable
func (hm *HashMap) Size() int {
	return hm.Data.Len()
}

func (hm *HashMap) First() Value {
	it := hm.Data.Iterator()
	if !it.HasElem() {
		return Nil{}
	}

	k, v := it.Elem()
	return NewVector().Conj(k.(Value), v.(Value))
}

func (hm *HashMap) Next() Seq {
	it := hm.Data.Iterator()
	if hm.Size() < 2 {
		return nil
	}

	k, _ := it.Elem()
	return &HashMap{
		Data: hm.Data.Dissoc(k),
	}
}

func (hm *HashMap) Cons(v Value) Seq {

	vec, ok := v.(*Vector)
	if !ok || vec.Size() != 2 {
		return hm
	}

	key, value := vec.Index(0), vec.Index(1)
	return &HashMap{
		Data: hm.Data.Assoc(key, value),
	}
}

func (hm *HashMap) Conj(vals ...Value) Seq {
	var h Seq = hm
	for _, v := range vals {
		h = h.Cons(v)
	}
	return h
}

func (hm *HashMap) Invoke(scope Scope, args ...Value) (Value, error) {

	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("invoking hash map requires 1 or 2 arguments")
	}

	key := args[0]
	value := hm.Get(key)

	if len(args) == 2 && value == nil {
		return value, nil
	}

	return value, nil
}

func (hm *HashMap) Delete(k Value) *HashMap {
	return &HashMap{Data: hm.Data.Dissoc(k)}
}

// Compare implements Comparable. It compares each key and value recursively,
// note that order is not important when comparing
func (hm *HashMap) Compare(other Value) bool {

	otherMap, ok := other.(*HashMap)
	if !ok {
		return false
	}

	if otherMap.Data.Len() != hm.Data.Len() {
		return false
	}

	for it := hm.Data.Iterator(); it.HasElem(); it.Next() {

		k1, v1 := it.Elem()

		v2, ok := otherMap.Data.Index(k1)
		if !ok {
			return false
		}

		if !compare(v1, v2) {
			return false
		}
	}

	return true
}

func (hm *HashMap) Eval(scope Scope) (Value, error) {
	res := &HashMap{Data: hashmap.New(compare, hasher)}

	for it := hm.Data.Iterator(); it.HasElem(); it.Next() {
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

func (hm HashMap) PrettyPrint(indent int) string {

	space := strings.Repeat(" ", indent)

	if hm.Size() == 0 {
		return fmt.Sprintf("{}")
	}

	str := strings.Builder{}
	fmt.Fprintf(&str, "{")

	for it := hm.Data.Iterator(); it.HasElem(); it.Next() {

		key, val := it.Elem()

		if v, ok := val.(PrettyPrinter); ok {
			val = v.PrettyPrint(indent + IndentLevel)
		}

		fmt.Fprintf(&str, "\n%s    %s %s", space, key, val)
	}

	fmt.Fprintf(&str, "\n%s}", space)
	return str.String()
}

func (hm HashMap) String() string {
	m := hm.Data
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

type Vector struct {
	Position
	Vec vector.Vector
}

func NewVector() *Vector {
	return &Vector{Vec: vector.Empty}
}

func (p *Vector) Eval(scope Scope) (Value, error) {
	var pv Seq = NewVector()
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

func (p *Vector) PrettyPrint(indent int) string {

	space := strings.Repeat(" ", indent)

	if p.Size() == 0 {
		return fmt.Sprintf("[]")
	}

	str := strings.Builder{}
	fmt.Fprintf(&str, "[")

	for it := p.Vec.Iterator(); it.HasElem(); it.Next() {
		elem := it.Elem()
		ppStr := elem

		if pp, ok := elem.(PrettyPrinter); ok {
			ppStr = pp.PrettyPrint(indent + IndentLevel)
		}

		fmt.Fprintf(&str, "\n    %s%s,", space, ppStr)
	}

	fmt.Fprintf(&str, "\n%s]", space)
	return str.String()
}

func (p *Vector) String() string {
	vals := make([]Value, 0, p.Vec.Len())
	for it := p.Vec.Iterator(); it.HasElem(); it.Next() {
		vals = append(vals, it.Elem().(Value))
	}
	return containerString(vals, "[", "]", " ")
}

func (p *Vector) First() Value {
	v, ok := p.Vec.Index(0)
	if !ok {
		return nil
	}
	return v.(Value)
}

func (p *Vector) Next() Seq {
	if p.Vec.Len() == 1 || p.Vec.Len() == 0 {
		return nil
	}
	return &Vector{
		Vec:      p.Vec.SubVector(1, p.Vec.Len()),
		Position: p.Position,
	}
}

func (p *Vector) Conj(vals ...Value) Seq {
	pv := &Vector{
		Vec:      p.Vec,
		Position: p.Position,
	}
	for _, v := range vals {
		pv.Vec = pv.Vec.Cons(v)
	}
	return pv
}

func (p *Vector) Cons(v Value) Seq {

	pv := NewVector()
	pv.Vec = pv.Vec.Cons(v)
	pv.SetPosition(p.Position)
	for it := p.Vec.Iterator(); it.HasElem(); it.Next() {
		val := it.Elem()
		pv.Vec = pv.Vec.Cons(val.(Value))
	}

	return pv
}

func (p *Vector) Set(i Value, v Value) Value {
	return &Vector{
		Vec:      p.Vec.Assoc(int(i.(Number)), v),
		Position: p.Position,
	}
}

func (p *Vector) Get(i Value) Value {
	value, ok := p.Vec.Index(int(i.(Number)))
	if !ok {
		return nil
	}
	return value.(Value)
}

func (p *Vector) SubVector(i, j int) Seq {
	return &Vector{
		Vec:      p.Vec.SubVector(i, j),
		Position: p.Position,
	}
}

func (p *Vector) Index(i int) Value {
	val, ok := p.Vec.Index(i)
	if !ok {
		panic("error out of bound")
	}
	return val.(Value)
}

func (p *Vector) SetPosition(pos Position) *Vector {
	p.Position = pos
	return p
}

// Compare implements Comparable which recursively compare values between
// other value. Order is important
func (p *Vector) Compare(other Value) bool {

	pv2, ok := other.(*Vector)
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

		if !compare(v1, v2) {
			return false
		}

		i++
	}

	return true
}

func (p *Vector) Size() int {
	return p.Vec.Len()
}

func (p *Vector) GetValues() []Value {
	vals := make([]Value, 0, p.Size())
	for it := p.Vec.Iterator(); it.HasElem(); it.Next() {
		vals = append(vals, it.Elem().(Value))
	}
	return vals
}

func (p *Vector) Invoke(scope Scope, args ...Value) (Value, error) {
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

// maximum stack to be stored
const MAX_STACK_CALL = 20

// Stack contains function call. When fn is called, Call will be pushed in Stack,
// when the fn exits, the stack is popped
type Stack []Call

// Add function call to stack
func (s *Stack) Push(call Call) {

	if s.Size() > MAX_STACK_CALL {
		// remove the bottom element
		*s = (*s)[1:]
	}

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
		if file != "" && line != 0 && col != 0 {
			fmt.Fprintf(&str, "\nat %s (%s:%d:%d)", 
				call.Name, file, line, col)
		}
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
	Min  int
	Max  int
	Step int
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
			Min:  l.Min + l.Step,
			Max:  l.Max,
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
	for curr = l; curr != nil && curr.First() != nil; curr = curr.Next() {
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
	return ((l.Max - l.Min) / l.Step)
}

func (l LazySeq) String() string {
	return fmt.Sprintf("#[%d %d %d]", l.Min, l.Max, l.Step)
}

func (l LazySeq) Eval(_ Scope) (Value, error) {
	return l, nil
}

type Class struct {
	Name          string
	Parent        *Class
	Members       *HashMap
	Methods       *HashMap
	StaticsMethod *HashMap
}

func (c Class) Eval(_ Scope) (Value, error) {
	return c, nil
}

func (c Class) PrettyPrint(indent int) string {

	// add whitespace based on indentation level
	space := strings.Repeat(" ", indent)

	if c.Members.Data.Len() == 0 {
		return fmt.Sprintf("%s {}", c.Name)
	}

	str := strings.Builder{}

	fmt.Fprintf(&str, "class %s {", c.Name)
	for name, member := range c.GetMembers() {

		if m, ok := member.(PrettyPrinter); ok {
			classStr := m.PrettyPrint(indent + IndentLevel)
			fmt.Fprintf(&str, "\n%s    %s: %s", space, name, classStr)

		} else {
			fmt.Fprintf(&str, "\n%s    %s: %s", space, name, member)

		}
	}

	fmt.Fprintf(&str, "\n%s}", space)
	return str.String()
}

func (c Class) String() string {
	return c.PrettyPrint(0)
}

func (c *Class) Inherit(parent *Class) {
	c.Parent = parent
}

func (c Class) GetMember(name Keyword) (Value, bool) {
	member := c.Members.Get(name)
	if member == nil && c.Parent == nil {
		return Nil{}, false

	} else if member == nil && c.Parent != nil {
		return c.Parent.GetMember(name)
	}
	return member, true
}

func (c Class) GetStaticMethod(name Keyword) (Invokable, bool) {
	static := c.StaticsMethod.Get(name)
	if static == nil && c.Parent == nil {
		return nil, false

	} else if static == nil && c.Parent != nil {
		return c.Parent.GetStaticMethod(name)
	}
	return static.(Invokable), true
}

func (c Class) GetMethod(name Keyword) (Invokable, bool) {
	method := c.Methods.Get(name)
	if method == nil && c.Parent == nil {
		return nil, false

	} else if method == nil && c.Parent != nil {
		return c.Parent.GetMethod(name)
	}
	return method.(Invokable), true
}

// finds whether member or method exists
func (c Class) Exists(name Keyword) bool {
	_, member := c.GetMember(name)
	_, method := c.GetMethod(name)
	return member || method
}

func (c Class) GetMembers() map[string]Value {
	members := make(map[string]Value)
	if c.Parent != nil {
		members = c.Parent.GetMembers()
	}

	for it := c.Members.Data.Iterator(); it.HasElem(); it.Next() {
		name, memberType := it.Elem()

		key := name.(Keyword)
		value := memberType.(Value)
		members[string(key)] = value
	}

	return members
}

func (c Class) GetMethods() map[string]Invokable {
	methods := make(map[string]Invokable)
	if c.Parent != nil {
		methods = c.Parent.GetMethods()
	}

	for it := c.Methods.Data.Iterator(); it.HasElem(); it.Next() {
		name, memberType := it.Elem()

		key := name.(Keyword)
		value := memberType.(Invokable)
		methods[string(key)] = value
	}
	return methods
}

func (c Class) Invoke(scope Scope, args ...Value) (Value, error) {

	if len(args) != 1 {
		return nil, fmt.Errorf("invalid arguments passed; expected 1")
	}

	arg, err := args[0].Eval(scope)
	if err != nil {
		return nil, err
	}

	passedMap, ok := arg.(*HashMap)
	if !ok {
		return nil, TypeError{
			Expected: TypeOf(NewHashMap()),
			Got:      TypeOf(arg),
		}
	}

	return Object{
		InstanceOf: c,
		Members:    passedMap,
	}, nil
}

type Object struct {
	InstanceOf Class
	Members    *HashMap
}

func (o Object) Eval(_ Scope) (Value, error) {
	return o, nil
}

func (o Object) Set(key, value Value) Value {
	return Object{
		InstanceOf: o.InstanceOf,
		Members:    o.Members.Set(key, value).(*HashMap),
	}
}

func (o Object) Get(key Value) Value {
	kw, ok := key.(Keyword)
	if !ok {
		return nil
	}

	val, ok := o.GetMember(kw)
	if ok {
		return val
	}

	val, ok = o.GetMethod(kw)
	if ok {
		return val
	}

	return nil
}

func (o Object) PrettyPrint(indent int) string {

	// add whitespace based on indentation level
	space := strings.Repeat(" ", indent)

	if o.Members.Data.Len() == 0 {
		return fmt.Sprintf("%s {}", o.InstanceOf.Name)
	}

	str := strings.Builder{}

	fmt.Fprintf(&str, "%s {", o.InstanceOf.Name)
	for it := o.Members.Data.Iterator(); it.HasElem(); it.Next() {

		name, member := it.Elem()

		if m, ok := member.(PrettyPrinter); ok {
			classStr := m.PrettyPrint(indent + IndentLevel)
			fmt.Fprintf(&str, "\n%s    %s: %s", space, name, classStr)

		} else {
			fmt.Fprintf(&str, "\n%s    %s: %s",
				space, string(name.(Keyword)), member)
		}

	}

	fmt.Fprintf(&str, "\n%s}", space)
	return str.String()
}

func (o Object) String() string {
	return o.PrettyPrint(0)
}

func (o Object) GetMember(name Keyword) (Value, bool) {
	val := o.Members.Get(name)
	if val == nil {
		return nil, false
	}
	return val, true
}

func (o Object) GetMethod(name Keyword) (Invokable, bool) {
	return o.InstanceOf.GetMethod(name)
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
// type, and return as if it has stackTrace
func addStackTrace(stack Stack, err error) error {
	if evalErr, ok := err.(EvalError); ok && evalErr.StackTrace == "" {
		evalErr.StackTrace = stack.StackTrace()
		return evalErr
	}
	return err
}

func ClearStack(stack *Stack) {
	for call := stack.Pop(); call != (Call{}); call = stack.Pop() {
	}
}
