package repl

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/issadarkthing/spirit/internal"
)

// Option implementations can be provided to New() to configure the REPL
// during initialization.
type Option func(repl *REPL)

// ReaderFactory should return an instance of reader when called. This might
// be called repeatedly. See WithReaderFactory()
type ReaderFactory interface {
	NewReader(r io.Reader) *internal.Reader
}

// ReaderFactoryFunc implements ReaderFactory using a function value.
type ReaderFactoryFunc func(r io.Reader) *internal.Reader

// NewReader simply calls the wrapped function value and returns the result.
func (factory ReaderFactoryFunc) NewReader(r io.Reader) *internal.Reader {
	return factory(r)
}

// ErrMapper should map a custom Input error to nil to indicate error that
// should be ignored by REPL, EOF to signal end of REPL session and any
// other error to indicate a irrecoverable failure.
type ErrMapper func(err error) error

// WithInput sets the REPL's input stream. `nil` defaults to bufio.Scanner
// backed by os.Stdin
func WithInput(in Input, mapErr ErrMapper) Option {
	if in == nil {
		in = &lineReader{
			scanner: bufio.NewScanner(os.Stdin),
			out:     os.Stdout,
		}
	}

	if mapErr == nil {
		mapErr = func(e error) error { return e }
	}

	return func(repl *REPL) {
		repl.input = in
		repl.mapInputErr = mapErr
	}
}

// WithOutput sets the REPL's output stream.`nil` defaults to stdout.
func WithOutput(w io.Writer) Option {
	if w == nil {
		w = os.Stdout
	}

	return func(repl *REPL) {
		repl.output = w
	}
}

// WithBanner sets the REPL's banner which is displayed once when the REPL
// starts.
func WithBanner(banner string) Option {
	return func(repl *REPL) {
		repl.banner = strings.TrimSpace(banner)
	}
}

// WithPrompts sets the prompt to be displayed when waiting for user input
// in the REPL.
func WithPrompts(oneLine, multiLine string) Option {
	return func(repl *REPL) {
		repl.prompt = oneLine
		repl.multiPrompt = multiLine
	}
}

// WithReaderFactory can be used set factory function for initializing internal
// Reader. This is useful when you want REPL to use custom reader instance.
func WithReaderFactory(factory ReaderFactory) Option {
	if factory == nil {
		factory = ReaderFactoryFunc(internal.NewReader)
	}

	return func(repl *REPL) {
		repl.factory = factory
	}
}

// WithPrinter sets the print function for the REPL.  It is useful for customizing
// how different types should be rendered into human-readable character streams.
func WithPrinter(f func(io.Writer, interface{}) error) Option {
	if f == nil {
		f = func(w io.Writer, v interface{}) (err error) {
			switch v.(type) {
			case error:
				_, err = fmt.Fprintf(w, "%+v\n", v)

			default:
				_, err = fmt.Fprintf(w, "%s\n", v)
			}

			return
		}
	}

	return func(repl *REPL) {
		repl.printer = f
	}
}

func withDefaults(opts []Option) []Option {
	return append([]Option{
		WithInput(nil, nil),
		WithOutput(nil),
		WithReaderFactory(nil),
		WithPrinter(nil),
	}, opts...)
}

type lineReader struct {
	scanner *bufio.Scanner
	out     io.Writer
	prompt  string
}

func (lr *lineReader) Readline() (string, error) {
	lr.out.Write([]byte(lr.prompt))

	if !lr.scanner.Scan() {
		if lr.scanner.Err() == nil { // scanner swallows EOF
			return lr.scanner.Text(), io.EOF
		}

		return "", lr.scanner.Err()
	}

	return lr.scanner.Text(), nil
}

// no-op
func (lr *lineReader) SetPrompt(p string) {
	lr.prompt = p
}
