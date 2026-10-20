package modul_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/seraphimhub/Nusa/internal/modul"
)

// buatModul menulis berkas modul dan mengembalikan direktori temptannya.
func buatModul(t *testing.T, nama string, isi string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, nama)
	if err := os.WriteFile(path, []byte(isi), 0o644); err != nil {
		t.Fatalf("tidak bisa menulis modul: %v", err)
	}
	return dir
}

func TestImporModulSederhana(t *testing.T) {
	dir := buatModul(t, "pembantu.nusa", `
tetap VERSI = "1.0"

fungsi sapa(nama) {
    kembali "Halo, " + nama
}
`)

	var out bytes.Buffer
	loader := modul.New(dir, &out, strings.NewReader(""))

	simbol, err := loader.Import("pembantu.nusa", 1)
	if err != nil {
		t.Fatalf("impor gagal: %v", err)
	}
	if len(simbol) != 2 {
		t.Fatalf("jumlah simbol = %d, ingin 2: %v", len(simbol), simbol)
	}

	nama := map[string]bool{}
	for _, s := range simbol {
		nama[s.Name] = true
	}
	if !nama["sapa"] || !nama["VERSI"] {
		t.Errorf("simbol yang diekspor kurang: %v", nama)
	}
}

func TestImporTanpaEkstensi(t *testing.T) {
	dir := buatModul(t, "pembantu.nusa", "fungsi sapa() { kembali 1 }")

	loader := modul.New(dir, &bytes.Buffer{}, strings.NewReader(""))

	simbol, err := loader.Import("pembantu", 1)
	if err != nil {
		t.Fatalf("impor gagal: %v", err)
	}
	if len(simbol) != 1 {
		t.Fatalf("jumlah simbol = %d, ingin 1", len(simbol))
	}
}

func TestImporMemakaiCache(t *testing.T) {
	dir := buatModul(t, "pembantu.nusa", "fungsi sapa() { kembali 1 }")

	loader := modul.New(dir, &bytes.Buffer{}, strings.NewReader(""))

	pertama, err := loader.Import("pembantu.nusa", 1)
	if err != nil {
		t.Fatalf("impor pertama gagal: %v", err)
	}
	kedua, err := loader.Import("pembantu.nusa", 2)
	if err != nil {
		t.Fatalf("impor kedua gagal: %v", err)
	}

	if len(pertama) != len(kedua) {
		t.Fatalf("hasil impor berbeda: %d vs %d", len(pertama), len(kedua))
	}
}

func TestImporMelingkarDitolak(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.nusa"), []byte("impor \"b.nusa\"\nfungsi fa() { kembali 1 }"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.nusa"), []byte("impor \"a.nusa\"\nfungsi fb() { kembali 2 }"), 0o644); err != nil {
		t.Fatal(err)
	}

	loader := modul.New(dir, &bytes.Buffer{}, strings.NewReader(""))

	_, err := loader.Import("a.nusa", 1)
	if err == nil {
		t.Fatal("impor melingkar seharusnya ditolak")
	}
	if !strings.Contains(err.Error(), "melingkar") {
		t.Errorf("pesan galat = %q, ingin mengandung 'melingkar'", err.Error())
	}
}

func TestImporModulTidakAda(t *testing.T) {
	loader := modul.New(t.TempDir(), &bytes.Buffer{}, strings.NewReader(""))

	_, err := loader.Import("hilang.nusa", 1)
	if err == nil {
		t.Fatal("impor modul yang tidak ada seharusnya gagal")
	}
	if !strings.Contains(err.Error(), "tidak ditemukan") {
		t.Errorf("pesan galat = %q, ingin mengandung 'tidak ditemukan'", err.Error())
	}
}

func TestImporModulRusak(t *testing.T) {
	dir := buatModul(t, "rusak.nusa", "tulis )")

	loader := modul.New(dir, &bytes.Buffer{}, strings.NewReader(""))

	_, err := loader.Import("rusak.nusa", 1)
	if err == nil {
		t.Fatal("modul yang gagal di-parse seharusnya ditolak")
	}
}

func TestNamaModul(t *testing.T) {
	if got := modul.NamaModul("/tmp/contoh/pembantu.nusa"); got != "pembantu" {
		t.Errorf("NamaModul = %q, ingin %q", got, "pembantu")
	}
}
