// Package nusa adalah antarmuka publik untuk menjalankan kode bahasa Nusa
// dari program Go.
//
// Paket ini sengaja berada di luar `internal/` supaya aplikasi Go di luar
// repositori ini bisa memakai Nusa sebagai bahasa render, template, atau
// logika bisnis. Seluruh detail internal (lexer, parser, AST, evaluator)
// disembunyikan di balik API yang kecil dan stabil:
//
//	program, err := nusa.CompileFile("halaman.nusa")
//	if err != nil {
//	    return err
//	}
//
//	var keluar strings.Builder
//	rt := nusa.NewRuntime(&keluar, nil)
//	rt.SetGlobal("nama", nusa.Text("Budi"))
//	if err := rt.Execute(program); err != nil {
//	    return err
//	}
//	fmt.Print(keluar.String())
//
// Program hasil Compile aman dipakai ulang untuk banyak eksekusi karena
// evaluator tidak mengubah node AST. Setiap Runtime punya lingkup variabel
// sendiri, sehingga dua runtime yang berbeda tidak saling memengaruhi.
package nusa

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/seraphimhub/Nusa/internal/ast"
	"github.com/seraphimhub/Nusa/internal/lexer"
	"github.com/seraphimhub/Nusa/internal/modul"
	"github.com/seraphimhub/Nusa/internal/parser"
	"github.com/seraphimhub/Nusa/internal/runtime"
	"github.com/seraphimhub/Nusa/internal/token"
)

// Program adalah program Nusa yang sudah di-parse dan siap dieksekusi.
// Program aman dipakai ulang untuk banyak pemanggilan Execute.
type Program struct {
	ast *ast.Program
}

// Compile mengurai sumber kode Nusa menjadi Program.
//
// Galat leksikal dan galat parsing dikembalikan sebagai error dengan
// nomor baris dan kolom bila bisa diketahui.
func Compile(source string) (*Program, error) {
	tokens := lexer.Tokenize(source)
	for _, tok := range tokens {
		if tok.Type == token.ILLEGAL {
			return nil, fmt.Errorf("galat leksikal pada baris %d: karakter tidak dikenal %q",
				tok.Line, tok.Literal)
		}
	}

	p := parser.New(tokens)
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		return nil, fmt.Errorf("galat parsing: %s", strings.Join(errs, "; "))
	}

	return &Program{ast: program}, nil
}

// CompileFile membaca berkas .nusa lalu mengurai isinya menjadi Program.
func CompileFile(path string) (*Program, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("berkas %s tidak bisa dibaca: %w", path, err)
	}

	program, err := Compile(string(data))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return program, nil
}

// Limits membatasi sumber daya eksekusi. Nilai nol berarti tanpa batas.
type Limits struct {
	// MaxDepth membatasi kedalaman pemanggilan fungsi.
	MaxDepth int
	// MaxSteps membatasi jumlah node yang dievaluasi.
	MaxSteps int64
}

// DefaultLimits mengembalikan batas bawaan yang dipakai runtime.
func DefaultLimits() Limits {
	l := runtime.DefaultLimits()
	return Limits{MaxDepth: l.MaxDepth, MaxSteps: l.MaxSteps}
}

// Error adalah galat runtime Nusa beserta posisi sumbernya.
type Error struct {
	// Message adalah pesan galat dalam bahasa Nusa.
	Message string
	// Line adalah nomor baris tempat galat muncul, 0 bila tidak diketahui.
	Line int
	// Column adalah kolom tempat galat muncul, 0 bila tidak diketahui.
	Column int
	// Trace adalah jejak pemanggilan fungsi, dari terdalam ke luar.
	Trace []string
}

// Error mengembalikan pesan galat agar tipe ini bisa dipakai sebagai error Go.
func (e *Error) Error() string {
	var out strings.Builder
	out.WriteString(e.Message)
	if e.Line > 0 {
		fmt.Fprintf(&out, " (baris %d)", e.Line)
	}
	for _, frame := range e.Trace {
		out.WriteString("\n  di ")
		out.WriteString(frame)
	}
	return out.String()
}

// Value adalah nilai Nusa yang bisa disuntikkan ke program lewat
// Runtime.SetGlobal. Gunakan fungsi pembuatnya: Text, Number, Decimal,
// Bool, List, atau Map.
type Value struct {
	obj runtime.Object
}

// Text membuat nilai bertipe teks.
func Text(s string) Value {
	return Value{obj: &runtime.String{Value: s}}
}

// Number membuat nilai bertipe angka bulat.
func Number(n int64) Value {
	return Value{obj: &runtime.Integer{Value: n}}
}

