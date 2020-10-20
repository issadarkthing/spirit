package spirit

import (
	"math"
	"strings"

	"github.com/issadarkthing/spirit/internal"
)

// BindAll binds all core functions into the given scope.
func bindAll(scope internal.Scope) error {
	core := map[string]internal.Value{

		// built-in
		"core/range": internal.ValueOf(slangRange),
		"core/future*": &internal.Fn{
			Args:     []string{"body"},
			Variadic: true,
			Func:     future,
		},
		"core/assoc":      internal.ValueOf(assoc),
		"core/keyword":    internal.ValueOf(keyword),
		"core/parse-json": internal.ValueOf(parsejson),

		"core/time": &internal.Fn{
			Args:     []string{"body"},
			Variadic: true,
			Func:     xlispTime,
		},
		"core/bounded?": internal.ValueOf(bound(scope)),
		"core/sleep":    internal.ValueOf(sleep),
		"core/deref*":   internal.ValueOf(deref(scope)),
		"core/doseq": &internal.Fn{
			Args:     []string{"vector", "exprs"},
			Variadic: true,
			Func:     doSeq,
		},
		"core/<>": &internal.Fn{
			Args: []string{"fn"},
			Variadic: true,
			Func: apply,
		},

		"unsafe/swap": &internal.Fn{
			Args: []string{"vector", "exprs"},
			Func: swap,
		},

		"core/atom": internal.ValueOf(internal.NewAtom),
		"core/swap!": &internal.Fn{
			Args: []string{"atom", "fn"},
			Func: safeSwap,
		},

		"core/and*": internal.ValueOf(and),
		"core/or*":  internal.ValueOf(or),
		"core/->": &internal.Fn{
			Args:     []string{"exprs"},
			Func:     threadFirst,
			Variadic: true,
		},
		"core/->>": &internal.Fn{
			Args:     []string{"exprs"},
			Func:     threadLast,
			Variadic: true,
		},
		"core/case": &internal.Fn{
			Args:     []string{"exprs", "clauses"},
			Func:     caseForm,
			Variadic: true,
		},
		"core/loop": internal.SpecialForm{
			Name:  "loop",
			Parse: parseLoop,
		},

		// special forms
		"core/do":           internal.Do,
		"core/def":          internal.Def,
		"core/if":           internal.If,
		"core/fn*":          internal.Lambda,
		"core/macro*":       internal.Macro,
		"core/let":          internal.Let,
		"core/quote":        internal.SimpleQuote,
		"core/syntax-quote": internal.SyntaxQuote,
		"core/recur":        internal.Recur,

		"core/macroexpand": internal.ValueOf(macroExpand),
		"core/eval":        internal.ValueOf(internal.Eval),
		"core/eval-string": internal.ValueOf(internal.ReadEvalStr),
		"core/type":        internal.ValueOf(typeOf),
		"core/to-type":     internal.ValueOf(toType),
		"core/impl?":       internal.ValueOf(implements),
		"core/realized*":   internal.ValueOf(futureRealize),
		"core/throw":       internal.ValueOf(throw),
		"core/substring":   internal.ValueOf(strings.Contains),
		"core/trim-suffix": internal.ValueOf(strings.TrimSuffix),
		"core/resolve":     internal.ValueOf(resolve(scope)),

		// Type system functions
		"core/str": internal.ValueOf(makeString),

		// Math functions
		"core/+":   internal.ValueOf(add),
		"core/-":   internal.ValueOf(sub),
		"core/*":   internal.ValueOf(multiply),
		"core//":   internal.ValueOf(divide),
		"core/mod": internal.ValueOf(math.Mod),
		"core/=":   internal.ValueOf(internal.Compare),
		"core/>":   internal.ValueOf(gt),
		"core/>=":  internal.ValueOf(gtE),
		"core/<":   internal.ValueOf(lt),
		"core/<=":  internal.ValueOf(ltE),

		// io functions
		"core/$":         internal.ValueOf(shell),
		"core/print":     internal.ValueOf(println),
		"core/printf":    internal.ValueOf(printf),
		"core/read*":     internal.ValueOf(read),
		"core/random":    internal.ValueOf(random),
		"core/shuffle":   internal.ValueOf(shuffle),
		"core/read-file": internal.ValueOf(readFile),

		"string/split": internal.ValueOf(splitString),

		"types/Seq":       typeOf((*internal.Seq)(nil)),
		"types/Invokable": typeOf((*internal.Invokable)(nil)),
	}

	for sym, val := range core {
		if err := scope.Bind(sym, val); err != nil {
			return err
		}
	}

	return nil
}
