package internal_test

import (
	"bytes"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/issadarkthing/spirit/internal"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		r        io.Reader
		fileName string
	}{
		{
			name:     "WithStringReader",
			r:        strings.NewReader(":test"),
			fileName: "<string>",
		},
		{
			name:     "WithBytesReader",
			r:        bytes.NewReader([]byte(":test")),
			fileName: "<bytes>",
		},
		{
			name:     "WihFile",
			r:        os.NewFile(0, "test.lisp"),
			fileName: "test.lisp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rd := internal.NewReader(tt.r)
			if rd == nil {
				t.Errorf("New() should return instance of Reader, got nil")
			} else if rd.File != tt.fileName {
				t.Errorf("internal.File = \"%s\", want = \"%s\"", rd.File, tt.name)
			}
		})
	}
}

func TestReader_SetMacro(t *testing.T) {
	t.Run("UnsetDefaultMacro", func(t *testing.T) {
		rd := internal.NewReader(strings.NewReader("~hello"))
		rd.SetMacro('~', nil, false) // remove unquote operator

		var want internal.Value
		want = internal.Symbol{
			Value: "~hello",
			Position: internal.Position{
				File:   "<string>",
				Line:   1,
				Column: 1,
			},
		}

		got, err := rd.One()
		if err != nil {
			t.Errorf("unexpected error: %#v", err)
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got = %#v, want = %#v", got, want)
		}
	})

	t.Run("CustomMacro", func(t *testing.T) {
		rd := internal.NewReader(strings.NewReader("~hello"))
		rd.SetMacro('~', func(rd *internal.Reader, _ rune) (internal.Value, error) {
			var ru []rune
			for {
				r, err := rd.NextRune()
				if err != nil {
					if err == io.EOF {
						break
					}
					return nil, err
				}

				if rd.IsTerminal(r) {
					break
				}
				ru = append(ru, r)
			}

			return internal.String(ru), nil
		}, false) // override unquote operator

		var want internal.Value
		want = internal.String("hello")

		got, err := rd.One()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got = %v, want = %v", got, want)
		}
	})
}

