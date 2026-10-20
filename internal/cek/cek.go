// Package cek berisi pemeriksa statis untuk program Nusa.
//
// Pemeriksa berjalan tanpa mengeksekusi program. Ia menelusuri pohon
// sintaks sambil menjaga tumpukan lingkup, lalu melaporkan:
//   - variabel yang dipakai tapi tidak pernah dideklarasikan,
//   - variabel lokal yang dideklarasikan tapi tidak pernah dipakai,
//   - nilai yang bertipe tidak cocok dengan anotasi tipe,
//   - deklarasi ganda pada lingkup yang sama.
//
// Lingkup Nusa bersifat leksikal: blok kurung kurawalintroduce lingkup
// baru dan fungsi introduce lingkup baru yang berisi parameter-nya.
package cek

import (
	"fmt"
	"sort"

	"github.com/seraphimhub/Nusa/internal/ast"
	"github.com/seraphimhub/Nusa/internal/runtime"
	"github.com/seraphimhub/Nusa/internal/tipe"
	"github.com/seraphimhub/Nusa/internal/token"
)

// Severity menandai seberapa serius sebuah temuan.
type Severity string

const (
	// Galat adalah temuan yang pasti membuat program salah.
	Galat Severity = "galat"
	// Peringatan adalah temuan yang mencurigakan tapi program mungkin
	// tetap benar, misalnya variabel yang dideklarasikan tapi belum dipakai.
	Peringatan Severity = "peringatan"
)

// Temuan adalah satu hasil pemeriksaan statis.
type Temuan struct {
	Severity Severity
	Line     int
	Column   int
	Pesan    string
}

// String merender temuan dalam satu baris.
func (t Temuan) String() string {
	return fmt.Sprintf("%s:%d:%d: %s", t.Severity, t.Line, t.Column, t.Pesan)
}

// lingkup menyimpan nama yang dikenal pada satu tingkat lingkup.
type lingkup struct {
	// nama berisi seluruh nama yang dideklarasikan di lingkup ini.
	nama map[string]bool
	// dipakai berisi nama yang benar-benar dirujuk di lingkup ini.
	dipakai map[string]bool
	// fungsi berisi nama yang dideklarasikan sebagai fungsi.
	fungsi map[string]bool
	// konstan berisi nama yang dideklarasikan dengan kata tetap.
	konstan map[string]bool
	// impor berisi nama yang berasal dari modul yang diimpor. Nama ini
	// tidak pernah dilaporkan sebagai tak terpakai karena dipakai di
	// berkas lain.
	impor map[string]bool
	// token menyimpan posisi deklarasi tiap nama untuk laporan.
	token map[string]token.Token
	// parent adalah lingkup luar. Nil untuk lingkup global.
	parent *lingkup
}

// baruLingkup membuat lingkup kosong dengan induk yang diberikan.
func baruLingkup(parent *lingkup) *lingkup {
	return &lingkup{
		nama:    map[string]bool{},
		dipakai: map[string]bool{},
		fungsi:  map[string]bool{},
		konstan: map[string]bool{},
		impor:   map[string]bool{},
		token:   map[string]token.Token{},
		parent:  parent,
	}
}

// ada melaporkan apakah nama dikenal pada lingkup ini atau lingkup luar.
func (l *lingkup) ada(nama string) bool {
	for s := l; s != nil; s = s.parent {
		if s.nama[nama] {
			return true
		}
	}
	return false
}

// tandaiDipakai mencatat bahwa nama dirujuk di lingkup tempat dideklarasi.
func (l *lingkup) tandaiDipakai(nama string) {
	for s := l; s != nil; s = s.parent {
		if s.nama[nama] {
			s.dipakai[nama] = true
			return
		}
	}
}

// Opsi adalah pilihan tambahan saat memeriksa program.
type Opsi struct {
	// Impor mengembalikan nama yang diekspor modul pada path tertentu.
	// Pemeriksa memakai daftar ini agar nama dari impor tidak dianggap
	// variabel tak terdefinisi. Bila nil, impor tidak bisa diselesaikan
	// sehingga nama dari impor tidak diperiksa.
	Impor func(path string) ([]string, error)
}

