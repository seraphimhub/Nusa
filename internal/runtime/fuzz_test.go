package runtime_test

import (
	"io"
	"strings"
	"testing"

	"github.com/seraphimhub/Nusa/internal/lexer"
	"github.com/seraphimhub/Nusa/internal/parser"
	"github.com/seraphimhub/Nusa/internal/runtime"
)

// FuzzExecute memastikan interpreter tidak panik dan selalu berhenti untuk
// masukan acak. Batas eksekusi dibuat ketat supaya fuzzer tidak tersangkut
// program yang lambat.
func FuzzExecute(f *testing.F) {
	seeds := []string{
		"",
		"tulis 1 + 2",
		"buat d = [1, 2, 3]\ntulis jumlah(d)",
		"fungsi fib(n) { jika n <= 1 { kembali n } kembali fib(n - 1) + fib(n - 2) }\ntulis fib(10)",
		"tetap PI = 3.14\ntulis PI",
		"buat p = peta { \"a\": 1 }\ntulis p[\"a\"]",
		"coba { lemparkan \"x\" } tangkap g { tulis pesan(g) }",
		"selama benar { }",
		"uji \"nama\" { cek 1 == 1 }",
		"cek 1 == 1",
		"tulis kudiag()",
		"buat x: angka = 1",
		"fungsi f() { kembali f() }\nf()",
		"tulis \"teks\" + 1",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		tokens := lexer.Tokenize(input)
		p := parser.New(tokens)
		program := p.ParseProgram()
		if len(p.Errors()) > 0 {
			return
		}

		rt := runtime.NewWithIO(io.Discard, strings.NewReader(""))
		rt.SetLimits(runtime.Limits{MaxDepth: 40, MaxSteps: 20000})
		result := rt.Execute(program)

		// Apa pun hasilnya, interpreter harus mengembalikan objek yang bisa
		// dirender tanpa panik (termasuk objek galat).
		if result != nil {
			_ = result.Type()
			_ = result.Inspect()
		}
	})
}
