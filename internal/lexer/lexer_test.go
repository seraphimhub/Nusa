package lexer

import (
	"testing"

	"github.com/seraphimhub/Nusa/internal/token"
)

// meaningfulTokens membuang NEWLINE agar tabel test lebih ringkas.
func meaningfulTokens(input string) []token.Token {
	var result []token.Token
	for _, tok := range Tokenize(input) {
		if tok.Type != token.NEWLINE {
			result = append(result, tok)
		}
	}
	return result
}

func TestOperatorsAndDelimiters(t *testing.T) {
	input := "= == != > >= < <= + += - -= * *= / /= % %= ( ) { } [ ] ,"

	expected := []struct {
		typ token.TokenType
		lit string
	}{
		{token.ASSIGN, "="},
		{token.EQ, "=="},
		{token.NEQ, "!="},
		{token.GREATER, ">"},
		{token.GTE, ">="},
		{token.LESS, "<"},
		{token.LTE, "<="},
		{token.PLUS, "+"},
		{token.PLUS_ASSIGN, "+="},
		{token.MINUS, "-"},
		{token.MINUS_ASSIGN, "-="},
		{token.MULTIPLY, "*"},
		{token.MULTIPLY_ASSIGN, "*="},
		{token.DIVIDE, "/"},
		{token.DIVIDE_ASSIGN, "/="},
		{token.MODULO, "%"},
		{token.MODULO_ASSIGN, "%="},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RBRACE, "}"},
		{token.LBRACKET, "["},
		{token.RBRACKET, "]"},
		{token.COMMA, ","},
		{token.EOF, ""},
	}

	got := meaningfulTokens(input)
	if len(got) != len(expected) {
		t.Fatalf("jumlah token = %d, ingin %d: %#v", len(got), len(expected), got)
	}

	for i, want := range expected {
		if got[i].Type != want.typ {
			t.Errorf("token %d: tipe = %s, ingin %s", i, got[i].Type, want.typ)
		}
		if got[i].Literal != want.lit {
			t.Errorf("token %d: literal = %q, ingin %q", i, got[i].Literal, want.lit)
		}
	}
}

func TestKeywords(t *testing.T) {
	keywords := []struct {
		input string
		typ   token.TokenType
	}{
		{"buat", token.BUAT},
		{"tulis", token.TULIS},
		{"jika", token.JIKA},
		{"kalau_tidak", token.KALAU_TIDAK},
		{"ulang", token.ULANG},
		{"selama", token.SELAMA},
		{"untuk", token.UNTUK},
		{"dari", token.DARI},
		{"sampai", token.SAMPAI},
		{"langkah", token.LANGKAH},
		{"dalam", token.DALAM},
		{"pilih", token.PILIH},
		{"kasus", token.KASUS},
		{"bawaan", token.BAWAAN},
		{"fungsi", token.FUNGSI},
		{"panggil", token.PANGGIL},
		{"kembali", token.KEMBALI},
		{"membaca", token.MEMBACA},
		{"berhenti", token.BERHENTI},
		{"lanjutkan", token.LANJUTKAN},
		{"dan", token.DAN},
		{"atau", token.ATAU},
		{"tidak", token.TIDAK},
		{"benar", token.BENAR},
		{"salah", token.SALAH},
		{"kosong", token.KOSONG},
	}

	for _, test := range keywords {
		t.Run(test.input, func(t *testing.T) {
			tokens := meaningfulTokens(test.input)
			if tokens[0].Type != test.typ {
				t.Errorf("tipe = %s, ingin %s", tokens[0].Type, test.typ)
			}
		})
	}
}

func TestIdentifiersAndNumbers(t *testing.T) {
	tokens := meaningfulTokens("nilai_1 42 3.14 _rahasia")

	expected := []struct {
		typ token.TokenType
		lit string
	}{
		{token.IDENT, "nilai_1"},
		{token.NUMBER, "42"},
		{token.FLOAT, "3.14"},
		{token.IDENT, "_rahasia"},
		{token.EOF, ""},
	}

	for i, want := range expected {
		if tokens[i].Type != want.typ || tokens[i].Literal != want.lit {
			t.Errorf("token %d = (%s, %q), ingin (%s, %q)", i,
				tokens[i].Type, tokens[i].Literal, want.typ, want.lit)
		}
	}
}

func TestStringEscapesAndQuotes(t *testing.T) {
	input := `"halo\n\"dunia\"\t\\" 'teks tunggal'`

	tokens := meaningfulTokens(input)
	if tokens[0].Type != token.STRING {
		t.Fatalf("token 0 bukan STRING: %s", tokens[0].Type)
	}
	if want := "halo\n\"dunia\"\t\\"; tokens[0].Literal != want {
		t.Errorf("string = %q, ingin %q", tokens[0].Literal, want)
	}
	if tokens[1].Type != token.STRING || tokens[1].Literal != "teks tunggal" {
		t.Errorf("kutip tunggal = (%s, %q)", tokens[1].Type, tokens[1].Literal)
	}
}

func TestCommentsAndLineNumbers(t *testing.T) {
	input := "# komentar\nbuat x = 1 // komentar\n/* blok\nkomentar */\ntulis x"

	tokens := meaningfulTokens(input)
	expected := []struct {
		typ  token.TokenType
		line int
	}{
		{token.BUAT, 2},
		{token.IDENT, 2},
		{token.ASSIGN, 2},
		{token.NUMBER, 2},
		{token.TULIS, 5},
		{token.IDENT, 5},
		{token.EOF, 5},
	}

	if len(tokens) != len(expected) {
		t.Fatalf("jumlah token = %d, ingin %d: %#v", len(tokens), len(expected), tokens)
	}
	for i, want := range expected {
		if tokens[i].Type != want.typ || tokens[i].Line != want.line {
			t.Errorf("token %d = (%s, baris %d), ingin (%s, baris %d)", i,
				tokens[i].Type, tokens[i].Line, want.typ, want.line)
		}
	}
}

func TestColumns(t *testing.T) {
	tokens := meaningfulTokens("buat x = 1")

	expected := []int{1, 6, 8, 10, 11}
	for i, column := range expected {
		if tokens[i].Column != column {
			t.Errorf("token %d kolom = %d, ingin %d", i, tokens[i].Column, column)
		}
	}
}

func TestIllegalCharacter(t *testing.T) {
	tokens := meaningfulTokens("buat @ = 1")

	if tokens[1].Type != token.ILLEGAL || tokens[1].Literal != "@" {
		t.Errorf("token illegal = (%s, %q)", tokens[1].Type, tokens[1].Literal)
	}
}

func TestUnterminatedStringStopsAtEOF(t *testing.T) {
	tokens := meaningfulTokens(`"belum ditutup`)

	if tokens[0].Type != token.STRING {
		t.Fatalf("token 0 = %s, ingin STRING", tokens[0].Type)
	}
	if tokens[0].Literal != "belum ditutup" {
		t.Errorf("literal = %q", tokens[0].Literal)
	}
	if tokens[1].Type != token.EOF {
		t.Errorf("token 1 = %s, ingin EOF", tokens[1].Type)
	}
}