// Pemeriksa menelusuri program dan mengumpulkan temuan.
type Pemeriksa struct {
	temuan []Temuan
	// builtins adalah nama fungsi bawaan yang sah.
	builtins map[string]bool
	// global adalah lingkup tingkat atas program.
	global *lingkup
	// opsi menyimpan pilihan pemeriksaan.
	opsi Opsi
	// imporGagal menandai ada impor yang tidak bisa diselesaikan. Setelah
	// itu, variabel yang tidak dikenal tidak dilaporkan agar tidak ada
	// false positive.
	imporGagal bool
}

// Periksa memeriksa program dan mengembalikan seluruh temuan. Program yang
// gagal parse tidak boleh dikirim ke sini; pemeriksa bertanggung jawab
// memastikan program valid secara sintaks.
func Periksa(program *ast.Program) []Temuan {
	return PeriksaDengan(program, Opsi{})
}

// PeriksaDengan memeriksa program dengan pilihan tambahan dan mengembalikan
// seluruh temuan.
func PeriksaDengan(program *ast.Program, opsi Opsi) []Temuan {
	builtins := make(map[string]bool)
	for _, nama := range runtime.BuiltinNames() {
		builtins[nama] = true
	}

	p := &Pemeriksa{
		builtins: builtins,
		global:   baruLingkup(nil),
		opsi:     opsi,
	}

	p.periksaBlok(p.global, program.Statements)
	p.laporTakTerpakai(p.global)

	return urut(p.temuan)
}

// urut mengurutkan temuan berdasarkan posisi di berkas.
func urut(temuan []Temuan) []Temuan {
	out := make([]Temuan, len(temuan))
	copy(out, temuan)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Line != out[j].Line {
			return out[i].Line < out[j].Line
		}
		return out[i].Column < out[j].Column
	})
	return out
}

// periksaBlok memeriksa daftar pernyataan pada satu lingkup.
func (p *Pemeriksa) periksaBlok(l *lingkup, statements []ast.Statement) {
	for _, stmt := range statements {
		p.periksaPernyataan(l, stmt)
	}
}

