package slang

import "fmt"

type invalidArgNumberError struct {
	expected int
	got      int
}

func (i invalidArgNumberError) Error() string {
	msg := "invalid number of argument given; expected (%d) got (%d)"
	return fmt.Sprintf(msg, i.expected, i.got)
}
