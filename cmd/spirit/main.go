package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"runtime/pprof"
	"os"
	"runtime"

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
	memProfile   = flag.String("memprofile", "", "memory profiling")
	cpuProfile   = flag.String("cpuprofile", "", "cpu profiling")
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

	} else if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
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


		if *memProfile != "" {
			f, err := os.Create(*memProfile)	
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()

			if err := pprof.WriteHeapProfile(f); err != nil {
				log.Fatal("could not write memory profile: ", err)
			}
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
		repl.WithPrompts(prompt, multiline),
	)

	if err := repl.Loop(context.Background()); err != nil {
		fatalf("REPL exited with error: %v", err)
	}
	fmt.Println("Bye!")
	}

		return e
		log.Fatalf("REPL exited with error: %v", err)
	}


func fatalf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
	os.Exit(1)
}
