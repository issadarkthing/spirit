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
		shortcut rune, selected sabre.Invokable,
	) (sabre.Value, error) {

		callBack := func() {
			_, err := selected.Invoke(scope)
			if err != nil {
				panic(err)
			}
		}
		return sabre.ValueOf(list.AddItem(first, second, shortcut, callBack)), nil
	}
}

// func ListSetChangedFunc(scope sabre.Scope) interface{} {
// 	return func(list *tview.List, handler sabre.Invokable) (sabre.Value, error) {

// 	}
// }

func AppSetBeforeDrawFunc(scope sabre.Scope) interface{} {
	return func(app *tview.Application, cb sabre.Invokable) (sabre.Value, error) {

		callBack := func(screen tcell.Screen) bool {
			val, err := cb.Invoke(scope, sabre.ValueOf(screen))
			if err != nil {
				panic(err)
			}
			return bool(val.(sabre.Bool))
		}

		return sabre.ValueOf(app.SetBeforeDrawFunc(callBack)), nil
	}
}

func AppSetInputCapture(scope sabre.Scope) interface{} {
	return func(app *tview.Application, cb sabre.Invokable) (sabre.Value, error) {

		callBack := func(e *tcell.EventKey) *tcell.EventKey {
			val, err := cb.Invoke(scope, sabre.ValueOf(e))
			if err != nil {
				panic(err)
			}
			result := val.(sabre.Any).V.Interface()
			return result.(*tcell.EventKey)
		}

		return sabre.ValueOf(app.SetInputCapture(callBack)), nil
	}
}
