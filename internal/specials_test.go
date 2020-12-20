package internal_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/issadarkthing/spirit/internal"
)

const src = `
(def temp (let [pi 3.1412]
			pi))

(def hello (fn* hello
	([arg] arg)
	([arg & rest] rest)))
`

func TestSpecials(t *testing.T) {
	scope := internal.NewSpirit()

	expected := internal.MultiFn{
		Name:    "hello",
		IsMacro: false,
		Methods: []internal.Fn{
			{
				Args:     []string{"arg", "rest"},
				Variadic: true,
				Body: internal.Module{
					internal.Symbol{Value: "rest"},
				},
			},
		},
	}

	res, err := internal.ReadEvalStr(scope, src)
	if err != nil {
		t.Errorf("Eval() unexpected error: %v", err)
	}
	if reflect.DeepEqual(res, expected) {
		t.Errorf("Eval() expected=%v, got=%v", expected, res)
	}
}

func TestDot(t *testing.T) {
	t.Parallel()

	table := []struct {
		name    string
		src     string
		want    internal.Value
		wantErr bool
	}{
		{
			name: "StringFieldAccess",
			src:  "foo.Name",
			want: internal.String("Bob"),
		},
		{
			name: "BoolFieldAccess",
			src:  "foo.Enabled",
			want: internal.Bool(false),
		},
		{
			name: "MethodAccess",
			src:  `(foo.Bar "Baz")`,
			want: internal.String("Bar(\"Baz\")"),
		},
		{
			name: "MethodAccessPtr",
			src:  `(foo.BarPtr "Bob")`,
			want: internal.String("BarPtr(\"Bob\")"),
		},
		{
			name:    "EvalFailed",
			src:     `blah.BarPtr`,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "NonExistentMember",
			src:     `foo.Baz`,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "PrivateMember",
			src:     `foo.privateMember`,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			scope := internal.NewSpirit()
			scope.BindGo("foo", &Foo{
				Name: "Bob",
			})

			form, err := internal.NewReader(strings.NewReader(tt.src)).All()
			if err != nil {
				t.Fatalf("failed to read source='%s': %+v", tt.src, err)
			}

			got, err := internal.Eval(scope, form)
			if (err != nil) != tt.wantErr {
				t.Errorf("Eval() unexpected error: %+v", err)
			}
			if !reflect.DeepEqual(tt.want, got) {
				t.Errorf("Eval() want=%#v, got=%#v", tt.want, got)
			}
		})
	}
}

// Foo is a dummy type for member access tests.
type Foo struct {
	Name          string
	Enabled       bool
	privateMember bool
}

func (foo *Foo) BarPtr(arg string) string {
	return fmt.Sprintf("BarPtr(\"%s\")", arg)
}

func (foo Foo) Bar(arg string) string {
	return fmt.Sprintf("Bar(\"%s\")", arg)
}
