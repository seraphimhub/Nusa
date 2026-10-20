package ast

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/seraphimhub/Nusa/internal/token"
)

// Node is the base interface for all AST nodes.
type Node interface {
	TokenLiteral() string
	String() string
}

// Statement is a node that doesn't produce a value.
type Statement interface {
	Node
	statementNode()
}

// TokenOf mengembalikan token utama sebuah node. Token ini dipakai runtime
// dan perkakas untuk menunjuk posisi galat dengan nomor baris dan kolom.
// Node yang tidak punya token sendiri (misalnya Program) mengembalikan token
// kosong.
func TokenOf(node Node) token.Token {
	switch n := node.(type) {
	case *LetStatement:
		return n.Token
	case *PrintStatement:
		return n.Token
	case *InputStatement:
		return n.Token
	case *ExpressionStatement:
		return n.Token
	case *BlockStatement:
		return n.Token
	case *IfStatement:
		return n.Token
	case *ForStatement:
		return n.Token
	case *WhileStatement:
		return n.Token
	case *FunctionStatement:
		return n.Token
	case *CallStatement:
		return n.Token
	case *ReturnStatement:
		return n.Token
	case *BreakStatement:
		return n.Token
	case *ContinueStatement:
		return n.Token
	case *Identifier:
		return n.Token
	case *IntegerLiteral:
		return n.Token
	case *FloatLiteral:
		return n.Token
	case *StringLiteral:
		return n.Token
	case *BooleanLiteral:
		return n.Token
	case *NilLiteral:
		return n.Token
	case *PrefixExpression:
		return n.Token
	case *InfixExpression:
		return n.Token
	case *CallExpression:
		return n.Token
	case *ListLiteral:
		return n.Token
	case *IndexExpression:
		return n.Token
	case *PetaLiteral:
		return n.Token
	case *AssignExpression:
		return n.Token
	case *UntukStatement:
		return n.Token
	case *CobaStatement:
		return n.Token
	case *LemparkanStatement:
		return n.Token
	case *PilihStatement:
		return n.Token
	case *UntukDalamStatement:
		return n.Token
	case *ImporStatement:
		return n.Token
	case *UjiStatement:
		return n.Token
	case *CekStatement:
		return n.Token
	}
	return token.Token{}
}

// Expression is a node that produces a value.
type Expression interface {
	Node
	expressionNode()
}

// Program is the root of the AST.
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	for _, s := range p.Statements {
		if s != nil {
			return s.TokenLiteral()
		}
	}
	return ""
}

func (p *Program) String() string {
	var out bytes.Buffer
	for _, s := range p.Statements {
		if s == nil {
			continue
		}
		out.WriteString(s.String())
		out.WriteString("\n")
	}
	return out.String()
}

// --- Statements ---

type LetStatement struct {
	Token token.Token // token.BUAT atau token.TETAP
	Name  *Identifier
	Value Expression
	Konst bool // true bila dideklarasikan dengan 'tetap'
	// TypeHint berisi nama tipe bila ditulis, misalnya "angka". Kosong
	// bila deklarasi tanpa anotasi tipe.
	TypeHint string
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }
func (ls *LetStatement) String() string {
	var out bytes.Buffer
	out.WriteString(ls.TokenLiteral() + " ")
	out.WriteString(ls.Name.String())
	if ls.TypeHint != "" {
		out.WriteString(": " + ls.TypeHint)
	}
	out.WriteString(" = ")
	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}
	return out.String()
}

type PrintStatement struct {
	Token token.Token // token.TULIS
	Value Expression
}

func (ps *PrintStatement) statementNode()       {}
func (ps *PrintStatement) TokenLiteral() string { return ps.Token.Literal }
func (ps *PrintStatement) String() string {
	var out bytes.Buffer
	out.WriteString(ps.TokenLiteral() + " ")
	if ps.Value != nil {
		out.WriteString(ps.Value.String())
	}
	return out.String()
}

type InputStatement struct {
	Token token.Token // token.MEMBACA
	Name  *Identifier
}

func (is *InputStatement) statementNode()       {}
func (is *InputStatement) TokenLiteral() string { return is.Token.Literal }
func (is *InputStatement) String() string {
	return is.TokenLiteral() + " " + is.Name.String()
}

