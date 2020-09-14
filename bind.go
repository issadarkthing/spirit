package xlisp

import (
	"math"

	"github.com/rivo/tview"
	"github.com/spy16/sabre"
)

// BindAll binds all core functions into the given scope.
func BindAll(scope sabre.Scope) error {
	core := map[string]sabre.Value{
		// gui frontend
		"tview/new-app":       sabre.ValueOf(tview.NewApplication),
		"tview/new-form":      sabre.ValueOf(tview.NewForm),
		"tview/new-box":       sabre.ValueOf(tview.NewBox),
		"tview/new-textview":  sabre.ValueOf(tview.NewTextView),
		"tview/new-list":      sabre.ValueOf(tview.NewList),
		"tview/list-add-item": sabre.ValueOf(ListAddItem(scope)),

		// built-in
		"core/range": sabre.ValueOf(slangRange),
		"core/future*": &sabre.Fn{
			Args: []string{"body"},
			Variadic: true,
			Func: future,
		},

		"core/sleep": sabre.ValueOf(sleep),
		"core/deref*": sabre.ValueOf(deref(scope)), 

		"core/doseq": &sabre.Fn{
			Args:     []string{"vector", "exprs"},
			Variadic: true,
			Func:     doSeq,
		},

		"unsafe/swap": &sabre.Fn{
			Args: []string{"vector", "exprs"},
			Func: swap,
		},

		"core/->": &sabre.Fn{
			Args:     []string{"exprs"},
			Func:     ThreadFirst,
			Variadic: true,
		},
		"core/->>": &sabre.Fn{
			Args:     []string{"exprs"},
			Func:     ThreadLast,
			Variadic: true,
		},
		"core/case": &sabre.Fn{
			Args:     []string{"exprs", "clauses"},
			Func:     Case,
			Variadic: true,
		},

		// special forms
		"core/do":           sabre.Do,
		"core/def":          sabre.Def,
		"core/if":           sabre.If,
		"core/fn*":          sabre.Lambda,
		"core/macro*":       sabre.Macro,
		"core/let":          sabre.Let,
		"core/quote":        sabre.SimpleQuote,
		"core/syntax-quote": sabre.SyntaxQuote,
		"core/recur":        sabre.Recur,

		"core/macroexpand": sabre.ValueOf(MacroExpand),
		"core/eval":        sabre.ValueOf(sabre.Eval),
		"core/type":        sabre.ValueOf(TypeOf),
		"core/to-type":     sabre.ValueOf(ToType),
		"core/impl?":       sabre.ValueOf(Implements),
		"core/realized*":   sabre.ValueOf(futureRealize),
		"core/throw":       sabre.ValueOf(Throw),

		// Type system functions
		"core/str": sabre.ValueOf(MakeString),

		// Math functions
		"core/+":   sabre.ValueOf(Add),
		"core/-":   sabre.ValueOf(Sub),
		"core/*":   sabre.ValueOf(Multiply),
		"core//":   sabre.ValueOf(Divide),
		"core/mod": sabre.ValueOf(math.Mod),
		"core/=":   sabre.ValueOf(sabre.Compare),
		"core/>":   sabre.ValueOf(Gt),
		"core/>=":  sabre.ValueOf(GtE),
		"core/<":   sabre.ValueOf(Lt),
		"core/<=":  sabre.ValueOf(LtE),

		// io functions
		"core/print":   sabre.ValueOf(Println),
		"core/printf":  sabre.ValueOf(Printf),
		"core/read*":   sabre.ValueOf(Read),
		"core/random":  sabre.ValueOf(Random),
		"core/shuffle": sabre.ValueOf(Shuffle),

		"types/Seq":       TypeOf((*sabre.Seq)(nil)),
		"types/Invokable": TypeOf((*sabre.Invokable)(nil)),
	}

	for sym, val := range core {
		if err := scope.Bind(sym, val); err != nil {
			return err
		}
	}

	return nil
}
