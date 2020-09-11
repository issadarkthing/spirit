package xlisp

import (
	"bufio"
	"fmt"
	"os"
)

// Println is an alias for fmt.Println which ignores the return values.
func Println(args ...interface{}) error {
	_, err := fmt.Println(args...)
	return err
}

// Printf is an alias for fmt.Printf which ignores the return values.
func Printf(format string, args ...interface{}) error {
	_, err := fmt.Printf(format, args...)
	return err
}

// Reads from stdin and returns string
func Read(prompt string) (string, error) {

	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return text[:len(text)-1], nil
}
