package runtime

import (
	"fmt"
	"strconv"

	"github.com/seraphimhub/Nusa/internal/parser"
	"github.com/seraphimhub/Nusa/internal/token"
)

type Runtime struct {
	Variables map[string]interface{}
}

func New() *Runtime {
	return &Runtime{
		Variables: make(map[string]interface{}),
	}
}

func parseValue(literal string) interface{} {
	if num, err := strconv.Atoi(literal); err == nil {
		return num
	}
	return literal
}

func toInt(v interface{}) int {
	switch val := v.(type) {
	case int:
		return val
	case string:
		n, _ := strconv.Atoi(val)
		return n
	default:
		return 0
	}
}

func (r *Runtime) printValue(value string) {
	if val, ok := r.Variables[value]; ok {
		fmt.Println(val)
	} else {
		fmt.Println(parseValue(value))
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
			r.Variables[name.Literal] = parseValue(value.Literal)

		case token.TULIS:
			value := p.Next()
			r.printValue(value.Literal)

		case token.JIKA:
			left := p.Next()
			operator := p.Next()
			right := p.Next()
			action := p.Next()
			value := p.Next()

			leftVal := toInt(r.Variables[left.Literal])
			rightVal := toInt(parseValue(right.Literal))

			if operator.Type == token.GREATER && leftVal > rightVal {
				if action.Type == token.TULIS {
					r.printValue(value.Literal)
				}
			}

		case token.ULANG:
			countTok := p.Next()
			action := p.Next()
			value := p.Next()

			count := toInt(parseValue(countTok.Literal))

			for i := 0; i < count; i++ {
				if action.Type == token.TULIS {
					r.printValue(value.Literal)
				}
			}

		case token.EOF:
			return
		}
	}
}
