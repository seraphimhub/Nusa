package lexer

import (
	"testing"

	"github.com/seraphimhub/Nusa/internal/token"
)

// FuzzTokenize memastikan lexer tidak pernah panik atau berputar tanpa henti
// untuk masukan apa pun, termasuk potongan UTF-8 yang rusak.
func FuzzTokenize(f *testing.F) {
	seeds := []string{
		"",
		"buat x = 10",
		"tetap PI = 3.14",
		`tulis "halo\n dunia"`,
		"jika x >= 5 { tulis x } kalau_tidak { tulis 0 }",
		"coba { lemparkan \"galat\" } tangkap g { tulis g[\"pesan\"] }",
		"peta { \"a\": 1, \"b\": 2 }",
		"/* komentar */ # komentar\n// komentar",
		"impor \"modul.nusa\"",
		"uji \"nama\" { cek 1 == 1 }",
		"untuk i dari 1 sampai 10 langkah 2 { tulis i }",
		"\"tidak ditutup",
		"'a'",
		"\x00\xff\xfe",
		"1..2.3.4",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		tokens := Tokenize(input)

		if len(tokens) == 0 {
			t.Fatalf("Tokenize(%q) tidak mengembalikan token apa pun", input)
		}
		if last := tokens[len(tokens)-1]; last.Type != token.EOF {
			t.Fatalf("token terakhir = %s, ingin EOF", last.Type)
		}
		// Setiap karakter paling banyak menghasilkan satu token, jadi jumlah
		// token tidak boleh melebihi panjang masukan. Ini menangkap
		// kemungkinan lexer berputar tanpa memajukan posisi.
		if max := len(input) + 1; len(tokens) > max {
			t.Fatalf("Tokenize(%q) menghasilkan %d token, batas wajar %d", input, len(tokens), max)
		}

		for i, tok := range tokens {
			if tok.Line < 1 || tok.Column < 1 {
				t.Fatalf("token %d (%s) punya posisi tidak valid: %d:%d", i, tok.Type, tok.Line, tok.Column)
			}
		}
	})
}
