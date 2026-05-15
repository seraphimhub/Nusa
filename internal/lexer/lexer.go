package lexer

import (
	"strings"
	"unicode"

	"github.com/seraphimhub/Nusa/internal/token"
)

func isNumber(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return len(s) > 0
}

func Tokenize(input string) []token.Token {
	var tokens []token.Token
	lines := strings.Split(input, "\n")

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") || strings.HasPrefix(trimmedLine, "//") {
			continue
		}

		parts := strings.Fields(line)

		for i := 0; i < len(parts); i++ {
			part := parts[i]

			if part == "#" || part == "//" {
				break
			}

			switch part {
			case "=":
				tokens = append(tokens, token.Token{Type: token.EQUAL, Literal: part})
				continue
			case ">":
				tokens = append(tokens, token.Token{Type: token.GREATER, Literal: part})
				continue
			case "<":
				tokens = append(tokens, token.Token{Type: token.LESS, Literal: part})
				continue
			case ">=":
				tokens = append(tokens, token.Token{Type: token.GTE, Literal: part})
				continue
			case "<=":
				tokens = append(tokens, token.Token{Type: token.LTE, Literal: part})
				continue
			case "==":
				tokens = append(tokens, token.Token{Type: token.EQ, Literal: part})
				continue
			case "!=":
				tokens = append(tokens, token.Token{Type: token.NEQ, Literal: part})
				continue
			case "+":
				tokens = append(tokens, token.Token{Type: token.PLUS, Literal: part})
				continue
			case "-":
				tokens = append(tokens, token.Token{Type: token.MINUS, Literal: part})
				continue
			case "*":
				tokens = append(tokens, token.Token{Type: token.MULTIPLY, Literal: part})
				continue
			case "/":
				tokens = append(tokens, token.Token{Type: token.DIVIDE, Literal: part})
				continue
			}

			if strings.HasPrefix(part, "\"") {
				value := part
				for !strings.HasSuffix(value, "\"") && i+1 < len(parts) {
					i++
					value += " " + parts[i]
				}
				value = strings.Trim(value, "\"")

				tokens = append(tokens, token.Token{
					Type:    token.STRING,
					Literal: value,
				})
				continue
			}

			if isNumber(part) {
				tokens = append(tokens, token.Token{
					Type:    token.NUMBER,
					Literal: part,
				})
				continue
			}

			tokens = append(tokens, token.Token{
				Type:    token.LookupIdent(part),
				Literal: part,
			})
		}
	}

	tokens = append(tokens, token.Token{Type: token.EOF})
	return tokens
}
