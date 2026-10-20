// Package format menyediakan pemformat kode Nusa.
//
// Pemformat bekerja di atas token, bukan atas teks mentah, sehingga hanya
// bagian tata letak yang diubah: indentasi, spasi di sekitar operator, dan
// spasi setelah koma. Komentar dan literal teks ditulis apa adanya supaya
// isi program tidak pernah berubah.
//
// Pemformat tidak mengubah struktur baris. Kalau penulis menulis ekspresi
// panjang di beberapa baris, baris itu tetap dipecah sama seperti aslinya.
package format

import (
	"strings"

	"github.com/seraphimhub/Nusa/internal/lexer"
	"github.com/seraphimhub/Nusa/internal/token"
)

// indentasi adalah jumlah spasi untuk satu tingkat indentasi.
const indentasi = 2

// Format memformat sumber Nusa dan mengembalikan hasilnya. Baris terakhir
// selalu diakhiri newline.
func Format(source string) string {
	p := &pemformat{
		src:       source,
		barisAwal: hitungAwalBaris(source),
	}
	p.jalankan()
	return p.out.String()
}

// pemformat menyimpan keadaan saat memformat.
type pemformat struct {
	src string
	// barisAwal berisi offset awal setiap baris sumber, dipakai untuk
	// mengambil literal teks apa adanya.
	barisAwal []int
	out       strings.Builder
	// baris adalah isi baris keluaran yang sedang disusun.
	baris strings.Builder
	// indentBaris adalah indentasi baris keluaran yang sedang disusun.
	indentBaris int
	// kedalaman adalah jumlah kurung kurawal yang belum tertutup.
	kedalaman int
	// prev adalah token terakhir yang sudah ditulis.
	prev token.Token
	// adaPrev menandai apakah prev sudah berisi token yang sah.
	adaPrev bool
	// unary menandai tanda minus terakhir bersifat unary.
	unary bool
}

// hitungAwalBaris menghitung offset awal setiap baris sumber.
func hitungAwalBaris(source string) []int {
	awal := []int{0}
	for i := 0; i < len(source); i++ {
		if source[i] == '\n' {
			awal = append(awal, i+1)
		}
	}
	return awal
}

// jalankan memproses seluruh token sumber.
func (p *pemformat) jalankan() {
	for _, tok := range lexer.TokenizeWithComments(p.src) {
		switch tok.Type {
		case token.EOF:
			// Selesai; sisanya ditangani oleh flush terakhir.
		case token.NEWLINE:
			p.tutupBaris()
		case token.KOMENTAR:
			p.tambahKomentar(tok)
		default:
			p.tambahToken(tok)
		}
	}
	p.tutupBarisTanpaBarisBaru()
}

// tutupBaris menulis baris keluaran yang sedang disusun lalu memulai
// baris baru.
func (p *pemformat) tutupBaris() {
	if p.baris.Len() > 0 {
		p.out.WriteString(strings.Repeat(" ", p.indentBaris))
		p.out.WriteString(p.baris.String())
		p.baris.Reset()
	}
	p.out.WriteString("\n")
	p.adaPrev = false
	p.unary = false
}

// tutupBarisTanpaBarisBaru menutup baris terakhir tanpa menambah baris
// kosong di akhir keluaran.
func (p *pemformat) tutupBarisTanpaBarisBaru() {
	if p.baris.Len() > 0 {
		p.tutupBaris()
	}
}

// tambahToken menulis satu token ke baris keluaran.
func (p *pemformat) tambahToken(tok token.Token) {
	if tok.Type == token.RBRACE {
		p.kedalaman--
		if p.kedalaman < 0 {
			p.kedalaman = 0
		}
	}

	if p.baris.Len() == 0 {
		p.indentBaris = p.kedalaman * indentasi
	} else if p.perluSpasi(p.prev, tok) {
		p.baris.WriteByte(' ')
	}

	if tok.Type == token.MINUS {
		p.unary = p.minusUnary()
	}

	p.baris.WriteString(p.teks(tok))
	p.prev = tok
	p.adaPrev = true

	if tok.Type == token.LBRACE {
		p.kedalaman++
	}
}

