package runtime

import (
	"fmt"
	"strconv"

	"github.com/seraphimhub/Nusa/internal/parser"
	"github.com/seraphimhub/Nusa/internal/token"
)

type Function struct {
	Action token.TokenType
	Value  string
}

type Runtime struct {
	Variables map[string]interface{}
	Functions map[string]Function
	LastIfResult bool
}

func New() *Runtime {
	return &Runtime{
		Variables: make(map[string]interface{}),
		Functions: make(map[string]Function),
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
		if num, err := strconv.Atoi(val); err == nil {
			return num
		}
	}
	return 0
}

func (r *Runtime) resolveValue(name string) interface{} {
	if val, ok := r.Variables[name]; ok {
		return val
	}
	return parseValue(name)
}

func (r *Runtime) printValue(value string) {
	fmt.Println(r.resolveValue(value))
}

func (r *Runtime) runFunction(name string) {
	fn, ok := r.Functions[name]
	if !ok {
		fmt.Println("fungsi tidak ditemukan:", name)
		return
	}
	if fn.Action == token.TULIS {
		r.printValue(fn.Value)
	}
}

func (r *Runtime) calculate(left int, op token.TokenType, right int) int {
	switch op {
	case token.PLUS:
		return left + right
	case token.MINUS:
		return left - right
	case token.MULTIPLY:
		return left * right
	case token.DIVIDE:
		if right == 0 {
			return 0
		}
		return left / right
	}
	return left
}

func (r *Runtime) compare(left int, op token.TokenType, right int) bool {
	switch op {
	case token.GREATER:
		return left > right
	case token.LESS:
		return left < right
	case token.GTE:
		return left >= right
	case token.LTE:
		return left <= right
	case token.EQ:
		return left == right
	case token.NEQ:
		return left != right
	}
	return false
}

func (r *Runtime) Execute(p *parser.Parser) {
	for {
		tok := p.Next()

		switch tok.Type {
		case token.BUAT:
			name := p.Next()
			p.Next()
			left := p.Next()

			if p.Pos < len(p.Tokens) {
				nextType := p.Tokens[p.Pos].Type
				if nextType == token.PLUS ||
					nextType == token.MINUS ||
					nextType == token.MULTIPLY ||
					nextType == token.DIVIDE {

					operator := p.Next()
					right := p.Next()

					leftVal := toInt(r.resolveValue(left.Literal))
					rightVal := toInt(r.resolveValue(right.Literal))

					r.Variables[name.Literal] = r.calculate(leftVal, operator.Type, rightVal)
					continue
				}
			}

			r.Variables[name.Literal] = r.resolveValue(left.Literal)

		case token.TULIS:
			value := p.Next()
			r.printValue(value.Literal)

		case token.JIKA:
			left := p.Next()
			operator := p.Next()
			right := p.Next()
			action := p.Next()
			value := p.Next()

			leftVal := toInt(r.resolveValue(left.Literal))
			rightVal := toInt(r.resolveValue(right.Literal))

			r.LastIfResult = r.compare(leftVal, operator.Type, rightVal)

			if r.LastIfResult {
				if action.Type == token.TULIS {
					r.printValue(value.Literal)
				}
			}

		case token.KALAU_TIDAK:
			action := p.Next()
			value := p.Next()

			if !r.LastIfResult {
				if action.Type == token.TULIS {
					r.printValue(value.Literal)
				}
			}

		case token.ULANG:
			countTok := p.Next()
			action := p.Next()
			value := p.Next()

			count := toInt(r.resolveValue(countTok.Literal))

			for i := 0; i < count; i++ {
				if action.Type == token.TULIS {
					r.printValue(value.Literal)
				}
			}

		case token.FUNGSI:
			name := p.Next()
			action := p.Next()
			value := p.Next()

			r.Functions[name.Literal] = Function{
				Action: action.Type,
				Value:  value.Literal,
			}

		case token.PANGGIL:
			name := p.Next()
			r.runFunction(name.Literal)

		case token.EOF:
			return
		}
	}
}
