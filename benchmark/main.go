package main

import (
	"flag"
	"fmt"
	"rafiki/compiler"
	"rafiki/eval"
	"rafiki/lexer"
	"rafiki/object"
	"rafiki/parser"
	"rafiki/vm"
	"time"
)

var engine = flag.String("engine", "vm", "use 'vm' or 'eval'")

var input = `
let fibonacci = fn(x) {
  if (x == 0) {
    0
  } else {
    if (x == 1) {
      return 1;
    } else {
      fibonacci(x - 1) + fibonacci(x - 2);
    }
  }
};
fibonacci(35);
`

func main() {
	flag.Parse()

	var duration time.Duration
	var result object.Object

	l := lexer.NewLexer(input)
	p := parser.NewParser(l)
	program := p.ParseProgram()

	if *engine == "vm" {
		comp := compiler.NewCompiler()
		err := comp.Compile(program)
		if err != nil {
			fmt.Printf("compiler error: %s", err)
			return
		}

		machine := vm.NewVm(comp.Bytecode())

		start := time.Now()

		err = machine.Run()
		if err != nil {
			fmt.Printf("vm error: %s", err)
			return
		}

		duration = time.Since(start)
		result = machine.LastPoppedStackElem()
	} else {
		env := object.NewEnvironment()
		start := time.Now()
		result = eval.Eval(program, env)
		duration = time.Since(start)
	}

	fmt.Printf(
		"engine=%s, result=%s, duration=%s\n",
		*engine,
		result.Inspect(),
		duration)
}
