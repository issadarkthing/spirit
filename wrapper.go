package xlisp

import (
	"reflect"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/spy16/sabre"
)

func checkIfNil(value interface{}) bool {
	return reflect.ValueOf(value) == reflect.ValueOf(sabre.Nil{})
}

// TODO wrap all functions in tview library that requires callback
func ListAddItem(scope sabre.Scope) interface{} {
	return func(
		list *tview.List, first, second string,
		shortcut rune, selected interface{},
	) (sabre.Value, error) {

		if checkIfNil(selected) {
			return sabre.ValueOf(list.AddItem(first, second, shortcut, nil)), nil
		}

		callBack := func() {
			_, err := selected.(sabre.Invokable).Invoke(scope)
			if err != nil {
				panic(err)
			}
		}

		return sabre.ValueOf(list.AddItem(first, second, shortcut, callBack)), nil
	}
}

func AppSetBeforeDrawFunc(scope sabre.Scope) interface{} {
	return func(app *tview.Application, cb interface{}) (sabre.Value, error) {

		if checkIfNil(cb) {
			return sabre.ValueOf(app.SetBeforeDrawFunc(nil)), nil
		}

		callBack := func(screen tcell.Screen) bool {
			val, err := cb.(sabre.Invokable).Invoke(scope, sabre.ValueOf(screen))
			if err != nil {
				panic(err)
			}
			return bool(val.(sabre.Bool))
		}

		return sabre.ValueOf(app.SetBeforeDrawFunc(callBack)), nil
	}
}

func AppSetInputCapture(scope sabre.Scope) interface{} {
	return func(app *tview.Application, cb interface{}) (sabre.Value, error) {

		if checkIfNil(cb) {
			return sabre.ValueOf(app.SetInputCapture(nil)), nil
		}

		callBack := func(e *tcell.EventKey) *tcell.EventKey {
			_, err := cb.(sabre.Invokable).Invoke(scope, sabre.ValueOf(e))
			if err != nil {
				panic(err)
			}
			return e
		}

		return sabre.ValueOf(app.SetInputCapture(callBack)), nil
	}
}
