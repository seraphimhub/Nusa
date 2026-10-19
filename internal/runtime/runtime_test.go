package runtime_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/seraphimhub/Nusa/internal/lexer"
	"github.com/seraphimhub/Nusa/internal/parser"
	"github.com/seraphimhub/Nusa/internal/runtime"
)

func execute(t *testing.T, source string) (string, runtime.Object) {
	t.Helper()
	return executeWithInput(t, source, "")
}

func executeWithInput(t *testing.T, source string, input string) (string, runtime.Object) {
	t.Helper()

	tokens := lexer.Tokenize(source)
	p := parser.New(tokens)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("galat parser: %v\nsource:\n%s", p.Errors(), source)
	}

	var out bytes.Buffer
	rt := runtime.NewWithIO(&out, strings.NewReader(input))
	result := rt.Execute(program)
	return out.String(), result
}

func mustRun(t *testing.T, source string) string {
	t.Helper()
	out, result := execute(t, source)
	if result != nil && result.Type() == runtime.ERROR_OBJ {
		t.Fatalf("galat runtime: %s", result.Inspect())
	}
	return strings.TrimRight(out, "\n")
}

func expectOutput(t *testing.T, source string, want string) {
	t.Helper()
	got := mustRun(t, source)
	if got != want {
		t.Errorf("output salah.\ningin:\n%s\ndapat:\n%s", want, got)
	}
}

func TestArithmeticAndOutput(t *testing.T) {
	expectOutput(t, `
tulis 1 + 2 * 3
tulis 10 / 3
tulis 10 % 3
tulis "umur " + 25
tulis 2.5 + 1
`, "7\n3\n1\numur 25\n3.5")
}

func TestIfAndLoops(t *testing.T) {
	expectOutput(t, `
buat total = 0
untuk i dari 1 sampai 5 {
    jika i == 3 {
        lanjutkan
    }
    total += i
}
tulis total

ulang 3 {
    total += 1
}
tulis total

buat n = 0
selama n < 2 {
    n += 1
}
tulis n
`, "12\n15\n2")
}

func TestBreakInFor(t *testing.T) {
	expectOutput(t, `
untuk i dari 1 sampai 100 {
    jika i == 4 {
        berhenti
    }
    tulis i
}
`, "1\n2\n3")
}

func TestForInAndLists(t *testing.T) {
	expectOutput(t, `
buat buah = ["apel", "jeruk"]
buah[0] = "mangga"
tulis buah
tulis buah[-1]
untuk x dalam buah {
    tulis x
}
untuk aksara dalam "ab" {
    tulis aksara
}
tulis panjang(buah)
`, `["mangga", "jeruk"]
jeruk
mangga
jeruk
a
b
2`)
}

func TestBreakContinueInForIn(t *testing.T) {
	expectOutput(t, `
buat hasil = ""
untuk x dalam [1, 2, 3, 4] {
    jika x == 2 {
        lanjutkan
    }
    jika x == 4 {
        berhenti
    }
    hasil += ke_teks(x)
}
tulis hasil
`, "13")
}

func TestFunctionsRecursionAndScope(t *testing.T) {
	expectOutput(t, `
fungsi faktorial(n) {
    jika n <= 1 {
        kembali 1
    }
    kembali n * faktorial(n - 1)
}
tulis faktorial(5)

fungsi ubah() {
    x = 10
}
buat x = 1
ubah()
tulis x
`, "120\n10")
}

func TestReturnFromLoop(t *testing.T) {
	expectOutput(t, `
fungsi cariPertama(daftar, target) {
    untuk x dalam daftar {
        jika x == target {
            kembali benar
        }
    }
    kembali salah
}
tulis cariPertama([1, 2, 3], 2)
tulis cariPertama([1], 5)
`, "benar\nsalah")
}

func TestStringAndNegativeIndex(t *testing.T) {
	expectOutput(t, `
tulis "nusa"[0]
tulis "nusa"[-1]
tulis potong("Halo Dunia", 5)
`, "n\na\nDunia")
}

