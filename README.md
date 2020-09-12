# Xlisp
Simple lisp written in go using [Sabre](https://github.com/spy16/sabre).


## Usage

1. `xlisp` for REPL
2. `xlisp -e "(+ 1 2 3)"` for executing string
3. `xlisp sample.lisp` for executing file.

## Documentation
Xlisp is highly inspired by Clojure with their syntax and semantics but it's
implemented in Go programming language. Some of the functions from clojure have 
been implemented in xlisp.

### Core
- range
- doseq
- ->
- ->>
- case
- do
- def
- defn
- defmacro
- if
- fn
- macro
- let
- quote
- syntax-quote
- recur
- macroexpand
- eval
- type
- to-type
- realize
- throw
- str
- \+
- \-
- \*
- \/
- mod
- \=
- \>
- <=
- <
- <=
- print
- printf
- read
- random
- shuffle
- first
- second
- next
- rest
- cons
- conj
- cons
- last
- reverse
- inc
- count
- reduce
- reduce-indexed
- map
- map-indexed
- filter
- filter-indexed
- concat
- apply-seq
- when
- when-not
- assert
- set
- list
- vector
- int
- float
- boolean
- not
- nil?
- seq?
- true?
- impl?
- is-type?
- set?
- list?
- vector?
- int?
- float?
- boolean?
- string?
- keyword?
- symbol?
- empty?
- number?
- even?
- odd?

### unsafe
- swap
