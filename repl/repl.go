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

		stackTop := machine.StackTop()
		io.WriteString(out, "Compiler Output:\n")
		io.WriteString(out, stackTop.Inspect())
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
