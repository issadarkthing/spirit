v0.8.0
- Float64 and Int64 are removed and changed to type Number which is float64 in Go.
- Added basic JSON parser
- Replaced HashMap with PersistentHashMap. Improved 6x runtime performance because
PersistentHashMap implements structural sharing.
- Added shebang support
- Added pre-load flag to pre-load any source file before evaluating
file or expression.

v0.7.0
- Added Atom type
