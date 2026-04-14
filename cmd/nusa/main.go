package main

import (
	"fmt"
	"os"

	"github.com/kamu/nusa/internal/lexer"
	"github.com/kamu/nusa/internal/runtime"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("pakai: nusa run file.nusa")
		return
	}

	filename := os.Args[2]

	data, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	tokens := lexer.Tokenize(string(data))

	rt := runtime.New()
	rt.Execute(tokens)
}