package lexer

import (
	"strings"

	"github.com/seraphimhub/Nusa/internal/token"
)

func Tokenize(input string) []token.Token {
	var tokens []token.Token
	lines := strings.Split(input, "\n")

	for _, line := range lines {
		parts := strings.Fields(line)

		for i := 0; i < len(parts); i++ {
			part := parts[i]

			if part == "=" {
				tokens = append(tokens, token.Token{
					Type:    token.EQUAL,
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

			tokens = append(tokens, token.Token{
				Type:    token.LookupIdent(part),
				Literal: part,
			})
		}
	}

	tokens = append(tokens, token.Token{
		Type: token.EOF,
	})

	return tokens
}
