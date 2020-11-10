package spirit

import (
	"fmt"

	"github.com/issadarkthing/spirit/internal"
)

func checkArity(expected, got int) error {

	msg := "invalid number of arguments passed; expected %d instead got %d"

	if expected != got {
		return fmt.Errorf(msg, expected, got)
	}

	return nil
}

func invalidType(expected, got internal.Value) error {
	return fmt.Errorf("invalid type; expected %T instead got %T", expected, got)
}

func doesNotImplementSeq(got internal.Value) error {
	return fmt.Errorf("%T does not implement Seq interface", got)
}

func doesNotImplementInvokable(got internal.Value) error {
	return fmt.Errorf("%v does not implement Invokable interface", got)
}

func checkArityAtLeast(atLeast, got int) error {

	msg := "invalid number of arguments passed; expected at least %d instead got %d"

	if got < atLeast {
		return fmt.Errorf(msg, atLeast, got)
	}

	return nil
}
