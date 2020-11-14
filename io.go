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
func println(args ...interface{}) error {
	result := []interface{}{}
	for _, v := range args {
		if str, ok := v.(internal.String); ok {
			result = append(result, removeSuffixPrefix(string(str), `"`))
			continue
		}
		result = append(result, v)
	}
	_, err := fmt.Println(result...)
	return err
}

func removeSuffixPrefix(str, cutset string) string {
	return strings.TrimSuffix(strings.TrimPrefix(str, cutset), cutset)
}

// Printf is an alias for fmt.Printf which ignores the return values.
func printf(format string, args ...interface{}) error {
	_, err := fmt.Printf(format, args...)
	return err
}

// Reads from stdin and returns string
func read(prompt string) (string, error) {

	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return text[:len(text)-1], nil
}

func random(max int) int {
	rand.Seed(time.Now().UnixNano())
	result := rand.Intn(max)
	return result
}

func shuffle(seq internal.Seq) internal.Seq {
	rand.Seed(time.Now().UnixNano())
	list := realize(seq)
	values := list.Values
	rand.Shuffle(len(list.Values), func(i, j int) {
		values[i], values[j] = values[j], values[i]
	})
	return list
}

func readFile(name string) (string, error) {

	content, err := ioutil.ReadFile(name)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func createShellOutput(out, err string, exit int) *internal.PersistentMap {
	m := internal.NewPersistentMap()
	m = m.Set(internal.Keyword("exit"), internal.Number(exit)).(*internal.PersistentMap)
	m = m.Set(internal.Keyword("out"), internal.String(out)).(*internal.PersistentMap)
	m = m.Set(internal.Keyword("err"), internal.String(err)).(*internal.PersistentMap)
	return m
}

func shell(command string) (*internal.PersistentMap, error) {

	cmd := exec.Command("bash", "-c", command)
	var cmdout, cmderr bytes.Buffer

	cmd.Stdout = &cmdout
	cmd.Stderr = &cmderr

	err := cmd.Run()
	if exitErr, ok := err.(*exec.ExitError); ok {
		errMsg := strings.TrimSpace(cmderr.String())
		return createShellOutput("", errMsg, exitErr.ExitCode()), nil
	} else if err != nil {
		return internal.NewPersistentMap(), err
	}

	output := strings.TrimSpace(cmdout.String())

	return createShellOutput(output, "", 0), nil
}
