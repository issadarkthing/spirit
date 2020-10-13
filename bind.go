package spirit

import (
	"math"
	"strings"

	"github.com/issadarkthing/spirit/internal"
)

// BindAll binds all core functions into the given scope.
func BindAll(scope internal.Scope) error {
	core := map[string]internal.Value{

		// built-in
		"core/range": internal.ValueOf(slangRange),
		"core/future*": &internal.Fn{
			Args:     []string{"body"},
			Variadic: true,
			Func:     future,
		},

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

		"unsafe/swap": &internal.Fn{
			Args: []string{"vector", "exprs"},
			Func: swap,
		},

		"core/atom": internal.ValueOf(newAtom),
		"core/swap!": &internal.Fn{
			Args: []string{"atom", "fn"},
			Func: safeSwap,
		},

		"core/and*": internal.ValueOf(and),
		"core/or*":  internal.ValueOf(or),
		"core/->": &internal.Fn{
			Args:     []string{"exprs"},
			Func:     ThreadFirst,
			Variadic: true,
		},
		"core/->>": &internal.Fn{
			Args:     []string{"exprs"},
			Func:     ThreadLast,
			Variadic: true,
		},
		"core/case": &internal.Fn{
			Args:     []string{"exprs", "clauses"},
			Func:     Case,
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

		"core/macroexpand": internal.ValueOf(MacroExpand),
		"core/eval":        internal.ValueOf(internal.Eval),
		"core/eval-string": internal.ValueOf(internal.ReadEvalStr),
		"core/type":        internal.ValueOf(TypeOf),
		"core/to-type":     internal.ValueOf(ToType),
		"core/impl?":       internal.ValueOf(Implements),
		"core/realized*":   internal.ValueOf(futureRealize),
		"core/throw":       internal.ValueOf(Throw),
		"core/substring":   internal.ValueOf(strings.Contains),
		"core/trim-suffix": internal.ValueOf(strings.TrimSuffix),
		"core/resolve":     internal.ValueOf(resolve(scope)),

		// Type system functions
		"core/str": internal.ValueOf(MakeString),

		// Math functions
		"core/+":   internal.ValueOf(Add),
		"core/-":   internal.ValueOf(Sub),
		"core/*":   internal.ValueOf(Multiply),
		"core//":   internal.ValueOf(Divide),
		"core/mod": internal.ValueOf(math.Mod),
		"core/=":   internal.ValueOf(internal.Compare),
		"core/>":   internal.ValueOf(Gt),
		"core/>=":  internal.ValueOf(GtE),
		"core/<":   internal.ValueOf(Lt),
		"core/<=":  internal.ValueOf(LtE),

		// io functions
		"core/$":         internal.ValueOf(Shell),
		"core/print":     internal.ValueOf(Println),
		"core/printf":    internal.ValueOf(Printf),
		"core/read*":     internal.ValueOf(Read),
		"core/random":    internal.ValueOf(Random),
		"core/shuffle":   internal.ValueOf(Shuffle),
		"core/read-file": internal.ValueOf(ReadFile),

		"string/split": internal.ValueOf(splitString),

		"types/Seq":       TypeOf((*internal.Seq)(nil)),
		"types/Invokable": TypeOf((*internal.Invokable)(nil)),
	}

	for sym, val := range core {
		if err := scope.Bind(sym, val); err != nil {
			return err
		}
	}

	return nil
}
