package internal_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/issadarkthing/spirit/internal"
)

var (
	_ internal.Seq = &internal.List{}
	_ internal.Seq = internal.Set{}
)

func TestList_Eval(t *testing.T) {
	executeEvalTests(t, []evalTestCase{
		{
			name:  "EmptyList",
			value: &internal.List{},
			want:  &internal.List{},
		},
		{
			name: "Invocation",
			value: &internal.List{
				Values: []internal.Value{internal.Symbol{Value: "greet"}, internal.String("Bob")},
			},
			getScope: func() internal.Scope {
				scope := internal.NewScope(nil)
				scope.BindGo("greet", func(name internal.String) string {
					return fmt.Sprintf("Hello %s!", string(name))
				})
				return scope
			},
			want: internal.String("Hello Bob!"),
		},
		{
			name: "NonInvokable",
			value: &internal.List{
				Values: []internal.Value{internal.Number(10), internal.Keyword("hello")},
			},
			wantErr: true,
		},
		{
			name: "EvalFailure",
			value: &internal.List{
				Values: []internal.Value{internal.Symbol{Value: "hello"}},
			},
			getScope: func() internal.Scope {
				return internal.NewScope(nil)
			},
			wantErr: true,
		},
	})
}

func TestModule_Eval(t *testing.T) {
	executeEvalTests(t, []evalTestCase{
		{
			name:  "NilModule",
			value: internal.Module(nil),
			want:  internal.Nil{},
		},
		{
			name:  "EmptyModule",
			value: internal.Module{},
			want:  internal.Nil{},
		},
		{
			name:  "SingleForm",
			value: internal.Module{internal.Number(10)},
			want:  internal.Number(10),
		},
		{
			name: "MultiForm",
			value: internal.Module{
				internal.Number(10),
				internal.String("hello"),
			},
			want: internal.String("hello"),
		},
		{
			name:     "Failure",
			getScope: func() internal.Scope { return internal.NewScope(nil) },
			value: internal.Module{
				internal.Symbol{Value: "hello"},
			},
			wantErr: true,
		},
	})
}

func TestVector_Eval(t *testing.T) {
	executeEvalTests(t, []evalTestCase{
		{
			name:  "EmptyVector",
			value: internal.NewPersistentVector(),
			want:  internal.NewPersistentVector(),
		},
		{
			name: "EvalFailure",
			getScope: func() internal.Scope {
				return internal.NewScope(nil)
			},
			value:   internal.NewPersistentVector().Cons(internal.Symbol{Value: "hello"}),
			wantErr: true,
		},
		{
			name: "Immutability",
			getScope: func() internal.Scope {
				list := internal.NewPersistentVector()
				list.Cons(internal.Keyword("hello"))
				scope := internal.NewScope(nil)
				scope.Bind("ls", list)
				return scope
			},
			value: internal.Symbol{Value: "ls"},
			want: internal.NewPersistentVector(),
		},
	})
}

func TestSet_Eval(t *testing.T) {
	executeEvalTests(t, []evalTestCase{
		{
			name:  "Empty",
			value: internal.Set{},
			want:  internal.Set{},
		},
		{
			name: "ValidWithoutDuplicates",
			getScope: func() internal.Scope {
				return internal.NewScope(nil)
			},
			value: internal.Set{Values: []internal.Value{internal.String("hello")}},
			want:  internal.Set{Values: []internal.Value{internal.String("hello")}},
		},
		{
			name: "ValidWithtDuplicates",
			getScope: func() internal.Scope {
				return internal.NewScope(nil)
			},
			value: internal.Set{Values: []internal.Value{
				internal.String("hello"),
				internal.String("hello"),
			}},
			want: internal.Set{Values: []internal.Value{internal.String("hello")}},
		},
		{
			name: "Failure",
			getScope: func() internal.Scope {
				return internal.NewScope(nil)
			},
			value:   internal.Set{Values: []internal.Value{internal.Symbol{Value: "hello"}}},
			wantErr: true,
		},
	})
}

