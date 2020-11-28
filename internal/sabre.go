// Package sabre provides data structures, reader for reading LISP source
// into data structures and functions for evluating forms against a context.
package internal

import (
	"fmt"
	"io"
	"reflect"
	"strings"
)

// Eval evaluates the given form against the scope and returns the result
// of evaluation.
func Eval(scope Scope, form Value) (Value, error) {
	if form == nil {
		return Nil{}, nil
	}

	v, err := form.Eval(scope)
	if err != nil {
		return v, err
	}

	return v, nil
}

// ReadEval consumes data from reader 'r' till EOF, parses into forms
// and evaluates all the forms obtained and returns the result.
func ReadEval(scope Scope, r io.Reader) (Value, error) {
	mod, err := NewReader(r).All()
	if err != nil {
		return nil, err
	}

	err = hoistValues(scope, mod)
	if err != nil {
		return nil, err
	}

	return Eval(scope, mod)
}

func hoistValues(scope Scope, value Value) error {

	hoistedVals := []string{"def", "defn", "defmacro", "defclass"}
	for _, form := range value.(Module) {
		list, ok := form.(*List); 
		if !ok {
			continue
		}

		if list.Size() < 2 {
			continue
		}

		def, isSymbol := list.Values[0].(Symbol)
		if !isSymbol {
			return fmt.Errorf("first argument must be symbol, not '%v'",
			reflect.TypeOf(list.Values[0]))
		}

		if !includes(def.String(), hoistedVals) {
			continue
		}

		sym, isSymbol := list.Values[1].(Symbol)
		if !isSymbol {
			return fmt.Errorf("first argument must be symbol, not '%v'",
			reflect.TypeOf(list.Values[1]))
		}
		symbol := sym.String()

		// only hoist for def that defines defn and defmacro
		if def.String() == "def" && (symbol == "defn" || symbol == "defmacro") {
			_, err := form.Eval(scope)
			if err != nil {
				return err
			}
			// do not hoist def
		} else if def.String() == "def" {
			continue
		}

		_, err := form.Eval(scope)
		if err != nil {
			return err
		}

	}
	return nil
}

func includes(search string, content []string) bool {
	for _, v := range content {
		if v == search {
			return true
		}
	}
	return false
}

// ReadEvalStr is a convenience wrapper for Eval that reads forms from
// string and evaluates for result.
func ReadEvalStr(scope Scope, src string) (Value, error) {
	return ReadEval(scope, strings.NewReader(src))
}

// Scope implementation is responsible for managing value bindings.
type Scope interface {
	Push(Call)
	Pop() Call
	StackTrace() string
	Parent() Scope
	Bind(symbol string, v Value) error
	Resolve(symbol string) (Value, error)
}

func newEvalErr(v Value, err error) EvalError {
	if ee, ok := err.(EvalError); ok {
		return ee
	} else if ee, ok := err.(*EvalError); ok && ee != nil {
		return *ee
	}

	return EvalError{
		Position: getPosition(v),
		Cause:    err,
		Form:     v,
	}
}

// EvalError represents error during evaluation.
type EvalError struct {
	Position
	Cause      error
	StackTrace string
	Form       Value
}

// Unwrap returns the underlying cause of this error.
func (ee EvalError) Unwrap() error { return ee.Cause }

func (ee EvalError) Error() string {
	return fmt.Sprintf("%s\nin '%s' (at line %d:%d) %v",
		ee.Cause, ee.File, ee.Line, ee.Column, ee.StackTrace,
	)
}
