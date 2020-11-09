package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/chzyer/readline"
	"github.com/issadarkthing/spirit"
	"github.com/issadarkthing/spirit/internal"
	"github.com/issadarkthing/spirit/internal/repl"
)

const (
	help = `spirit %s [Commit: %s] [Compiled with %s]
Visit https://github.com/issadarkthing/spirit for more.`
	prompt = " λ >>"
	multiline = "|"
)

var (
	version = "N/A"
	commit  = "N/A"

	executeStr   = flag.String("e", "", "Execute string")
	unload       = flag.Bool("u", false, "Unload core library")
	preload      = flag.String("p", "", "Pre-loads file")
	printVersion = flag.Bool("v", false, "Prints slang version and exit")
	matcher      = map[rune]rune{
		'(': ')',
		'[': ']',
		'{': '}',
		'"': '"',
	}
)

func main() {
	flag.Parse()

	if *printVersion {
		fmt.Println(version)
		return
	}

	xl := spirit.New()
	xl.BindGo("*version*", version)

	var result internal.Value
	var err error

	home, err := os.UserHomeDir()
	if err != nil {
		fatalf("error: %v\n", err)
	}

	libPath := home + "/.local/lib/spirit/core.st"

	core, err := os.Open(libPath)
	if err != nil {
		fatalf("error: %v\n", err)
	}
	defer core.Close()

	if !*unload {
		_, err = xl.ReadEval(core)
	}

	if *preload != "" {
		preloadFile, err := os.Open(*preload)
		if err != nil {
			fatalf("error: %v\n", err)
		}
		defer preloadFile.Close()

		_, err = xl.ReadEval(preloadFile)
		if err != nil {
			fatalf("error: %v\n", err)
		}
	}

	xl.SwitchNS(internal.Symbol{Value: "user"})

	if len(flag.Args()) > 0 {

		fh, err := os.Open(flag.Arg(0))
		if err != nil {
			fatalf("error: %v\n", err)
		}
		defer fh.Close()

		xl.BindGo("*file*", fh.Name())
		_, err = xl.ReadEval(fh)
		if err != nil {
			fatalf("error: %v\n", err)
		}
		return
	}

	if *executeStr != "" {
		result, err = xl.ReadEvalStr(*executeStr)
		fmt.Println(result)
		if err != nil {
			fatalf("error: %v\n", err)
		}
		return
	}

	lr, errMapper := readlineInstance(xl)

	repl := repl.New(xl,
		repl.WithBanner(fmt.Sprintf(help, version, commit, runtime.Version())),
		repl.WithInput(lr, errMapper),
		repl.WithOutput(lr.Stdout()),
		repl.WithPrompts(prompt, multiline),
	)

	if err := repl.Loop(context.Background()); err != nil {
		fatalf("REPL exited with error: %v", err)
	}
	fmt.Println("Bye!")
}


func insertAt(index int, value rune, slice []rune) []rune {
	newSlice := make([]rune, 0, len(slice)+1)

	for i, v := range slice {
		if i == index {
			newSlice = append(newSlice, value, v)
			continue
		}
		newSlice = append(newSlice, v)
	}

	if index > len(slice)-1 {
		newSlice = append(newSlice, value)
	}

	return newSlice
}

func removeAt(index int, slice []rune) []rune {
	return append(slice[:index], slice[index+1:]...)
}

func listener(line []rune, pos int, key rune) ([]rune, int, bool) {

	switch key {
	case '(', '{', '[':
		line = insertAt(pos, matcher[key], line)
		return line, pos, true
	case '"':
		if len(line) > pos && line[pos] == '"' {
			return removeAt(pos, line), pos, true
		}
		line = insertAt(pos, '"', line)
		return line, pos, true
	case ')', '}', ']':
		if len(line) > pos && line[pos] == key {
			return removeAt(pos, line), pos, true
		}
	}

	return line, pos, false
}

func readlineInstance(scope internal.Scope) (*readline.Instance, func(error) error) {
	completer := readline.NewPrefixCompleter(
		readline.PcItem("(source *file*)"),
	)

	lr, err := readline.NewEx(&readline.Config{
		HistoryFile: "/tmp/spirit-repl.tmp",
		AutoComplete: completer,
		Listener: readline.FuncListener(listener),
	})


	if err != nil {
		fatalf("readline: %v", err)
	}

	errMapper := func(e error) error {
		if errors.Is(e, readline.ErrInterrupt) {
			return nil
		}

		return e
	}

	return lr, errMapper
}

func fatalf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
	os.Exit(1)
}
