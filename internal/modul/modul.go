// Package modul memuat modul Nusa dari berkas lain.
//
// Modul adalah berkas .nusa yang dieksekusi di lingkup globalnya sendiri.
// Semua fungsi dan konstanta yang dideklarasikan di tingkat atas modul
// diekspor, lalu dimasukkan ke lingkup berkas pemanggil lewat pernyataan
// `impor "modul.nusa"`.
package modul

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/seraphimhub/Nusa/internal/ast"
	"github.com/seraphimhub/Nusa/internal/lexer"
	"github.com/seraphimhub/Nusa/internal/parser"
	"github.com/seraphimhub/Nusa/internal/runtime"
	"github.com/seraphimhub/Nusa/internal/token"
)

// Loader memuat modul Nusa. Beberapa loader dapat berbagi keadaan melalui
// NewWithState sehingga modul yang sama hanya dimuat satu kali.
type Loader struct {
	// Basis adalah direktori tempat path relatif diselesaikan.
	Basis string
	// Out dan In dipakai modul saat dieksekusi.
	Out io.Writer
	In  io.Reader

	state *state
}

// state adalah keadaan bersama antar loader: cache modul yang sudah dimuat
// dan daftar modul yang sedang dimuat untuk deteksi impor melingkar.
type state struct {
	cache  map[string][]runtime.Symbol
	dimuat map[string]bool
}

// New membuat loader baru dengan keadaan sendiri.
func New(basis string, out io.Writer, in io.Reader) *Loader {
	return &Loader{
		Basis: basis,
		Out:   out,
		In:    in,
		state: &state{
			cache:  make(map[string][]runtime.Symbol),
			dimuat: make(map[string]bool),
		},
	}
}

// NewWithState membuat loader yang berbagi keadaan dengan loader lain.
func NewWithState(basis string, out io.Writer, in io.Reader, shared *Loader) *Loader {
	ldr := New(basis, out, in)
	if shared != nil {
		ldr.state = shared.state
	}
	return ldr
}

// Import memuat modul pada path dan mengembalikan simbol yang diekspor.
// Path relatif diselesaikan terhadap Basis loader ini.
func (l *Loader) Import(path string, line int) ([]runtime.Symbol, error) {
	berkas, err := l.resolve(path)
	if err != nil {
		return nil, err
	}

	kunci, err := filepath.Abs(berkas)
	if err != nil {
		kunci = berkas
	}

	if simbol, ok := l.state.cache[kunci]; ok {
		return simbol, nil
	}
	if l.state.dimuat[kunci] {
		return nil, fmt.Errorf("impor modul melingkar pada %s", berkas)
	}
	l.state.dimuat[kunci] = true
	defer delete(l.state.dimuat, kunci)

	simbol, err := l.muat(kunci, berkas)
	if err != nil {
		return nil, err
	}

	l.state.cache[kunci] = simbol
	return simbol, nil
}

// Ekspor mengembalikan nama yang diekspor modul pada path tanpa menjalankan
// modul. Fungsi ini dipakai pemeriksa statis sehingga nama dari impor bisa
// dikenali tanpa efek samping.
func (l *Loader) Ekspor(path string) ([]string, error) {
	berkas, err := l.resolve(path)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(berkas)
	if err != nil {
		return nil, fmt.Errorf("modul '%s' tidak bisa dibaca: %v", berkas, err)
	}

	tokens := lexer.Tokenize(string(data))
	p := parser.New(tokens)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		return nil, fmt.Errorf("galat parsing pada %s: %s", berkas, p.Errors()[0])
	}

	var nama []string
	for _, stmt := range program.Statements {
		switch s := stmt.(type) {
		case *ast.FunctionStatement:
			nama = append(nama, s.Name.Value)
		case *ast.LetStatement:
			// Hanya konstanta yang diekspor.
			if s.Konst {
				nama = append(nama, s.Name.Value)
			}
		}
	}

	return nama, nil
}

// resolve menentukan lokasi berkas modul. Path absolut dipakai apa adanya,
// sedangkan path relatif diselesaikan terhadap Basis.
func (l *Loader) resolve(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("nama berkas modul kosong")
	}

	kandidat := path
	if !filepath.IsAbs(kandidat) {
		kandidat = filepath.Join(l.Basis, kandidat)
	}

	// Ekstensi .nusa boleh dikesampingkan demi kenyamanan.
	if filepath.Ext(kandidat) == "" {
		kandidat += ".nusa"
	}

	if _, err := os.Stat(kandidat); err != nil {
		return "", fmt.Errorf("modul '%s' tidak ditemukan", path)
	}
	return kandidat, nil
}

// muat membaca, mem-parse, dan mengeksekusi modul lalu mengumpulkan
// simbol yang diekspornya.
func (l *Loader) muat(kunci string, berkas string) ([]runtime.Symbol, error) {
	data, err := os.ReadFile(berkas)
	if err != nil {
		return nil, fmt.Errorf("modul '%s' tidak bisa dibaca: %v", berkas, err)
	}
	sumber := string(data)

	tokens := lexer.Tokenize(sumber)
	for _, tok := range tokens {
		if tok.Type == token.ILLEGAL {
			return nil, fmt.Errorf("galat leksikal pada %s baris %d: karakter tidak dikenal %q",
				berkas, tok.Line, tok.Literal)
		}
	}

	p := parser.New(tokens)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		return nil, fmt.Errorf("galat parsing pada %s: %s", berkas, p.Errors()[0])
	}

	// Modul dijalankan di lingkup globalnya sendiri agar tidak berbagi
	// variabel dengan berkas pemanggil.
	anak := NewWithState(filepath.Dir(berkas), l.Out, l.In, l)
	rt := runtime.NewWithIO(l.Out, l.In)
	rt.SetModuleLoader(anak.Import)

	hasil := rt.Execute(program)
	if hasil != nil && hasil.Type() == runtime.ERROR_OBJ {
		return nil, fmt.Errorf("modul '%s' gagal dijalankan: %s", berkas, hasil.Inspect())
	}

	return ekspor(program, rt), nil
}

// ekspor mengumpulkan fungsi dan konstanta tingkat atas modul beserta
// nilainya. Urutan deklarasi dijaga agar hasil impor selalu sama.
func ekspor(program *ast.Program, rt *runtime.Runtime) []runtime.Symbol {
	var simbol []runtime.Symbol

	for _, stmt := range program.Statements {
		switch s := stmt.(type) {
		case *ast.FunctionStatement:
			if nilai, ok := rt.Global(s.Name.Value); ok {
				simbol = append(simbol, runtime.Symbol{Name: s.Name.Value, Value: nilai})
			}
		case *ast.LetStatement:
			// Hanya konstanta yang diekspor; variabel biasa tetap
			// menjadi milik modul.
			if !s.Konst {
				continue
			}
			if nilai, ok := rt.Global(s.Name.Value); ok {
				simbol = append(simbol, runtime.Symbol{Name: s.Name.Value, Value: nilai})
			}
		}
	}

	return simbol
}

// NamaModul mengembalikan nama berkas modul tanpa direktori, dipakai untuk
// pesan galat yang lebih ramah.
func NamaModul(path string) string {
	base := filepath.Base(path)
	if strings.HasSuffix(base, ".nusa") {
		return strings.TrimSuffix(base, ".nusa")
	}
	return base
}
