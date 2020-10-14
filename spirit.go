package spirit

import (
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/issadarkthing/spirit/internal"
)

const (
	nsSeparator = '/'
	defaultNS   = "user"
)

// returns new Spirit instance
func New() *Spirit {
	sl := &Spirit{
		mu:       &sync.RWMutex{},
		bindings: map[nsSymbol]internal.Value{},
	}

	if err := bindAll(sl); err != nil {
		panic(err)
	}
	sl.checkNS = true

	_ = sl.SwitchNS(internal.Symbol{Value: defaultNS})
	_ = sl.BindGo("ns", sl.SwitchNS)
	return sl
}

// Spirit instance
type Spirit struct {
	mu        *sync.RWMutex
	currentNS string
	checkNS   bool
	bindings  map[nsSymbol]internal.Value
}

// Eval evaluates the given value in spirit context.
func (spirit *Spirit) Eval(v internal.Value) (internal.Value, error) {
	return internal.Eval(spirit, v)
}

// ReadEval reads from the given reader and evaluates all the forms
// obtained in spirit context.
func (spirit *Spirit) ReadEval(r io.Reader) (internal.Value, error) {
	internalReader := internal.NewReader(r)
	mod, err := internalReader.All()
	if err != nil {
		return nil, err
	}
	return internal.Eval(spirit, mod)
}


// ReadEvalStr reads the source and evaluates it in spirit context.
func (spirit *Spirit) ReadEvalStr(src string) (internal.Value, error) {
	return spirit.ReadEval(strings.NewReader(src))
}

// Bind binds the given name to the given Value into the spirit interpreter
// context.
func (spirit *Spirit) Bind(symbol string, v internal.Value) error {
	spirit.mu.Lock()
	defer spirit.mu.Unlock()

	nsSym, err := spirit.splitSymbol(symbol)
	if err != nil {
		return err
	}

	if spirit.checkNS && nsSym.NS != spirit.currentNS {
		return fmt.Errorf("cannot bind outside current namespace")
	}

	spirit.bindings[*nsSym] = v
	return nil
}

// Resolve finds the value bound to the given symbol and returns it if
// found in the spirit context and returns it.
func (spirit *Spirit) Resolve(symbol string) (internal.Value, error) {
	spirit.mu.RLock()
	defer spirit.mu.RUnlock()

	if symbol == "ns" {
		symbol = "user/ns"
	}

	nsSym, err := spirit.splitSymbol(symbol)
	if err != nil {
		return nil, err
	}

	return spirit.resolveAny(symbol, *nsSym, nsSym.WithNS("core"))
}

// BindGo is similar to Bind but handles conversion of Go value 'v' to
// internal Value type.
func (spirit *Spirit) BindGo(symbol string, v interface{}) error {
	return spirit.Bind(symbol, internal.ValueOf(v))
}

// SwitchNS changes the current namespace to the string value of given symbol.
func (spirit *Spirit) SwitchNS(sym internal.Symbol) error {
	spirit.mu.Lock()
	spirit.currentNS = sym.String()
	spirit.mu.Unlock()

	return spirit.Bind("*ns*", sym)
}

// CurrentNS returns the current active namespace.
func (spirit *Spirit) CurrentNS() string {
	spirit.mu.RLock()
	defer spirit.mu.RUnlock()

	return spirit.currentNS
}

// Parent always returns nil to represent this is the root scope.
func (spirit *Spirit) Parent() internal.Scope {
	return nil
}

func (spirit *Spirit) resolveAny(symbol string, syms ...nsSymbol) (internal.Value, error) {
	for _, s := range syms {
		v, found := spirit.bindings[s]
		if found {
			return v, nil
		}
	}

	return nil, fmt.Errorf("unable to resolve symbol: %v", symbol)
}

func (spirit *Spirit) splitSymbol(symbol string) (*nsSymbol, error) {
	sep := string(nsSeparator)
	if symbol == sep {
		return &nsSymbol{
			NS:   spirit.currentNS,
			Name: symbol,
		}, nil
	}

	parts := strings.SplitN(symbol, sep, 2)
	if len(parts) < 2 {
		return &nsSymbol{
			NS:   spirit.currentNS,
			Name: symbol,
		}, nil
	}

	if strings.Contains(parts[1], sep) && parts[1] != sep {
		return nil, fmt.Errorf("invalid qualified symbol: '%s'", symbol)
	}

	return &nsSymbol{
		NS:   parts[0],
		Name: parts[1],
	}, nil
}

type nsSymbol struct {
	NS   string
	Name string
}

func (s nsSymbol) WithNS(ns string) nsSymbol {
	s.NS = ns
	return s
}
