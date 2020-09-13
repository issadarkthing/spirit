# Xlisp
Simple lisp written in go using [Sabre](https://github.com/spy16/sabre).
Xlisp is a DSL for [tview](https://github.com/rivo/tview) library.


## Usage

1. `xlisp` for REPL
2. `xlisp -e "(+ 1 2 3)"` for executing string
3. `xlisp sample.lisp` for executing file.

## Documentation
Xlisp is highly inspired by Clojure with their syntax and semantics but it's
implemented in Go programming language. Some of the functions from clojure have 
been implemented in xlisp.

## Example
First you have to compile the xlisp.
```sh
$ make
```
Then, you can run this example using `bin/xlisp example/sample.lisp`
```clojure
(def app (tview/new-app))
(def text-view (tview/new-textview))

(def sample-text "Leverage agile frameworks to provide a robust synopsis for high level overviews. Iterative approaches to corporate strategy foster collaborative thinking to further the overall value proposition. Organically grow the holistic world view of disruptive innovation via workplace diversity and empowerment.
                 Bring to the table win-win survival strategies to ensure proactive domination. At the end of the day, going forward, a new normal that has evolved from generation X is on the runway heading towards a streamlined cloud solution. User generated content in real-time will have multiple touchpoints for offshoring.
                 Capitalize on low hanging fruit to identify a ballpark value added activity to beta test. Override the digital divide with additional clickthroughs from DevOps. Nanotechnology immersion along the information highway will close the loop on focusing solely on the bottom line")

(text-view.SetWordWrap true)
(text-view.SetBorder true)
(text-view.SetTitle "This is just an example")
(text-view.SetText sample-text)
(app.SetRoot text-view true)

; main loop
(app.Run)
```
Then it will fill up your terminal with box and texts. You can `ctrl + c` to cancel
the application.
![terminal](/example.png)
