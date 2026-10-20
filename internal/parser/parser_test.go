package parser_test

import (
	"testing"

	"github.com/seraphimhub/Nusa/internal/ast"
	"github.com/seraphimhub/Nusa/internal/lexer"
	"github.com/seraphimhub/Nusa/internal/parser"
)

func parseProgram(t *testing.T, input string) (*ast.Program, []string) {
	t.Helper()
	tokens := lexer.Tokenize(input)
	p := parser.New(tokens)
	program := p.ParseProgram()
	return program, p.Errors()
}

func mustParse(t *testing.T, input string) *ast.Program {
	t.Helper()
	program, errors := parseProgram(t, input)
	if len(errors) > 0 {
		t.Fatalf("galat parser tidak terduga: %v", errors)
	}
	return program
}

func TestParseLetAndAssignments(t *testing.T) {
	program := mustParse(t, `
buat x = 1
x = 2
x += 3
`)

	if len(program.Statements) != 3 {
		t.Fatalf("jumlah statement = %d, ingin 3", len(program.Statements))
	}

	let, ok := program.Statements[0].(*ast.LetStatement)
	if !ok {
		t.Fatalf("statement 0 bukan LetStatement: %T", program.Statements[0])
	}
	if let.Name.Value != "x" {
		t.Errorf("nama = %q, ingin x", let.Name.Value)
	}

	for i, operator := range []string{"=", "+="} {
		stmt, ok := program.Statements[i+1].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("statement %d bukan ExpressionStatement: %T", i+1, program.Statements[i+1])
		}
		assign, ok := stmt.Expression.(*ast.AssignExpression)
		if !ok {
			t.Fatalf("statement %d bukan AssignExpression: %T", i+1, stmt.Expression)
		}
		if assign.Operator != operator {
			t.Errorf("operator = %q, ingin %q", assign.Operator, operator)
		}
		ident, ok := assign.Target.(*ast.Identifier)
		if !ok || ident.Value != "x" {
			t.Errorf("target bukan identifier x: %#v", assign.Target)
		}
	}
}

func TestParseTetap(t *testing.T) {
	program := mustParse(t, `
tetap PI = 3.14
buat x = 1
`)

	tetap, ok := program.Statements[0].(*ast.LetStatement)
	if !ok {
		t.Fatalf("statement 0 bukan LetStatement: %T", program.Statements[0])
	}
	if !tetap.Konst {
		t.Error("Konst = false, ingin true")
	}
	if tetap.TokenLiteral() != "tetap" {
		t.Errorf("TokenLiteral = %q, ingin tetap", tetap.TokenLiteral())
	}

	buat := program.Statements[1].(*ast.LetStatement)
	if buat.Konst {
		t.Error("deklarasi 'buat' tidak boleh bertanda konstanta")
	}
}

func TestParseListAndIndexAssignment(t *testing.T) {
	program := mustParse(t, `
buat buah = ["apel", "jeruk", 3]
buah[1] = "pisang"
`)

	let := program.Statements[0].(*ast.LetStatement)
	list, ok := let.Value.(*ast.ListLiteral)
	if !ok {
		t.Fatalf("nilai bukan ListLiteral: %T", let.Value)
	}
	if len(list.Elements) != 3 {
		t.Errorf("jumlah elemen = %d, ingin 3", len(list.Elements))
	}

	stmt := program.Statements[1].(*ast.ExpressionStatement)
	assign, ok := stmt.Expression.(*ast.AssignExpression)
	if !ok {
		t.Fatalf("bukan AssignExpression: %T", stmt.Expression)
	}
	index, ok := assign.Target.(*ast.IndexExpression)
	if !ok {
		t.Fatalf("target bukan IndexExpression: %T", assign.Target)
	}
	if index.Left.String() != "buah" {
		t.Errorf("left = %q, ingin buah", index.Left.String())
	}
	if index.Index.String() != "1" {
		t.Errorf("index = %q, ingin 1", index.Index.String())
	}
}

