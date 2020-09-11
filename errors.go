package xlisp

import "fmt"

func checkArity(expected, got int) error {

	msg := "invalid number of arguments passed; expected %d instead got %d"

	if expected != got {
		return fmt.Errorf(msg, expected, got)
	}

	return nil
}

func checkArityAtLeast(atLeast, got int) error {

	msg := "invalid number of arguments passed; expected at least %d instead got %d"

	if got < atLeast {
		return fmt.Errorf(msg, atLeast, got)
	}

	return nil
}
