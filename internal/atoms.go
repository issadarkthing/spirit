package internal

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

// Nil represents a nil value.
type Nil struct{}

// Eval returns the underlying value.
func (n Nil) Eval(_ Scope) (Value, error) { return n, nil }

func (n Nil) String() string { return "nil" }

// Bool represents a boolean value.
type Bool bool

// Eval returns the underlying value.
func (b Bool) Eval(_ Scope) (Value, error) { return b, nil }

func (b Bool) String() string { return fmt.Sprintf("%t", b) }

// Number represents double precision floating point numbers
type Number float64

// Eval simply returns itself since Floats evaluate to themselves.
func (n Number) Eval(_ Scope) (Value, error) { return n, nil }

func (n Number) String() string { 
	return strconv.FormatFloat(float64(n), 'f', -1, 64) 
}


// String represents double-quoted string literals. String Form represents
// the true string value obtained from the reader. Escape sequences are not
// applicable at this level.
type String string

// Eval simply returns itself since Strings evaluate to themselves.
func (se String) Eval(_ Scope) (Value, error) { return se, nil }

func (se String) String() string { return fmt.Sprintf("\"%s\"", string(se)) }

// First returns the first character if string is not empty, nil otherwise.
func (se String) First() Value {
	if len(se) == 0 {
		return Nil{}
	}

	return Character(se[0])
}

// Next slices the string by excluding first character and returns the
// remainder.
func (se String) Next() Seq { return se.chars().Next() }

// Cons converts the string to character sequence and adds the given value
// to the beginning of the list.
func (se String) Cons(v Value) Seq { return se.chars().Cons(v) }

// Conj joins the given values to list of characters of the string and returns
// the new sequence.
func (se String) Conj(vals ...Value) Seq { return se.chars().Conj(vals...) }

func (se String) chars() Values {
	var vals Values
	for _, r := range se {
		vals = append(vals, Character(r))
	}
	return vals
}

// Character represents a character literal.  For example, \a, \b, \1, \∂ etc
// are valid character literals. In addition, special literals like \newline,
// \space etc are supported by the reader.
type Character rune

// Eval simply returns itself since Chracters evaluate to themselves.
func (char Character) Eval(_ Scope) (Value, error) { return char, nil }

func (char Character) String() string { return fmt.Sprintf("\\%c", rune(char)) }

// Keyword represents a keyword literal.
type Keyword string

// Eval simply returns itself since Keywords evaluate to themselves.
func (kw Keyword) Eval(_ Scope) (Value, error) { return kw, nil }

func (kw Keyword) String() string { return fmt.Sprintf(":%s", string(kw)) }

// Invoke enables keyword lookup for maps.
func (kw Keyword) Invoke(scope Scope, args ...Value) (Value, error) {
	if err := verifyArgCount([]int{1, 2}, args); err != nil {
		return nil, err
	}

	argVals, err := evalValueList(scope, args)
	if err != nil {
		return nil, err
	}

	hm, ok := argVals[0].(*PersistentMap)
	if !ok {
		return Nil{}, nil
	}

	def := Value(Nil{})
	if len(argVals) == 2 {
		def = argVals[1]
	}

	return hm.Get(kw, def), nil
}

// Symbol represents a name given to a value in memory.
type Symbol struct {
	Position
	Value string
}

// Eval returns the value bound to this symbol in current context. If the
// symbol is in fully qualified form (i.e., separated by '.'), eval does
// recursive member access.
func (sym Symbol) Eval(scope Scope) (Value, error) {
	target, err := sym.resolveValue(scope)
	if err != nil {
		return nil, err
	}

	if _, isSpecial := target.(SpecialForm); isSpecial {
		return nil, fmt.Errorf("can't take value of special form '%s'", sym.Value)
	}

	if isMacro(target) {
		return nil, fmt.Errorf("can't take value of macro '%s'", sym.Value)
	}

	return target, nil
}

// Compare compares this symbol to the given value. Returns true if
// the given value is a symbol with same data.
func (sym Symbol) Compare(v Value) bool {
	other, ok := v.(Symbol)
	if !ok {
		return false
	}

	return other.Value == sym.Value
}

func (sym Symbol) String() string { return sym.Value }

func (sym Symbol) resolveValue(scope Scope) (Value, error) {
	fields := strings.Split(sym.Value, ".")

	if sym.Value == "." {
		fields = []string{"."}
	}

	target, err := scope.Resolve(fields[0])
	if len(fields) == 1 || err != nil {
		return target, err
	}

	rv := reflect.ValueOf(target)
	for i := 1; i < len(fields); i++ {
		if rv.Type() == reflect.TypeOf(Any{}) {
			rv = rv.Interface().(Any).V
		}

		rv, err = accessMember(rv, fields[i])
		if err != nil {
			return nil, err
		}
	}

	if isKind(rv.Type(), reflect.Chan, reflect.Array,
		reflect.Func, reflect.Ptr) && rv.IsNil() {
		return Nil{}, nil
	}

	return ValueOf(rv.Interface()), nil
}

func resolveSpecial(scope Scope, v Value) (*SpecialForm, error) {
	sym, isSymbol := v.(Symbol)
	if !isSymbol {
		return nil, nil
	}

	v, err := sym.resolveValue(scope)
	if err != nil {
		return nil, nil
	}

	sf, ok := v.(SpecialForm)
	if !ok {
		return nil, nil
	}

	return &sf, nil
}

// Atom is a thread-safe reference type
type Atom struct {
	mu  sync.RWMutex
	Val Value
}

func (a *Atom) UpdateState(scope Scope, fn Invokable) (Value, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	res, err := fn.Invoke(scope, a.Val)
	if err != nil {
		return nil, err
	}

	a.Val = res
	return res, nil
}

func (a *Atom) GetVal() Value {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Val
}

func (a *Atom) String() string {
	return fmt.Sprintf("(atom %v)", a.GetVal())
}

func (a *Atom) Eval(_ Scope) (Value, error) {
	return ValueOf(a), nil
}

func NewAtom(val Value) *Atom {
	return &Atom{Val: val}
}
