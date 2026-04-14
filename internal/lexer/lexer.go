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
		parts := strings.Fields(line)

		for i := 0; i < len(parts); i++ {
			part := parts[i]

			switch part {
			case "=":
				tokens = append(tokens, token.Token{
					Type:    token.EQUAL,
					Literal: part,
				})
				continue
			case ">":
				tokens = append(tokens, token.Token{
					Type:    token.GREATER,
					Literal: part,
				})
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
