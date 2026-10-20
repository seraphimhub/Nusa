package runtime

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/seraphimhub/Nusa/internal/ast"
	"github.com/seraphimhub/Nusa/internal/tipe"
	"github.com/seraphimhub/Nusa/internal/token"
)

// --- Object System ---

type ObjectType string

const (
	INTEGER_OBJ  ObjectType = "INTEGER"
	FLOAT_OBJ    ObjectType = "FLOAT"
	STRING_OBJ   ObjectType = "STRING"
	BOOLEAN_OBJ  ObjectType = "BOOLEAN"
	NIL_OBJ      ObjectType = "NIL"
	LIST_OBJ     ObjectType = "LIST"
	PETA_OBJ     ObjectType = "PETA"
	RETURN_OBJ   ObjectType = "RETURN"
	BREAK_OBJ    ObjectType = "BREAK"
	CONTINUE_OBJ ObjectType = "CONTINUE"
	FUNCTION_OBJ ObjectType = "FUNCTION"
	BUILTIN_OBJ  ObjectType = "BUILTIN"
	ERROR_OBJ    ObjectType = "ERROR"
	GALAT_OBJ    ObjectType = "GALAT"
)

type Object interface {
	Type() ObjectType
	Inspect() string
}

type Integer struct{ Value int64 }

func (i *Integer) Type() ObjectType { return INTEGER_OBJ }
func (i *Integer) Inspect() string  { return strconv.FormatInt(i.Value, 10) }

type Float struct{ Value float64 }

func (f *Float) Type() ObjectType { return FLOAT_OBJ }
func (f *Float) Inspect() string  { return strconv.FormatFloat(f.Value, 'f', -1, 64) }

type String struct{ Value string }

func (s *String) Type() ObjectType { return STRING_OBJ }
func (s *String) Inspect() string  { return s.Value }

type Boolean struct{ Value bool }

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) Inspect() string {
	if b.Value {
		return "benar"
	}
	return "salah"
}

type Nil struct{}

func (n *Nil) Type() ObjectType { return NIL_OBJ }
func (n *Nil) Inspect() string  { return "kosong" }

// List bersifat mutable: variabel menyimpan referensi ke daftar yang sama.
type List struct{ Elements []Object }