func TestParseUntukNumeric(t *testing.T) {
	program := mustParse(t, `
untuk i dari 1 sampai 10 langkah 2 {
    tulis i
}
`)

	forStmt, ok := program.Statements[0].(*ast.UntukStatement)
	if !ok {
		t.Fatalf("bukan UntukStatement: %T", program.Statements[0])
	}
	if forStmt.Name.Value != "i" {
		t.Errorf("nama = %q, ingin i", forStmt.Name.Value)
	}
	if forStmt.Start.String() != "1" || forStmt.End.String() != "10" || forStmt.Step.String() != "2" {
		t.Errorf("batas loop salah: %s %s %s", forStmt.Start, forStmt.End, forStmt.Step)
	}
	if len(forStmt.Body.Statements) != 1 {
		t.Errorf("jumlah statement body = %d, ingin 1", len(forStmt.Body.Statements))
	}
}

func TestParseUntukDalam(t *testing.T) {
	program := mustParse(t, `
untuk x dalam [1, 2, 3] {
    tulis x
}
`)

	forStmt, ok := program.Statements[0].(*ast.UntukDalamStatement)
	if !ok {
		t.Fatalf("bukan UntukDalamStatement: %T", program.Statements[0])
	}
	if len(forStmt.Names) != 1 || forStmt.Names[0].Value != "x" {
		t.Errorf("nama = %#v, ingin [x]", forStmt.Names)
	}
	if _, ok := forStmt.Iterable.(*ast.ListLiteral); !ok {
		t.Errorf("iterable bukan ListLiteral: %T", forStmt.Iterable)
	}
}

func TestParsePetaLiteral(t *testing.T) {
	program := mustParse(t, `
buat siswa = peta { "nama": "Budi", "umur": 17 }
buat petaKosong = peta {}
`)

	peta, ok := program.Statements[0].(*ast.LetStatement).Value.(*ast.PetaLiteral)
	if !ok {
		t.Fatalf("nilai bukan PetaLiteral: %T", program.Statements[0])
	}
	if len(peta.Entri) != 2 {
		t.Fatalf("jumlah entri = %d, ingin 2", len(peta.Entri))
	}
	if peta.Entri[0].Kunci.String() != `"nama"` || peta.Entri[0].Nilai.String() != `"Budi"` {
		t.Errorf("entri pertama salah: %s: %s", peta.Entri[0].Kunci, peta.Entri[0].Nilai)
	}

	petaKosong, ok := program.Statements[1].(*ast.LetStatement).Value.(*ast.PetaLiteral)
	if !ok {
		t.Fatalf("nilai bukan PetaLiteral: %T", program.Statements[1])
	}
	if len(petaKosong.Entri) != 0 {
		t.Errorf("peta kosong punya %d entri", len(petaKosong.Entri))
	}
}

func TestParsePetaBersarangDanKomaAkhir(t *testing.T) {
	program := mustParse(t, `
buat data = peta {
    "a": 1,
    "b": peta { "c": 2 },
}
`)

	peta := program.Statements[0].(*ast.LetStatement).Value.(*ast.PetaLiteral)
	if len(peta.Entri) != 2 {
		t.Fatalf("jumlah entri = %d, ingin 2", len(peta.Entri))
	}
	if _, ok := peta.Entri[1].Nilai.(*ast.PetaLiteral); !ok {
		t.Errorf("entri kedua bukan peta bersarang: %T", peta.Entri[1].Nilai)
	}
}

func TestParseUntukDuaVariabel(t *testing.T) {
	program := mustParse(t, `
untuk k, v dalam peta { "a": 1 } {
    tulis k
}
`)

	forStmt, ok := program.Statements[0].(*ast.UntukDalamStatement)
	if !ok {
		t.Fatalf("bukan UntukDalamStatement: %T", program.Statements[0])
	}
	if len(forStmt.Names) != 2 || forStmt.Names[0].Value != "k" || forStmt.Names[1].Value != "v" {
		t.Errorf("nama = %#v, ingin [k v]", forStmt.Names)
	}
}

func TestParseCobaTangkap(t *testing.T) {
	program := mustParse(t, `
coba {
    tulis 1 / 0
} tangkap e {
    tulis e["pesan"]
}

coba {
    tulis "aman"
} tangkap {
    tulis "galat"
}
`)

	pertama, ok := program.Statements[0].(*ast.CobaStatement)
	if !ok {
		t.Fatalf("statement 0 bukan CobaStatement: %T", program.Statements[0])
	}
	if pertama.ErrName == nil || pertama.ErrName.Value != "e" {
		t.Errorf("nama galat = %#v, ingin e", pertama.ErrName)
	}
	if len(pertama.Try.Statements) != 1 || len(pertama.Catch.Statements) != 1 {
		t.Errorf("jumlah statement coba=%d tangkap=%d", len(pertama.Try.Statements), len(pertama.Catch.Statements))
	}

	kedua := program.Statements[1].(*ast.CobaStatement)
	if kedua.ErrName != nil {
		t.Errorf("tangkap tanpa nama seharusnya ErrName nil, dapat %#v", kedua.ErrName)
	}
}

