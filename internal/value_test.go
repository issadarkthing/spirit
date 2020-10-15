package internal_test

import (
	"reflect"
	"testing"

	"github.com/issadarkthing/spirit/internal"
)

var _ internal.Seq = internal.Values(nil)

func TestValues_First(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		vals := internal.Values{}

		want := internal.Value(nil)
		got := vals.First()

		if !reflect.DeepEqual(got, want) {
			t.Errorf("First() want=%#v, got=%#v", want, got)
		}
	})

	t.Run("Nil", func(t *testing.T) {
		vals := internal.Values(nil)
		want := internal.Value(nil)
		got := vals.First()

		if !reflect.DeepEqual(got, want) {
			t.Errorf("First() want=%#v, got=%#v", want, got)
		}
	})

	t.Run("NonEmpty", func(t *testing.T) {
		vals := internal.Values{internal.Number(10)}

		want := internal.Number(10)
		got := vals.First()

		if !reflect.DeepEqual(got, want) {
			t.Errorf("First() want=%#v, got=%#v", want, got)
		}
	})
}

func TestValues_Next(t *testing.T) {
	t.Parallel()

	table := []struct {
		name string
		vals []internal.Value
		want internal.Seq
	}{
		{
			name: "Nil",
			vals: []internal.Value(nil),
			want: nil,
		},
		{
			name: "Empty",
			vals: []internal.Value{},
			want: nil,
		},
		{
			name: "SingleItem",
			vals: []internal.Value{internal.Number(10)},
			want: nil,
		},
		{
			name: "MultiItem",
			vals: []internal.Value{internal.Number(10), internal.String("hello"), internal.Bool(true)},
			want: &internal.List{Values: internal.Values{internal.String("hello"), internal.Bool(true)}},
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			got := internal.Values(tt.vals).Next()

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Next() want=%#v, got=%#v", tt.want, got)
			}
		})
	}
}

func TestValues_Cons(t *testing.T) {
	t.Parallel()

	table := []struct {
		name string
		vals []internal.Value
		item internal.Value
		want internal.Values
	}{
		{
			name: "Nil",
			vals: []internal.Value(nil),
			item: internal.Number(10),
			want: internal.Values{internal.Number(10)},
		},
		{
			name: "Empty",
			vals: []internal.Value{},
			item: internal.Number(10),
			want: internal.Values{internal.Number(10)},
		},
		{
			name: "SingleItem",
			vals: []internal.Value{internal.Number(10)},
			item: internal.String("hello"),
			want: internal.Values{internal.String("hello"), internal.Number(10)},
		},
		{
			name: "MultiItem",
			vals: []internal.Value{internal.Number(10), internal.String("hello")},
			item: internal.Bool(true),
			want: internal.Values{internal.Bool(true), internal.Number(10), internal.String("hello")},
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			got := internal.Values(tt.vals).Cons(tt.item)

			if !reflect.DeepEqual(got, &internal.List{Values: tt.want}) {
				t.Errorf("Next() want=%#v, got=%#v", tt.want, got)
			}
		})
	}
}