func (l *List) Type() ObjectType { return LIST_OBJ }
func (l *List) Inspect() string {
	parts := make([]string, len(l.Elements))
	for i, e := range l.Elements {
		parts[i] = repr(e)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// PetaEntri adalah satu pasangan kunci-nilai di dalam peta.
type PetaEntri struct {
	Kunci Object
	Nilai Object
}

// Peta adalah kamus kunci-nilai dengan urutan insersi yang stabil sehingga
// keluaran program selalu sama untuk program yang sama.
type Peta struct {
	entri map[string]PetaEntri
	urut  []string
}

func NewPeta() *Peta {
	return &Peta{entri: make(map[string]PetaEntri), urut: []string{}}
}

func (p *Peta) Type() ObjectType { return PETA_OBJ }

func (p *Peta) Inspect() string {
	parts := make([]string, 0, len(p.urut))
	for _, hash := range p.urut {
		entri := p.entri[hash]
		parts = append(parts, repr(entri.Kunci)+": "+repr(entri.Nilai))
	}
	return "peta{" + strings.Join(parts, ", ") + "}"
}

// Panjang mengembalikan jumlah entri.
func (p *Peta) Panjang() int { return len(p.urut) }

// Set menyimpan nilai untuk kunci, menambahkannya bila belum ada.
func (p *Peta) Set(kunci Object, nilai Object) *Error {
	hash, errObj := kunciHash(kunci)
	if errObj != nil {
		return errObj
	}
	if _, ada := p.entri[hash]; !ada {
		p.urut = append(p.urut, hash)
	}
	p.entri[hash] = PetaEntri{Kunci: kunci, Nilai: nilai}
	return nil
}

// Ambil mengembalikan nilai dan apakah kunci ditemukan.
func (p *Peta) Ambil(kunci Object) (Object, bool, *Error) {
	hash, errObj := kunciHash(kunci)
	if errObj != nil {
		return nil, false, errObj
	}
	entri, ada := p.entri[hash]
	if !ada {
		return nil, false, nil
	}
	return entri.Nilai, true, nil
}

// Hapus membuang satu entri, mengembalikan apakah kunci ditemukan.
func (p *Peta) Hapus(kunci Object) (bool, *Error) {
	hash, errObj := kunciHash(kunci)
	if errObj != nil {
		return false, errObj
	}
	if _, ada := p.entri[hash]; !ada {
		return false, nil
	}
	delete(p.entri, hash)
	for i, h := range p.urut {
		if h == hash {
			p.urut = append(p.urut[:i], p.urut[i+1:]...)
			break
		}
	}
	return true, nil
}

// Kunci mengembalikan seluruh kunci sesuai urutan insersi.
func (p *Peta) Kunci() []Object {
	keys := make([]Object, 0, len(p.urut))
	for _, hash := range p.urut {
		keys = append(keys, p.entri[hash].Kunci)
	}
	return keys
}

// Nilai mengembalikan seluruh nilai sesuai urutan insersi.
func (p *Peta) Nilai() []Object {
	values := make([]Object, 0, len(p.urut))
	for _, hash := range p.urut {
		values = append(values, p.entri[hash].Nilai)
	}
	return values
}

// kunciHash memvalidasi jenis kunci dan mengubahnya menjadi penanda unik.
func kunciHash(kunci Object) (string, *Error) {
	switch obj := kunci.(type) {
	case *String:
		return "teks\x00" + obj.Value, nil
	case *Integer:
		return "angka\x00" + strconv.FormatInt(obj.Value, 10), nil
	case *Float:
		return "desimal\x00" + strconv.FormatFloat(obj.Value, 'g', -1, 64), nil
	case *Boolean:
		return "logika\x00" + strconv.FormatBool(obj.Value), nil
	case *Nil:
		return "kosong\x00", nil
	default:
		return "", &Error{Message: fmt.Sprintf("kunci peta harus angka, desimal, teks, logika, atau kosong; mendapat %s", typeName(kunci))}
	}
}

// repr menambahkan tanda kutip pada teks di dalam daftar agar batas
// antar-elemen tetap jelas.
func repr(obj Object) string {
	if s, ok := obj.(*String); ok {
		return strconv.Quote(s.Value)
	}
	return obj.Inspect()
}

type ReturnValue struct{ Value Object }

func (rv *ReturnValue) Type() ObjectType { return RETURN_OBJ }
func (rv *ReturnValue) Inspect() string  { return rv.Value.Inspect() }

type BreakValue struct{}

func (bv *BreakValue) Type() ObjectType { return BREAK_OBJ }
func (bv *BreakValue) Inspect() string  { return "berhenti" }

type ContinueValue struct{}

func (cv *ContinueValue) Type() ObjectType { return CONTINUE_OBJ }
func (cv *ContinueValue) Inspect() string  { return "lanjutkan" }

type FunctionObject struct {
	Name   string
	Params []*ast.Param
	Body   *ast.BlockStatement
	Env    *Environment
	// ReturnType berisi nama tipe kembalian bila fungsi punya anotasi.
	ReturnType string
}

func (fo *FunctionObject) Type() ObjectType { return FUNCTION_OBJ }
func (fo *FunctionObject) Inspect() string {
	if fo.Name == "" {
		return fmt.Sprintf("fungsi(%d parameter)", len(fo.Params))
	}
	return fmt.Sprintf("fungsi %s(%d parameter)", fo.Name, len(fo.Params))
}

// Frame adalah satu pemanggilan fungsi pada jejak galat.
type Frame struct {
	Name string
	Line int
}

// Error membawa pesan, nomor baris, kolom, dan jejak pemanggilan fungsi
// sehingga galat bisa dilacak sampai ke pemanggilnya. Kolom dipakai untuk
// menunjuk letak yang tepat pada cuplikan sumber.
type Error struct {
	Message string
	Line    int
	Column  int
	Trace   []Frame
}

func (e *Error) Type() ObjectType { return ERROR_OBJ }

// Galat adalah nilai galat yang bisa dipegang program, misalnya variabel pada
// blok 'tangkap'. Tipe ini terpisah dari ERROR_OBJ agar tidak tertukar dengan
// galat yang sedang dipropagasikan runtime.
type GalatValue struct {
	Err *Error
}

func (g *GalatValue) Type() ObjectType { return GALAT_OBJ }
func (g *GalatValue) Inspect() string  { return g.Err.Inspect() }

// maxTraceFrames membatasi panjang jejak yang ditampilkan agar galat pada
// rekursi dalam tetap terbaca.
const maxTraceFrames = 10

func (e *Error) Inspect() string {
	var out strings.Builder
	out.WriteString("ERROR: ")
	if e.Line > 0 {
		fmt.Fprintf(&out, "baris %d: ", e.Line)
	}
	out.WriteString(e.Message)

	shown := e.Trace
	if len(shown) > maxTraceFrames {
		shown = shown[:maxTraceFrames]
	}
	for _, frame := range shown {
		fmt.Fprintf(&out, "\n  di %s (dipanggil pada baris %d)", frame.Name, frame.Line)
	}
	if len(e.Trace) > maxTraceFrames {
		fmt.Fprintf(&out, "\n  ... dan %d frame lainnya", len(e.Trace)-maxTraceFrames)
	}
	return out.String()
}

// --- Environment ---

type Environment struct {
	store  map[string]Object
	consts map[string]bool
	outer  *Environment
	// returnType dan namaFungsi mencatat fungsi yang sedang berjalan.
	// Keduanya dipakai untuk memeriksa nilai kembalian terhadap anotasi
	// tipe tepat di pernyataan 'kembali'.
	returnType string
	namaFungsi string
}

func NewEnvironment() *Environment {
	s := make(map[string]Object)
	return &Environment{store: s, consts: make(map[string]bool), outer: nil}
}

func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

func (e *Environment) Set(name string, val Object) Object {
	e.store[name] = val
	return val
}

// SetConst mendeklarasikan konstanta pada scope ini.
func (e *Environment) SetConst(name string, val Object) Object {
	e.store[name] = val
	e.consts[name] = true
	return val
}

// IsConstLocal melaporkan apakah nama dideklarasikan sebagai konstanta pada
// scope ini saja, sehingga deklarasi ulang bisa dicegah.
func (e *Environment) IsConstLocal(name string) bool {
	return e.consts[name]
}

// FindConst mencari konstanta pada scope ini maupun scope luar.
func (e *Environment) FindConst(name string) bool {
	if e.consts[name] {
		return true
	}
	if e.outer != nil {
		return e.outer.FindConst(name)
	}
	return false
}

// Assign memperbarui variabel di scope tempat ia dideklarasikan.
func (e *Environment) Assign(name string, val Object) bool {
	if _, ok := e.store[name]; ok {
		e.store[name] = val
		return true
	}
	if e.outer != nil {
		return e.outer.Assign(name, val)
	}
	return false
}

// --- Runtime (Evaluator) ---

// Limits membatasi sumber daya eksekusi skrip agar program yang salah tidak
// membuat interpreter kehabisan memori atau tumpukan.
type Limits struct {
	MaxDepth int   // kedalaman pemanggilan fungsi, 0 berarti tanpa batas
	MaxSteps int64 // jumlah node yang dievaluasi, 0 berarti tanpa batas
}

func DefaultLimits() Limits {
	return Limits{MaxDepth: 1000, MaxSteps: 5_000_000}
}

type Runtime struct {
	env      *Environment
	Out      io.Writer
	In       io.Reader
	reader   *bufio.Reader
	builtins map[string]*Builtin
	limits   Limits
	steps    int64
	stack    []Frame
	// pemuatModul dipakai untuk memuat modul pada pernyataan 'impor'.
	pemuatModul ModuleLoader
	// hasilCek mengumpulkan hasil dari pernyataan 'cek'.
	hasilCek []HasilCek
	// ujiSekarang adalah nama blok uji yang sedang berjalan.
	ujiSekarang string
	// dalamUji menandai bahwa eksekusi sedang berada di dalam blok
	// 'uji'. Di dalam blok uji, pemeriksaan yang gagal dicatat sebagai
	// kegagalan tanpa menghentikan program.
	dalamUji bool
}

// HasilCek menyimpan hasil satu pernyataan cek.
type HasilCek struct {
	// Uji adalah nama blok uji pemilik pemeriksaan. Kosong untuk
	// pemeriksaan yang ditulis di luar blok uji.
	Uji string
	// Lolos bernilai true bila ekspresi yang diperiksa bernilai benar.
	Lolos bool
	// Pesan menjelaskan kegagalan. Kosong bila pemeriksaan lulus.
	Pesan string
	// Ekspresi adalah ekspresi yang diperiksa dalam bentuk sumber.
	Ekspresi string
	// Line adalah baris tempat pemeriksaan ditulis.
	Line int
	// Column adalah kolom tempat pemeriksaan ditulis.
	Column int
}

func New() *Runtime {
	return NewWithIO(os.Stdout, os.Stdin)
}

func NewWithIO(out io.Writer, in io.Reader) *Runtime {
	if out == nil {
		out = io.Discard
	}
	return &Runtime{
		env:      NewEnvironment(),
		Out:      out,
		In:       in,
		builtins: newBuiltins(),
		limits:   DefaultLimits(),
	}
}

// SetLimits mengganti batas eksekusi runtime.
func (r *Runtime) SetLimits(limits Limits) {
	r.limits = limits
}

// Symbol adalah satu nama yang diekspor modul.
type Symbol struct {
	Name  string
	Value Object
}

// ModuleLoader memuat modul dari path dan mengembalikan simbol yang
// diekspor modul tersebut. Path relatif diselesaikan oleh pemuat.
type ModuleLoader func(path string, line int) ([]Symbol, error)

// SetModuleLoader memasang pemuat modul pada runtime. Tanpa pemuat,
// pernyataan 'impor' akan menghasilkan galat yang jelas.
func (r *Runtime) SetModuleLoader(loader ModuleLoader) {
	r.pemuatModul = loader
}

// Global mengembalikan nilai variabel tingkat atas dengan nama tersebut.
func (r *Runtime) Global(name string) (Object, bool) {
	obj, ok := r.env.Get(name)
	return obj, ok
}

// SetGlobal menyuntik nilai ke lingkup global runtime.
func (r *Runtime) SetGlobal(name string, obj Object) {
	env := r.env
	if env == nil {
		return
	}
	env.store[name] = obj
}

// Execute menjalankan program. Nilai kembalian berupa Object: galat
// dikembalikan sebagai *Error agar pemanggil bisa menentukan exit code.
// Anggaran langkah direset setiap pemanggilan agar REPL tetap responsif.
func (r *Runtime) Execute(p *ast.Program) Object {
	r.steps = 0
	r.stack = r.stack[:0]
	return r.eval(p, r.env)
}

func (r *Runtime) eval(node ast.Node, env *Environment) Object {
	r.steps++
	if r.limits.MaxSteps > 0 && r.steps > r.limits.MaxSteps {
		return &Error{Message: fmt.Sprintf("batas eksekusi terlampaui: lebih dari %d langkah; periksa kemungkinan perulangan tak berujung", r.limits.MaxSteps)}
	}

	if node == nil {
		return nil
	}

	switch n := node.(type) {

	// --- Program ---
	case *ast.Program:
		return r.evalProgram(n, env)

	// --- Statements ---
	case *ast.BlockStatement:
		return r.evalBlockStatement(n, env)

	case *ast.LetStatement:
		if env.IsConstLocal(n.Name.Value) {
			return newErrorTok(n.Token, "'%s' adalah konstanta dan tidak bisa dideklarasikan ulang", n.Name.Value)
		}

		val := r.eval(n.Value, env)
		if isError(val) {
			return val
		}
		if val == nil {
			return newErrorTok(n.Token, "ekspresi tidak menghasilkan nilai")
		}

		if n.TypeHint != "" && !cocokTipe(n.TypeHint, val) {
			return newErrorTok(ast.TokenOf(n.Value),
				"nilai untuk '%s' harus bertipe %s, mendapat %s",
				n.Name.Value, n.TypeHint, typeName(val))
		}

		if n.Konst {
			env.SetConst(n.Name.Value, val)
		} else {
			env.Set(n.Name.Value, val)
		}

	case *ast.PrintStatement:
		val := r.eval(n.Value, env)
		if isError(val) {
			return val
		}
		if val == nil {
			return newErrorTok(n.Token, "tidak ada nilai untuk ditulis")
		}
		fmt.Fprintln(r.Out, val.Inspect())

	case *ast.IfStatement:
		return r.evalIfExpression(n, env)

	case *ast.ForStatement:
		return r.evalForStatement(n, env)

	case *ast.UntukStatement:
		return r.evalUntukStatement(n, env)

	case *ast.UntukDalamStatement:
		return r.evalUntukDalamStatement(n, env)

	case *ast.PilihStatement:
		return r.evalPilihStatement(n, env)

	case *ast.CobaStatement:
		return r.evalCobaStatement(n, env)

	case *ast.ImporStatement:
		return r.evalImporStatement(n, env)

	case *ast.UjiStatement:
		return r.evalUjiStatement(n, env)

	case *ast.CekStatement:
		return r.evalCekStatement(n, env)

	case *ast.LemparkanStatement:
		value := r.eval(n.Value, env)
		if isError(value) {
			return value
		}
		if value == nil {
			return newErrorTok(n.Token, "lemparkan membutuhkan nilai galat")
		}
		switch errValue := value.(type) {
		case *GalatValue:
			// Galat yang ditangkap bisa dilemparkan ulang apa adanya.
			return errValue.Err
		case *Error:
			return errValue
		}
		return newErrorTok(n.Token, "%s", value.Inspect())

	case *ast.WhileStatement:
		return r.evalWhileStatement(n, env)

	case *ast.FunctionStatement:
		fn := &FunctionObject{
			Name:       n.Name.Value,
			Params:     n.Params,
			Body:       n.Body,
			Env:        env,
			ReturnType: n.ReturnType,
		}
		env.Set(n.Name.Value, fn)

	case *ast.CallStatement:
		result := r.evalCallExpression(n.CallExpr, env)
		if isError(result) {
			return result
		}

	case *ast.ReturnStatement:
		val := r.eval(n.ReturnValue, env)
		if isError(val) {
			return val
		}
		// Anotasi tipe kembalian diperiksa tepat di pernyataan 'kembali'
		// supaya galat menunjuk ke sumber yang salah.
		if env.returnType != "" && !cocokTipe(env.returnType, val) {
			return newErrorTok(n.Token, "fungsi '%s' harus mengembalikan nilai bertipe %s, mendapat %s",
				env.namaFungsi, env.returnType, typeName(val))
		}
		return &ReturnValue{Value: val}

	case *ast.BreakStatement:
		return &BreakValue{}

	case *ast.ContinueStatement:
		return &ContinueValue{}

	case *ast.InputStatement:
		if n.Name == nil {
			return newErrorTok(n.Token, "membaca membutuhkan nama variabel")
		}
		env.Set(n.Name.Value, &String{Value: r.readInput()})

	// --- Expressions ---
	case *ast.AssignExpression:
		return r.evalAssign(n, env)

	case *ast.InfixExpression:
		return r.evalInfix(n, env)

	case *ast.IntegerLiteral:
		return &Integer{Value: n.Value}

	case *ast.FloatLiteral:
		return &Float{Value: n.Value}

	case *ast.StringLiteral:
		return &String{Value: n.Value}

	case *ast.BooleanLiteral:
		return &Boolean{Value: n.Value}

	case *ast.NilLiteral:
		return &Nil{}

	case *ast.Identifier:
		return r.evalIdentifier(n, env)

	case *ast.PrefixExpression:
		right := r.eval(n.Right, env)
		if isError(right) {
			return right
		}
		return r.evalPrefixExpression(n.Operator, right, n.Token)

	case *ast.CallExpression:
		return r.evalCallExpression(n, env)

	case *ast.PetaLiteral:
		peta := NewPeta()
		for _, entri := range n.Entri {
			kunci := r.eval(entri.Kunci, env)
			if isError(kunci) {
				return kunci
			}
			nilai := r.eval(entri.Nilai, env)
			if isError(nilai) {
				return nilai
			}
			if errObj := peta.Set(kunci, nilai); errObj != nil {
				errObj.Line = n.Token.Line
				return errObj
			}
		}
		return peta

	case *ast.ListLiteral:
		elements := make([]Object, 0, len(n.Elements))
		for _, e := range n.Elements {
			val := r.eval(e, env)
			if isError(val) {
				return val
			}
			elements = append(elements, val)
		}
		return &List{Elements: elements}

	case *ast.IndexExpression:
		return r.evalIndexExpression(n, env)

	case *ast.ExpressionStatement:
		return r.eval(n.Expression, env)
	}

	return nil
}

func (r *Runtime) evalProgram(program *ast.Program, env *Environment) Object {
	var result Object

	for _, statement := range program.Statements {
		result = r.eval(statement, env)

		if errObj, ok := result.(*Error); ok {
			attachPosition(errObj, statement)
		}

		switch result := result.(type) {
		case *ReturnValue:
			return result.Value
		case *BreakValue:
			return result
		case *ContinueValue:
			return result
		case *Error:
			return result
		}
	}

	return result
}

func (r *Runtime) evalBlockStatement(block *ast.BlockStatement, env *Environment) Object {
	var result Object

	for _, statement := range block.Statements {
		result = r.eval(statement, env)

		if errObj, ok := result.(*Error); ok {
			attachPosition(errObj, statement)
		}

		if result != nil {
			rt := result.Type()
			if rt == RETURN_OBJ || rt == BREAK_OBJ || rt == CONTINUE_OBJ || rt == ERROR_OBJ {
				return result
			}
		}
	}

	return result
}

// --- If Expression ---

func (r *Runtime) evalIfExpression(ie *ast.IfStatement, env *Environment) Object {
	condition := r.eval(ie.Condition, env)
	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return r.eval(ie.Consequence, env)
	} else if ie.Alternative != nil {
		return r.eval(ie.Alternative, env)
	}

	return &Nil{}
}

// --- For Statement (ulang) ---

func (r *Runtime) evalForStatement(fs *ast.ForStatement, env *Environment) Object {
	countObj := r.eval(fs.Count, env)
	if isError(countObj) {
		return countObj
	}

	count, ok := countObj.(*Integer)
	if !ok {
		return newErrorTok(fs.Token, "ulang membutuhkan angka bulat, mendapat %s", typeName(countObj))
	}

	for i := int64(0); i < count.Value; i++ {
		result := r.eval(fs.Body, env)
		if stop, propagate := loopSignal(result); stop {
			return propagate
		}
	}

	return nil
}

// --- Untuk Statement (numeric) ---

func (r *Runtime) evalUntukStatement(us *ast.UntukStatement, env *Environment) Object {
	startObj := r.eval(us.Start, env)
	if isError(startObj) {
		return startObj
	}
	endObj := r.eval(us.End, env)
	if isError(endObj) {
		return endObj
	}

	stepObj := Object(&Integer{Value: 1})
	if us.Step != nil {
		stepObj = r.eval(us.Step, env)
		if isError(stepObj) {
			return stepObj
		}
	}

	// Jalur integer mempertahankan tipe angka bulat selama iterasi.
	if start, ok := startObj.(*Integer); ok {
		if end, ok := endObj.(*Integer); ok {
			if step, ok := stepObj.(*Integer); ok {
				if step.Value == 0 {
					return newErrorTok(us.Token, "langkah tidak boleh nol")
				}
				for i := start.Value; (step.Value > 0 && i <= end.Value) || (step.Value < 0 && i >= end.Value); i += step.Value {
					env.Set(us.Name.Value, &Integer{Value: i})
					result := r.eval(us.Body, env)
					if stop, propagate := loopSignal(result); stop {
						return propagate
					}
				}
				return nil
			}
		}
	}

	start, ok := toFloat64(startObj)
	if !ok {
		return newErrorTok(us.Token, "nilai awal harus berupa angka, mendapat %s", typeName(startObj))
	}
	end, ok := toFloat64(endObj)
	if !ok {
		return newErrorTok(us.Token, "nilai akhir harus berupa angka, mendapat %s", typeName(endObj))
	}
	step, ok := toFloat64(stepObj)
	if !ok {
		return newErrorTok(us.Token, "langkah harus berupa angka, mendapat %s", typeName(stepObj))
	}
	if step == 0 {
		return newErrorTok(us.Token, "langkah tidak boleh nol")
	}

	for i := start; (step > 0 && i <= end) || (step < 0 && i >= end); i += step {
		env.Set(us.Name.Value, &Float{Value: i})
		result := r.eval(us.Body, env)
		if stop, propagate := loopSignal(result); stop {
			return propagate
		}
	}

	return nil
}

// --- Untuk Dalam Statement ---

func (r *Runtime) evalUntukDalamStatement(ud *ast.UntukDalamStatement, env *Environment) Object {
	iterable := r.eval(ud.Iterable, env)
	if isError(iterable) {
		return iterable
	}

	// Satu nama menerima nilai untuk daftar/teks dan menerima kunci untuk
	// peta. Dua nama menerima kunci/index dan nilainya.
	run := func(satu, kunci, nilai Object) (bool, Object) {
		if len(ud.Names) > 1 {
			env.Set(ud.Names[0].Value, kunci)
			env.Set(ud.Names[1].Value, nilai)
		} else {
			env.Set(ud.Names[0].Value, satu)
		}
		result := r.eval(ud.Body, env)
		return loopSignal(result)
	}

	switch obj := iterable.(type) {
	case *List:
		for i, item := range obj.Elements {
			if stop, propagate := run(item, &Integer{Value: int64(i)}, item); stop {
				return propagate
			}
		}
	case *String:
		for i, ch := range []rune(obj.Value) {
			aksara := &String{Value: string(ch)}
			if stop, propagate := run(aksara, &Integer{Value: int64(i)}, aksara); stop {
				return propagate
			}
		}
	case *Peta:
		for _, kunci := range obj.Kunci() {
			nilai, _, errObj := obj.Ambil(kunci)
			if errObj != nil {
				errObj.Line = ud.Token.Line
				return errObj
			}
			if stop, propagate := run(kunci, kunci, nilai); stop {
				return propagate
			}
		}
	default:
		return newErrorTok(ud.Token, "untuk ... dalam hanya bisa mengulang daftar, teks, atau peta; mendapat %s", typeName(iterable))
	}

	return nil
}

// --- Coba / Tangkap ---

// evalCobaStatement menjalankan blok coba; bila hasilnya galat, blok tangkap
// dijalankan dengan objek galat tersedia bagi program.
func (r *Runtime) evalCobaStatement(cs *ast.CobaStatement, env *Environment) Object {
	result := r.eval(cs.Try, env)

	errObj, tertangkap := result.(*Error)
	if !tertangkap {
		// Termasuk return/break/continue yang harus tetap diteruskan.
		return result
	}

	if cs.ErrName != nil {
		env.Set(cs.ErrName.Value, &GalatValue{Err: errObj})
	}
	return r.eval(cs.Catch, env)
}

// --- Impor ---

// evalImporStatement memuat modul dan memasukkan simbol yang diekspor modul
// ke dalam lingkup saat ini.
func (r *Runtime) evalImporStatement(n *ast.ImporStatement, env *Environment) Object {
	if r.pemuatModul == nil {
		return newErrorTok(n.Token, "impor modul tidak tersedia pada runtime ini")
	}

	simbol, err := r.pemuatModul(n.Path, n.Token.Line)
	if err != nil {
		return newErrorTok(n.Token, "%s", err.Error())
	}

	for _, s := range simbol {
		if env.IsConstLocal(s.Name) {
			return newErrorTok(n.Token, "impor modul mendefinisikan '%s' yang sudah ada sebagai konstanta", s.Name)
		}
		env.Set(s.Name, s.Value)
	}

	return nil
}

// --- Uji / Cek ---

// evalUjiStatement menjalankan isi satu blok uji. Semua pemeriksaan di dalam
// blok dicatat; galat lain yang muncul selama blok berjalan juga dicatat
// sebagai kegagalan agar satu tes yang rusak tidak menghentikan tes lain.
func (r *Runtime) evalUjiStatement(n *ast.UjiStatement, env *Environment) Object {
	sebelumnya := r.ujiSekarang
	dalamUji := r.dalamUji

	r.ujiSekarang = n.Nama
	r.dalamUji = true

	result := r.eval(n.Tubuh, env)

	r.ujiSekarang = sebelumnya
	r.dalamUji = dalamUji

	if errObj, ok := result.(*Error); ok {
		r.catatGagalUji(n, errObj)
		return nil
	}

	return nil
}

// evalCekStatement memeriksa bahwa ekspresi bernilai benar. Pemeriksaan
// yang gagal dicatat pada runtime sehingga penjalur uji bisa melaporkannya.
func (r *Runtime) evalCekStatement(n *ast.CekStatement, env *Environment) Object {
	nilai := r.eval(n.Ekspresi, env)

	hasil := HasilCek{
		Uji:      r.ujiSekarang,
		Ekspresi: n.Ekspresi.String(),
		Line:     n.Token.Line,
		Column:   n.Token.Column,
	}

	if errObj, ok := nilai.(*Error); ok {
		hasil.Pesan = errObj.Message
		r.hasilCek = append(r.hasilCek, hasil)
		if r.dalamUji {
			return nil
		}
		return errObj
	}

	if !isTruthy(nilai) {
		hasil.Pesan = n.Pesan
		if hasil.Pesan == "" {
			hasil.Pesan = fmt.Sprintf("ekspresi '%s' harus bernilai benar", hasil.Ekspresi)
		}
		r.hasilCek = append(r.hasilCek, hasil)
		if r.dalamUji {
			return nil
		}
		return newErrorTok(n.Token, "%s", hasil.Pesan)
	}

	hasil.Lolos = true
	r.hasilCek = append(r.hasilCek, hasil)

	return nil
}

// catatGagalUji mencatat galat yang terjadi di luar pemeriksaan.
func (r *Runtime) catatGagalUji(n *ast.UjiStatement, errObj *Error) {
	r.hasilCek = append(r.hasilCek, HasilCek{
		Uji:      n.Nama,
		Pesan:    errObj.Message,
		Ekspresi: errObj.Message,
		Line:     errObj.Line,
		Column:   errObj.Column,
	})
}

// HasilCek mengembalikan hasil seluruh pemeriksaan yang sudah dijalankan.
func (r *Runtime) HasilCek() []HasilCek {
	return r.hasilCek
}

// --- Pilih Statement ---
// evalPilihStatement mencocokkan subjek dengan setiap kasus memakai
// kesetaraan nilai. Cabang pertama yang cocok langsung dijalankan tanpa
// fallthrough; bawaan dipakai bila tidak ada kasus yang cocok.
func (r *Runtime) evalPilihStatement(ps *ast.PilihStatement, env *Environment) Object {
	subjek := r.eval(ps.Subjek, env)
	if isError(subjek) {
		return subjek
	}

	for _, kasus := range ps.Kasus {
		nilai := r.eval(kasus.Nilai, env)
		if isError(nilai) {
			return nilai
		}
		if objectsEqual(subjek, nilai) {
			return r.eval(kasus.Body, env)
		}
	}

	if ps.Bawaan != nil {
		return r.eval(ps.Bawaan, env)
	}

	return nil
}

// loopSignal mengembalikan (berhenti, objek yang harus diteruskan).
func loopSignal(result Object) (bool, Object) {
	if result == nil {
		return false, nil
	}
	switch result.Type() {
	case BREAK_OBJ:
		return true, nil
	case RETURN_OBJ, ERROR_OBJ:
		return true, result
	}
	return false, nil
}

// --- While Statement ---

func (r *Runtime) evalWhileStatement(ws *ast.WhileStatement, env *Environment) Object {
	for {
		condition := r.eval(ws.Condition, env)
		if isError(condition) {
			return condition
		}

		if !isTruthy(condition) {
			break
		}

		result := r.eval(ws.Body, env)
		if stop, propagate := loopSignal(result); stop {
			return propagate
		}
	}

	return nil
}

// --- Identifier ---

func (r *Runtime) evalIdentifier(node *ast.Identifier, env *Environment) Object {
	val, ok := env.Get(node.Value)
	if ok {
		return val
	}
	if builtin, ok := r.builtins[node.Value]; ok {
		return builtin
	}
	return newErrorTok(node.Token, "variabel '%s' tidak ditemukan", node.Value)
}

// --- Assignment ---

func (r *Runtime) evalAssign(ae *ast.AssignExpression, env *Environment) Object {
	if ident, ok := ae.Target.(*ast.Identifier); ok {
		if env.FindConst(ident.Value) {
			return newErrorTok(ae.Token, "'%s' adalah konstanta dan tidak bisa diubah", ident.Value)
		}
	}

	// Untuk target indeks, kontainer dan kuncinya dievaluasi sekali saja
	// agar penugasan gabungan membaca nilai lama dengan benar.
	var container Object
	var indexObj Object
	var current Object

	if index, ok := ae.Target.(*ast.IndexExpression); ok {
		left := r.eval(index.Left, env)
		if isError(left) {
			return left
		}
		key := r.eval(index.Index, env)
		if isError(key) {
			return key
		}
		container, indexObj = left, key
	}

	if ae.Operator != "=" {
		if container != nil {
			value, errObj := readIndex(container, indexObj, indexToken(ae.Target))
			if errObj != nil {
				return errObj
			}
			current = value
		} else {
			value := r.eval(ae.Target, env)
			if isError(value) {
				return value
			}
			current = value
		}
	}

	value := r.eval(ae.Value, env)
	if isError(value) {
		return value
	}
	if value == nil {
		return newErrorTok(ae.Token, "ekspresi tidak menghasilkan nilai")
	}

	if ae.Operator != "=" {
		op := strings.TrimSuffix(ae.Operator, "=")
		value = r.evalInfixExpression(op, current, value, ae.Token)
		if isError(value) {
			return value
		}
	}

	switch target := ae.Target.(type) {
	case *ast.Identifier:
		if !env.Assign(target.Value, value) {
			env.Set(target.Value, value)
		}
		return value
	case *ast.IndexExpression:
		if errObj := writeIndex(container, indexObj, value, indexToken(ae.Target)); errObj != nil {
			return errObj
		}
		return value
	default:
		return newErrorTok(ae.Token, "target penugasan harus variabel, elemen daftar, atau kunci peta")
	}
}

// --- Index ---

func (r *Runtime) evalIndexExpression(node *ast.IndexExpression, env *Environment) Object {
	left := r.eval(node.Left, env)
	if isError(left) {
		return left
	}
	indexObj := r.eval(node.Index, env)
	if isError(indexObj) {
		return indexObj
	}

	value, errObj := readIndex(left, indexObj, node.Token)
	if errObj != nil {
		return errObj
	}
	return value
}

// readIndex mengambil elemen daftar/teks, entri peta, atau bidang galat.
func readIndex(container Object, indexObj Object, tok token.Token) (Object, *Error) {
	switch obj := container.(type) {
	case *List:
		index, err := resolveIndex(indexObj, len(obj.Elements))
		if err != nil {
			return nil, newErrorTok(tok, "%s", err.Error())
		}
		return obj.Elements[index], nil
	case *String:
		runes := []rune(obj.Value)
		index, err := resolveIndex(indexObj, len(runes))
		if err != nil {
			return nil, newErrorTok(tok, "%s", err.Error())
		}
		return &String{Value: string(runes[index])}, nil
	case *Peta:
		nilai, ada, errObj := obj.Ambil(indexObj)
		if errObj != nil {
			errObj.Line = tok.Line
			errObj.Column = tok.Column
			return nil, errObj
		}
		if !ada {
			return nil, newErrorTok(tok, "kunci %s tidak ada di peta; gunakan ada(peta, kunci) atau ambil(peta, kunci, bawaan)", repr(indexObj))
		}
		return nilai, nil
	case *GalatValue:
		key, ok := indexObj.(*String)
		if !ok {
			return nil, newErrorTok(tok, "galat hanya bisa diindeks dengan teks: \"pesan\", \"baris\", atau \"jejak\"")
		}
		switch key.Value {
		case "pesan":
			return &String{Value: obj.Err.Message}, nil
		case "baris":
			return &Integer{Value: int64(obj.Err.Line)}, nil
		case "jejak":
			frames := make([]Object, 0, len(obj.Err.Trace))
			for _, frame := range obj.Err.Trace {
				frames = append(frames, &String{Value: fmt.Sprintf("%s (baris %d)", frame.Name, frame.Line)})
			}
			return &List{Elements: frames}, nil
		default:
			return nil, newErrorTok(tok, "kunci galat '%s' tidak dikenal; pakai \"pesan\", \"baris\", atau \"jejak\"", key.Value)
		}
	default:
		return nil, newErrorTok(tok, "tipe %s tidak bisa diindeks", typeName(container))
	}
}

// writeIndex menulis elemen daftar atau entri peta.
func writeIndex(container Object, indexObj Object, value Object, tok token.Token) *Error {
	switch obj := container.(type) {
	case *List:
		index, err := resolveIndex(indexObj, len(obj.Elements))
		if err != nil {
			return newErrorTok(tok, "%s", err.Error())
		}
		obj.Elements[index] = value
		return nil
	case *Peta:
		if errObj := obj.Set(indexObj, value); errObj != nil {
			errObj.Line = tok.Line
			errObj.Column = tok.Column
			return errObj
		}
		return nil
	default:
		return newErrorTok(tok, "tipe %s tidak bisa diindeks", typeName(container))
	}
}

// resolveIndex mendukung indeks negatif dari akhir daftar/teks.
func resolveIndex(indexObj Object, length int) (int, error) {
	index, ok := indexObj.(*Integer)
	if !ok {
		return 0, fmt.Errorf("indeks harus angka bulat, mendapat %s", typeName(indexObj))
	}

	idx := int(index.Value)
	if idx < 0 {
		idx = length + idx
	}
	if idx < 0 || idx >= length {
		return 0, fmt.Errorf("indeks %d di luar batas (panjang %d)", index.Value, length)
	}
	return idx, nil
}

// --- Prefix Expression ---

func (r *Runtime) evalPrefixExpression(operator string, right Object, tok token.Token) Object {
	switch operator {
	case "!", "tidak":
		return r.evalBangOperatorExpression(right)
	case "-":
		return r.evalMinusPrefixOperatorExpression(right, tok)
	default:
		return newErrorTok(tok, "operator '%s' tidak dikenal", operator)
	}
}

func (r *Runtime) evalBangOperatorExpression(right Object) Object {
	switch obj := right.(type) {
	case *Boolean:
		return &Boolean{Value: !obj.Value}
	case *Nil:
		return &Boolean{Value: true}
	default:
		return &Boolean{Value: false}
	}
}

func (r *Runtime) evalMinusPrefixOperatorExpression(right Object, tok token.Token) Object {
	switch obj := right.(type) {
	case *Integer:
		return &Integer{Value: -obj.Value}
	case *Float:
		return &Float{Value: -obj.Value}
	default:
		return newErrorTok(tok, "operator '-' tidak bisa dipakai untuk %s", typeName(right))
	}
}

// --- Infix Expression ---

func (r *Runtime) evalInfix(n *ast.InfixExpression, env *Environment) Object {
	// dan/atau memakai evaluasi short-circuit: sisi kanan hanya
	// dievaluasi bila hasilnya belum bisa ditentukan.
	if n.Operator == "dan" || n.Operator == "atau" {
		left := r.eval(n.Left, env)
		if isError(left) {
			return left
		}
		if n.Operator == "dan" && !isTruthy(left) {
			return &Boolean{Value: false}
		}
		if n.Operator == "atau" && isTruthy(left) {
			return &Boolean{Value: true}
		}

		right := r.eval(n.Right, env)
		if isError(right) {
			return right
		}
		return &Boolean{Value: isTruthy(right)}
	}

	left := r.eval(n.Left, env)
	if isError(left) {
		return left
	}
	right := r.eval(n.Right, env)
	if isError(right) {
		return right
	}

	return r.evalInfixExpression(n.Operator, left, right, n.Token)
}

func (r *Runtime) evalInfixExpression(operator string, left Object, right Object, tok token.Token) Object {
	// Perbandingan kesetaraan berlaku lintas tipe tanpa error.
	if operator == "==" {
		return nativeBoolToBooleanObject(objectsEqual(left, right))
	}
	if operator == "!=" {
		return nativeBoolToBooleanObject(!objectsEqual(left, right))
	}

	// Penggabungan teks dengan tipe apa pun.
	if operator == "+" && (left.Type() == STRING_OBJ || right.Type() == STRING_OBJ) {
		return &String{Value: left.Inspect() + right.Inspect()}
	}

	if isNumeric(left) && isNumeric(right) {
		if li, ok := left.(*Integer); ok {
			if ri, ok := right.(*Integer); ok {
				return r.evalIntegerInfixExpression(operator, li.Value, ri.Value, tok)
			}
		}
		lf, _ := toFloat64(left)
		rf, _ := toFloat64(right)
		return r.evalFloatInfixExpression(operator, lf, rf, tok)
	}

	if left.Type() == STRING_OBJ && right.Type() == STRING_OBJ {
		return r.evalStringInfixExpression(operator, left.(*String).Value, right.(*String).Value, tok)
	}

	if left.Type() == BOOLEAN_OBJ && right.Type() == BOOLEAN_OBJ {
		return r.evalBooleanInfixExpression(operator, left.(*Boolean).Value, right.(*Boolean).Value, tok)
	}

	return newErrorTok(tok, "operator '%s' tidak bisa dipakai untuk %s dan %s",
		operator, typeName(left), typeName(right))
}

func (r *Runtime) evalIntegerInfixExpression(operator string, left, right int64, tok token.Token) Object {
	switch operator {
	case "+":
		return &Integer{Value: left + right}
	case "-":
		return &Integer{Value: left - right}
	case "*":
		return &Integer{Value: left * right}
	case "/":
		if right == 0 {
			return newErrorTok(tok, "pembagian dengan nol")
		}
		return &Integer{Value: left / right}
	case "%":
		if right == 0 {
			return newErrorTok(tok, "modulo dengan nol")
		}
		return &Integer{Value: left % right}
	case "<":
		return nativeBoolToBooleanObject(left < right)
	case ">":
		return nativeBoolToBooleanObject(left > right)
	case "<=":
		return nativeBoolToBooleanObject(left <= right)
	case ">=":
		return nativeBoolToBooleanObject(left >= right)
	default:
		return newErrorTok(tok, "operator '%s' tidak dikenal untuk angka", operator)
	}
}

func (r *Runtime) evalFloatInfixExpression(operator string, left, right float64, tok token.Token) Object {
	switch operator {
	case "+":
		return &Float{Value: left + right}
	case "-":
		return &Float{Value: left - right}
	case "*":
		return &Float{Value: left * right}
	case "/":
		if right == 0 {
			return newErrorTok(tok, "pembagian dengan nol")
		}
		return &Float{Value: left / right}
	case "<":
		return nativeBoolToBooleanObject(left < right)
	case ">":
		return nativeBoolToBooleanObject(left > right)
	case "<=":
		return nativeBoolToBooleanObject(left <= right)
	case ">=":
		return nativeBoolToBooleanObject(left >= right)
	default:
		return newErrorTok(tok, "operator '%s' tidak dikenal untuk angka desimal", operator)
	}
}

func (r *Runtime) evalStringInfixExpression(operator string, left, right string, tok token.Token) Object {
	switch operator {
	case "+":
		return &String{Value: left + right}
	case "<":
		return nativeBoolToBooleanObject(left < right)
	case ">":
		return nativeBoolToBooleanObject(left > right)
	case "<=":
		return nativeBoolToBooleanObject(left <= right)
	case ">=":
		return nativeBoolToBooleanObject(left >= right)
	default:
		return newErrorTok(tok, "teks tidak mendukung operator '%s'", operator)
	}
}

func (r *Runtime) evalBooleanInfixExpression(operator string, left, right bool, tok token.Token) Object {
	switch operator {
	case "dan":
		return nativeBoolToBooleanObject(left && right)
	case "atau":
		return nativeBoolToBooleanObject(left || right)
	default:
		return newErrorTok(tok, "logika tidak mendukung operator '%s'", operator)
	}
}

// objectsEqual membandingkan nilai, termasuk daftar secara rekursif.
func objectsEqual(a, b Object) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a.Type() != b.Type() {
		if isNumeric(a) && isNumeric(b) {
			af, _ := toFloat64(a)
			bf, _ := toFloat64(b)
			return af == bf
		}
		return false
	}

	switch av := a.(type) {
	case *Integer:
		return av.Value == b.(*Integer).Value
	case *Float:
		return av.Value == b.(*Float).Value
	case *String:
		return av.Value == b.(*String).Value
	case *Boolean:
		return av.Value == b.(*Boolean).Value
	case *Nil:
		return true
	case *List:
		bv := b.(*List)
		if len(av.Elements) != len(bv.Elements) {
			return false
		}
		for i := range av.Elements {
			if !objectsEqual(av.Elements[i], bv.Elements[i]) {
				return false
			}
		}
		return true
	case *GalatValue:
		bv := b.(*GalatValue)
		return av.Err.Message == bv.Err.Message && av.Err.Line == bv.Err.Line
	case *Peta:
		bv := b.(*Peta)
		if av.Panjang() != bv.Panjang() {
			return false
		}
		for _, hash := range av.urut {
			entri := av.entri[hash]
			nilai, ada, errObj := bv.Ambil(entri.Kunci)
			if errObj != nil || !ada {
				return false
			}
			if !objectsEqual(entri.Nilai, nilai) {
				return false
			}
		}
		return true
	default:
		return a == b
	}
}

