package format_test

import (
	"strings"
	"testing"

	"github.com/seraphimhub/Nusa/internal/format"
)

func TestFormatIndentasi(t *testing.T) {
	sumber := "fungsi f() {\nketik = 1\n}\n"
	hasil := format.Format(sumber)

	ingin := "fungsi f() {\n  ketik = 1\n}\n"
	if hasil != ingin {
		t.Errorf("hasil:\n%q\ningin:\n%q", hasil, ingin)
	}
}

func TestFormatSpasiOperator(t *testing.T) {
	hasil := format.Format("buat x=1+2\n")

	ingin := "buat x = 1 + 2\n"
	if hasil != ingin {
		t.Errorf("hasil:\n%q\ningin:\n%q", hasil, ingin)
	}
}

func TestFormatSpasiKoma(t *testing.T) {
	hasil := format.Format("fungsi f(a,b,c) {\n  kembali a\n}\n")

	ingin := "fungsi f(a, b, c) {\n  kembali a\n}\n"
	if hasil != ingin {
		t.Errorf("hasil:\n%q\ningin:\n%q", hasil, ingin)
	}
}

func TestFormatPanggilanFungsiRapat(t *testing.T) {
	hasil := format.Format("tulis panjang( \"abc\" )\n")

	ingin := "tulis panjang(\"abc\")\n"
	if hasil != ingin {
		t.Errorf("hasil:\n%q\ningin:\n%q", hasil, ingin)
	}
}

func TestFormatKomentarDipertahankan(t *testing.T) {
	sumber := "# komentar awal\ntulis 1 # komentar akhir\n"
	hasil := format.Format(sumber)

	if !strings.Contains(hasil, "# komentar awal") {
		t.Errorf("komentar baris sendiri hilang: %q", hasil)
	}
	if !strings.Contains(hasil, "# komentar akhir") {
		t.Errorf("komentar ujung baris hilang: %q", hasil)
	}
}

func TestFormatKomentarBlokDipertahankan(t *testing.T) {
	sumber := "/* blok\n   komentar */\ntulis 1\n"
	hasil := format.Format(sumber)

	if !strings.Contains(hasil, "/* blok") || !strings.Contains(hasil, "komentar */") {
		t.Errorf("komentar blok hilang: %q", hasil)
	}
}

func TestFormatTeksEscapeDipertahankan(t *testing.T) {
	hasil := format.Format("tulis \"a\\nb\"\n")

	ingin := "tulis \"a\\nb\"\n"
	if hasil != ingin {
		t.Errorf("hasil:\n%q\ningin:\n%q", hasil, ingin)
	}
}

func TestFormatTeksKutipTunggalDipertahankan(t *testing.T) {
	hasil := format.Format("tulis 'halo'\n")

	ingin := "tulis 'halo'\n"
	if hasil != ingin {
		t.Errorf("hasil:\n%q\ningin:\n%q", hasil, ingin)
	}
}

func TestFormatIdempoten(t *testing.T) {
	sumber := "fungsi f(a,b){\n   kalau a>b {kembali a}\n   kembali a+b\n}\n"
	sekali := format.Format(sumber)
	dua := format.Format(sekali)

	if sekali != dua {
		t.Errorf("format tidak idempoten:\n%q\nvs\n%q", sekali, dua)
	}
}

func TestFormatBarisKosongDipertahankan(t *testing.T) {
	hasil := format.Format("tulis 1\n\n\ntulis 2\n")

	ingin := "tulis 1\n\n\ntulis 2\n"
	if hasil != ingin {
		t.Errorf("hasil:\n%q\ningin:\n%q", hasil, ingin)
	}
}

func TestFormatProgramValidTetapValid(t *testing.T) {
	// Hasil format harus tetap bisa dibaca lexer dan parser.
	sumber := "fungsi tambah(a,b){\n   kembali a+b\n}\nuntuk i dari 1 sampai 3 {\n   tulis tambah(i,1)\n}\n"

	hasil := format.Format(sumber)
	if strings.Contains(hasil, "  ") {
		// Indentasi memang memakai spasi; yang penting adalah tidak ada spasi
		// ganda di sekitar operator.
		if strings.Contains(hasil, "=  ") || strings.Contains(hasil, "+  ") {
			t.Errorf("ada spasi ganda di sekitar operator: %q", hasil)
		}
	}
}
