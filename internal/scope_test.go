package internal_test

import (
	"reflect"
	"testing"

	"github.com/issadarkthing/spirit/internal"
)

var _ internal.Scope = (*internal.MapScope)(nil)

func TestMapScope_Resolve(t *testing.T) {
	tests := []struct {
		name     string
		symbol   string
		getScope func() *internal.MapScope
		want     internal.Value
		wantErr  bool
	}{
		{
			name:   "WithBinding",
			symbol: "hello",
			getScope: func() *internal.MapScope {
				scope := internal.NewScope(nil)
				_ = scope.Bind("hello", internal.String("Hello World!"))
				return scope
			},
			want: internal.String("Hello World!"),
		},
		{
			name:   "WithBindingInParent",
			symbol: "pi",
			getScope: func() *internal.MapScope {
				parent := internal.NewScope(nil)
				_ = parent.Bind("pi", internal.Number(3.1412))
				return internal.NewScope(parent)
			},
			want: internal.Number(3.1412),
		},
		{
			name:   "WithNoBinding",
			symbol: "hello",
			getScope: func() *internal.MapScope {
				return internal.NewScope(nil)
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scope := tt.getScope()

			got, err := scope.Resolve(tt.symbol)
			if (err != nil) != tt.wantErr {
				t.Errorf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Resolve() got = %v, want %v", got, tt.want)
			}
		})
	}
}