type ExpressionStatement struct {
	Token      token.Token
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

type BlockStatement struct {
	Token      token.Token // token.LBRACE
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) String() string {
	var out bytes.Buffer
	for _, s := range bs.Statements {
		if s == nil {
			continue
		}
		out.WriteString(s.String())
		out.WriteString("\n")
	}
	return out.String()
}

type IfStatement struct {
	Token       token.Token // token.JIKA
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (is *IfStatement) statementNode()       {}
func (is *IfStatement) TokenLiteral() string { return is.Token.Literal }
func (is *IfStatement) String() string {
	var out bytes.Buffer
	out.WriteString("jika ")
	out.WriteString(is.Condition.String())
	out.WriteString(" ")
	out.WriteString(is.Consequence.String())
	if is.Alternative != nil {
		out.WriteString("kalau_tidak ")
		out.WriteString(is.Alternative.String())
	}
	return out.String()
}

type ForStatement struct {
	Token   token.Token // token.ULANG
	Count   Expression
	Body    *BlockStatement
	VarName string // optional: iteration variable name
}

func (fs *ForStatement) statementNode()       {}
func (fs *ForStatement) TokenLiteral() string { return fs.Token.Literal }
func (fs *ForStatement) String() string {
	var out bytes.Buffer
	out.WriteString("ulang ")
	out.WriteString(fs.Count.String())
	out.WriteString(" ")
	out.WriteString(fs.Body.String())
	return out.String()
}

type WhileStatement struct {
	Token     token.Token // token.SELAMA
	Condition Expression
	Body      *BlockStatement
}

func (ws *WhileStatement) statementNode()       {}
func (ws *WhileStatement) TokenLiteral() string { return ws.Token.Literal }
func (ws *WhileStatement) String() string {
	var out bytes.Buffer
	out.WriteString("selama ")
	out.WriteString(ws.Condition.String())
	out.WriteString(" ")
	out.WriteString(ws.Body.String())
	return out.String()
}

type FunctionStatement struct {
	Token  token.Token // token.FUNGSI
	Name   *Identifier
	Params []*Param
	Body   *BlockStatement
	// ReturnType berisi nama tipe kembalian bila ditulis. Kosong bila
	// fungsi tidak punya anotasi tipe kembalian.
	ReturnType string
}

func (fs *FunctionStatement) statementNode()       {}
func (fs *FunctionStatement) TokenLiteral() string { return fs.Token.Literal }
func (fs *FunctionStatement) String() string {
	var out bytes.Buffer
	out.WriteString("fungsi ")
	out.WriteString(fs.Name.String())
	out.WriteString("(")
	params := []string{}
	for _, p := range fs.Params {
		params = append(params, p.String())
	}
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(")")
	if fs.ReturnType != "" {
		out.WriteString(": " + fs.ReturnType)
	}
	out.WriteString(" ")
	out.WriteString(fs.Body.String())
	return out.String()
}

// Param adalah satu parameter fungsi beserta anotasi tipenya bila ada.
type Param struct {
	Name     *Identifier
	TypeHint string
}

func (p *Param) String() string {
	if p.TypeHint == "" {
		return p.Name.String()
	}
	return p.Name.String() + ": " + p.TypeHint
}

type CallStatement struct {
	Token    token.Token // token.PANGGIL
	CallExpr *CallExpression
}

func (cs *CallStatement) statementNode()       {}
func (cs *CallStatement) TokenLiteral() string { return cs.Token.Literal }
func (cs *CallStatement) String() string {
	return cs.TokenLiteral() + " " + cs.CallExpr.String()
}

type ReturnStatement struct {
	Token       token.Token // token.KEMBALI
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer
	out.WriteString(rs.TokenLiteral())
	if rs.ReturnValue != nil {
		out.WriteString(" ")
		out.WriteString(rs.ReturnValue.String())
	}
	return out.String()
}

type BreakStatement struct {
	Token token.Token
}

func (bs *BreakStatement) statementNode()       {}
func (bs *BreakStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BreakStatement) String() string       { return "berhenti" }

type ContinueStatement struct {
	Token token.Token
}

func (cs *ContinueStatement) statementNode()       {}
func (cs *ContinueStatement) TokenLiteral() string { return cs.Token.Literal }
func (cs *ContinueStatement) String() string       { return "lanjutkan" }

// --- Expressions ---

type Identifier struct {
	Token token.Token // token.IDENT
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

type FloatLiteral struct {
	Token token.Token
	Value float64
}

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FloatLiteral) String() string       { return fl.Token.Literal }

type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return fmt.Sprintf("\"%s\"", sl.Value) }

type BooleanLiteral struct {
	Token token.Token
	Value bool
}

func (bl *BooleanLiteral) expressionNode()      {}
func (bl *BooleanLiteral) TokenLiteral() string { return bl.Token.Literal }
func (bl *BooleanLiteral) String() string {
	if bl.Value {
		return "benar"
	}
	return "salah"
}

type NilLiteral struct {
	Token token.Token
}

func (nl *NilLiteral) expressionNode()      {}
func (nl *NilLiteral) TokenLiteral() string { return nl.Token.Literal }
func (nl *NilLiteral) String() string       { return "kosong" }

// Prefix expression: !expr or -expr
type PrefixExpression struct {
	Token    token.Token // token.TIDAK or token.MINUS
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")
	return out.String()
}

// Infix expression: left OPERATOR right
type InfixExpression struct {
	Token    token.Token
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" " + ie.Operator + " ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")
	return out.String()
}