// --- Call Expression ---

func (r *Runtime) evalCallExpression(call *ast.CallExpression, env *Environment) Object {
	fn, errObj := r.resolveFunction(call, env)
	if errObj != nil {
		return errObj
	}

	args := r.evalExpressions(call.Arguments, env)
	if len(args) == 1 && isError(args[0]) {
		return args[0]
	}

	switch f := fn.(type) {
	case *Builtin:
		var result Object
		if f.Konteks != nil {
			// Builtin yang butuh akses ke runtime, misalnya kudiag.
			result = f.Konteks(r, args...)
		} else {
			result = f.Fn(args...)
		}
		if errObj, ok := result.(*Error); ok {
			r.attachTrace(errObj, call.Token)
		}
		return result
	case *FunctionObject:
		name := f.Name
		if name == "" {
			name = call.Function.String()
		}

		if len(args) != len(f.Params) {
			return newErrorTok(call.Token, "fungsi '%s' membutuhkan %d parameter, tapi diberikan %d",
				name, len(f.Params), len(args))
		}

		// Anotasi tipe parameter diperiksa sebelum pemanggilan dijalankan
		// agar galat menunjuk ke tempat argumen salah ditulis.
		for i, param := range f.Params {
			if param.TypeHint == "" {
				continue
			}
			if !cocokTipe(param.TypeHint, args[i]) {
				return newErrorTok(call.Token,
					"argumen ke-%d fungsi '%s' harus bertipe %s, mendapat %s",
					i+1, name, param.TypeHint, typeName(args[i]))
			}
		}

		// Batas kedalaman mencegah tumpukan Go meluap pada rekursi tak
		// berujung, yang jika terjadi membuat proses mati total.
		if r.limits.MaxDepth > 0 && len(r.stack) >= r.limits.MaxDepth {
			return newErrorTok(call.Token, "kedalaman pemanggilan melebihi batas %d pada fungsi '%s'; periksa rekursi tanpa kondisi berhenti",
				r.limits.MaxDepth, name)
		}

		enclosed := NewEnclosedEnvironment(f.Env)
		enclosed.returnType = f.ReturnType
		enclosed.namaFungsi = name
		for i, param := range f.Params {
			enclosed.Set(param.Name.Value, args[i])
		}

		r.stack = append(r.stack, Frame{Name: name, Line: call.Token.Line})
		result := r.eval(f.Body, enclosed)
		if errObj, ok := result.(*Error); ok {
			r.attachTrace(errObj, call.Token)
		}
		r.stack = r.stack[:len(r.stack)-1]

		if result != nil && result.Type() == RETURN_OBJ {
			result = result.(*ReturnValue).Value
		}

		// Pemeriksaan di sini hanya berlaku untuk fungsi yang selesai
		// tanpa pernyataan 'kembali'; galat dari 'kembali' sudah membawa
		// pesan yang lebih tepat.
		if f.ReturnType != "" && !isError(result) && !cocokTipe(f.ReturnType, result) {
			return newErrorTok(call.Token,
				"fungsi '%s' harus mengembalikan nilai bertipe %s, mendapat %s",
				name, f.ReturnType, typeName(result))
		}

		return result
	default:
		return newErrorTok(call.Token, "'%s' bukan sebuah fungsi", call.Function.String())
	}
}

