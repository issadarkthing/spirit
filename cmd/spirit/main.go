package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"

	"github.com/issadarkthing/spirit/internal"
	"github.com/issadarkthing/spirit/internal/repl"
)

const (
	help = `spirit %s [Commit: %s] [Compiled with %s]
Visit https://github.com/issadarkthing/spirit for more.`
	prompt = " λ >>"
	multiline = "|"
	stdpath = "/.local/lib/spirit/core.st"
)

var (
	version = "N/A"
	commit  = "N/A"

	executeStr   = flag.String("e", "", "Execute string")
	unload       = flag.Bool("u", false, "Unload core library")
	preload      = flag.String("p", "", "Pre-loads file")
	printVersion = flag.Bool("v", false, "Prints spirit version and exit")
	memProfile   = flag.String("memprofile", "", "memory profiling")
	cpuProfile   = flag.String("cpuprofile", "", "cpu profiling")
)

func main() {
	flag.Parse()

	if *printVersion {
		fmt.Println(version)
		return

	} else if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		defer f.Close()

		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}


	sp := internal.NewSpirit()
	sp.BindGo("*version*", version)

	var result internal.Value
	var err error

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}

	libPath := home + stdpath

	core, err := os.Open(libPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}
	defer core.Close()

	// do not load standard library
	if !*unload {
		_, err = sp.ReadEval(core)
	}

	// pre-load file
	if *preload != "" {
		preloadFile, err := os.Open(*preload)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
		defer preloadFile.Close()

		_, err = sp.ReadEval(preloadFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
	}

	sp.SwitchNS(internal.Symbol{Value: "user"})

	if len(flag.Args()) > 0 {

		fh, err := os.Open(flag.Arg(0))
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return
		}
		defer fh.Close()

		abs, err := filepath.Abs(fh.Name())
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return
		}

		err = os.Chdir(filepath.Dir(abs))
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return
		}

		sp.BindGo("*file*", fh.Name())
		sp.BindGo("*path*", abs)
		sp.BindGo("*argv*", flag.Args())
		_, err = sp.ReadEval(fh)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}


		if *memProfile != "" {
			f, err := os.Create(*memProfile)	
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			defer f.Close()

			if err := pprof.WriteHeapProfile(f); err != nil {
				fmt.Fprintln(os.Stderr, "could not write memory profile: ", err)
			}
		}

		return
	}

	if *executeStr != "" {
		result, err = sp.ReadEvalStr(*executeStr)
		fmt.Println(result)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
		return
	}


	repl := repl.New(sp,
		repl.WithBanner(fmt.Sprintf(help, version, commit, runtime.Version())),
		repl.WithPrompts(prompt, multiline),
	)

	if err := repl.Loop(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "REPL exited with error: %v", err)
	}

	fmt.Println("Bye!")
}