// CallExpression: panggil fn(args)
type CallExpression struct {
	Token     token.Token // token.PANGGIL
	Function  Expression  // Identifier
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var out bytes.Buffer
	args := []string{}
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}
	out.WriteString(ce.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")
	return out.String()
}

// ListLiteral: [1, 2, 3]
type ListLiteral struct {
	Token    token.Token // token.LBRACKET
	Elements []Expression
}

func (ll *ListLiteral) expressionNode()      {}
func (ll *ListLiteral) TokenLiteral() string { return ll.Token.Literal }
func (ll *ListLiteral) String() string {
	elems := make([]string, 0, len(ll.Elements))
	for _, e := range ll.Elements {
		elems = append(elems, e.String())
	}
	return "[" + strings.Join(elems, ", ") + "]"
}

// IndexExpression: daftar[0]
type IndexExpression struct {
	Token token.Token // token.LBRACKET
	Left  Expression
	Index Expression
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IndexExpression) String() string {
	return "(" + ie.Left.String() + "[" + ie.Index.String() + "])"
}

// PetaEntri adalah satu pasangan kunci-nilai pada literal peta.
type PetaEntri struct {
	Kunci Expression
	Nilai Expression
}

// PetaLiteral: peta { "a": 1, "b": 2 }
type PetaLiteral struct {
	Token token.Token // token.PETA
	Entri []PetaEntri
}

func (pl *PetaLiteral) expressionNode()      {}
func (pl *PetaLiteral) TokenLiteral() string { return pl.Token.Literal }
func (pl *PetaLiteral) String() string {
	parts := make([]string, 0, len(pl.Entri))
	for _, entri := range pl.Entri {
		parts = append(parts, entri.Kunci.String()+": "+entri.Nilai.String())
	}
	return "peta {" + strings.Join(parts, ", ") + "}"
}

// AssignExpression: target = value atau target += value
type AssignExpression struct {
	Token    token.Token // token.ASSIGN atau operator gabungan
	Target   Expression
	Operator string
	Value    Expression
}

func (ae *AssignExpression) expressionNode()      {}
func (ae *AssignExpression) TokenLiteral() string { return ae.Token.Literal }
func (ae *AssignExpression) String() string {
	return ae.Target.String() + " " + ae.Operator + " " + ae.Value.String()
}

// UntukStatement: untuk i dari 1 sampai 10 langkah 2 { ... }
type UntukStatement struct {
	Token token.Token // token.UNTUK
	Name  *Identifier
	Start Expression
	End   Expression
	Step  Expression // opsional, default 1
	Body  *BlockStatement
}

func (us *UntukStatement) statementNode()       {}
func (us *UntukStatement) TokenLiteral() string { return us.Token.Literal }
func (us *UntukStatement) String() string {
	var out bytes.Buffer
	out.WriteString("untuk ")
	out.WriteString(us.Name.String())
	out.WriteString(" dari ")
	out.WriteString(us.Start.String())
	out.WriteString(" sampai ")
	out.WriteString(us.End.String())
	if us.Step != nil {
		out.WriteString(" langkah ")
		out.WriteString(us.Step.String())
	}
	out.WriteString(" ")
	out.WriteString(us.Body.String())
	return out.String()
}

// CobaStatement: coba { ... } tangkap [nama] { ... }
type CobaStatement struct {
	Token   token.Token // token.COBA
	Try     *BlockStatement
	ErrName *Identifier // opsional, menerima objek galat
	Catch   *BlockStatement
}

