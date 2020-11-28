package spirit

import (
	"fmt"

	"github.com/issadarkthing/spirit/internal"
)



func invalidType(expected, got internal.Value) error {
	return fmt.Errorf("invalid type; expected %T instead got %T", expected, got)
}

func doesNotImplementSeq(got internal.Value) error {
	return fmt.Errorf("%T does not implement Seq interface", got)
}

func doesNotImplementInvokable(got internal.Value) error {
	return fmt.Errorf("%v does not implement Invokable interface", got)
}