// periksaPernyataan memeriksa satu pernyataan.
func (p *Pemeriksa) periksaPernyataan(l *lingkup, stmt ast.Statement) {
	switch n := stmt.(type) {
	case nil:
		return

	case *ast.LetStatement:
		p.periksaEkspresi(l, n.Value)
		if n.TypeHint != "" {
			p.cekTipeHint(n.Name, n.TypeHint, n.Value)
		}
		p.deklarasi(l, n.Name, n.Konst, false)

	case *ast.PrintStatement:
		p.periksaEkspresi(l, n.Value)

	case *ast.InputStatement:
		p.deklarasi(l, n.Name, false, false)

	case *ast.ExpressionStatement:
		p.periksaEkspresi(l, n.Expression)

	case *ast.BlockStatement:
		// Blok di dalam tubuh fungsi atau percabangan memakai lingkup
		// yang sama; lingkup baru dibuat oleh pemanggil yang memahami
		// batasnya.
		p.periksaBlok(l, n.Statements)

	case *ast.IfStatement:
		p.periksaEkspresi(l, n.Condition)
		p.periksaBlok(l, blokStatements(n.Consequence))
		p.periksaBlok(l, blokStatements(n.Alternative))

	case *ast.ForStatement:
		p.periksaEkspresi(l, n.Count)
		if n.VarName != "" {
			l.nama[n.VarName] = true
		}
		p.periksaBlok(l, blokStatements(n.Body))

	case *ast.WhileStatement:
		p.periksaEkspresi(l, n.Condition)
		p.periksaBlok(l, blokStatements(n.Body))

	case *ast.UntukStatement:
		p.periksaEkspresi(l, n.Start)
		p.periksaEkspresi(l, n.End)
		p.periksaEkspresi(l, n.Step)
		l.nama[n.Name.Value] = true
		p.periksaBlok(l, blokStatements(n.Body))

	case *ast.UntukDalamStatement:
		p.periksaEkspresi(l, n.Iterable)
		for _, nama := range n.Names {
			l.nama[nama.Value] = true
		}
		p.periksaBlok(l, blokStatements(n.Body))

	case *ast.PilihStatement:
		p.periksaEkspresi(l, n.Subjek)
		for _, kasus := range n.Kasus {
			p.periksaEkspresi(l, kasus.Nilai)
			p.periksaBlok(l, blokStatements(kasus.Body))
		}
		p.periksaBlok(l, blokStatements(n.Bawaan))

	case *ast.CobaStatement:
		p.periksaBlok(l, blokStatements(n.Try))
		if n.ErrName != nil {
			l.nama[n.ErrName.Value] = true
		}
		p.periksaBlok(l, blokStatements(n.Catch))

	case *ast.LemparkanStatement:
		p.periksaEkspresi(l, n.Value)

	case *ast.ImporStatement:
		p.periksaImpor(l, n)

	case *ast.FunctionStatement:
		if l.nama[n.Name.Value] {
			p.gandakan(l, n.Name, "fungsi")
			return
		}
		l.nama[n.Name.Value] = true
		l.fungsi[n.Name.Value] = true
		l.token[n.Name.Value] = n.Name.Token

		// Fungsi membuat lingkup baru berisi parameter-parameternya.
		inner := baruLingkup(l)
		for _, param := range n.Params {
			inner.nama[param.Name.Value] = true
			inner.token[param.Name.Value] = param.Name.Token
		}
		p.periksaBlok(inner, blokStatements(n.Body))
		p.laporTakTerpakai(inner)
		p.laporParameterTakTerpakai(inner, n.Params)

	case *ast.CallStatement:
		p.periksaEkspresi(l, n.CallExpr)

	case *ast.ReturnStatement:
		p.periksaEkspresi(l, n.ReturnValue)

	case *ast.BreakStatement, *ast.ContinueStatement:
		// Tidak ada ekspresi yang perlu diperiksa.
	}
}

// periksaImpor mendaftarkan nama yang diekspor modul ke lingkup saat ini.
// Bila modul tidak bisa di-resolve, pemeriksaan variabel tak terdefinisi
// dimatikan untuk sisa berkas agar tidak ada false positive.
func (p *Pemeriksa) periksaImpor(l *lingkup, n *ast.ImporStatement) {
	if p.opsi.Impor == nil {
		p.imporGagal = true
		return
	}

	nama, err := p.opsi.Impor(n.Path)
	if err != nil {
		p.imporGagal = true
		p.temuan = append(p.temuan, Temuan{
			Severity: Galat,
			Line:     n.Token.Line,
			Column:   n.Token.Column,
			Pesan:    err.Error(),
		})
		return
	}

	for _, namaModul := range nama {
		l.nama[namaModul] = true
		l.impor[namaModul] = true
		l.token[namaModul] = n.Token
	}
}

// deklarasi mencatat deklarasi baru dan memberi tahu bila nama ganda.
func (p *Pemeriksa) deklarasi(l *lingkup, id *ast.Identifier, konstan bool, fungsi bool) {
	if l.nama[id.Value] {
		p.gandakan(l, id, "variabel")
		return
	}
	l.nama[id.Value] = true
	l.konstan[id.Value] = konstan
	l.fungsi[id.Value] = fungsi
	l.token[id.Value] = id.Token
}

// gandakan melaporkan deklarasi ganda pada lingkup yang sama.
func (p *Pemeriksa) gandakan(l *lingkup, id *ast.Identifier, jenis string) {
	p.temuan = append(p.temuan, Temuan{
		Severity: Galat,
		Line:     id.Token.Line,
		Column:   id.Token.Column,
		Pesan: fmt.Sprintf("%s '%s' sudah dideklarasikan di lingkup ini",
			jenis, id.Value),
	})
}

