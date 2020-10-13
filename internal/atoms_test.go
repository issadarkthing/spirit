package internal_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/issadarkthing/spirit/internal"
)

var _ internal.Seq = internal.String("")

func TestBool_Eval(t *testing.T) {
	executeEvalTests(t, []evalTestCase{
		{
			name:     "Success",
			getScope: nil,
			value:    internal.Bool(true),
			want:     internal.Bool(true),
		},
	})
}

func TestNil_Eval(t *testing.T) {
	executeEvalTests(t, []evalTestCase{
		{
			name:     "Success",
			getScope: nil,
			value:    internal.Nil{},
			want:     internal.Nil{},
		},
	})
}

func TestString_Eval(t *testing.T) {
	executeEvalTests(t, []evalTestCase{
		{
			name:     "Success",
			getScope: nil,
			value:    internal.String("hello"),
			want:     internal.String("hello"),
		},
	})
}

func TestKeyword_Eval(t *testing.T) {
	executeEvalTests(t, []evalTestCase{
		{
			name:     "Success",
			getScope: nil,
			value:    internal.Keyword("hello"),
			want:     internal.Keyword("hello"),
		},
	})
}

func TestSymbol_Eval(t *testing.T) {
	executeEvalTests(t, []evalTestCase{
		{
			name: "Success",
			getScope: func() internal.Scope {
				scope := internal.NewScope(nil)
				scope.Bind("hello", internal.String("world"))

				return scope
			},
			value: internal.Symbol{Value: "hello"},
			want:  internal.String("world"),
		},
	})
}

func TestCharacter_Eval(t *testing.T) {
	executeEvalTests(t, []evalTestCase{
		{
			name:     "Success",
			getScope: nil,
			value:    internal.Character('a'),
			want:     internal.Character('a'),
		},
	})
}

func TestNil_String(t *testing.T) {
	executeStringTestCase(t, []stringTestCase{
		{
			value: internal.Nil{},
			want:  "nil",
		},
	})
}

func TestInt64_String(t *testing.T) {
	executeStringTestCase(t, []stringTestCase{
		{
			value: internal.Int64(10),
			want:  "10",
		},
		{
			value: internal.Int64(-10),
			want:  "-10",
		},
	})
}

func TestFloat64_String(t *testing.T) {
	executeStringTestCase(t, []stringTestCase{
		{
			value: internal.Float64(10.3),
			want:  "10.300000",
		},
		{
			value: internal.Float64(-10.3),
			want:  "-10.300000",
		},
	})
}

func TestBool_String(t *testing.T) {
	executeStringTestCase(t, []stringTestCase{
		{
			value: internal.Bool(true),
			want:  "true",
		},
		{
			value: internal.Bool(false),
			want:  "false",
		},
	})
}

func TestKeyword_String(t *testing.T) {
	executeStringTestCase(t, []stringTestCase{
		{
			value: internal.Keyword("hello"),
			want:  ":hello",
		},
	})
}

func TestSymbol_String(t *testing.T) {
	executeStringTestCase(t, []stringTestCase{
		{
			value: internal.Symbol{Value: "hello"},
			want:  "hello",
		},
	})
}

func TestCharacter_String(t *testing.T) {
	executeStringTestCase(t, []stringTestCase{
		{
			value: internal.Character('a'),
			want:  "\\a",
		},
	})
}

func TestString_String(t *testing.T) {
	executeStringTestCase(t, []stringTestCase{
		{
			value: internal.String("hello world"),
			want:  `"hello world"`,
		},
		{
			value: internal.String("hello\tworld"),
			want: `"hello	world"`,
		},
	})
}

type stringTestCase struct {
	value internal.Value
	want  string
}

type evalTestCase struct {
	name     string
	getScope func() internal.Scope
	value    internal.Value
	want     internal.Value
	wantErr  bool
}

func executeStringTestCase(t *testing.T, tests []stringTestCase) {
	t.Parallel()

	for _, tt := range tests {
		t.Run(reflect.TypeOf(tt.value).Name(), func(t *testing.T) {
			got := strings.TrimSpace(tt.value.String())
			if got != tt.want {
				t.Errorf("String() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func executeEvalTests(t *testing.T, tests []evalTestCase) {
	t.Parallel()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var scope internal.Scope
			if tt.getScope != nil {
				scope = tt.getScope()
			}

			got, err := internal.Eval(scope, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Eval() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Eval() got = %v, want %v", got, tt.want)
			}
		})
	}
}
