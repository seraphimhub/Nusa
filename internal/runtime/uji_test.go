package runtime_test

import (
	"strings"
	"testing"

	"github.com/seraphimhub/Nusa/internal/lexer"
	"github.com/seraphimhub/Nusa/internal/parser"
	"github.com/seraphimhub/Nusa/internal/runtime"
)

// jalankanUji mengeksekusi program dan mengembalikan hasil pemeriksaan.
func jalankanUji(t *testing.T, source string) []runtime.HasilCek {
	t.Helper()

	tokens := lexer.Tokenize(source)
	p := parser.New(tokens)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("galat parser: %v\nsource:\n%s", p.Errors(), source)
	}

	rt := runtime.New()
	rt.Execute(program)

	return rt.HasilCek()
}

func TestUjiSemuaLulus(t *testing.T) {
	hasil := jalankanUji(t, `
fungsi tambah(a, b) {
    kembali a + b
}

uji "penjumlahan" {
    cek tambah(1, 2) == 3
    cek tambah(1, 2) != 4
}
`)

	if len(hasil) != 2 {
		t.Fatalf("jumlah hasil = %d, ingin 2", len(hasil))
	}
	for _, h := range hasil {
		if !h.Lolos {
			t.Errorf("pemeriksaan '%s' gagal: %s", h.Ekspresi, h.Pesan)
		}
		if h.Uji != "penjumlahan" {
			t.Errorf("nama uji = %q, ingin %q", h.Uji, "penjumlahan")
		}
	}
}

func TestUjiAdaKegagalan(t *testing.T) {
	hasil := jalankanUji(t, `
uji "salah" {
    cek 1 == 1
    cek 1 == 2, "satu tidak sama dengan dua"
}
`)

	if len(hasil) != 2 {
		t.Fatalf("jumlah hasil = %d, ingin 2", len(hasil))
	}
	if !hasil[0].Lolos {
		t.Errorf("pemeriksaan pertama seharusnya lulus")
	}
	if hasil[1].Lolos {
		t.Errorf("pemeriksaan kedua seharusnya gagal")
	}
	if hasil[1].Pesan != "satu tidak sama dengan dua" {
		t.Errorf("pesan = %q, ingin %q", hasil[1].Pesan, "satu tidak sama dengan dua")
	}
}

func TestUjiKegagalanTidakMenghentikanBlok(t *testing.T) {
	// Pemeriksaan yang gagal di dalam blok uji tidak boleh menghentikan
	// pemeriksaan berikutnya.
	hasil := jalankanUji(t, `
uji "berseveral" {
    cek 1 == 2
    cek 2 == 2
    cek 3 == 4
}
`)

	if len(hasil) != 3 {
		t.Fatalf("jumlah hasil = %d, ingin 3", len(hasil))
	}
	if hasil[0].Lolos || !hasil[1].Lolos || hasil[2].Lolos {
		t.Errorf("status pemeriksaan salah: %v", hasil)
	}
}

func TestCekDiLuarBlokUjiGagalStop(t *testing.T) {
	// Di luar blok uji, pemeriksaan yang gagal menghentikan program.
	tokens := lexer.Tokenize("cek 1 == 2\ntulis \"tidak tercapai\"\n")
	p := parser.New(tokens)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("galat parser: %v", p.Errors())
	}

	rt := runtime.New()
	result := rt.Execute(program)

	if result == nil || result.Type() != runtime.ERROR_OBJ {
		t.Fatalf("ingin galat runtime, dapat %#v", result)
	}
	if !strings.Contains(result.Inspect(), "harus bernilai benar") {
		t.Errorf("pesan galat = %q", result.Inspect())
	}
}

func TestUjiGalatDiLuarPemeriksaanDicatat(t *testing.T) {
	// Galat di dalam blok uji tapi di luar pemeriksaan dicatat sebagai
	// kegagalan. Blok dihentikan oleh galat itu, tetapi blok uji lain
	// tetap berjalan.
	hasil := jalankanUji(t, `
uji "panggil tak terdefinisi" {
    cek 1 == 1
    panggil_fungsi_yang_tidak_ada()
}

uji "masih berjalan" {
    cek 2 == 2
}
`)

	if len(hasil) != 3 {
		t.Fatalf("jumlah hasil = %d, ingin 3", len(hasil))
	}
	if !hasil[0].Lolos {
		t.Errorf("pemeriksaan pertama seharusnya lulus: %v", hasil[0])
	}
	if hasil[1].Lolos {
		t.Errorf("galat seharusnya dicatat sebagai kegagalan: %v", hasil[1])
	}
	if !strings.Contains(hasil[1].Pesan, "tidak ditemukan") {
		t.Errorf("pesan kegagalan = %q", hasil[1].Pesan)
	}
	if !hasil[2].Lolos {
		t.Errorf("blok uji berikutnya harus tetap berjalan: %v", hasil[2])
	}
}

func TestKudiagMemberiRingkasan(t *testing.T) {
	output, _ := execute(t, "tulis kudiag()")

	for _, ingin := range []string{"langkah:", "kedalaman:", "variabel aktif:"} {
		if !strings.Contains(output, ingin) {
			t.Errorf("keluaran %q tidak mengandung %q", output, ingin)
		}
	}
}

func TestKudiagMenampilkanJejakFungsi(t *testing.T) {
	output, _ := execute(t, `
fungsi panggil_kudiag() {
    kembali kudiag()
}
fungsi panggil_luar() {
    kembali panggil_kudiag()
}
tulis panggil_luar()
`)

	if !strings.Contains(output, "jejak: panggil_luar -> panggil_kudiag") {
		t.Errorf("jejak fungsi tidak tampil: %q", output)
	}
}
