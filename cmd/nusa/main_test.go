package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeProgram(t *testing.T, name string, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("gagal menulis file uji: %v", err)
	}
	return path
}

// captureRun menjalankan CLI sambil menangkap stdout dan stderr.
func captureRun(t *testing.T, args ...string) (int, string) {
	t.Helper()

	oldOut, oldErr := os.Stdout, os.Stderr
	outR, outW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	errR, errW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	os.Stdout, os.Stderr = outW, errW
	code := run(args)
	outW.Close()
	errW.Close()
	os.Stdout, os.Stderr = oldOut, oldErr

	out, _ := io.ReadAll(outR)
	errOut, _ := io.ReadAll(errR)
	return code, string(out) + string(errOut)
}

func TestRunValidProgram(t *testing.T) {
	path := writeProgram(t, "halo.nusa", "buat x = 5\ntulis x\n")

	code, output := captureRun(t, "run", path)
	if code != 0 {
		t.Fatalf("exit code = %d, ingin 0", code)
	}
	if output != "5\n" {
		t.Errorf("output = %q, ingin %q", output, "5\n")
	}
}

func TestRunSyntaxError(t *testing.T) {
	path := writeProgram(t, "salah.nusa", "buat = 5\n")

	code, output := captureRun(t, "run", path)
	if code != 1 {
		t.Fatalf("exit code = %d, ingin 1", code)
	}
	if !strings.Contains(output, "Error parsing") {
		t.Errorf("output = %q, ingin mengandung Error parsing", output)
	}
}

func TestRunRuntimeError(t *testing.T) {
	path := writeProgram(t, "galat.nusa", "tulis 1 / 0\n")

	code, output := captureRun(t, "run", path)
	if code != 1 {
		t.Fatalf("exit code = %d, ingin 1", code)
	}
	if !strings.Contains(output, "pembagian dengan nol") {
		t.Errorf("output = %q, ingin mengandung pesan pembagian nol", output)
	}
}

func TestRunLexicalError(t *testing.T) {
	path := writeProgram(t, "ilegal.nusa", "tulis @\n")

	code, output := captureRun(t, "run", path)
	if code != 1 {
		t.Fatalf("exit code = %d, ingin 1", code)
	}
	if !strings.Contains(output, "karakter tidak dikenal") {
		t.Errorf("output = %q, ingin mengandung karakter tidak dikenal", output)
	}
}

func TestRunMissingFile(t *testing.T) {
	code, output := captureRun(t, "run", filepath.Join(t.TempDir(), "tidak-ada.nusa"))
	if code != 1 {
		t.Fatalf("exit code = %d, ingin 1", code)
	}
	if !strings.Contains(output, "tidak bisa membaca file") {
		t.Errorf("output = %q", output)
	}
}

func TestVersionAndHelp(t *testing.T) {
	code, output := captureRun(t, "version")
	if code != 0 || !strings.Contains(output, "Nusa v") {
		t.Errorf("version: code=%d output=%q", code, output)
	}

	code, output = captureRun(t, "help")
	if code != 0 || !strings.Contains(output, "PENGGUNAAN") {
		t.Errorf("help: code=%d output=%q", code, output)
	}
}

func TestASTAndTokenCommands(t *testing.T) {
	path := writeProgram(t, "contoh.nusa", "buat x = 1\n")

	code, output := captureRun(t, "ast", path)
	if code != 0 || !strings.Contains(output, "buat x = 1") {
		t.Errorf("ast: code=%d output=%q", code, output)
	}

	code, output = captureRun(t, "token", path)
	if code != 0 || !strings.Contains(output, "BUAT") {
		t.Errorf("token: code=%d output=%q", code, output)
	}
}

func TestUnknownCommand(t *testing.T) {
	code, output := captureRun(t, "terbang")
	if code != 1 || !strings.Contains(output, "command tidak dikenal") {
		t.Errorf("code=%d output=%q", code, output)
	}
}

func TestNeedsMoreInput(t *testing.T) {
	tests := []struct {
		source string
		want   bool
	}{
		{"tulis 1", false},
		{"jika benar {", true},
		{"jika benar {\n    tulis 1\n}", false},
		{"buat x = [1, 2,", true},
		{"buat x = [1, 2]", false},
		{`buat s = "abc`, true},
		{`buat s = "abc"`, false},
		{`buat s = "a\""`, false},
		{`tulis "{"`, false},
		{"/* belum selesai", true},
		{"/* selesai */ tulis 1", false},
		{"# komentar {", false},
		{"// komentar [", false},
	}

	for _, test := range tests {
		if got := needsMoreInput(test.source); got != test.want {
			t.Errorf("needsMoreInput(%q) = %v, ingin %v", test.source, got, test.want)
		}
	}
}