func (r *Runtime) resolveFunction(call *ast.CallExpression, env *Environment) (Object, Object) {
	if ident, ok := call.Function.(*ast.Identifier); ok {
		if fn, found := env.Get(ident.Value); found {
			return fn, nil
		}
		if builtin, found := r.builtins[ident.Value]; found {
			return builtin, nil
		}
		return nil, newErrorTok(call.Token, "fungsi '%s' tidak ditemukan", ident.Value)
	}

	fn := r.eval(call.Function, env)
	if isError(fn) {
		return nil, fn
	}
	return fn, nil
}

func (r *Runtime) evalExpressions(exps []ast.Expression, env *Environment) []Object {
	var result []Object

	for _, e := range exps {
		evaluated := r.eval(e, env)
		if isError(evaluated) {
			return []Object{evaluated}
		}
		result = append(result, evaluated)
	}

	return result
}

// --- Input ---

func (r *Runtime) readInput() string {
	if r.reader == nil {
		if r.In == nil {
			r.In = strings.NewReader("")
		}
		// Jangan bungkus ulang bufio.Reader agar buffer yang sudah ada
		// (misalnya dipakai REPL) tidak tertinggal.
		if buffered, ok := r.In.(*bufio.Reader); ok {
			r.reader = buffered
		} else {
			r.reader = bufio.NewReader(r.In)
		}
	}
	fmt.Fprint(r.Out, "> ")
	line, err := r.reader.ReadString('\n')
	if err != nil && line == "" {
		return ""
	}
	return strings.TrimRight(line, "\r\n")
}

