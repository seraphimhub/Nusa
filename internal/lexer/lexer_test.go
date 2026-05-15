package lexer

import (
	"testing"

	"github.com/seraphimhub/Nusa/internal/token"
)

func TestTokenize_StringAndOperators(t *testing.T) {
	input := "buat x = 10\njika x >= 10 tulis \"halo dunia\""
	tokens := Tokenize(input)

	wantTypes := []token.TokenType{
		token.BUAT, token.IDENT, token.EQUAL, token.NUMBER,
		token.JIKA, token.IDENT, token.GTE, token.NUMBER, token.TULIS, token.STRING,
		token.EOF,
	}

	if len(tokens) != len(wantTypes) {
		t.Fatalf("jumlah token tidak sesuai: got=%d want=%d", len(tokens), len(wantTypes))
	}

	for i, want := range wantTypes {
		if tokens[i].Type != want {
			t.Fatalf("token[%d] type mismatch: got=%s want=%s", i, tokens[i].Type, want)
		}
	}

	if tokens[9].Literal != "halo dunia" {
		t.Fatalf("literal string salah: got=%q", tokens[9].Literal)
	}
}

func TestTokenize_SkipComments(t *testing.T) {
	input := "# komentar\nbuat x = 1 // inline\ntulis x"
	tokens := Tokenize(input)

	wantTypes := []token.TokenType{token.BUAT, token.IDENT, token.EQUAL, token.NUMBER, token.TULIS, token.IDENT, token.EOF}
	if len(tokens) != len(wantTypes) {
		t.Fatalf("jumlah token tidak sesuai: got=%d want=%d", len(tokens), len(wantTypes))
	}
	for i, want := range wantTypes {
		if tokens[i].Type != want {
			t.Fatalf("token[%d] type mismatch: got=%s want=%s", i, tokens[i].Type, want)
		}
	}
}
