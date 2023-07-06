package repl

import (
	"bufio"
	"fmt"
	"io"
	"rafiki/compiler"
	"rafiki/eval"
	"rafiki/lexer"
	"rafiki/object"
	"rafiki/parser"
	"rafiki/quotes"
	"rafiki/vm"
)

// TODO - find an equivalent package to readline and implement

const WELCOME = "Rafiki Version 0.1\nPress Ctrl+C to Exit\n"
const PROMPT = "rafiki >> "

const RAFIKI = `
$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$    ?$$$$$$$$$
$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$*     9$$$$$$$$
$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$' :X:  '$$$$$$$$
$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$E' !%!>  $$$$$$$$
$$$$$$$$$$$$$$$$$$$$$$#"$$$$$$$$$$$$$$$$$$$$$$$$. '"* .@$$$$$$$$
$$$$$$$$$$$$$$$$$$$$#  d$$$$$$$$$$$$$$$$$$$$$$R**   **$$$$$$$$$$
$$$$$$$$$$$$$$**$R"  z@***$$$$$$$$$$$$$$$$$$$$N/ -~<~z$$$$$$$$$$
$$$$$$$$$$$$$$  " .. <.()d$$$$$$$$$$$$$$$$$$$$$P '!"J$$$$$$$$$$$
$$$$$$$$$$$$$$   '~ '?*R$$$$$$$$$$$$$$$$$$$$$$$ .Xf:$$$$$$$$$$$$
$$$$$$$$$$$$$F   (^.uU(.$$$$$$$$Z #$$$$$$$$$$$$b '~4$$$$$$$$$$$$
$$$$$$$$$$$$$  :~uiCJ8$$$$**$^*$"  "*$$$$$$$$$$$    $$$$$$$$$$$$
$$$$$$$$$$$$$   $$$$$$$$$$$hx@$~  :!hc#**$$$$$$$    '$$$$$$$$$$$
$$$$$$$$$$$$$  '$$$$$$$$#(tWE'' .~~ '4"*%/$$$$$$  t  #$$$$$$$$$$
$$$$$$$$$$$$$  '$$$$$$$( s$$E  x  '+     .$$$$$$  $   $$$$$$$$$$
$$$$$$$$$$$$$   $$$$$$$Fd\$$$ <$    :    "$$$$$$  $k  '$$$$$$$$$
$$$$$$$$$$$$$>  #$$$$$$  8$*" d$   - '     $$$$f <$$   #$$$$$$$$
$$$$$$$$$$$$$L  '$$$$$$@ "   4$$    '     :4$$$  9$$    #$$$$$$$
$$$$$$$$$$$$$k   ?$*#"  -    " $e         '  ^"  ""      #$$$$$$
$$$$$$$$$$$$$F                '"$N        .               $$$$$$
$$$$$$$$$$$$$                   $$k 'c  +$$>            '.9$$$$$
$$$$$$$$$$$$$        .uedW@Wc   '$$  '   #$     uuu...  9$$$$$$$
$$$$$$$$$*#*R$r   .e$$$$$$$$$    ^$$L    d^ $  4$$$$$Lo$$$$$$$$$
$$$$$$$$#'   "N.:$$$$$$$$$$$$      #N'F'' .@F  8$$$$$$$$$$$$$$$$
$$$$$$R'      "$$$$$$$$$$$$$$     9u"  .@\$$   $$$$$$$$$$$$$$$$$
$$$$$#   :$N.  #$$$$$$$$$$$$F     $$B  $F9$$  '$$$$$$$$$$$$$$$$$
$$$F    d$$$$c  "$$$$$$$$$$*     d$**NJFd$$$  '$$$$$$$$$$$$$$$$$
$$$Ncud$$$$$$$c  'R$$$$$$R"    x$GoWW@"d$$$$k '$$$$$$$$$$$$$$$$$
$$$$$$$$$$$$$$$i   "**#""  ~ s$$$$$*"z$$$$$$F  $$$$$$$$$$$$$$$$$
$$$$$$$$$$$$$$$$k         .u$$$$#"Lo$$$$$$$$F  $$$$$$$$$$$$$$$$$
$$$$$$$$$$$$$$$$$k'      d$$$"  @$$$$$$$$$$$K  R$$$$$$$$$$$$$$$$
$$$$$$$$$$$$$$$$P      :$$$*    ^*$$$$$$$$$$&  3$$$$$$$$$$$$$$$$
$$$$$$$$$$$$$$$R      @$$$F        ^#$$$$$$$$  3$$$$$$$$$$$$$$$$
$$$$$$$$$$$$$$$      ~)LuLL.          '$$$$$$> '$$$$$$$$$$$$$$$$
$$$$$$$$$$$$$$"     x$$$$$$$$$e.        "R$$$> '$$$$$$$$$$$$$$$$
$$$$$$$$$$$$$P     s$$$$$$$$$$$$c        @$$$K '$$$$$$$$$$$$$$$$
$$$$$$$$$$$$$      @$$$$$$$$$$$$$k       $$$$> '$$$$$$$$$$$$$$$$
$$$$$$$$$$$$$.    '$$$$$$$$$$$$$$F     .d$$$$> <$$$$$$$$$$$$$$$$
$$$$$$$$$$$$$$    '$$$$$$$$$$$$$$     :$$$$$$> 9$$$$$$$$$$$$$$$$
$$$$$$$$$$$$$$    x$$$$$$$$$$$$P   d@$$$$$$$$L 9$$$$$$$$$$$$$$$$
$$$$$$$$$$$$$$k:  '$$$$$$$$$$$"   """"***$$$$E @$$$$$$$$$$$$$$$$
$$$$$$$$$$$$$$$@b  '$$$$$$$$$N/=+.  ~:.%: #$$$ 9$$$$$$$$$$$$$$$$
$$$$$$$$$$$$$**""  ..#$$$$$$$$$$bed$$o(.Lx@$$$$$$$$$$$$$$$$$$$$$
$$$$$$$$$$$$   > <~ ~d$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$`

