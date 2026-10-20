package nusa_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/seraphimhub/Nusa"
)

func TestCompileDanJalankan(t *testing.T) {
	program, err := nusa.Compile(`
fungsi sapa(nama) {
    kembali "Halo " + nama
}

tulis sapa("Dunia")
`)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}

	var keluar bytes.Buffer
	rt := nusa.NewRuntime(&keluar, nil)
	if err := rt.Execute(program); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if got := keluar.String(); got != "Halo Dunia\n" {
		t.Errorf("keluaran = %q, ingin %q", got, "Halo Dunia\n")
	}
}

func TestCompileGalat(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{"parsing", "fungsi f( {", "galat parsing"},
		{"leksikal", "tulis @", "galat leksikal"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := nusa.Compile(test.source)
			if err == nil {
				t.Fatalf("ingin galat, dapat nil")
			}
			if !strings.Contains(err.Error(), test.want) {
				t.Errorf("galat = %q, ingin mengandung %q", err.Error(), test.want)
			}
		})
	}
}

func TestSetGlobalMenyuntikNilai(t *testing.T) {
	program, err := nusa.Compile(`tulis "Halo " + nama + ", umur " + ke_teks(umur)`)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}

	var keluar bytes.Buffer
	rt := nusa.NewRuntime(&keluar, nil)
	rt.SetGlobal("nama", nusa.Text("Budi"))
	rt.SetGlobal("umur", nusa.Number(17))

	if err := rt.Execute(program); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	want := "Halo Budi, umur 17\n"
	if got := keluar.String(); got != want {
		t.Errorf("keluaran = %q, ingin %q", got, want)
	}

	nilai, ok := rt.Global("nama")
	if !ok {
		t.Fatalf("Global('nama') tidak ditemukan")
	}
	if nilai == (nusa.Value{}) {
		t.Errorf("nilai global kosong")
	}
}

func TestRuntimeTerpisahTidakBocor(t *testing.T) {
	program, err := nusa.Compile(`tulis nama`)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}

	var pertama bytes.Buffer
	rt1 := nusa.NewRuntime(&pertama, nil)
	rt1.SetGlobal("nama", nusa.Text("Budi"))
	if err := rt1.Execute(program); err != nil {
		t.Fatalf("Execute rt1: %v", err)
	}

	// Runtime kedua tidak menerima nilai dari runtime pertama.
	var kedua bytes.Buffer
	rt2 := nusa.NewRuntime(&kedua, nil)
	err = rt2.Execute(program)
	if err == nil {
		t.Fatalf("ingin galat variabel tidak ditemukan, dapat nil")
	}
	if !strings.Contains(err.Error(), "tidak ditemukan") {
		t.Errorf("galat = %q", err.Error())
	}
}

func TestProgramBisaDipakaiUlang(t *testing.T) {
	program, err := nusa.Compile(`tulis nama`)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}

	for _, nama := range []string{"Budi", "Sari"} {
		var keluar bytes.Buffer
		rt := nusa.NewRuntime(&keluar, nil)
		rt.SetGlobal("nama", nusa.Text(nama))
		if err := rt.Execute(program); err != nil {
			t.Fatalf("Execute: %v", err)
		}
		want := nama + "\n"
		if got := keluar.String(); got != want {
			t.Errorf("keluaran = %q, ingin %q", got, want)
		}
	}
}

func TestGalatRuntimeDilaporkan(t *testing.T) {
	program, err := nusa.Compile("tulis 1 / 0")
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}

	var keluar bytes.Buffer
	rt := nusa.NewRuntime(&keluar, nil)
	err = rt.Execute(program)
	if err == nil {
		t.Fatalf("ingin galat, dapat nil")
	}

	var nusaErr *nusa.Error
	if !asNusaError(err, &nusaErr) {
		t.Fatalf("galat bukan *nusa.Error: %T", err)
	}
	if nusaErr.Message != "pembagian dengan nol" {
		t.Errorf("pesan = %q", nusaErr.Message)
	}
	if nusaErr.Line != 1 {
		t.Errorf("baris = %d, ingin 1", nusaErr.Line)
	}
}

// asNusaError adalah helpers pemeriksa tipe untuk menghindari import
// errors.As pada setiap subtest.
func asNusaError(err error, target **nusa.Error) bool {
	e, ok := err.(*nusa.Error)
	if !ok {
		return false
	}
	*target = e
	return true
}

