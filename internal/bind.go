package internal

import (
	"math"
	"strings"
)

// BindAll binds all core functions into the given scope.
func bindAll(scope Scope) error {
	core := map[string]Value{

		// built-in
		"core/lazy-range*": ValueOf(lazyRange),
		"core/future*": &Fn{
			Args:     []string{"body"},
			Variadic: true,
			Func:     future,
		},
		"core/assoc*":     ValueOf(assoc),
		"core/keyword":    ValueOf(keyword),
		"core/parse-json": ValueOf(parsejson),
		"core/round":      ValueOf(math.Round),

		"core/time": &Fn{
			Args:     []string{"body"},
			Variadic: true,
			Func:     xlispTime,
		},
		"core/bounded?": ValueOf(bound(scope)),
		"core/sleep":    ValueOf(sleep),
		"core/deref":    ValueOf(deref(scope)),
		"core/doseq": &Fn{
			Args:     []string{"vector", "exprs"},
			Variadic: true,
			Func:     doSeq,
		},
		"core/<>": &Fn{
			Args:     []string{"fn"},
			Variadic: true,
			Func:     apply,
		},

		"unsafe/swap": &Fn{
			Args: []string{"vector", "exprs"},
			Func: swap,
		},

		"core/atom": ValueOf(NewAtom),
		"core/swap!": &Fn{
			Args: []string{"atom", "fn"},
			Func: safeSwap,
		},

		"core/and*": ValueOf(and),
		"core/or*":  ValueOf(or),
		"core/case": &Fn{
			Args:     []string{"exprs", "clauses"},
			Func:     caseForm,
			Variadic: true,
		},
		"core/eval": &Fn{
			Args: []string{"exprs"},
			Func: eval,
		},
		"core/eval-string": &Fn{
			Args: []string{"exprs"},
			Func: evalStr,
		},
		"core/loop": SpecialForm{
			Name:  "loop",
			Parse: parseLoop,
		},
		"core/defclass": &Fn{
			Args: []string{"hash-map", "methods"},
			Func: defClass,
		},
		"core/mem": &Fn{
			Args:     []string{"exprs"},
			Variadic: true,
			Func:     mem,
		},
		"core/recur": &Fn{
			Args:     []string{"bindings"},
			Variadic: true,
			Func:     recur,
		},

        "core/exception": ValueOf(Exception{}),

		// special forms
		"core/do":           Do,
		"core/def":          Def,
		"core/if":           If,
		"core/fn*":          Lambda,
		"core/macro*":       Macro,
		"core/let":          Let,
		"core/try":          Try,
		"core/quote":        SimpleQuote,
		"core/syntax-quote": SyntaxQuote,

		"core/in-ns":       ValueOf(scope.(*Spirit).SwitchNS),
		"core/memory":      ValueOf(memory),
		"core/macroexpand": ValueOf(macroExpand),
		"core/type":        ValueOf(typeOf),
		"core/to-type":     ValueOf(toType),
		"core/impl?":       ValueOf(implements),
		"core/realized*":   ValueOf(futureRealize),
		"core/throw":       ValueOf(throw),
		"core/error-is":    ValueOf(errorIs),
		"core/substring":   ValueOf(strings.Contains),
		"core/trim-suffix": ValueOf(strings.TrimSuffix),
		"core/resolve":     ValueOf(resolve(scope)),
		"core/force-gc":    ValueOf(forceGC),

		// Type system functions
		"core/str": ValueOf(makeString),

		// Math functions
		"core/+":      ValueOf(add),
		"core/-":      ValueOf(sub),
		"core/*":      ValueOf(multiply),
		"core//":      ValueOf(divide),
		"core/mod":    ValueOf(math.Mod),
		"core/=":      ValueOf(Compare),
		"core/>":      ValueOf(gt),
		"core/>=":     ValueOf(gtE),
		"core/<":      ValueOf(lt),
		"core/<=":     ValueOf(ltE),
		"core/sqrt":   ValueOf(math.Sqrt),
		"core/prime?": ValueOf(isPrime),

		// io functions
		"core/$":         ValueOf(shell),
		"core/print":     ValueOf(println),
		"core/printf":    ValueOf(printf),
		"core/pprint":    ValueOf(pprint),
		"core/read*":     ValueOf(read),
		"core/random":    ValueOf(random),
		"core/shuffle":   ValueOf(shuffle),
		"core/read-file": ValueOf(readFile),
		"core/import":    ValueOf(spiritImport(scope)),

		"core/split": ValueOf(strings.Split),
		"core/trim":  ValueOf(strings.Trim),

		"types/Seq":       typeOf((*Seq)(nil)),
		"types/Invokable": typeOf((*Invokable)(nil)),
		"types/Assoc":     typeOf((*Assoc)(nil)),
	}

	for sym, val := range core {
		if err := scope.Bind(sym, val); err != nil {
			return err
		}
	}

	return nil
}
