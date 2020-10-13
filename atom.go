package spirit

import (
	"fmt"
	"sync"

	"github.com/issadarkthing/spirit/internal"
)

type Atom struct {
	mu  sync.RWMutex
	Val internal.Value
}

func (a *Atom) UpdateState(scope internal.Scope, fn internal.Invokable) (internal.Value, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	res, err := fn.Invoke(scope, a.Val)
	if err != nil {
		return nil, err
	}

	a.Val = res
	return res, nil
}

func (a *Atom) GetVal() internal.Value {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Val
}

func (a *Atom) String() string {
	return fmt.Sprintf("(atom %v)", a.GetVal())
}

func (a *Atom) Eval(_ internal.Scope) (internal.Value, error) {
	return internal.ValueOf(a), nil
}

func newAtom(val internal.Value) *Atom {
	return &Atom{Val: val}
}