func (cs *CobaStatement) statementNode()       {}
func (cs *CobaStatement) TokenLiteral() string { return cs.Token.Literal }
func (cs *CobaStatement) String() string {
	out := "coba " + cs.Try.String() + " tangkap"
	if cs.ErrName != nil {
		out += " " + cs.ErrName.String()
	}
	return out + " " + cs.Catch.String()
}

// LemparkanStatement: lemparkan nilai
type LemparkanStatement struct {
	Token token.Token // token.LEMPAR
	Value Expression
}

// ImporStatement: impor "modul.nusa"
//
// Path adalah lokasi berkas modul relatif terhadap berkas pemanggil.
type ImporStatement struct {
	Token token.Token // token.IMPOR
	Path  string
}

func (is *ImporStatement) statementNode()       {}
func (is *ImporStatement) TokenLiteral() string { return is.Token.Literal }
func (is *ImporStatement) String() string {
	return "impor " + strconv.Quote(is.Path)
}

func (ls *LemparkanStatement) statementNode()       {}
func (ls *LemparkanStatement) TokenLiteral() string { return ls.Token.Literal }
func (ls *LemparkanStatement) String() string {
	return "lemparkan " + ls.Value.String()
}

// UjiStatement: uji "nama tes" { ... }
//
// Tubuh berisi pernyataan cek yang dijalankan oleh penjalur uji.
type UjiStatement struct {
	Token token.Token // token.UJI
	Nama  string
	Tubuh *BlockStatement
}

func (us *UjiStatement) statementNode()       {}
func (us *UjiStatement) TokenLiteral() string { return us.Token.Literal }
func (us *UjiStatement) String() string {
	return "uji " + strconv.Quote(us.Nama) + " " + us.Tubuh.String()
}

// CekStatement: cek ekspresi [, "pesan"]
//
// Pernyataan ini memeriksa bahwa ekspresi bernilai benar. Bila ada pesan
// khusus, pesan itu dipakai sebagai penjelasan kegagalan.
type CekStatement struct {
	Token    token.Token // token.CEK
	Ekspresi Expression
	Pesan    string
}

func (cs *CekStatement) statementNode()       {}
func (cs *CekStatement) TokenLiteral() string { return cs.Token.Literal }
func (cs *CekStatement) String() string {
	out := "cek " + cs.Ekspresi.String()
	if cs.Pesan != "" {
		out += ", " + strconv.Quote(cs.Pesan)
	}
	return out
}

// KasusBlok adalah satu cabang di dalam pernyataan pilih.
type KasusBlok struct {
	Token token.Token // token.KASUS
	Nilai Expression
	Body  *BlockStatement
}

func (kb *KasusBlok) String() string {
	return "kasus " + kb.Nilai.String() + " " + kb.Body.String()
}

// PilihStatement: pilih nilai { kasus 1 { ... } bawaan { ... } }
type PilihStatement struct {
	Token  token.Token // token.PILIH
	Subjek Expression
	Kasus  []KasusBlok
	Bawaan *BlockStatement // opsional
}

func (ps *PilihStatement) statementNode()       {}
func (ps *PilihStatement) TokenLiteral() string { return ps.Token.Literal }
func (ps *PilihStatement) String() string {
	var out bytes.Buffer
	out.WriteString("pilih ")
	out.WriteString(ps.Subjek.String())
	out.WriteString(" {")
	for _, kasus := range ps.Kasus {
		out.WriteString("\n")
		out.WriteString(kasus.String())
	}
	if ps.Bawaan != nil {
		out.WriteString("\nbawaan ")
		out.WriteString(ps.Bawaan.String())
	}
	out.WriteString("\n}")
	return out.String()
}

// UntukDalamStatement: untuk x dalam daftar { ... }
// Bila ada dua nama, keduanya menerima kunci/index dan nilai.
type UntukDalamStatement struct {
	Token    token.Token // token.UNTUK
	Names    []*Identifier
	Iterable Expression
	Body     *BlockStatement
}

func (ud *UntukDalamStatement) statementNode()       {}
func (ud *UntukDalamStatement) TokenLiteral() string { return ud.Token.Literal }
func (ud *UntukDalamStatement) String() string {
	nama := make([]string, len(ud.Names))
	for i, n := range ud.Names {
		nama[i] = n.String()
	}
	return "untuk " + strings.Join(nama, ", ") + " dalam " + ud.Iterable.String() + " " + ud.Body.String()
}
