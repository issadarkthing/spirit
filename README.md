# Spirit
Spirit is a scripting, functional language inspired by Clojure. Spirit
does not target JVM machines instead it's interpreted line by line. This has
the advantage of fast startup and suitable for scripting environment.

## Background
This programming language is just an experiment for me to tinker and mess around
in building programming language. I always wanted to build my own programming language
and implement some of the things that may have on another language to this language.
Lisp-like syntax is chosen because it is easy to parse, consistent and elegant.
But, as you see the syntax is not directly inherited from Lisp rather the syntax
is much, much similar to Clojure.

## Data Types
- String
- Number
- List
- Vector
- HashMap
- Set
- Atom
- Future

## Differences
These are the differences that I deliberately made to differ from Clojure.

- Function hoisting for `defn` and `defmacro`
- Function `apply` in clojure is equivalent to `<>` in spirit
- All functions that acts on Seq returns the same concrete type Seq.
	For example, `map` on vector returns vector instead of list
- Keyword is used instead of String as key on JSON object