// --- Helpers ---

func newError(line int, format string, args ...interface{}) *Error {
	return &Error{Message: fmt.Sprintf(format, args...), Line: line}
}

// newErrorTok membuat galat yang menunjuk posisi token secara presisi.
func newErrorTok(tok token.Token, format string, args ...interface{}) *Error {
	return &Error{Message: fmt.Sprintf(format, args...), Line: tok.Line, Column: tok.Column}
}

// attachPosition melengkapi galat dengan posisi node bila belum terisi.
// Dipanggil pada tingkat pernyataan agar setiap galat punya titik yang bisa
// ditunjuk pada cuplikan sumber.
func attachPosition(err *Error, node ast.Node) {
	if err == nil || err.Column != 0 {
		return
	}
	tok := ast.TokenOf(node)
	if tok.Line == 0 {
		return
	}
	if err.Line == 0 {
		err.Line = tok.Line
	}
	err.Column = tok.Column
}

// indexToken mengembalikan token tanda kurung siku dari target indeks agar
// galat indeks menunjuk posisi yang tepat.
func indexToken(target ast.Expression) token.Token {
	if index, ok := target.(*ast.IndexExpression); ok {
		return index.Token
	}
	return token.Token{}
}

// snapshot menyalin jejak pemanggilan dengan urutan terdalam lebih dulu.
func snapshot(stack []Frame) []Frame {
	if len(stack) == 0 {
		return nil
	}
	trace := make([]Frame, 0, len(stack))
	for i := len(stack) - 1; i >= 0; i-- {
		trace = append(trace, stack[i])
	}
	return trace
}

