package lexer

import (
	"strings"

	"github.com/kamu/nusa/internal/token"
)

func Tokenize(input string) []token.Token {
	words := strings.Fields(input)
	var tokens []token.Token

	for i := 0; i < len(words); i++ {
		word := words[i]

		switch word {
		case "buat":
			tokens = append(tokens, token.Token{Type: token.BUAT, Literal: word})
		case "tulis":
			tokens = append(tokens, token.Token{Type: token.TULIS, Literal: word})
		case "=":
			tokens = append(tokens, token.Token{Type: token.EQUAL, Literal: word})
		default:
			tokens = append(tokens, token.Token{Type: token.IDENT, Literal: word})
		}
	}

	return tokens
}