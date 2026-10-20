// Command hostweb adalah contoh host HTTP untuk bahasa Nusa.
//
// Pola ini dipakai saat Nusa belum punya operasi jaringan sendiri:
// Go melayani HTTP, sedangkan berkas .nusa menulis logika dan markup.
//
// Jalankan dengan:
//
//	go run ./cmd/hostweb
//
// lalu buka http://localhost:8080/ di peramban.
//
// Program ini juga membuktikan bahwa Nusa bisa dipakai dari program Go
// di luar repositori ini, lewat paket publik "github.com/seraphimhub/Nusa".
package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/seraphimhub/Nusa"
)

// port adalah porta server yang dipakai contoh ini.
const port = 8080

// template memuat template Nusa yang sudah di-parse. Program di-parse satu
// kali di awal lalu dieksekusi ulang untuk setiap request, karena AST Nusa
// tidak dimutasi saat dievaluasi.
type template struct {
	program   *nusa.Program
	env       *nusa.Environment
}

// muatTemplate membaca dan mengurai berkas template Nusa.
func muatTemplate(path string) (*template, error) {
	program, err := nusa.CompileFile(path)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat template %s: %w", path, err)
	}
	return &template{program: program, env: nusa.NewEnvironment()}, nil
}

// render menjalankan template dengan nilai yang diberikan.
//
// Setiap request memakai runtime sendiri sehingga variabel per-request tidak
// bocur ke request lain.
func (t *template) render(nama string) (string, error) {
	var keluar strings.Builder

	rt := nusa.NewRuntime(&keluar, nil)

	// Setel environment dengan variabel global 'nama'
	env := nusa.NewEnvironment()
	env.store["nama"] = nusa.Text(nama)
	rt.env = env

	if err := rt.Execute(t.program); err != nil {
		return keluar.String(), err
	}

	return keluar.String(), nil
}

func main() {
	halaman, err := muatTemplate("examples/web/halaman.nusa")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	api, err := muatTemplate("examples/web/api.nusa")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", handlerHalaman(halaman))
	mux.HandleFunc("/api/sapa", handlerAPI(api))

	alamat := fmt.Sprintf("http://localhost:%d/", port)
	fmt.Printf("Nusa web host berjalan di %s\n", alamat)
	fmt.Println("  /          halaman HTML (template examples/web/halaman.nusa)")
	fmt.Println("  /api/sapa  respons JSON (template examples/web/api.nusa)")
	fmt.Println("Tekan Ctrl+C untuk berhenti.")

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err != nil {
		fmt.Fprintln(os.Stderr, "server gagal:", err)
		os.Exit(1)
	}
}

// handlerHalaman melayani route HTML.
func handlerHalaman(tpl *template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nama := namaDari(r)

		keluaran, err := tpl.render(nama)
		if err != nil {
			http.Error(w, "galat: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, keluaran)
	}
}

// handlerAPI melayani route JSON.
func handlerAPI(tpl *template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nama := namaDari(r)

		keluaran, err := tpl.render(nama)
		if err != nil {
			http.Error(w, "galat: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprint(w, keluaran)
	}
}

// namaDari membaca parameter "nama" dari query string dengan nilai
// bawaan bila kosong.
func namaDari(r *http.Request) string {
	nama := strings.TrimSpace(r.URL.Query().Get("nama"))
	if nama == "" {
		return "Dunia"
	}
	return nama
}
