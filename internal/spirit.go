package internal

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	nsSeparator = '/'
	defaultNS   = "user"
)

var _ Scope = (*Spirit)(nil)

// returns new Spirit instance
func NewSpirit() *Spirit {
	sl := &Spirit{
		Bindings: map[nsSymbol]Value{},
	}

	if err := bindAll(sl); err != nil {
		panic(err)
	}
	sl.checkNS = true

	_ = sl.SwitchNS(Symbol{Value: defaultNS})
	_ = sl.BindGo("ns", sl.SwitchNS)
	return sl
}

// Spirit instance
type Spirit struct {
	Stack
	currentNS string
	checkNS   bool
	Bindings  map[nsSymbol]Value
	Files     []string
}

// Eval evaluates the given value in spirit context.
func (spirit *Spirit) Eval(v Value) (Value, error) {
	return Eval(spirit, v)
}

// ReadEval reads from the given reader and evaluates all the forms
// obtained in spirit context.
func (spirit *Spirit) ReadEval(r io.Reader) (Value, error) {
	return ReadEval(spirit, r)
}

// ReadFile reads the content of the filename given. Use this to
// prevent recursive source
func (spirit *Spirit) ReadFile(filePath string) (Value, error) {

	if spirit.FileImported(filePath) {
		return nil, nil
	}

	spirit.AddFile(filePath)

	f, err := os.Open(filePath)
	defer f.Close()
	if err != nil {
		return nil, ImportError{err}
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, OSError{err}
	}

	dir := filepath.Dir(filePath)
	spirit.BindGo("*cwd*", dir)

	os.Chdir(dir)
	value, err := spirit.ReadEval(f)
	if err != nil {
		return nil, err
	}
	os.Chdir(cwd)

	return value, nil
}

// AddFile adds file to slice of imported files to prevent circular dependency.
func (s *Spirit) AddFile(file string) {
	s.Files = append(s.Files, file)
}

func (s *Spirit) FileImported(file string) bool {
	for _, v := range s.Files {
		if v == file {
			return true
		}
	}
	return false
}

// ReadEvalStr reads the source and evaluates it in spirit context.
func (spirit *Spirit) ReadEvalStr(src string) (Value, error) {
	return spirit.ReadEval(strings.NewReader(src))
}

// Bind binds the given name to the given Value into the spirit interpreter
// context.
func (spirit *Spirit) Bind(symbol string, v Value) error {

	nsSym, err := spirit.splitSymbol(symbol)
	if err != nil {
		return err
	}

	if spirit.checkNS && nsSym.NS != spirit.currentNS {
		return fmt.Errorf("cannot bind outside current namespace")
	}

	spirit.Bindings[*nsSym] = v
	return nil
}

// Resolve finds the value bound to the given symbol and returns it if
// found in the spirit context and returns it.
func (spirit *Spirit) Resolve(symbol string) (Value, error) {

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
	return spirit.Bind(symbol, ValueOf(v))
}

// SwitchNS changes the current namespace to the string value of given symbol.
func (spirit *Spirit) SwitchNS(sym Symbol) error {
	spirit.currentNS = sym.String()
	return spirit.Bind("*ns*", sym)
}

// CurrentNS returns the current active namespace.
func (spirit *Spirit) CurrentNS() string {
	return spirit.currentNS
}

// Parent always returns nil to represent this is the root scope.
func (spirit *Spirit) Parent() Scope {
	return nil
}

func (spirit *Spirit) resolveAny(symbol string, syms ...nsSymbol) (Value, error) {
	for _, s := range syms {
		v, found := spirit.Bindings[s]
		if found {
			return v, nil
		}
	}

	return nil, ResolveError{Sym: Symbol{Value: symbol}}
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