// minusUnary melaporkan apakah tanda minus terakhir dipakai sebagai operator
// unary, yaitu bila token sebelumnya adalah operator, tanda buka, atau awal
// baris.
func (p *pemformat) minusUnary() bool {
	if !p.adaPrev {
		return true
	}
	switch p.prev.Type {
	case token.ASSIGN, token.PLUS_ASSIGN, token.MINUS_ASSIGN,
		token.MULTIPLY_ASSIGN, token.DIVIDE_ASSIGN, token.MODULO_ASSIGN,
		token.PLUS, token.MINUS, token.MULTIPLY, token.DIVIDE, token.MODULO,
		token.GREATER, token.LESS, token.GTE, token.LTE, token.EQ, token.NEQ,
		token.BANG, token.DAN, token.ATAU, token.TIDAK,
		token.LPAREN, token.LBRACKET, token.LBRACE, token.COMMA, token.TITIK_DUA:
		return true
	}
	return false
}

// tambahKomentar menulis komentar ke baris keluaran. Komentar yang berdiri
// sendiri tetap menjadi baris sendiri; komentar di ujung baris menempel
// pada baris itu.
func (p *pemformat) tambahKomentar(tok token.Token) {
	teks := p.teks(tok)

	if p.baris.Len() == 0 {
		p.indentBaris = p.kedalaman * indentasi
		p.baris.WriteString(teks)
	} else {
		p.baris.WriteByte(' ')
		p.baris.WriteString(teks)
	}

	p.prev = tok
	p.adaPrev = true
}

// perluSpasi menentukan apakah perlu spasi antara dua token yang berurutan
// pada baris yang sama.
func (p *pemformat) perluSpasi(prev token.Token, cur token.Token) bool {
	// Tanda baca yang menempel ke kiri tidak didahului spasi.
	switch cur.Type {
	case token.COMMA, token.RPAREN, token.RBRACKET, token.TITIK_DUA:
		return false
	case token.LPAREN:
		// Panggilan fungsi dan pengelompokan ditulis rapat, tetapi
		// setelah kata kunci tetap ada spasi.
		switch prev.Type {
		case token.IDENT, token.RPAREN, token.RBRACKET:
			return false
		}
		return true
	case token.LBRACKET:
		switch prev.Type {
		case token.IDENT, token.RPAREN, token.RBRACKET, token.STRING:
			return false
		}
		return true
	}

	// Tanda buka tidak diikuti spasi.
	switch prev.Type {
	case token.LPAREN, token.LBRACKET:
		return false
	case token.BANG:
		return false
	case token.MINUS:
		if p.unary {
			return false
		}
	}

	return true
}

// teks mengembalikan teks sumber untuk sebuah token. Literal teks ditulis
// apa adanya supaya escape tetap utuh.
func (p *pemformat) teks(tok token.Token) string {
	if tok.Type != token.STRING {
		return tok.Literal
	}

	awal := p.offset(tok.Line, tok.Column)
	if awal < 0 || awal >= len(p.src) {
		return tok.Literal
	}

	quote := p.src[awal]
	if quote != '"' && quote != '\'' {
		return tok.Literal
	}

	for i := awal + 1; i < len(p.src); i++ {
		if p.src[i] == '\\' {
			i++
			continue
		}
		if p.src[i] == quote {
			return p.src[awal : i+1]
		}
	}

	return p.src[awal:]
}

// offset mengubah posisi baris dan kolom menjadi offset di dalam sumber.
func (p *pemformat) offset(line int, column int) int {
	if line < 1 || line > len(p.barisAwal) || column < 1 {
		return -1
	}
	return p.barisAwal[line-1] + column - 1
}