/*
	The execution order of the program with the Interpreter looks like this:

	Lexer: 		Take the input code and look it as a pure string.
				Look through that string for recognizable tokens - keywords, variables names, numbers, arrays, etc.
				This output will be an array of tokens that are in a format which we can reliably parse.

	Parser: 	Take the tokens from the Lexer.  Use those tokens to build an abstract syntax tree that relates the tokens to eachother.
				Instead of the Lexer's array of "1 + 2", we'll now have "+" as a parent node with "1" and "2" as child nodes.
				In the parser, we find our program's structure and draw up a map for evaluation.

	Evaluator:	Take the AST from the Parser. Walk through the tree from top to bottom. Turn nodes into real values and execute them.

	The Interpreter is a point-translator of code into execution.  It executes the code you give it right here, right now.

	The Compiler, on the hand, has a more complex set of operations.

	The execution order for the Compiler looks like this:

	Lexer:		Same behavior.

	Parser: 	Same behavior.

	Compiler:	Take in the AST from the Parser.  Turn it into different slices of data. One slice is the variables and constants declared
				in the program, located in arrays.  The other slice is the functions we'll perform on that data, with the "addresses"
				of where that data is located in the constant and variables arrays.  The functions are called OpCodes, the addresses, Operands.

				The Compiler, on the surface, will have a very similar structure to the Evaluator.  But instead of truly evaluating,
				we're translating our high level language into something much simpler to evaluate.

	VM:			The VM then takes these OpCodes and Operands and evaluate them in a very similar manner to the Evaluator.
*/

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	// We want our environment to persist between REPL calls
	e := object.NewEnvironment()
	macroEnv := object.NewEnvironment()

	fmt.Printf("\n")
	io.WriteString(out, RAFIKI)
	fmt.Printf("\n\n")
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
		p := parser.NewParser(l)

		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			printParserErrors(out, p.Errors())
			continue
		}

		eval.DefineMacros(program, macroEnv)
		expandedProgram := eval.ExpandMacros(program, macroEnv)

		compiler := compiler.NewCompiler()
		err := compiler.Compile(expandedProgram)
		if err != nil {
			fmt.Fprintf(out, "Woops! Compilation failed:\n %s\n", err)
			continue
		}

		machine := vm.NewVm(compiler.Bytecode())
		err = machine.Run()
		if err != nil {
			fmt.Fprintf(out, "Woops! Executing bytecode failed:\n %s\n", err)
			continue
		}

		lastPopped := machine.LastPoppedStackElem()
		io.WriteString(out, "Compiler Output:\n")
		io.WriteString(out, lastPopped.Inspect())
		io.WriteString(out, "\n")

		// Interpreted Output
		result := eval.Eval(expandedProgram, e)
		io.WriteString(out, "Interpreted Output:\n")
		io.WriteString(out, result.Inspect())
		io.WriteString(out, "\n")
	}
}

func printParserErrors(out io.Writer, errors []string) {
	// io.WriteString(out, RAFIKI)
	io.WriteString(out, "\n\n")
	io.WriteString(out, "Whoops! We ran into some monkey business here!\n\n")
	io.WriteString(out, "\tparser errors:\n")
	for _, msg := range errors {
		io.WriteString(out, "\t\t"+msg+"\n")
	}
	io.WriteString(out, "\n\n")
}