// attachTrace melengkapi galat dengan nomor baris dan jejak pemanggilan bila
// belum diisi oleh pembuat galat asli.
func (r *Runtime) attachTrace(err *Error, tok token.Token) {
	if err.Line == 0 {
		err.Line = tok.Line
		err.Column = tok.Column
	}
	if len(err.Trace) == 0 {
		err.Trace = snapshot(r.stack)
	}
}

func typeName(obj Object) string {
	if obj == nil {
		return "kosong"
	}
	switch obj.Type() {
	case INTEGER_OBJ:
		return "angka"
	case FLOAT_OBJ:
		return "desimal"
	case STRING_OBJ:
		return "teks"
	case BOOLEAN_OBJ:
		return "logika"
	case NIL_OBJ:
		return "kosong"
	case LIST_OBJ:
		return "daftar"
	case PETA_OBJ:
		return "peta"
	case GALAT_OBJ:
		return "galat"
	case FUNCTION_OBJ, BUILTIN_OBJ:
		return "fungsi"
	case ERROR_OBJ:
		return "galat"
	}
	return string(obj.Type())
}

func isTruthy(obj Object) bool {
	if obj == nil {
		return false
	}
	switch o := obj.(type) {
	case *Nil:
		return false
	case *Boolean:
		return o.Value
	case *Integer:
		return o.Value != 0
	case *Float:
		return o.Value != 0
	case *String:
		return o.Value != ""
	default:
		return true
	}
}

