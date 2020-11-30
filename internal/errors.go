package internal

import (
	"fmt"
	"strings"
	"reflect"
)

type TypeError struct {
	Expected reflect.Type
	Got      reflect.Type
}

func (t TypeError) removePrefix(str string) string {
	str = strings.TrimLeft(str, "*")
	str = strings.TrimLeft(str, "internal.")
	return str
}

func (t TypeError) Error() string {
	expected := t.Expected.String()
	got := t.Got.String()

	expected = t.removePrefix(expected)
	got = t.removePrefix(got)

	return fmt.Sprintf("invalid type, expected '%s' got '%s'", expected, got)
}
