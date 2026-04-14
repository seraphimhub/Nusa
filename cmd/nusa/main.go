package main

import (
	"fmt"
	"os"

	"github.com/seraphimhub/Nusa/internal/lexer"
	"github.com/seraphimhub/Nusa/internal/parser"
	"github.com/seraphimhub/Nusa/internal/runtime"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("pakai: nusa run file.nusa")
		return
	}

	command := os.Args[1]
	filename := os.Args[2]

	if command != "run" {
		fmt.Println("command hanya support: run")
		return
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	tokens := lexer.Tokenize(string(data))
	p := parser.New(tokens)

	rt := runtime.New()
	rt.Execute(p)
}
