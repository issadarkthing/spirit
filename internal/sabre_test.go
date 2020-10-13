package internal_test

import (
	"reflect"
	"testing"

	"github.com/issadarkthing/spirit/internal"
)

const sampleProgram = `
(def v [1 2 3])
(def pi 3.1412)
(def echo (fn* [arg] arg))
(echo pi)

(def int-num 10)
(def float-num 10.1234)
(def list '(nil 1 []))
(def vector ["hello" nil])
(def set #{1 2 3})
(def empty-set #{})

(def complex-calc (let* [sample '(1 2 3 4 [])]
					(sample.First)))

(assert (= int-num 10)
		(= float-num 10.1234)
		(= pi 3.1412)
		(= list '(nil 1 []))
		(= vector ["hello" nil])
		(= empty-set #{})
		(= echo (fn* [arg] arg))
		(= complex-calc 1))

(echo pi)
`

func BenchmarkEval(b *testing.B) {
	scope := internal.NewScope(nil)
	_ = scope.BindGo("inc", func(a int) int {
		return a + 1
	})

	f := &internal.List{
		Values: internal.Values{
			internal.Symbol{Value: "inc"},
			internal.Int64(10),
		},
	}

	for i := 0; i < b.N; i++ {
		_, _ = internal.Eval(scope, f)
	}
}

func BenchmarkGoCall(b *testing.B) {
	inc := func(a int) int {
		return a + 1
	}

	for i := 0; i < b.N; i++ {
		_ = inc(10)
	}
}

func TestEval(t *testing.T) {
	t.Parallel()

	table := []struct {
		name     string
		src      string
		getScope func() internal.Scope
		want     internal.Value
		wantErr  bool
	}{
		{
			name: "Empty",
			src:  "",
			want: internal.Nil{},
		},
		{
			name: "SingleForm",
			src:  "123",
			want: internal.Int64(123),
		},
		{
			name: "MultiForm",
			src:  `123 [] ()`,
			want: &internal.List{
				Values:   internal.Values(nil),
				Position: internal.Position{File: "<string>", Line: 1, Column: 8},
			},
		},
		{
			name: "WithFunctionCalls",
			getScope: func() internal.Scope {
				scope := internal.NewScope(nil)
				_ = scope.BindGo("ten?", func(i internal.Int64) bool {
					return i == 10
				})
				return scope
			},
			src:  `(ten? 10)`,
			want: internal.Bool(true),
		},
		{
			name:    "ReadError",
			src:     `123 [] (`,
			want:    nil,
			wantErr: true,
		},
		{
			name: "Program",
			src:  sampleProgram,
			want: internal.Float64(3.1412),
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			scope := internal.Scope(internal.New())
			if tt.getScope != nil {
				scope = tt.getScope()
			}

			scope.Bind("=", internal.ValueOf(internal.Compare))
			scope.Bind("assert", &internal.Fn{Func: asserter(t)})

			got, err := internal.ReadEvalStr(scope, tt.src)
			if (err != nil) != tt.wantErr {
				t.Errorf("Eval() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Eval() got = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func asserter(t *testing.T) func(internal.Scope, []internal.Value) (internal.Value, error) {
	return func(scope internal.Scope, exprs []internal.Value) (internal.Value, error) {
		var res internal.Value
		var err error

		for _, expr := range exprs {
			res, err = expr.Eval(scope)
			if err != nil {
				t.Errorf("%s: %s", expr, err)
			}

			if !isTruthy(res) {
				t.Errorf("assertion failed: %s (result=%v)", expr, res)
			}
		}

		return res, err
	}
}

func isTruthy(v internal.Value) bool {
	if v == nil || v == (internal.Nil{}) {
		return false
	}
	if b, ok := v.(internal.Bool); ok {
		return bool(b)
	}
	return true
}