func TestReader_All(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		want    internal.Value
		wantErr bool
	}{
		{
			name: "ValidLiteralSample",
			src:  `'hello #{} 123 "Hello\tWorld" 12.34 -0xF +010 true nil 0b1010 \a :hello`,
			want: internal.Module{
				&internal.List{
					Values: []internal.Value{
						internal.Symbol{Value: "quote"},
						internal.Symbol{
							Value: "hello",
							Position: internal.Position{
								File:   "<string>",
								Line:   1,
								Column: 2,
							},
						},
					},
				},
				internal.Set{
					Position: internal.Position{
						File:   "<string>",
						Line:   1,
						Column: 9,
					},
				},
				internal.Number(123),
				internal.String("Hello\tWorld"),
				internal.Number(12.34),
				internal.Number(-15),
				internal.Number(8),
				internal.Bool(true),
				internal.Nil{},
				internal.Number(10),
				internal.Character('a'),
				internal.Keyword("hello"),
			},
		},
		{
			name: "WithComment",
			src:  `:valid-keyword ; comment should return errSkip`,
			want: internal.Module{internal.Keyword("valid-keyword")},
		},
		{
			name:    "UnterminatedString",
			src:     `:valid-keyword "unterminated string literal`,
			wantErr: true,
		},
		{
			name: "CommentFollowedByForm",
			src:  `; comment should return errSkip` + "\n" + `:valid-keyword`,
			want: internal.Module{internal.Keyword("valid-keyword")},
		},
		{
			name:    "UnterminatedList",
			src:     `:valid-keyword (add 1 2`,
			wantErr: true,
		},
		{
			name:    "UnterminatedVector",
			src:     `:valid-keyword [1 2`,
			wantErr: true,
		},
		{
			name:    "EOFAfterQuote",
			src:     `:valid-keyword '`,
			wantErr: true,
		},
		{
			name:    "CommentAfterQuote",
			src:     `:valid-keyword ';hello world`,
			wantErr: true,
		},
		{
			name:    "UnbalancedParenthesis",
			src:     `())`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := internal.NewReader(strings.NewReader(tt.src)).All()
			if (err != nil) != tt.wantErr {
				t.Errorf("All() error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("All() got = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestReader_One(t *testing.T) {
	executeReaderTests(t, []readerTestCase{
		{
			name:    "Empty",
			src:     "",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "QuotedEOF",
			src:     "';comment is a no-op form\n",
			wantErr: true,
		},
		{
			name:    "ListEOF",
			src:     "( 1",
			wantErr: true,
		},
		{
			name: "UnQuote",
			src:  "~(x 3)",
			want: &internal.List{
				Values: []internal.Value{
					internal.Symbol{Value: "unquote"},
					&internal.List{
						Values: []internal.Value{
							internal.Symbol{
								Value: "x",
								Position: internal.Position{
									File:   "<string>",
									Line:   1,
									Column: 3,
								},
							},
							internal.Number(3),
						},
						Position: internal.Position{
							File:   "<string>",
							Line:   1,
							Column: 2,
						},
					},
				},
			},
		},
	})
}

func TestReader_One_Number(t *testing.T) {
	executeReaderTests(t, []readerTestCase{
		{
			name: "NumberWithLeadingSpaces",
			src:  "    +1234",
			want: internal.Number(1234),
		},
		{
			name: "PositiveInt",
			src:  "+1245",
			want: internal.Number(1245),
		},
		{
			name: "NegativeInt",
			src:  "-234",
			want: internal.Number(-234),
		},
		{
			name: "PositiveFloat",
			src:  "+1.334",
			want: internal.Number(1.334),
		},
		{
			name: "NegativeFloat",
			src:  "-1.334",
			want: internal.Number(-1.334),
		},
		{
			name: "PositiveHex",
			src:  "0x124",
			want: internal.Number(0x124),
		},
		{
			name: "NegativeHex",
			src:  "-0x124",
			want: internal.Number(-0x124),
		},
		{
			name: "PositiveOctal",
			src:  "0123",
			want: internal.Number(0123),
		},
		{
			name: "NegativeOctal",
			src:  "-0123",
			want: internal.Number(-0123),
		},
		{
			name: "PositiveBinary",
			src:  "0b10",
			want: internal.Number(2),
		},
		{
			name: "NegativeBinary",
			src:  "-0b10",
			want: internal.Number(-2),
		},
		{
			name: "PositiveBase2Radix",
			src:  "2r10",
			want: internal.Number(2),
		},
		{
			name: "NegativeBase2Radix",
			src:  "-2r10",
			want: internal.Number(-2),
		},
		{
			name: "PositiveBase4Radix",
			src:  "4r123",
			want: internal.Number(27),
		},
		{
			name: "NegativeBase4Radix",
			src:  "-4r123",
			want: internal.Number(-27),
		},
		{
			name: "ScientificSimple",
			src:  "1e10",
			want: internal.Number(1e10),
		},
		{
			name: "ScientificNegativeExponent",
			src:  "1e-10",
			want: internal.Number(1e-10),
		},
		{
			name: "ScientificWithDecimal",
			src:  "1.5e10",
			want: internal.Number(1.5e+10),
		},
		{
			name:    "FloatStartingWith0",
			src:     "012.3",
			want:    internal.Number(012.3),
			wantErr: false,
		},
		{
			name:    "InvalidValue",
			src:     "1ABe13",
			wantErr: true,
		},
		{
			name:    "InvalidScientificFormat",
			src:     "1e13e10",
			wantErr: true,
		},
		{
			name:    "InvalidExponent",
			src:     "1e1.3",
			wantErr: true,
		},
		{
			name:    "InvalidRadixFormat",
			src:     "1r2r3",
			wantErr: true,
		},
		{
			name:    "RadixBase3WithDigit4",
			src:     "-3r1234",
			wantErr: true,
		},
		{
			name:    "RadixMissingValue",
			src:     "2r",
			wantErr: true,
		},
		{
			name:    "RadixInvalidBase",
			src:     "2ar",
			wantErr: true,
		},
		{
			name:    "RadixWithFloat",
			src:     "2.3r4",
			wantErr: true,
		},
		{
			name:    "DecimalPointInBinary",
			src:     "0b1.0101",
			wantErr: true,
		},
		{
			name:    "InvalidDigitForOctal",
			src:     "08",
			wantErr: true,
		},
		{
			name:    "IllegalNumberFormat",
			src:     "9.3.2",
			wantErr: true,
		},
	})
}

func TestReader_One_String(t *testing.T) {
	executeReaderTests(t, []readerTestCase{
		{
			name: "SimpleString",
			src:  `"hello"`,
			want: internal.String("hello"),
		},
		{
			name: "EscapeQuote",
			src:  `"double quote is \""`,
			want: internal.String(`double quote is "`),
		},
		{
			name: "EscapeSlash",
			src:  `"hello\\world"`,
			want: internal.String(`hello\world`),
		},
		{
			name:    "UnexpectedEOF",
			src:     `"double quote is`,
			wantErr: true,
		},
		{
			name:    "InvalidEscape",
			src:     `"hello \x world"`,
			wantErr: true,
		},
		{
			name:    "EscapeEOF",
			src:     `"hello\`,
			wantErr: true,
		},
	})
}

func TestReader_One_Keyword(t *testing.T) {
	executeReaderTests(t, []readerTestCase{
		{
			name: "SimpleASCII",
			src:  `:test`,
			want: internal.Keyword("test"),
		},
		{
			name: "LeadingTrailingSpaces",
			src:  "          :test          ",
			want: internal.Keyword("test"),
		},
		{
			name: "SimpleUnicode",
			src:  `:∂`,
			want: internal.Keyword("∂"),
		},
		{
			name: "WithSpecialChars",
			src:  `:this-is-valid?`,
			want: internal.Keyword("this-is-valid?"),
		},
		{
			name: "FollowedByMacroChar",
			src:  `:this-is-valid'hello`,
			want: internal.Keyword("this-is-valid"),
		},
	})
}

func TestReader_One_Character(t *testing.T) {
	executeReaderTests(t, []readerTestCase{
		{
			name: "ASCIILetter",
			src:  `\a`,
			want: internal.Character('a'),
		},
		{
			name: "ASCIIDigit",
			src:  `\1`,
			want: internal.Character('1'),
		},
		{
			name: "Unicode",
			src:  `\∂`,
			want: internal.Character('∂'),
		},
		{
			name: "Newline",
			src:  `\newline`,
			want: internal.Character('\n'),
		},
		{
			name: "FormFeed",
			src:  `\formfeed`,
			want: internal.Character('\f'),
		},
		{
			name: "Unicode",
			src:  `\u00AE`,
			want: internal.Character('®'),
		},
		{
			name:    "InvalidUnicode",
			src:     `\uHELLO`,
			wantErr: true,
		},
		{
			name:    "OutOfRangeUnicode",
			src:     `\u-100`,
			wantErr: true,
		},
		{
			name:    "UnknownSpecial",
			src:     `\hello`,
			wantErr: true,
		},
		{
			name:    "EOF",
			src:     `\`,
			wantErr: true,
		},
	})
}

func TestReader_One_Symbol(t *testing.T) {
	executeReaderTests(t, []readerTestCase{
		{
			name: "SimpleASCII",
			src:  `hello`,
			want: internal.Symbol{
				Value: "hello",
				Position: internal.Position{
					File:   "<string>",
					Line:   1,
					Column: 1,
				},
			},
		},
		{
			name: "Unicode",
			src:  `find-∂`,
			want: internal.Symbol{
				Value: "find-∂",
				Position: internal.Position{
					File:   "<string>",
					Line:   1,
					Column: 1,
				},
			},
		},
		{
			name: "SingleChar",
			src:  `+`,
			want: internal.Symbol{
				Value: "+",
				Position: internal.Position{
					File:   "<string>",
					Line:   1,
					Column: 1,
				},
			},
		},
	})
}

func TestReader_One_List(t *testing.T) {
	executeReaderTests(t, []readerTestCase{
		{
			name: "EmptyList",
			src:  `()`,
			want: &internal.List{
				Values: nil,
				Position: internal.Position{
					File:   "<string>",
					Line:   1,
					Column: 1,
				},
			},
		},
		{
			name: "ListWithOneEntry",
			src:  `(help)`,
			want: &internal.List{
				Values: []internal.Value{
					internal.Symbol{
						Value: "help",
						Position: internal.Position{
							File:   "<string>",
							Line:   1,
							Column: 2,
						},
					},
				},
				Position: internal.Position{
					File:   "<string>",
					Line:   1,
					Column: 1,
				},
			},
		},
		{
			name: "ListWithMultipleEntry",
			src:  `(+ 0xF 3.1413)`,
			want: &internal.List{
				Values: []internal.Value{
					internal.Symbol{
						Value: "+",
						Position: internal.Position{
							File:   "<string>",
							Line:   1,
							Column: 2,
						},
					},
					internal.Number(15),
					internal.Number(3.1413),
				},
				Position: internal.Position{
					File:   "<string>",
					Line:   1,
					Column: 1,
				},
			},
		},
		{
			name: "ListWithCommaSeparator",
			src:  `(+,0xF,3.1413)`,
			want: &internal.List{
				Values: []internal.Value{
					internal.Symbol{
						Value: "+",
						Position: internal.Position{
							File:   "<string>",
							Line:   1,
							Column: 2,
						},
					},
					internal.Number(15),
					internal.Number(3.1413),
				},
				Position: internal.Position{
					File:   "<string>",
					Line:   1,
					Column: 1,
				},
			},
		},
		{
			name: "MultiLine",
			src: `(+
                      0xF
                      3.1413
					)`,
			want: &internal.List{
				Values: []internal.Value{
					internal.Symbol{
						Value: "+",
						Position: internal.Position{
							File:   "<string>",
							Line:   1,
							Column: 2,
						},
					},
					internal.Number(15),
					internal.Number(3.1413),
				},
				Position: internal.Position{
					File:   "<string>",
					Line:   1,
					Column: 1,
				},
			},
		},
		{
			name: "MultiLineWithComments",
			src: `(+         ; plus operator adds numerical values
                      0xF    ; hex representation of 15
                      3.1413 ; value of math constant pi
                  )`,
			want: &internal.List{
				Values: []internal.Value{
					internal.Symbol{
						Value: "+",
						Position: internal.Position{
							File:   "<string>",
							Line:   1,
							Column: 2,
						},
					},
					internal.Number(15),
					internal.Number(3.1413),
				},
				Position: internal.Position{
					File:   "<string>",
					Line:   1,
					Column: 1,
				},
			},
		},
		{
			name:    "UnexpectedEOF",
			src:     "(+ 1 2 ",
			wantErr: true,
		},
	})
}

func TestReader_One_Vector(t *testing.T) {
	executeReaderTests(t, []readerTestCase{
		{
			name: "Empty",
			src:  `[]`,
			want: internal.Vector{
				Values: nil,
				Position: internal.Position{
					File:   "<string>",
					Line:   1,
					Column: 1,
				},
			},
		},
		{
			name: "WithOneEntry",
			src:  `[help]`,
			want: internal.Vector{
				Values: []internal.Value{
					internal.Symbol{
						Value: "help",
						Position: internal.Position{
							File:   "<string>",
							Line:   1,
							Column: 2,
						},
					},
				},
				Position: internal.Position{
					File:   "<string>",
					Line:   1,
					Column: 1,
				},
			},
		},
		{
			name: "WithMultipleEntry",
			src:  `[+ 0xF 3.1413]`,
			want: internal.Vector{
				Values: []internal.Value{
					internal.Symbol{
						Value: "+",
						Position: internal.Position{
							File:   "<string>",
							Line:   1,
							Column: 2,
						},
					},
					internal.Number(15),
					internal.Number(3.1413),
				},
				Position: internal.Position{
					File:   "<string>",
					Line:   1,
					Column: 1,
				},
			},
		},
		{
			name: "WithCommaSeparator",
			src:  `[+,0xF,3.1413]`,
			want: internal.Vector{
				Values: []internal.Value{
					internal.Symbol{
						Value: "+",
						Position: internal.Position{
							File:   "<string>",
							Line:   1,
							Column: 2,
						},
					},
					internal.Number(15),
					internal.Number(3.1413),
				},
				Position: internal.Position{
					File:   "<string>",
					Line:   1,
					Column: 1,
				},
			},
		},
		{
			name: "MultiLine",
			src: `[+
                      0xF
                      3.1413
					]`,
			want: internal.Vector{
				Values: []internal.Value{
					internal.Symbol{
						Value: "+",
						Position: internal.Position{
							File:   "<string>",
							Line:   1,
							Column: 2,
						},
					},
					internal.Number(15),
					internal.Number(3.1413),
				},
				Position: internal.Position{
					File:   "<string>",
					Line:   1,
					Column: 1,
				},
			},
		},
		{
			name: "MultiLineWithComments",
			src: `[+         ; plus operator adds numerical values
                      0xF    ; hex representation of 15
                      3.1413 ; value of math constant pi
                  ]`,
			want: internal.Vector{
				Values: []internal.Value{
					internal.Symbol{
						Value: "+",
						Position: internal.Position{
							File:   "<string>",
							Line:   1,
							Column: 2,
						},
					},
					internal.Number(15),
					internal.Number(3.1413),
				},
				Position: internal.Position{
					File:   "<string>",
					Line:   1,
					Column: 1,
				},
			},
		},
		{
			name:    "UnexpectedEOF",
			src:     "[+ 1 2 ",
			wantErr: true,
		},
	})
}

func TestReader_One_Set(t *testing.T) {
	executeReaderTests(t, []readerTestCase{
		{
			name: "Empty",
			src:  "#{}",
			want: internal.Set{
				Values: nil,
				Position: internal.Position{
					File:   "<string>",
					Line:   1,
					Column: 2,
				},
			},
		},
		{
			name: "Valid",
			src:  "#{1 2 []}",
			want: internal.Set{
				Values: []internal.Value{internal.Number(1),
					internal.Number(2),
					internal.Vector{
						Position: internal.Position{
							File:   "<string>",
							Column: 7,
							Line:   1,
						},
					},
				},
				Position: internal.Position{
					File:   "<string>",
					Line:   1,
					Column: 2,
				},
			},
		},
		{
			name:    "HasDuplicate",
			src:     "#{1 2 2}",
			wantErr: true,
		},
	})
}

func TestReader_One_HashMap(t *testing.T) {

	executeReaderTests(t, []readerTestCase{
		{
			name:    "NonHashableKey",
			src:     `{[] 10}`,
			wantErr: true,
		},
		{
			name:    "OddNumberOfForms",
			src:     "{:hello 10 :age}",
			wantErr: true,
		},
	})
}

type readerTestCase struct {
	name    string
	src     string
	want    internal.Value
	wantErr bool
}

func executeReaderTests(t *testing.T, tests []readerTestCase) {
	t.Parallel()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := internal.NewReader(strings.NewReader(tt.src)).One()
			if (err != nil) != tt.wantErr {
				t.Errorf("One() error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("One() got = %#v, want %#v", got, tt.want)
			}
		})
	}
}