func TestParseLemparkan(t *testing.T) {
	program := mustParse(t, `
lemparkan "gagal"
`)

	lemparkan, ok := program.Statements[0].(*ast.LemparkanStatement)
	if !ok {
		t.Fatalf("bukan LemparkanStatement: %T", program.Statements[0])
	}
	if lemparkan.Value.String() != `"gagal"` {
		t.Errorf("nilai = %q", lemparkan.Value.String())
	}
}

func TestParseGalatBentukBaru(t *testing.T) {
	tests := []string{
		`peta "a": 1 }`,
		`peta { "a" 1 }`,
		`untuk i, j dari 1 sampai 3 { tulis i }`,
		`untuk i, dalam [1] { tulis i }`,
		`coba { tulis 1 }`,
		`coba { tulis 1 } tangkap 1 { tulis 2 }`,
		`coba x { tulis 1 } tangkap { tulis 2 }`,
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, errors := parseProgram(t, input)
			if len(errors) == 0 {
				t.Fatalf("ingin galat parser untuk %q", input)
			}
		})
	}
}

func TestParsePilih(t *testing.T) {
	program := mustParse(t, `
pilih nilai {
    kasus 1 {
        tulis "satu"
    }
    kasus 2 { tulis "dua" }
    bawaan {
        tulis "lain"
    }
}
`)

	pilih, ok := program.Statements[0].(*ast.PilihStatement)
	if !ok {
		t.Fatalf("bukan PilihStatement: %T", program.Statements[0])
	}
	if pilih.Subjek.String() != "nilai" {
		t.Errorf("subjek = %q, ingin nilai", pilih.Subjek.String())
	}
	if len(pilih.Kasus) != 2 {
		t.Fatalf("jumlah kasus = %d, ingin 2", len(pilih.Kasus))
	}
	if pilih.Kasus[0].Nilai.String() != "1" || pilih.Kasus[1].Nilai.String() != "2" {
		t.Errorf("nilai kasus salah: %s, %s", pilih.Kasus[0].Nilai, pilih.Kasus[1].Nilai)
	}
	if len(pilih.Kasus[0].Body.Statements) != 1 {
		t.Errorf("body kasus pertama = %d statement", len(pilih.Kasus[0].Body.Statements))
	}
	if pilih.Bawaan == nil {
		t.Errorf("bawaan nil")
	}
}

func TestParsePilihErrors(t *testing.T) {
	tests := []string{
		`pilih { kasus 1 { tulis 1 } }`,
		`pilih 1 { tulis 1 }`,
		`pilih 1 { kasus 1 tulis 1 }`,
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, errors := parseProgram(t, input)
			if len(errors) == 0 {
				t.Fatalf("ingin galat parser untuk %q", input)
			}
		})
	}
}

func TestParseFunctionAndCalls(t *testing.T) {
	program := mustParse(t, `
fungsi tambah(a, b) {
    kembali a + b
}
buat hasil = tambah(1, 2)
panggil tambah(3, 4)
`)

	fn, ok := program.Statements[0].(*ast.FunctionStatement)
	if !ok {
		t.Fatalf("statement 0 bukan FunctionStatement: %T", program.Statements[0])
	}
	if fn.Name.Value != "tambah" || len(fn.Params) != 2 {
		t.Errorf("fungsi = %s dengan %d parameter", fn.Name.Value, len(fn.Params))
	}

	let := program.Statements[1].(*ast.LetStatement)
	call, ok := let.Value.(*ast.CallExpression)
	if !ok {
		t.Fatalf("nilai bukan CallExpression: %T", let.Value)
	}
	if len(call.Arguments) != 2 {
		t.Errorf("jumlah argumen = %d, ingin 2", len(call.Arguments))
	}

	callStmt, ok := program.Statements[2].(*ast.CallStatement)
	if !ok {
		t.Fatalf("statement 2 bukan CallStatement: %T", program.Statements[2])
	}
	if len(callStmt.CallExpr.Arguments) != 2 {
		t.Errorf("jumlah argumen panggil = %d, ingin 2", len(callStmt.CallExpr.Arguments))
	}
}

