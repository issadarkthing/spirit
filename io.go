package xlisp

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/spy16/sabre"
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

func Random(max int) int {
	rand.Seed(time.Now().UnixNano())
	result := rand.Intn(max)
	return result
}

func Shuffle(seq sabre.Seq) sabre.Seq {
	rand.Seed(time.Now().UnixNano())
	list := Realize(seq)
	values := list.Values
	rand.Shuffle(len(list.Values), func(i, j int) {
		values[i], values[j] = values[j], values[i]
	})
	return list
}
