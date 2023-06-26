// main.go

package main

import (
	"os"
	"rafiki/repl"
)

func main() {
	repl.Start(os.Stdin, os.Stdout)
}
