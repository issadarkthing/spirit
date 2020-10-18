package internal_test

import (
	"reflect"
	"testing"

	"github.com/issadarkthing/spirit/internal"
	"github.com/kr/pretty"
)

type stackTestCase struct {
	name     string
	getScope func() internal.Scope
	src      string
	want     internal.Stack
}

func TestStack_Eval(t *testing.T) {
	executeStackTests(t, []stackTestCase{
		{
			name: "Simple",
			getScope: func() internal.Scope {
				return internal.New()
			},
			src: "(do (if true (bruh)))",
			want: internal.Stack{
				internal.Call{
					Name: "do",
					Position: internal.Position{
						File: "<string>",
						Line: 1,
						Column: 1,
					},
				},
				internal.Call{
					Name: "if",
					Position: internal.Position{
						File:  "<string>",
						Line: 1,
						Column: 5,
					},
				},
			},
		},
		{
			name: "MultipleCalls",
			getScope: func() internal.Scope {
				return internal.New()
			},
			src: `(do (def x 100) 
					(if true (bruh)))`,
			want: internal.Stack{
				internal.Call{
					Name: "do",
					Position: internal.Position{
						File: "<string>",
						Line: 1,
						Column: 1,
					},
				},
				internal.Call{
					Name: "if",
					Position: internal.Position{
						File:  "<string>",
						Line: 2,
						Column: 6,
					},
				},
			},
		},
		{
			name: "NestedMultiLine",
			getScope: func() internal.Scope {
				return internal.New()
			},
			want: internal.Stack{
				internal.Call{
					Name: "do",
					Position: internal.Position{
						File: "<string>",
						Line: 1,
						Column: 1,
					},
				},
				internal.Call{
					Name: "do",
					Position: internal.Position{
						File: "<string>",
						Line: 2,
						Column: 6,
					},
				},
				internal.Call{
					Name: "if",
					Position: internal.Position{
						File: "<string>",
						Line: 3,
						Column: 8,
					},
				},
			},
			src: `(do
					(do
					  (if true
						(case))))`,
		},
	})
}

func executeStackTests(t *testing.T, tests []stackTestCase) {
	t.Parallel()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			scope, _ := tt.getScope().(*internal.MapScope)
			internal.ReadEvalStr(scope, tt.src)

			if scope.Size() != len(tt.want) {
				t.Errorf("Mismatch stack size, expected %d, got %d", 
					 len(tt.want), scope.Size(),
				)				
				return
			}

			for i, want := range tt.want {
				got := scope.Stack[i]
				if !reflect.DeepEqual(got, want) {
					t.Errorf("Mismatch stack item, \nexpected \n%s, \ngot \n%s",
						pretty.Sprint(want), pretty.Sprint(got),
					)
				}
			}

		})
	}
}