func TestParseIfElseIf(t *testing.T) {
	program := mustParse(t, `
jika nilai >= 90 {
    tulis "A"
}
kalau_tidak jika nilai >= 80 {
    tulis "B"
}
kalau_tidak {
    tulis "C"
}
`)

	ifStmt, ok := program.Statements[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("bukan IfStatement: %T", program.Statements[0])
	}
	if ifStmt.Alternative == nil {
		t.Fatalf("alternative nil")
	}
	if len(ifStmt.Alternative.Statements) != 1 {
		t.Fatalf("alternative statements = %d", len(ifStmt.Alternative.Statements))
	}
	if _, ok := ifStmt.Alternative.Statements[0].(*ast.IfStatement); !ok {
		t.Errorf("else-if bukan IfStatement: %T", ifStmt.Alternative.Statements[0])
	}
}

func TestParsePrecedence(t *testing.T) {
	program := mustParse(t, `buat x = 1 + 2 * 3`)

	let := program.Statements[0].(*ast.LetStatement)
	if got, want := let.Value.String(), "(1 + (2 * 3))"; got != want {
		t.Errorf("precedence = %q, ingin %q", got, want)
	}
}

func TestParseErrors(t *testing.T) {
	tests := []string{
		"buat = 5",
		"untuk i 1 sampai 2 { tulis i }",
		"jika { tulis \"x\" }",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			program, errors := parseProgram(t, input)
			if len(errors) == 0 {
				t.Fatalf("ingin galat parser, program: %#v", program)
			}
		})
	}
}

func TestParseAnotasiTipe(t *testing.T) {
	program := mustParse(t, `
buat nama: teks = "Budi"
tetap PI: desimal = 3.14
`)

	pertama := program.Statements[0].(*ast.LetStatement)
	if pertama.TypeHint != "teks" {
		t.Errorf("type hint = %q, ingin teks", pertama.TypeHint)
	}
	if kedua := program.Statements[1].(*ast.LetStatement); kedua.TypeHint != "desimal" || !kedua.Konst {
		t.Errorf("konstanta = %#v", kedua)
	}
}

func TestParseAnotasiTipeFungsi(t *testing.T) {
	program := mustParse(t, `
fungsi sapa(nama: teks, umur: angka): teks {
    kembali nama
}
`)

	fn := program.Statements[0].(*ast.FunctionStatement)
	if fn.ReturnType != "teks" {
		t.Errorf("tipe kembalian = %q, ingin teks", fn.ReturnType)
	}
	if len(fn.Params) != 2 {
		t.Fatalf("jumlah parameter = %d", len(fn.Params))
	}
	if fn.Params[0].TypeHint != "teks" || fn.Params[1].TypeHint != "angka" {
		t.Errorf("tipe parameter = %q, %q", fn.Params[0].TypeHint, fn.Params[1].TypeHint)
	}
}

func TestParseTipeTidakDikenal(t *testing.T) {
	for _, input := range []string{
		"buat x: tidak_ada = 1",
		"fungsi f(a: apa_ada) { kembali a }",
		"fungsi f(): tidak_ada { kembali 1 }",
	} {
		t.Run(input, func(t *testing.T) {
			_, errors := parseProgram(t, input)
			if len(errors) == 0 {
				t.Fatalf("ingin galat tipe tidak dikenal")
			}
		})
	}
}

func TestParsePemulihanGalatTidakMenelanPernyataanBerikutnya(t *testing.T) {
	program, errors := parseProgram(t, `
buat 1 = 2
tulis "masih terbaca"
`)

	if len(errors) == 0 {
		t.Fatalf("ingin galat parser")
	}
	if len(program.Statements) != 1 {
		t.Fatalf("jumlah pernyataan = %d, ingin 1", len(program.Statements))
	}
	if _, ok := program.Statements[0].(*ast.PrintStatement); !ok {
		t.Errorf("pernyataan bukan PrintStatement: %T", program.Statements[0])
	}
}
