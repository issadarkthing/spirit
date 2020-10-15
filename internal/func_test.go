package internal_test

import (
	"reflect"
	"testing"

	"github.com/issadarkthing/spirit/internal"
)

func TestMultiFn_Eval(t *testing.T) {
	executeEvalTests(t, []evalTestCase{
		{
			name:  "Valid",
			value: internal.MultiFn{},
			want:  internal.MultiFn{},
		},
	})
}

func TestMultiFn_String(t *testing.T) {
	executeStringTestCase(t, []stringTestCase{
		{
			value: internal.MultiFn{
				Name: "hello",
			},
			want: "(hello)",
		},
	})
}

func TestMultiFn_Invoke(t *testing.T) {
	t.Parallel()

	table := []struct {
		name     string
		getScope func() internal.Scope
		multiFn  internal.MultiFn
		args     []internal.Value
		want     internal.Value
		wantErr  bool
	}{
		{
			name: "WrongArity",
			multiFn: internal.MultiFn{
				Name: "arityOne",
				Methods: []internal.Fn{
					{
						Args: []string{"arg1"},
					},
				},
			},
			args:    []internal.Value{},
			wantErr: true,
		},
		{
			name: "VariadicArity",
			multiFn: internal.MultiFn{
				Name: "arityMany",
				Methods: []internal.Fn{
					{
						Args:     []string{"args"},
						Variadic: true,
					},
				},
			},
			args: []internal.Value{},
			want: internal.Nil{},
		},
		{
			name:     "ArgEvalFailure",
			getScope: func() internal.Scope { return internal.NewScope(nil) },
			multiFn: internal.MultiFn{
				Name: "arityOne",
				Methods: []internal.Fn{
					{
						Args: []string{"arg1"},
					},
				},
			},
			args:    []internal.Value{internal.Symbol{Value: "argVal"}},
			wantErr: true,
		},
		{
			name: "Macro",
			getScope: func() internal.Scope {
				scope := internal.NewScope(nil)
				scope.Bind("argVal", internal.String("hello"))
				return scope
			},
			multiFn: internal.MultiFn{
				Name:    "arityOne",
				IsMacro: true,
				Methods: []internal.Fn{
					{
						Args: []string{"arg1"},
						Body: internal.Number(10),
					},
				},
			},
			args: []internal.Value{internal.Symbol{Value: "argVal"}},
			want: internal.Number(10),
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			var scope internal.Scope
			if tt.getScope != nil {
				scope = tt.getScope()
			}

			got, err := tt.multiFn.Invoke(scope, tt.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Invoke() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Invoke() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFn_Invoke(t *testing.T) {
	t.Parallel()

	table := []struct {
		name     string
		getScope func() internal.Scope
		fn       internal.Fn
		args     []internal.Value
		want     internal.Value
		wantErr  bool
	}{
		{
			name: "GoFuncWrap",
			fn: internal.Fn{
				Func: func(scope internal.Scope, args []internal.Value) (internal.Value, error) {
					return internal.Number(10), nil
				},
			},
			want: internal.Number(10),
		},
		{
			name: "NoBody",
			fn: internal.Fn{
				Args: []string{"test"},
			},
			args: []internal.Value{internal.Bool(true)},
			want: internal.Nil{},
		},
		{
			name: "VariadicMatch",
			fn: internal.Fn{
				Args:     []string{"test"},
				Variadic: true,
			},
			args: []internal.Value{},
			want: internal.Nil{},
		},
		{
			name: "VariadicMatch",
			fn: internal.Fn{
				Args:     []string{"test"},
				Variadic: true,
			},
			args: []internal.Value{internal.Number(10), internal.Bool(true)},
			want: internal.Nil{},
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			var scope internal.Scope
			if tt.getScope != nil {
				scope = tt.getScope()
			}

			got, err := tt.fn.Invoke(scope, tt.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Invoke() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Invoke() got = %v, want %v", got, tt.want)
			}
		})
	}
}