func isError(obj Object) bool {
	if obj != nil {
		return obj.Type() == ERROR_OBJ
	}
	return false
}

// cocokTipe memeriksa apakah nilai sebuah objek sesuai dengan anotasi tipe.
// Anotasi 'apa_saja' menerima semua nilai, dan nilai kosong (tanpa
// kembalian) cocok dengan tipe 'kosong'.
func cocokTipe(hint string, obj Object) bool {
	if hint == "" || hint == tipe.ApaSaja {
		return true
	}
	if obj == nil {
		return hint == tipe.Kosong
	}

	switch hint {
	case tipe.Angka:
		_, ok := obj.(*Integer)
		return ok
	case tipe.Desimal:
		_, ok := obj.(*Float)
		return ok
	case tipe.Teks:
		_, ok := obj.(*String)
		return ok
	case tipe.Logika:
		_, ok := obj.(*Boolean)
		return ok
	case tipe.Kosong:
		_, ok := obj.(*Nil)
		return ok
	case tipe.Daftar:
		_, ok := obj.(*List)
		return ok
	case tipe.Peta:
		_, ok := obj.(*Peta)
		return ok
	case tipe.Fungsi:
		switch obj.(type) {
		case *FunctionObject, *Builtin:
			return true
		}
		return false
	}
	return false
}

func isNumeric(obj Object) bool {
	if obj == nil {
		return false
	}
	return obj.Type() == INTEGER_OBJ || obj.Type() == FLOAT_OBJ
}

func nativeBoolToBooleanObject(input bool) *Boolean {
	return &Boolean{Value: input}
}

func toFloat64(obj Object) (float64, bool) {
	switch o := obj.(type) {
	case *Integer:
		return float64(o.Value), true
	case *Float:
		return o.Value, true
	}
	return 0, false
}

// runeLen dipakai builtin dan pengukuran panjang teks.
func runeLen(s string) int {
	return utf8.RuneCountInString(s)
}
