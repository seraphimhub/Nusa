package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCekProgramBersih(t *testing.T) {
	path := writeProgram(t, "bersih.nusa", "buat x = 1\ntulis x\n")

	code, output := captureRun(t, "cek", path)
	if code != 0 {
		t.Fatalf("exit code = %d, ingin 0; output: %s", code, output)
	}
	if !strings.Contains(output, "tidak ada temuan") {
		t.Errorf("output = %q, ingin mengandung 'tidak ada temuan'", output)
	}
}

func TestCekVariabelTakTerdefinisi(t *testing.T) {
	path := writeProgram(t, "salah.nusa", "tulis tidak_ada\n")

	code, output := captureRun(t, "cek", path)
	if code != 1 {
		t.Fatalf("exit code = %d, ingin 1", code)
	}
	if !strings.Contains(output, "tidak ditemukan") {
		t.Errorf("output = %q, ingin mengandung 'tidak ditemukan'", output)
	}
	if !strings.Contains(output, "^") {
		t.Errorf("output = %q, ingin mengandung penanda caret", output)
	}
}

func TestCekGalatParsing(t *testing.T) {
	path := writeProgram(t, "salah.nusa", "buat = 5\n")

	code, output := captureRun(t, "cek", path)
	if code != 1 {
		t.Fatalf("exit code = %d, ingin 1", code)
	}
	if !strings.Contains(output, "Error parsing") {
		t.Errorf("output = %q, ingin mengandung Error parsing", output)
	}
}

func TestFmtKeStdout(t *testing.T) {
	path := writeProgram(t, "acak.nusa", "buat x=1\nfungsi f(){\nkembali x\n}\n")

	code, output := captureRun(t, "fmt", path)
	if code != 0 {
		t.Fatalf("exit code = %d, ingin 0", code)
	}
	ingin := "buat x = 1\nfungsi f() {\n  kembali x\n}\n"
	if output != ingin {
		t.Errorf("output = %q, ingin %q", output, ingin)
	}
}

func TestFmtTulisBerkas(t *testing.T) {
	path := writeProgram(t, "acak.nusa", "buat x=1\n")

	code, output := captureRun(t, "fmt", "-w", path)
	if code != 0 {
		t.Fatalf("exit code = %d, ingin 0; output: %s", code, output)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "buat x = 1\n" {
		t.Errorf("isi berkas = %q, ingin %q", string(data), "buat x = 1\n")
	}
}

func TestFmtModeCek(t *testing.T) {
	path := writeProgram(t, "acak.nusa", "buat x=1\n")

	code, output := captureRun(t, "fmt", "--cek", path)
	if code != 1 {
		t.Fatalf("exit code = %d, ingin 1; output: %s", code, output)
	}
	if !strings.Contains(output, "belum rapi") {
		t.Errorf("output = %q, ingin mengandung 'belum rapi'", output)
	}

	// Setelah dirapikan, mode cek harus berhasil.
	if code, output := captureRun(t, "fmt", "-w", path); code != 0 {
		t.Fatalf("penulisan gagal: %d %s", code, output)
	}
	if code, output := captureRun(t, "fmt", "--cek", path); code != 0 {
		t.Fatalf("exit code = %d, ingin 0; output: %s", code, output)
	}
}

func TestFmtMempertahankanKomentar(t *testing.T) {
	path := writeProgram(t, "komentar.nusa", "# catatan\ntulis 1 # catatan kecil\n")

	code, output := captureRun(t, "fmt", path)
	if code != 0 {
		t.Fatalf("exit code = %d, ingin 0", code)
	}
	if !strings.Contains(output, "# catatan") || !strings.Contains(output, "# catatan kecil") {
		t.Errorf("komentar hilang: %q", output)
	}
}

func TestUjiSemuaLulus(t *testing.T) {
	path := writeProgram(t, "matematika_uji.nusa", `
fungsi tambah(a, b) {
    kembali a + b
}

uji "penjumlahan" {
    cek tambah(1, 2) == 3
}
`)

	code, output := captureRun(t, "uji", path)
	if code != 0 {
		t.Fatalf("exit code = %d, ingin 0; output: %s", code, output)
	}
	if !strings.Contains(output, "LULUS") {
		t.Errorf("output = %q, ingin mengandung LULUS", output)
	}
}

func TestUjiAdaKegagalan(t *testing.T) {
	path := writeProgram(t, "salah_uji.nusa", `
uji "salah" {
    cek 1 == 2, "satu bukan dua"
}
`)

	code, output := captureRun(t, "uji", path)
	if code != 1 {
		t.Fatalf("exit code = %d, ingin 1", code)
	}
	if !strings.Contains(output, "GAGAL") {
		t.Errorf("output = %q, ingin mengandung GAGAL", output)
	}
	if !strings.Contains(output, "satu bukan dua") {
		t.Errorf("output = %q, ingin mengandung pesan kegagalan", output)
	}
}

func TestUjiMenemukanBerkasOtomatis(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "otomatis_uji.nusa")
	if err := os.WriteFile(path, []byte("uji \"otomatis\" {\n    cek 1 == 1\n}\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)

	code, output := captureRun(t, "uji")
	if code != 0 {
		t.Fatalf("exit code = %d, ingin 0; output: %s", code, output)
	}
	if !strings.Contains(output, "LULUS") {
		t.Errorf("output = %q, ingin mengandung LULUS", output)
	}
}

func TestImporModulLewatCLI(t *testing.T) {
	dir := t.TempDir()
	modulPath := filepath.Join(dir, "pembantu.nusa")
	if err := os.WriteFile(modulPath, []byte("fungsi sapa(nama) {\n  kembali \"Halo \" + nama\n}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	utamaPath := filepath.Join(dir, "utama.nusa")
	if err := os.WriteFile(utamaPath, []byte("impor \"pembantu.nusa\"\ntulis sapa(\"Dunia\")\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	code, output := captureRun(t, "run", utamaPath)
	if code != 0 {
		t.Fatalf("exit code = %d, ingin 0; output: %s", code, output)
	}
	if !strings.Contains(output, "Halo Dunia") {
		t.Errorf("output = %q, ingin mengandung 'Halo Dunia'", output)
	}
}