// Decimal membuat nilai bertipe desimal.
func Decimal(f float64) Value {
	return Value{obj: &runtime.Float{Value: f}}
}

// Bool membuat nilai bertipe logika.
func Bool(b bool) Value {
	return Value{obj: &runtime.Boolean{Value: b}}
}

// List membuat nilai bertipe daftar. Nilai Earned yang tidak diberikan
// menjadi daftar kosong.
func List(values ...Value) Value {
	elems := make([]runtime.Object, 0, len(values))
	for _, v := range values {
		elems = append(elems, v.obj)
	}
	return Value{obj: &runtime.List{Elements: elems}}
}

// Map membuat nilai bertipe peta. Kunci diurutkan agar isi peta selalu sama
// untuk masukan yang sama.
func Map(entries map[string]Value) Value {
	peta := runtime.NewPeta()
	kunci := make([]string, 0, len(entries))
	for k := range entries {
		kunci = append(kunci, k)
	}
	sort.Strings(kunci)
	for _, k := range kunci {
		_ = peta.Set(&runtime.String{Value: k}, entries[k].obj)
	}
	return Value{obj: peta}
}

// Runtime adalah lingkungan eksekusi program Nusa.
//
// Setiap runtime punya lingkup variabel sendiri, sehingga nilai yang
// disuntik lewat SetGlobal tidak bocor ke runtime lain.
type Runtime struct {
	rt *runtime.Runtime
}

// NewRuntime membuat runtime yang menulis keluaran ke out dan membaca
// masukan dari in. Parameter nil berarti keluaran dibuang dan masukan
// dianggap kosong.
func NewRuntime(out io.Writer, in io.Reader) *Runtime {
	if out == nil {
		out = io.Discard
	}
	if in == nil {
		in = strings.NewReader("")
	}
	return &Runtime{rt: runtime.NewWithIO(out, in)}
}

// SetGlobal menyuntik nilai ke lingkup global runtime. Nilai disuntik
// sebelum program dijalankan, sehingga program Nusa membacanya seperti
// variabel biasa. Nama yang bentrok dengan konstanta global diabaikan.
func (r *Runtime) SetGlobal(name string, value Value) {
	r.rt.env.store[name] = value.obj
}

// Global mengembalikan nilai variabel tingkat atas dengan nama tersebut.
func (r *Runtime) Global(name string) (Value, bool) {
	obj, ok := r.rt.env.Get(name)
	if !ok {
		return Value{}, false
	}
	return Value{obj: obj}, true
}

// SetLimits mengganti batas eksekusi runtime.
func (r *Runtime) SetLimits(limits Limits) {
	r.rt.SetLimits(runtime.Limits{MaxDepth: limits.MaxDepth, MaxSteps: limits.MaxSteps})
}

// SetPromptWriter memisahkan stream prompt "> " dari stream keluaran
// program. Tanpa pemisahan ini, prompt ikut tercetak di dalam keluaran
// yang akan dikirim ke pengguna, misalnya halaman HTML.
// Di versi ini, penanda prompt tidak digunakan (no-op).
func (r *Runtime) SetPromptWriter(w io.Writer) {
}

// EnableModules mengaktifkan pernyataan `impor` dengan berkas modul dibaca
// relatif terhadap direktori dir. Modul yang sama hanya dieksekusi satu kali
// dan hasilnya di-cache antar request.
func (r *Runtime) EnableModules(dir string) {
	if dir == "" {
		dir = "."
	}
	loader := modul.New(filepath.Clean(dir), r.rt.Out, r.rt.In)
	r.rt.SetModuleLoader(loader.Import)
}

// Execute menjalankan program dan mengembalikan galat bila eksekusi gagal.
// Keluaran program ditulis ke stream yang diberikan saat runtime dibuat.
func (r *Runtime) Execute(program *Program) error {
	if program == nil {
		return fmt.Errorf("program Nusa kosong")
	}

	hasil := r.rt.Execute(program.ast)
	if hasil == nil {
		return nil
	}
	if errObj, ok := hasil.(*runtime.Error); ok {
		return &Error{
			Message: errObj.Message,
			Line:    errObj.Line,
			Column:  errObj.Column,
			Trace:   jejak(errObj),
		}
	}
	return nil
}

// jejak mengubah jejak pemanggilan runtime menjadi daftar teks.
func jejak(errObj *runtime.Error) []string {
	if len(errObj.Trace) == 0 {
		return nil
	}
	frames := make([]string, 0, len(errObj.Trace))
	for _, frame := range errObj.Trace {
		frames = append(frames, fmt.Sprintf("%s (baris %d)", frame.Name, frame.Line))
	}
	return frames
}
