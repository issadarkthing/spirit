package spirit

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/issadarkthing/spirit/internal"
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

func Shuffle(seq internal.Seq) internal.Seq {
	rand.Seed(time.Now().UnixNano())
	list := Realize(seq)
	values := list.Values
	rand.Shuffle(len(list.Values), func(i, j int) {
		values[i], values[j] = values[j], values[i]
	})
	return list
}

func ReadFile(name string) (string, error) {

	content, err := ioutil.ReadFile(name)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func createShellOutput(out, err string, exit int) *internal.HashMap {
	return &internal.HashMap{
		Data: map[internal.Value]internal.Value{
			internal.Keyword("exit"): internal.Int64(exit),
			internal.Keyword("out"):  internal.String(out),
			internal.Keyword("err"):  internal.String(err),
		},
	}
}

func Shell(command string) (*internal.HashMap, error) {

	cmd := exec.Command("bash", "-c", command)
	var cmdout, cmderr bytes.Buffer

	cmd.Stdout = &cmdout
	cmd.Stderr = &cmderr

	err := cmd.Run()
	if exitErr, ok := err.(*exec.ExitError); ok {
		errMsg := strings.TrimSpace(cmderr.String())
		return createShellOutput("", errMsg, exitErr.ExitCode()), nil
	} else if err != nil {
		return &internal.HashMap{}, err
	}

	output := strings.TrimSpace(cmdout.String())

	return createShellOutput(output, "", 0), nil
}
