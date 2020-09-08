# Slang

Simple lisp written in go using [Sabre](https://github.com/spy16/sabre).


## Usage

1. `slang` for REPL
2. `slang -e "(+ 1 2 3)"` for executing string
3. `slang examples/simple.lisp` for executing file.

## Goal
LSP for text editor that will be written in [go](https://golang.org) using
[tview](https://github.com/rivo/tview). Quite possibly the library itself will
be ported to this language and able to manipulate the widgets using this language.
