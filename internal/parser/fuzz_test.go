package parser_test

import (
	"testing"

	"github.com/seraphimhub/Nusa/internal/lexer"
	"github.com/seraphimhub/Nusa/internal/parser"
)

// FuzzParse memastikan parser selalu berhenti dan tidak panik, baik untuk
// program valid maupun masukan acak. Bila parsing berhasil, AST juga harus
// bisa dicetak ulang tanpa panik.
func FuzzParse(f *testing.F) {
	seeds := []string{
		"",
		"buat x = 1",
		"fungsi f(a, b) { kembali a + b }",
		"untuk k, v dalam peta { \"a\": 1 } { tulis k + v }",
		"coba { lemparkan \"x\" } tangkap g { tulis g[\"pesan\"] }",
		"pilih x { kasus 1 { tulis 1 } bawaan { tulis 2 } }",
		"selama benar { berhenti }",
		"impor \"modul.nusa\"",
		"uji \"nama\" { cek 1 == 1 }",
		"cek 1 == 1, \"pesan\"",
		"buat x: angka = 1",
		"fungsi f(a: teks): angka { kembali 1 }",
		"[[[[[[[[",
		") ) ) ((((",
		"\"teks",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		tokens := lexer.Tokenize(input)
		p := parser.New(tokens)
		program := p.ParseProgram()

		if program == nil {
			t.Fatal("ParseProgram mengembalikan nil")
		}
		// Pemulihan galat harus berhenti; batas ini menangkap loop tak
		// berujung pada strategi sinkronisasi parser.
		if max := len(tokens) + 10; len(p.Errors()) > max {
			t.Fatalf("parser menghasilkan %d galat untuk %d token", len(p.Errors()), len(tokens))
		}

		_ = program.TokenLiteral()
		if len(p.Errors()) == 0 {
			// Program kosong (misalnya hanya spasi atau komentar) sah
			// saja; yang tidak boleh terjadi adalah program yang punya
			// pernyataan tetapi tidak bisa dicetak ulang.
			if len(program.Statements) > 0 && program.String() == "" {
				t.Fatalf("AST kosong untuk program yang berhasil di-parse: %q", input)
			}
		}
	})
}
