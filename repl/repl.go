package repl

import (
	"bufio"
	"fmt"
	"io"
	"rafiki/lexer"
	"rafiki/quotes"
	"rafiki/token"
)

const WELCOME = "Rafiki Version 0.1\nPress Ctrl+C to Exit\n"
const PROMPT = "rafiki >> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	fmt.Printf("\n")
	fmt.Fprintf(out, WELCOME)
	quotes.PrintQuote()
	fmt.Printf("\n")

	for {
		fmt.Fprintf(out, PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		l := lexer.NewLexer(line)

		for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
			fmt.Fprintf(out, "%+v\n", tok)
		}
	}
}