// blokStatements mengembalikan daftar pernyataan sebuah blok, atau daftar
// kosong bila blok tidak ada.
func blokStatements(b *ast.BlockStatement) []ast.Statement {
	if b == nil {
		return nil
	}
	return b.Statements
}

// periksaEkspresi memeriksa satu ekspresi dan semua ekspresi turunan.
func (p *Pemeriksa) periksaEkspresi(l *lingkup, e ast.Expression) {
	switch n := e.(type) {
	case nil:
		return

	case *ast.Identifier:
		p.cekIdentifier(l, n)

	case *ast.PrefixExpression:
		p.periksaEkspresi(l, n.Right)

	case *ast.InfixExpression:
		p.periksaEkspresi(l, n.Left)
		p.periksaEkspresi(l, n.Right)

	case *ast.CallExpression:
		p.periksaEkspresi(l, n.Function)
		for _, arg := range n.Arguments {
			p.periksaEkspresi(l, arg)
		}

	case *ast.ListLiteral:
		for _, el := range n.Elements {
			p.periksaEkspresi(l, el)
		}

	case *ast.IndexExpression:
		p.periksaEkspresi(l, n.Left)
		p.periksaEkspresi(l, n.Index)

	case *ast.PetaLiteral:
		for _, entri := range n.Entri {
			p.periksaEkspresi(l, entri.Kunci)
			p.periksaEkspresi(l, entri.Nilai)
		}

	case *ast.AssignExpression:
		p.periksaEkspresi(l, n.Value)
		// Target penugasan dibaca, bukan ditulis, jadi harus sudah ada.
		switch target := n.Target.(type) {
		case *ast.Identifier:
			p.cekIdentifier(l, target)
		default:
			p.periksaEkspresi(l, n.Target)
		}
	}
}

// cekIdentifier memastikan nama yang dirujuk memang dikenal.
func (p *Pemeriksa) cekIdentifier(l *lingkup, id *ast.Identifier) {
	if l.ada(id.Value) {
		l.tandaiDipakai(id.Value)
		return
	}
	if p.builtins[id.Value] {
		return
	}
	// Bila ada impor yang gagal di-resolve, nama dari impor tidak bisa
	// dikenali sehingga laporkan saja sebagai peringatan.
	if p.imporGagal {
		p.temuan = append(p.temuan, Temuan{
			Severity: Peringatan,
			Line:     id.Token.Line,
			Column:   id.Token.Column,
			Pesan:    fmt.Sprintf("variabel '%s' tidak dapat dipastikan ada karena impor gagal dimuat", id.Value),
		})
		return
	}
	p.temuan = append(p.temuan, Temuan{
		Severity: Galat,
		Line:     id.Token.Line,
		Column:   id.Token.Column,
		Pesan:    fmt.Sprintf("variabel '%s' tidak ditemukan", id.Value),
	})
}

// cekTipeHint memeriksa apakah nilai ekspresi cocok dengan anotasi tipe yang
// ditulis. Hanya nilai yang tipenya bisa dipastikan secara statis yang
// dilaporkan; nilai lain dilewati agar tidak ada false positive.
func (p *Pemeriksa) cekTipeHint(nama *ast.Identifier, hint string, value ast.Expression) {
	if hint == tipe.ApaSaja {
		return
	}

	aktual, pasti := tipeStatis(value)
	if !pasti {
		return
	}

	if !cocok(aktual, hint) {
		// Tanda panah menunjuk nilai yang salah, bukan nama variabel,
		// supaya letak masalahnya jelas.
		tok := ast.TokenOf(value)
		if tok.Line == 0 {
			tok = nama.Token
		}
		p.temuan = append(p.temuan, Temuan{
			Severity: Galat,
			Line:     tok.Line,
			Column:   tok.Column,
			Pesan: fmt.Sprintf("nilai untuk '%s' harus bertipe %s, mendapat %s",
				nama.Value, hint, aktual),
		})
	}
}