func TestShortCircuitEvaluation(t *testing.T) {
	expectOutput(t, `
buat aman = salah dan (1 / 0 == 0)
tulis aman
tulis benar atau (1 / 0 == 0)
`, "salah\nbenar")
}

func TestEqualityAcrossTypes(t *testing.T) {
	expectOutput(t, `
tulis 1 == 1
tulis 1 == "1"
tulis kosong == kosong
tulis [1, 2] == [1, 2]
tulis [1] != [2]
tulis 1 == 1.0
`, "benar\nsalah\nbenar\nbenar\nbenar\nbenar")
}

func TestBuiltins(t *testing.T) {
	expectOutput(t, `
tulis huruf_besar("nusa")
tulis huruf_kecil("NUSA")
tulis cari("Halo Dunia", "Dunia")
tulis gabung(["a", "b"], "-")
tulis jumlah([1, 2, 3])
tulis jumlah([1, 2.5])
tulis tipe(1) + " " + tipe(1.5) + " " + tipe("a") + " " + tipe(benar) + " " + tipe(kosong) + " " + tipe([])
tulis urutkan([3, 1, 2])
tulis balik([1, 2, 3])
tulis rentang(1, 5, 2)
tulis ke_angka("42") + 1
tulis ke_teks(benar)
tulis ke_desimal("2.5") + 1
tulis mutlak(-3)
tulis bulat(2.6) + " " + lantai(2.6) + " " + langit(2.1)
tulis pangkat(2, 8)
tulis akar(9)
`, `NUSA
nusa
5
a-b
6
3.5
angka desimal teks logika kosong daftar
[1, 2, 3]
[3, 2, 1]
[1, 3]
43
benar
3.5
3
3 2 3
256
3`)
}

func TestListMutation(t *testing.T) {
	expectOutput(t, `
buat angka = [1, 2]
tambahkan(angka, 3)
tulis angka
hapus(angka, 0)
tulis angka
angka[0] *= 10
tulis angka
`, "[1, 2, 3]\n[2, 3]\n[20, 3]")
}

func TestInput(t *testing.T) {
	out, result := executeWithInput(t, `
membaca nama
tulis "Halo " + nama
`, "Budi\n")
	if result != nil && result.Type() == runtime.ERROR_OBJ {
		t.Fatalf("galat runtime: %s", result.Inspect())
	}
	want := "> Halo Budi\n"
	if out != want {
		t.Errorf("output = %q, ingin %q", out, want)
	}
}

func TestRuntimeErrors(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{"bagi nol", `tulis 1 / 0`, "pembagian dengan nol"},
		{"variabel hilang", `tulis x`, "variabel 'x' tidak ditemukan"},
		{"indeks luar batas", `tulis [1][5]`, "di luar batas"},
		{"indeks negatif luar batas", `tulis [1][-2]`, "di luar batas"},
		{"argumen builtin", `tulis panjang()`, "membutuhkan 1 argumen"},
		{"operasi tipe salah", `tulis "a" - 1`, "tidak bisa dipakai"},
		{"ulang bukan angka", `ulang "x" { tulis 1 }`, "angka bulat"},
		{"konversi gagal", `tulis ke_angka("abc")`, "tidak bisa mengubah"},
		{"akar negatif", `tulis akar(-1)`, "negatif"},
		{"indeks tipe salah", `tulis 1[0]`, "tidak bisa diindeks"},
		{"parameter kurang", `fungsi f(a) { kembali a }
tulis f()`, "membutuhkan 1 parameter"},
		{"langkah nol", `untuk i dari 1 sampai 2 langkah 0 { tulis i }`, "langkah tidak boleh nol"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, result := execute(t, test.source)
			if result == nil || result.Type() != runtime.ERROR_OBJ {
				t.Fatalf("ingin galat runtime, dapat %#v", result)
			}
			if !strings.Contains(result.Inspect(), test.want) {
				t.Errorf("galat = %q, ingin mengandung %q", result.Inspect(), test.want)
			}
		})
	}
}