func TestList_String(t *testing.T) {
	executeStringTestCase(t, []stringTestCase{
		{
			value: &internal.List{},
			want:  "()",
		},
		{
			value: &internal.List{
				Values: []internal.Value{internal.Keyword("hello")},
			},
			want: "(:hello)",
		},
		{
			value: &internal.List{
				Values: []internal.Value{internal.Keyword("hello"), &internal.List{}},
			},
			want: "(:hello ())",
		},
		{
			value: &internal.List{
				Values: []internal.Value{internal.Symbol{Value: "quote"}, internal.Symbol{Value: "hello"}},
			},
			want: "(quote hello)",
		},
		{
			value: &internal.List{
				Values: []internal.Value{
					internal.Symbol{Value: "quote"},
					&internal.List{Values: []internal.Value{internal.Symbol{Value: "hello"}}}},
			},
			want: "(quote (hello))",
		},
	})
}

func TestVector_String(t *testing.T) {
	executeStringTestCase(t, []stringTestCase{
		{
			value: internal.NewPersistentVector(),
			want:  "[]",
		},
		{
			value: internal.NewPersistentVector().
			Cons(internal.Keyword("hello")),
			want:  "[:hello]",
		},
		{
			value: internal.NewPersistentVector().
			Cons(internal.Keyword("hello")).
			Cons(&internal.List{}),
			want:  "[:hello ()]",
		},
	})
}

func TestModule_String(t *testing.T) {
	executeStringTestCase(t, []stringTestCase{
		{
			value: internal.Module(nil),
			want:  "",
		},
		{
			value: internal.Module{internal.Symbol{Value: "hello"}},
			want:  "hello",
		},
		{
			value: internal.Module{internal.Symbol{Value: "hello"}, internal.Keyword("world")},
			want:  "hello\n:world",
		},
	})
}

func TestVector_Invoke(t *testing.T) {
	t.Parallel()

	vector := internal.NewPersistentVector().Cons(internal.Keyword("hello"))

	table := []struct {
		name     string
		getScope func() internal.Scope
		args     []internal.Value
		want     internal.Value
		wantErr  bool
	}{
		{
			name:    "NoArgs",
			args:    []internal.Value{},
			wantErr: true,
		},
		{
			name:    "InvalidIndex",
			args:    []internal.Value{internal.Number(10)},
			wantErr: true,
		},
		{
			name:    "ValidIndex",
			args:    []internal.Value{internal.Number(0)},
			want:    internal.Keyword("hello"),
			wantErr: false,
		},
		{
			name:    "NonIntegerArg",
			args:    []internal.Value{internal.Keyword("h")},
			wantErr: true,
		},
		{
			name: "EvalFailure",
			getScope: func() internal.Scope {
				return internal.NewScope(nil)
			},
			args:    []internal.Value{internal.Symbol{Value: "hello"}},
			wantErr: true,
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			var scope internal.Scope
			if tt.getScope != nil {
				scope = tt.getScope()
			}

			got, err := vector.(*internal.PersistentVector).Invoke(scope, tt.args...)
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

func TestHashMap_Eval(t *testing.T) {
	executeEvalTests(t, []evalTestCase{
		{
			name: "Simple",
			value: internal.NewPersistentMap().
				Set(internal.Keyword("name"), internal.String("Bob")),
			want: internal.NewPersistentMap().
				Set(internal.Keyword("name"), internal.String("Bob")),
		},
		{
			name: "Immutability",
			getScope: func() internal.Scope {
				scope := internal.NewScope(nil)
				hm := internal.NewPersistentMap()
				hm.Set(internal.Keyword("bob"), internal.Number(12))
				scope.Bind("my-hashmap", hm)
				return scope
			},
			value: internal.Symbol{Value: "my-hashmap"},
			want: internal.NewPersistentMap(),
		},
	})
}

func TestHashMap_String(t *testing.T) {
	executeStringTestCase(t, []stringTestCase{
		{
			value: internal.NewPersistentMap().
				Set(internal.Keyword("name"), internal.String("Bob")),
			want: `{:name "Bob"}`,
		},
	})
}
