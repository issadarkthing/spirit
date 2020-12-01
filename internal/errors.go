package internal

import (
	"fmt"
	"strings"
)

type TypeError struct {
	Expected Value
	Got      Value
}

func (t TypeError) Error() string {

	expected := TypeOf(t.Expected)
	got := TypeOf(t.Got)

	return fmt.Sprintf(
		"TypeError: expected %s instead got %s",
		expected, got,
	)
}

type ArgumentError struct {
	Fn  string
	Got int
}

func (a ArgumentError) Error() string {
	return fmt.Sprintf(
		"ArgumentError: wrong number of args (%d) to '%s'",
		a.Got, a.Fn,
	)
}

type ResolveError struct {
	Sym Symbol
}

func (r ResolveError) Error() string {
	return fmt.Sprintf(
		"ResolveError: unable to resolve symbol '%s'", r.Sym,
	)
}

func RemovePrefix(str string) string {
	str = strings.TrimLeft(str, "*")
	str = strings.TrimLeft(str, "internal.")
	return str
}