// tipeStatis menebak tipe nilai dari bentuk literalnya. Nilai kedua bernilai
// salah bila tipenya belum bisa dipastikan.
func tipeStatis(e ast.Expression) (string, bool) {
	switch n := e.(type) {
	case nil:
		return "", false
	case *ast.IntegerLiteral:
		return tipe.Angka, true
	case *ast.FloatLiteral:
		return tipe.Desimal, true
	case *ast.StringLiteral:
		return tipe.Teks, true
	case *ast.BooleanLiteral:
		return tipe.Logika, true
	case *ast.NilLiteral:
		return tipe.Kosong, true
	case *ast.ListLiteral:
		return tipe.Daftar, true
	case *ast.PetaLiteral:
		return tipe.Peta, true
	case *ast.InfixExpression:
		// Operator aritmetika pada dua angka menghasilkan angka; untuk
		// angka dan desimal hasilnya desimal.
		left, lok := tipeStatis(n.Left)
		right, rok := tipeStatis(n.Right)
		if !lok || !rok {
			return "", false
		}
		switch n.Operator {
		case "+", "-", "*", "/", "%":
			if left == tipe.Angka && right == tipe.Angka {
				return tipe.Angka, true
			}
			if (left == tipe.Angka || left == tipe.Desimal) &&
				(right == tipe.Angka || right == tipe.Desimal) {
				return tipe.Desimal, true
			}
			return "", false
		case "<", ">", "<=", ">=", "==", "!=", "dan", "atau":
			return tipe.Logika, true
		}
		return "", false
	case *ast.PrefixExpression:
		inner, ok := tipeStatis(n.Right)
		if !ok {
			return "", false
		}
		switch n.Operator {
		case "!", "tidak":
			return tipe.Logika, true
		case "-":
			return inner, true
		}
		return "", false
	}
	return "", false
}

// cocok memeriksa apakah tipe aktual memenuhi anotasi tipe.
func cocok(aktual string, hint string) bool {
	if hint == tipe.ApaSaja || aktual == "" {
		return true
	}
	if aktual == hint {
		return true
	}
	// Angka bulat bisa dipakai di tempat desimal karena bisa dikonversi
	// tanpa kehilangan ketelitian.
	if hint == tipe.Desimal && aktual == tipe.Angka {
		return true
	}
	return false
}

// laporTakTerpakai memeriksa nama yang tidak pernah dirujuk.
func (p *Pemeriksa) laporTakTerpakai(l *lingkup) {
	nama := make([]string, 0, len(l.nama))
	for n := range l.nama {
		nama = append(nama, n)
	}
	sort.Strings(nama)

	for _, n := range nama {
		if l.dipakai[n] {
			continue
		}
		// Fungsi dan konstanta tingkat atas boleh tidak dipakai karena
		// bisa dipanggil dari berkas lain lewat impor.
		if l.parent == nil && (l.fungsi[n] || l.konstan[n]) {
			continue
		}
		// Nama dari impor dipakai di berkas lain.
		if l.impor[n] {
			continue
		}
		tok := l.token[n]
		p.temuan = append(p.temuan, Temuan{
			Severity: Peringatan,
			Line:     tok.Line,
			Column:   tok.Column,
			Pesan:    fmt.Sprintf("'%s' dideklarasikan tapi tidak pernah dipakai", n),
		})
	}
}

// laporParameterTakTerpakai memeriksa parameter fungsi yang tidak dipakai.
func (p *Pemeriksa) laporParameterTakTerpakai(l *lingkup, params []*ast.Param) {
	for _, param := range params {
		if l.dipakai[param.Name.Value] {
			continue
		}
		p.temuan = append(p.temuan, Temuan{
			Severity: Peringatan,
			Line:     param.Name.Token.Line,
			Column:   param.Name.Token.Column,
			Pesan: fmt.Sprintf("parameter '%s' tidak pernah dipakai",
				param.Name.Value),
		})
	}
}