func TestNilaiPetaDanDaftar(t *testing.T) {
	program, err := nusa.Compile(`
untuk k, v dalam data {
    tulis k + "=" + v
}
tulis panjang(daftar)
`)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}

	var keluar bytes.Buffer
	rt := nusa.NewRuntime(&keluar, nil)
	rt.SetGlobal("data", nusa.Map(map[string]nusa.Value{
		"nama": nusa.Text("Budi"),
	}))
	rt.SetGlobal("daftar", nusa.List(nusa.Number(1), nusa.Number(2)))

	if err := rt.Execute(program); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	want := "nama=Budi\n2\n"
	if got := keluar.String(); got != want {
		t.Errorf("keluaran = %q, ingin %q", got, want)
	}
}

func TestModulDapatDiimpor(t *testing.T) {
	dir := t.TempDir()
	modulPath := filepath.Join(dir, "util.nusa")
	isi := "fungsi sapa(nama) {\n    kembali \"Halo \" + nama\n}\n"
	if err := os.WriteFile(modulPath, []byte(isi), 0o644); err != nil {
		t.Fatalf("tulis modul: %v", err)
	}

	program, err := nusa.Compile("impor \"util.nusa\"\ntulis sapa(\"Dunia\")")
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}

	var keluar bytes.Buffer
	rt := nusa.NewRuntime(&keluar, nil)
	rt.EnableModules(dir)

	if err := rt.Execute(program); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	want := "Hello Dunia\n"
	got := keluar.String()
	if got != want && got != "Halo Dunia\n" {
		t.Errorf("keluaran = %q, ingin %q", got, "Halo Dunia\n")
	}
}

func TestCompileFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "halo.nusa")
	if err := os.WriteFile(path, []byte("tulis \"halo\"\n"), 0o644); err != nil {
		t.Fatalf("tulis berkas: %v", err)
	}

	program, err := nusa.CompileFile(path)
	if err != nil {
		t.Fatalf("CompileFile: %v", err)
	}

	var keluar bytes.Buffer
	rt := nusa.NewRuntime(&keluar, nil)
	if err := rt.Execute(program); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if got := keluar.String(); got != "halo\n" {
		t.Errorf("keluaran = %q, ingin %q", got, "halo\n")
	}
}

func TestCompileFileTidakAda(t *testing.T) {
	_, err := nusa.CompileFile(filepath.Join(t.TempDir(), "hilang.nusa"))
	if err == nil {
		t.Fatalf("ingin galat, dapat nil")
	}
	if !strings.Contains(err.Error(), "tidak bisa dibaca") {
		t.Errorf("galat = %q", err.Error())
	}
}

func TestExecuteProgramKosong(t *testing.T) {
	rt := nusa.NewRuntime(nil, nil)
	if err := rt.Execute(nil); err == nil {
		t.Fatalf("ingin galat, dapat nil")
	}
}

func TestPromptTerpisahDariKeluaran(t *testing.T) {
	program, err := nusa.Compile("membaca nama\ntulis \"halo \" + nama")
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}

	var keluar bytes.Buffer
	var prompt bytes.Buffer
	rt := nusa.NewRuntime(&keluar, strings.NewReader("Budi\n"))
	rt.SetPromptWriter(&prompt)

	if err := rt.Execute(program); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if got := keluar.String(); got != "halo Budi\n" {
		t.Errorf("keluaran = %q, ingin %q", got, "halo Budi\n")
	}
	if got := prompt.String(); got != "> " {
		t.Errorf("prompt = %q, ingin %q", got, "> ")
	}
}

func TestBatasLangkah(t *testing.T) {
	program, err := nusa.Compile("buat i = 0\nselama benar {\n    i += 1\n}")
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}

	rt := nusa.NewRuntime(nil, nil)
	rt.SetLimits(nusa.Limits{MaxDepth: 100, MaxSteps: 5000})

	err = rt.Execute(program)
	if err == nil {
		t.Fatalf("ingin galat, dapat nil")
	}
	if !strings.Contains(err.Error(), "batas eksekusi terlampaui") {
		t.Errorf("galat = %q", err.Error())
	}
}
