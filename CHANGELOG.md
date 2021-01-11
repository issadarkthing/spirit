v0.9.0
- Add: add ExceptionError
- Add: add documentation string
- Fix: remove stack trace with no postion information
- Fix: unsafe/swap does not change reference of symbol at the correct scope
- Fix: reverse function does not work as intended when Seq with singel element
	is passed


v0.8.0
- Float64 and Int64 are removed and changed to type `Number` which is float64 in Go.
- Added basic JSON parser
- Replaced HashMap with PersistentHashMap. Improved 6x runtime performance because
PersistentHashMap implements structural sharing.
- Added shebang support
- Added pre-load flag to pre-load any source file before evaluating
file or expression.
- Replaced Vector with PersistentVector with crazy amount performance gained when
updating Vector
```
Performance in updating a value at the middle of the vector with 10000 elements
Vector           Elapsed time: 1.108098484s
PersistentVector Elapsed time: 13.867µs
```
- Added stack trace
- Function hoisting for `defn` and `defmacro`
- Added support for unquote splicing
- Added `Future`
- Added Object-Oriented Programming `defclass` and `defmethod`

v0.7.0
- Added Atom type (experimental)
