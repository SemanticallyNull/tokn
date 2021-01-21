package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

func main() {
	Main(os.Args, os.Stdout)
}

func Main(s []string, w io.Writer) {
	log := ioutil.Discard

	file, err := os.Open(s[1])
	if err != nil {
		panic(err)
	}
	defer file.Close()

	b := bufio.NewReader(file)

	var (
		tokens     []Token
		line       int
		col        int
		cur        []byte
		skipTokens int
		vars       map[string]interface{}
	)
	vars = make(map[string]interface{})
	line = 1

	for {
		fmt.Fprintln(log, "ReadByte")
		c, err := b.ReadByte()
		if err != nil {
			fmt.Fprintln(log, "Top EOF")
			if err == io.EOF {
				if len(cur) > 0 {
					tokens = append(tokens, Token{
						Pos: Pos{
							Line: line,
							Col:  col - (len(cur) - 1),
						},
						Type:  TTypeFromBytes(cur),
						Value: cur,
					})
				}
				break
			}
			panic(err)
		}

		fmt.Fprintln(log, "Incr IDX by 1")
		col++

		fmt.Fprintln(log, "Switch c")
		switch c {
		case '"':
			fmt.Fprintln(log, "Case \"")

			fmt.Fprintf(log, "Read to %s\n", string(c))
			toToken, err := b.ReadBytes(c)
			if err != nil {
				if err == io.EOF {
					panic(fmt.Sprintf("could not find matching quote at line %d, col %d", line, col))
				}
				panic(err)
			}

			fmt.Fprintf(log, "Incr IDX by %d\n", len(toToken))
			col += len(toToken)

			fmt.Fprintf(log, "Append %s to tokens\n", toToken[:len(toToken)-1])
			tokens = append(tokens, Token{
				Pos: Pos{
					Line: line,
					Col:  col - len(toToken),
				},
				Type:  STRING,
				Value: toToken[:len(toToken)-1],
			})

			continue
		case ' ', '\t', '\n':
			fmt.Fprintln(log, "Case space, \\n or \\t")

			fmt.Fprintf(log, "Append %s to tokens\n", cur)
			if len(cur) > 0 {
				tokens = append(tokens, Token{
					Pos: Pos{
						Line: line,
						Col:  col - (len(cur) + 1),
					},
					Type:  TTypeFromBytes(cur),
					Value: cur,
				})
			}
			cur = make([]byte, 0)

			if c == '\n' {
				line++
				col = 0
			}

			continue
		}

		fmt.Fprintf(log, "Append %s to cur\n", string(c))
		cur = append(cur, c)
	}

	fmt.Fprintf(log, "%s\n", tokens)

	for idx, b := range tokens {
		switch b.Type {
		case VAR:
			if len(tokens) <= idx+2 {
				panic(fmt.Sprintf("expected IDENT, got EOF at %s", b.Pos))
			}
			if tokens[idx+1].Type != IDENT {
				panic(fmt.Sprintf("expected IDENT, got %s at %s", tokens[idx+1].Type, b.Pos))
			}
			if tokens[idx+2].Type != STRING {
				panic(fmt.Sprintf("expected IDENT, got %s at %s", tokens[idx+2].Type, b.Pos))
			}

			vars[string(tokens[idx+1].Value)] = tokens[idx+2].Value
			skipTokens = 2
		case P:
			print(w, tokens, idx, b, vars, "")
			skipTokens = 1
		case PLN:
			print(w, tokens, idx, b, vars, "\n")
			skipTokens = 1
		default:
			if skipTokens > 0 {
				skipTokens--
				continue
			}
			panic(fmt.Sprintf("unexpected token at %s", b.Pos))
		}
	}
}

func print(w io.Writer, tokens []Token, idx int, b Token, vars map[string]interface{}, suffix string) {
	if len(tokens) <= idx+1 {
		panic(fmt.Sprintf("expected IDENT, got EOF at %s", b.Pos))
	}
	switch tokens[idx+1].Type {
	case STRING:
		fmt.Fprintf(w, "%s%s", tokens[idx+1].Value, suffix)
	case IDENT:
		fmt.Fprintf(w, "%s%s", vars[string(tokens[idx+1].Value)], suffix)
	default:
		panic(fmt.Sprintf("expected STRING, got %s at %s", tokens[idx+1].Type, b.Pos))
	}
}

type TType int

const (
	IDENT TType = iota
	STRING

	P
	PLN
	VAR
)

var stringToTType = map[string]TType{
	"p":   P,
	"pln": PLN,
	"var": VAR,
}
var tTypeToString = map[TType]string{
	P:      "p",
	PLN:    "pln",
	IDENT:  "IDENT",
	STRING: "string",
	VAR:    "var",
}

func TTypeFromBytes(b []byte) TType {
	if tt, ok := stringToTType[string(b)]; ok {
		return tt
	}
	return IDENT
}
func (t TType) String() string {
	if str, ok := tTypeToString[t]; ok {
		return str
	}

	return "unknown"
}

type Token struct {
	Pos
	Type  TType
	Value []byte
}

type Pos struct {
	Line int
	Col  int
}

func (p Pos) String() string {
	return fmt.Sprintf("line %d, col %d", p.Line, p.Col)
}
