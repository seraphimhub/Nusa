package cek_test

import (
	"strings"
	"testing"

	"github.com/seraphimhub/Nusa/internal/cek"
	"github.com/seraphimhub/Nusa/internal/lexer"
	"github.com/seraphimhub/Nusa/internal/parser"
)

// periksa mem-parse dan memeriksa sebuah program Nusa.
func periksa(t *testing.T, sumber string) []cek.Temuan {
	t.Helper()

	tokens := lexer.Tokenize(sumber)
	p := parser.New(tokens)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("program gagal di-parse: %v", p.Errors())
	}

	return cek.Periksa(program)
}

// cariPesan mencari temuan yang pesannya memuat potongan tertentu.
func cariPesan(temuan []cek.Temuan, potongan string) *cek.Temuan {
	for i := range temuan {
		if strings.Contains(temuan[i].Pesan, potongan) {
			return &temuan[i]
		}
	}
	return nil
}

func TestCekVariabelTidakDitemukan(t *testing.T) {
	temuan := periksa(t, `tulis nama`)

	hasil := cariPesan(temuan, "tidak ditemukan")
	if hasil == nil {
		t.Fatalf("tidak ada temuan variabel tak terdefinisi; dapat: %v", temuan)
	}
	if hasil.Severity != cek.Galat {
		t.Errorf("severity = %q, ingin %q", hasil.Severity, cek.Galat)
	}
	if hasil.Line != 1 {
		t.Errorf("baris = %d, ingin 1", hasil.Line)
	}
}

func TestCekVariabelTakTerpakai(t *testing.T) {
	temuan := periksa(t, "buat x = 1\ntulis 2")

	hasil := cariPesan(temuan, "tidak pernah dipakai")
	if hasil == nil {
		t.Fatalf("tidak ada temuan variabel tak terpakai; dapat: %v", temuan)
	}
	if hasil.Severity != cek.Peringatan {
		t.Errorf("severity = %q, ingin %q", hasil.Severity, cek.Peringatan)
	}
}

func TestCekVariabelTerpakaiBersih(t *testing.T) {
	temuan := periksa(t, "buat x = 1\ntulis x")

	if len(temuan) != 0 {
		t.Fatalf("program bersih menghasilkan temuan: %v", temuan)
	}
}

func TestCekTipeTidakCocok(t *testing.T) {
	temuan := periksa(t, `buat umur: angka = "dua puluh"`)

	hasil := cariPesan(temuan, "harus bertipe angka")
	if hasil == nil {
		t.Fatalf("tidak ada temuan tipe; dapat: %v", temuan)
	}
	if hasil.Severity != cek.Galat {
		t.Errorf("severity = %q, ingin %q", hasil.Severity, cek.Galat)
	}
}

func TestCekTipeCocokTidakDITEMukan(t *testing.T) {
	temuan := periksa(t, "buat umur: angka = 20\ntulis umur")

	if len(temuan) != 0 {
		t.Fatalf("tipe yang cocok seharusnya bersih; dapat: %v", temuan)
	}
}

func TestCekAngkaBolehJadiDesimal(t *testing.T) {
	temuan := periksa(t, "buat x: desimal = 10\ntulis x")

	if len(temuan) != 0 {
		t.Fatalf("angka bulat boleh dipakai sebagai desimal; dapat: %v", temuan)
	}
}

func TestCekDeklarasiGanda(t *testing.T) {
	temuan := periksa(t, "buat x = 1\nbuat x = 2\ntulis x")

	hasil := cariPesan(temuan, "sudah dideklarasikan")
	if hasil == nil {
		t.Fatalf("tidak ada temuan deklarasi ganda; dapat: %v", temuan)
	}
	if hasil.Severity != cek.Galat {
		t.Errorf("severity = %q, ingin %q", hasil.Severity, cek.Galat)
	}
}

func TestCekFungsiTidakTerdefinisi(t *testing.T) {
	temuan := periksa(t, "tambah(1, 2)")

	if cariPesan(temuan, "tidak ditemukan") == nil {
		t.Fatalf("fungsi tak terdefinisi seharusnya dilaporkan; dapat: %v", temuan)
	}
}

func TestCekFungsiBawaanDikenal(t *testing.T) {
	temuan := periksa(t, "tulis panjang([1, 2, 3])")

	if len(temuan) != 0 {
		t.Fatalf("fungsi bawaan harus dikenal; dapat: %v", temuan)
	}
}

func TestCekParameterTakTerpakai(t *testing.T) {
	temuan := periksa(t, "fungsi sapa(nama) {\n    kembali \"halo\"\n}\nsapa(\"Budi\")")

	hasil := cariPesan(temuan, "parameter 'nama' tidak pernah dipakai")
	if hasil == nil {
		t.Fatalf("tidak ada temuan parameter tak terpakai; dapat: %v", temuan)
	}
	if hasil.Severity != cek.Peringatan {
		t.Errorf("severity = %q, ingin %q", hasil.Severity, cek.Peringatan)
	}
}

func TestCekParameterTerpakaiBersih(t *testing.T) {
	temuan := periksa(t, "fungsi sapa(nama) {\n    kembali \"halo \" + nama\n}\nsapa(\"Budi\")")

	if len(temuan) != 0 {
		t.Fatalf("parameter yang dipakai seharusnya bersih; dapat: %v", temuan)
	}
}

func TestCekLingkupFungsiTerpisah(t *testing.T) {
	// Variabel di luar fungsi tidak boleh terlihat di dalam fungsi.
	temuan := periksa(t, "fungsi f() {\n    return luar\n}\nf()")

	if cariPesan(temuan, "tidak ditemukan") == nil {
		t.Fatalf("variabel luar tidak boleh terlihat di fungsi; dapat: %v", temuan)
	}
}

func TestCekLingkupBlokBerbagi(t *testing.T) {
	// Blok di Nusa berbagi lingkup dengan lingkup luarnya, sehingga
	// variabel yang dibuat di dalam blok tetap terlihat setelahnya.
	temuan := periksa(t, "jika benar {\n    buat nilai = 1\n}\ntulis nilai")

	if len(temuan) != 0 {
		t.Fatalf("variabel blok harus terlihat di luar blok; dapat: %v", temuan)
	}
}

func TestCekTeksTidakDisebutSebagaiGalat(t *testing.T) {
	// Nama fungsi bawaan yang dipakai sebagai nilai harus tetap dikenal.
	temuan := periksa(t, "buat f = panjang\ntulis f(\"abc\")")

	if len(temuan) != 0 {
		t.Fatalf("builtin sebagai nilai seharusnya bersih; dapat: %v", temuan)
	}
}

func TestCekProgramKosong(t *testing.T) {
	temuan := periksa(t, "")

	if len(temuan) != 0 {
		t.Fatalf("program kosong seharusnya bersih; dapat: %v", temuan)
	}
}
