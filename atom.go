package xlisp

import (
	"fmt"
	"sync"

	"github.com/spy16/sabre"
)

type Atom struct {
	mu  sync.RWMutex
	Val sabre.Value
}

func (a *Atom) UpdateState(scope sabre.Scope, fn sabre.Invokable) (sabre.Value, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	res, err := fn.Invoke(scope, a.Val)
	if err != nil {
		return nil, err
	}

	a.Val = res
	return res, nil
}

func (a *Atom) GetVal() sabre.Value {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Val
}

func (a *Atom) String() string {
	return fmt.Sprintf("(atom %v)", a.GetVal())
}

func (a *Atom) Eval(_ sabre.Scope) (sabre.Value, error) {
	return sabre.ValueOf(a), nil
}

func newAtom(val sabre.Value) *Atom {
	return &Atom{Val: val}
}
