package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/seraphimhub/Nusa/internal/lexer"
	"github.com/seraphimhub/Nusa/internal/parser"
	"github.com/seraphimhub/Nusa/internal/runtime"
)

// runUji menjalankan penjalur uji. Tanpa argumen, seluruh berkas uji pada
// direktori sekarang dijalankan.
func runUji(args []string) int {
	if len(args) == 0 {
		berkas, err := kumpulkanBerkasUji(".")
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return 1
		}
		if len(berkas) == 0 {
			fmt.Fprintln(os.Stderr, "error: tidak ada berkas uji yang ditemukan")
			fmt.Fprintln(os.Stderr, "       cari berkas berakhiran _uji.nusa atau _test.nusa")
			return 1
		}
		return ujiFiles(berkas)
	}

	return ujiFiles(args)
}

// ujiFiles menjalankan seluruh berkas uji yang diberikan.
func ujiFiles(filenames []string) int {
	totalLulus := 0
	totalGagal := 0
	adaGalatBerkas := false

	for _, filename := range filenames {
		source, ok := readSource(filename)
		if !ok {
			return 1
		}

		tokens := lexer.Tokenize(source)
		if reportLexErrors(tokens, source) {
			return 1
		}

		p := parser.New(tokens)
		program := p.ParseProgram()
		if len(p.Errors()) > 0 {
			fmt.Fprintln(os.Stderr, "Error parsing:")
			for _, err := range p.Errors() {
				fmt.Fprint(os.Stderr, "  "+formatGalatTeks(err, source))
			}
			return 1
		}

		rt := newRuntime(filename)
		result := rt.Execute(program)
		if result != nil && result.Type() == runtime.ERROR_OBJ {
			// Galat di luar blok uji menghentikan berkas ini, tetapi
			// berkas lain tetap dijalankan.
			fmt.Fprint(os.Stderr, formatGalat(result.(*runtime.Error), source))
			adaGalatBerkas = true
		}

		lulus, gagal := tampilkanHasilUji(rt, filename, source)
		totalLulus += lulus
		totalGagal += gagal
	}

	fmt.Println()
	if totalGagal == 0 && !adaGalatBerkas {
		fmt.Printf("%d pemeriksaan lulus\n", totalLulus)
		return 0
	}

	fmt.Printf("%d pemeriksaan lulus, %d gagal\n", totalLulus, totalGagal)
	return 1
}

// tampilkanHasilUji mencetak ringkasan hasil uji sebuah berkas dan
// mengembalikan jumlah pemeriksaan yang lulus serta gagal.
func tampilkanHasilUji(rt *runtime.Runtime, filename string, source string) (int, int) {
	hasil := rt.HasilCek()
	if len(hasil) == 0 {
		fmt.Printf("%s: tidak ada pemeriksaan\n", filename)
		return 0, 0
	}

	// Kelompokkan hasil berdasarkan nama blok uji sehingga laporan tetap
	// rapi walaupun pemeriksaan tidak berurutan.
	urut := []string{}
	perUji := map[string][]runtime.HasilCek{}
	for _, h := range hasil {
		if _, ada := perUji[h.Uji]; !ada {
			urut = append(urut, h.Uji)
		}
		perUji[h.Uji] = append(perUji[h.Uji], h)
	}

	lulus := 0
	gagal := 0

	for _, nama := range urut {
		daftar := perUji[nama]
		jumlahLulus := 0
		jumlahGagal := 0
		for _, h := range daftar {
			if h.Lolos {
				jumlahLulus++
			} else {
				jumlahGagal++
			}
		}

		lulus += jumlahLulus
		gagal += jumlahGagal

		status := "LULUS"
		if jumlahGagal > 0 {
			status = "GAGAL"
		}
		fmt.Printf("%-6s %s (%d cek)\n", status, namaUji(nama), len(daftar))

		for _, h := range daftar {
			if h.Lolos {
				continue
			}
			fmt.Fprintf(os.Stderr, "       baris %d: %s\n", h.Line, h.Pesan)
			if cuplikan := cuplikanSumber(source, h.Line, h.Column); cuplikan != "" {
				fmt.Fprint(os.Stderr, cuplikan)
			}
		}
	}

	return lulus, gagal
}

// namaUji memberi nama yang enak dibaca untuk blok uji tanpa nama.
func namaUji(nama string) string {
	if nama == "" {
		return "(tanpa nama)"
	}
	return fmt.Sprintf("%q", nama)
}

// kumpulkanBerkasUji mencari berkas uji pada direktori. Pola nama yang
// dikenali adalah sufiks _uji.nusa dan _test.nusa.
func kumpulkanBerkasUji(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("tidak bisa membaca direktori '%s': %v", dir, err)
	}

	var keluar []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".nusa") {
			continue
		}
		if strings.HasSuffix(name, "_uji.nusa") || strings.HasSuffix(name, "_test.nusa") {
			keluar = append(keluar, filepath.Join(dir, name))
		}
	}

	sort.Strings(keluar)
	return keluar, nil
}
