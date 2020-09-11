package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/chzyer/readline"
	"github.com/spy16/sabre"
	"github.com/spy16/sabre/repl"
	"github.com/spy16/slang"
)

const help = `Slang %s [Commit: %s] [Compiled with %s]
Visit https://github.com/spy16/sabre for more.`

var (
	version = "N/A"
	commit  = "N/A"

	executeStr   = flag.String("e", "", "Execute string")
	printVersion = flag.Bool("v", false, "Prints slang version and exit")
)


func main() {
	flag.Parse()

	if *printVersion {
		fmt.Println(version)
		return
	}

	sl := slang.New()
	sl.BindGo("*version*", version)

	var result sabre.Value
	var err error

	if len(os.Args) > 1 {

		resolvedPath, err := filepath.Abs("lib/core.lisp")
		if err != nil {
			fatalf("error: %v\n", err)
		}

		core, err := os.Open(resolvedPath)
		if err != nil {
			fatalf("error: %v\n", err)
		}
		defer core.Close()

		_, err = sl.ReadEval(core)

		fh, err := os.Open(os.Args[1])
		if err != nil {
			fatalf("error: %v\n", err)
		}
		defer fh.Close()

		sl.SwitchNS(sabre.Symbol{Value: "user"})

		_, err = sl.ReadEval(fh)
		if err != nil {
			fatalf("error: %v\n", err)
		}
		return
	}

	if *executeStr != "" {
		result, err = sl.ReadEvalStr(*executeStr)
		fmt.Println(result)
		if err != nil {
			fatalf("error: %v\n", err)
		}
		return
	}

	lr, errMapper := readlineInstance()

	repl := repl.New(sl,
		repl.WithBanner(fmt.Sprintf(help, version, commit, runtime.Version())),
		repl.WithInput(lr, errMapper),
		repl.WithOutput(lr.Stdout()),
		repl.WithPrompts("=>", "|"),
	)

	if err := repl.Loop(context.Background()); err != nil {
		fatalf("REPL exited with error: %v", err)
	}
	fmt.Println("Bye!")
}

func readlineInstance() (*readline.Instance, func(error) error) {
	lr, err := readline.New("")
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
