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
func (s *Spirit) Eval(v Value) (Value, error) {
	return Eval(s, v)
}

// ReadEval reads from the given reader and evaluates all the forms
// obtained in spirit context.
func (s *Spirit) ReadEval(r io.Reader) (Value, error) {
	return ReadEval(s, r)
}

// ReadFile reads the content of the filename given. Use this to
// prevent recursive source
func (s *Spirit) ReadFile(filePath string) (Value, error) {

	if s.FileImported(filePath) {
		return nil, nil
	}

	s.AddFile(filePath)

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
	s.BindGo("*cwd*", dir)

	os.Chdir(dir)
	value, err := s.ReadEval(f)
	if err != nil {
		return nil, err
	}
	os.Chdir(cwd)

	return value, nil
}

func (s *Spirit) Has(symbol string) bool {

	if symbol == "ns" {
		symbol = "user/ns"
	}

	nsSym, err := s.splitSymbol(symbol)
	if err != nil {
		return false
	}

	_, found := s.Bindings[*nsSym]
	return found
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
func (s *Spirit) ReadEvalStr(src string) (Value, error) {
	return s.ReadEval(strings.NewReader(src))
}

// Bind binds the given name to the given Value into the spirit interpreter
// context.
func (s *Spirit) Bind(symbol string, v Value) error {

	nsSym, err := s.splitSymbol(symbol)
	if err != nil {
		return err
	}

	if s.checkNS && nsSym.NS != s.currentNS {
		return fmt.Errorf("cannot bind outside current namespace")
	}

	s.Bindings[*nsSym] = v
	return nil
}

// Resolve finds the value bound to the given symbol and returns it if
// found in the spirit context and returns it.
func (s *Spirit) Resolve(symbol string) (Value, error) {

	if symbol == "ns" {
		symbol = "user/ns"
	}

	nsSym, err := s.splitSymbol(symbol)
	if err != nil {
		return nil, err
	}

	return s.resolveAny(symbol, *nsSym, nsSym.WithNS("core"))
}

// BindGo is similar to Bind but handles conversion of Go value 'v' to
// internal Value type.
func (s *Spirit) BindGo(symbol string, v interface{}) error {
	return s.Bind(symbol, ValueOf(v))
}

// SwitchNS changes the current namespace to the string value of given symbol.
func (s *Spirit) SwitchNS(sym Symbol) error {
	s.currentNS = sym.String()
	return s.Bind("*ns*", sym)
}

// CurrentNS returns the current active namespace.
func (s *Spirit) CurrentNS() string {
	return s.currentNS
}

// Parent always returns nil to represent this is the root scope.
func (s *Spirit) Parent() Scope {
	return nil
}

func (s *Spirit) resolveAny(symbol string, syms ...nsSymbol) (Value, error) {
	for _, symbol := range syms {
		v, found := s.Bindings[symbol]
		if found {
			return v, nil
		}
	}

	return nil, ResolveError{Sym: Symbol{Value: symbol}}
}

func (s *Spirit) splitSymbol(symbol string) (*nsSymbol, error) {
	sep := string(nsSeparator)
	if symbol == sep {
		return &nsSymbol{
			NS:   s.currentNS,
			Name: symbol,
		}, nil
	}

	parts := strings.SplitN(symbol, sep, 2)
	if len(parts) < 2 {
		return &nsSymbol{
			NS:   s.currentNS,
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
