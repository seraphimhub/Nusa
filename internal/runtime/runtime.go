package runtime

import (
	"fmt"

	"github.com/kamu/nusa/internal/parser"
	"github.com/kamu/nusa/internal/token"
)

type Runtime struct {
	Variables map[string]string
}

func New() *Runtime {
	return &Runtime{
		Variables: make(map[string]string),
	}
}

func (r *Runtime) Execute(p *parser.Parser) {
	for {
		tok := p.Next()

		switch tok.Type {
		case token.BUAT:
			name := p.Next()
			p.Next() // skip =
			value := p.Next()

			r.Variables[name.Literal] = value.Literal

		case token.TULIS:
			value := p.Next()

			if val, ok := r.Variables[value.Literal]; ok {
				fmt.Println(val)
			} else {
				fmt.Println(value.Literal)
			}

		case token.EOF:
			return
		}
	}
}