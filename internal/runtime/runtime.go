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
	RETURN_OBJ   ObjectType = "RETURN"
	BREAK_OBJ    ObjectType = "BREAK"
	CONTINUE_OBJ ObjectType = "CONTINUE"
	FUNCTION_OBJ ObjectType = "FUNCTION"
	BUILTIN_OBJ  ObjectType = "BUILTIN"
	ERROR_OBJ    ObjectType = "ERROR"
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
	Params []*ast.Identifier
	Body   *ast.BlockStatement
	Env    *Environment
}

func (fo *FunctionObject) Type() ObjectType { return FUNCTION_OBJ }
func (fo *FunctionObject) Inspect() string {
	if fo.Name == "" {
		return fmt.Sprintf("fungsi(%d parameter)", len(fo.Params))
	}
	return fmt.Sprintf("fungsi %s(%d parameter)", fo.Name, len(fo.Params))
}

type Error struct{ Message string }

func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string  { return "ERROR: " + e.Message }

// --- Environment ---

type Environment struct {
	store map[string]Object
	outer *Environment
}

func NewEnvironment() *Environment {
	s := make(map[string]Object)
	return &Environment{store: s, outer: nil}
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

type Runtime struct {
	env      *Environment
	Out      io.Writer
	In       io.Reader
	reader   *bufio.Reader
	builtins map[string]*Builtin
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
	}
}

// Execute menjalankan program. Nilai kembalian berupa Object: galat
// dikembalikan sebagai *Error agar pemanggil bisa menentukan exit code.
func (r *Runtime) Execute(p *ast.Program) Object {
	return r.eval(p, r.env)
}

func (r *Runtime) eval(node ast.Node, env *Environment) Object {
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
		val := r.eval(n.Value, env)
		if isError(val) {
			return val
		}
		if val == nil {
			return newError(n.Token.Line, "ekspresi tidak menghasilkan nilai")
		}
		env.Set(n.Name.Value, val)

	case *ast.PrintStatement:
		val := r.eval(n.Value, env)
		if isError(val) {
			return val
		}
		if val == nil {
			return newError(n.Token.Line, "tidak ada nilai untuk ditulis")
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

	case *ast.WhileStatement:
		return r.evalWhileStatement(n, env)

	case *ast.FunctionStatement:
		fn := &FunctionObject{
			Name:   n.Name.Value,
			Params: n.Params,
			Body:   n.Body,
			Env:    env,
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
		return &ReturnValue{Value: val}

	case *ast.BreakStatement:
		return &BreakValue{}

	case *ast.ContinueStatement:
		return &ContinueValue{}

	case *ast.InputStatement:
		if n.Name == nil {
			return newError(n.Token.Line, "membaca membutuhkan nama variabel")
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
		return r.evalPrefixExpression(n.Operator, right, n.Token.Line)

	case *ast.CallExpression:
		return r.evalCallExpression(n, env)

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
		return newError(fs.Token.Line, "ulang membutuhkan angka bulat, mendapat %s", typeName(countObj))
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
					return newError(us.Token.Line, "langkah tidak boleh nol")
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
		return newError(us.Token.Line, "nilai awal harus berupa angka, mendapat %s", typeName(startObj))
	}
	end, ok := toFloat64(endObj)
	if !ok {
		return newError(us.Token.Line, "nilai akhir harus berupa angka, mendapat %s", typeName(endObj))
	}
	step, ok := toFloat64(stepObj)
	if !ok {
		return newError(us.Token.Line, "langkah harus berupa angka, mendapat %s", typeName(stepObj))
	}
	if step == 0 {
		return newError(us.Token.Line, "langkah tidak boleh nol")
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

	run := func(item Object) (bool, Object) {
		env.Set(ud.Name.Value, item)
		result := r.eval(ud.Body, env)
		return loopSignal(result)
	}

	switch obj := iterable.(type) {
	case *List:
		for _, item := range obj.Elements {
			if stop, propagate := run(item); stop {
				return propagate
			}
		}
	case *String:
		for _, ch := range obj.Value {
			if stop, propagate := run(&String{Value: string(ch)}); stop {
				return propagate
			}
		}
	default:
		return newError(ud.Token.Line, "untuk ... dalam hanya bisa mengulang daftar atau teks, mendapat %s", typeName(iterable))
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
	return newError(node.Token.Line, "variabel '%s' tidak ditemukan", node.Value)
}

// --- Assignment ---

func (r *Runtime) evalAssign(ae *ast.AssignExpression, env *Environment) Object {
	var current Object
	var list *List
	index := -1

	if ae.Operator != "=" {
		switch target := ae.Target.(type) {
		case *ast.Identifier:
			obj, ok := env.Get(target.Value)
			if !ok {
				return newError(ae.Token.Line, "variabel '%s' tidak ditemukan", target.Value)
			}
			current = obj
		case *ast.IndexExpression:
			l, i, errObj := r.resolveListIndex(target, env)
			if errObj != nil {
				return errObj
			}
			list, index = l, i
			current = l.Elements[i]
		default:
			return newError(ae.Token.Line, "target penugasan harus variabel atau elemen daftar")
		}
	}

	value := r.eval(ae.Value, env)
	if isError(value) {
		return value
	}
	if value == nil {
		return newError(ae.Token.Line, "ekspresi tidak menghasilkan nilai")
	}

	if ae.Operator != "=" {
		op := strings.TrimSuffix(ae.Operator, "=")
		value = r.evalInfixExpression(op, current, value, ae.Token.Line)
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
		if list == nil {
			l, i, errObj := r.resolveListIndex(target, env)
			if errObj != nil {
				return errObj
			}
			list, index = l, i
		}
		list.Elements[index] = value
		return value
	default:
		return newError(ae.Token.Line, "target penugasan harus variabel atau elemen daftar")
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

	switch obj := left.(type) {
	case *List:
		index, err := resolveIndex(indexObj, len(obj.Elements))
		if err != nil {
			return newError(node.Token.Line, "%s", err.Error())
		}
		return obj.Elements[index]
	case *String:
		runes := []rune(obj.Value)
		index, err := resolveIndex(indexObj, len(runes))
		if err != nil {
			return newError(node.Token.Line, "%s", err.Error())
		}
		return &String{Value: string(runes[index])}
	default:
		return newError(node.Token.Line, "tipe %s tidak bisa diindeks", typeName(left))
	}
}

func (r *Runtime) resolveListIndex(node *ast.IndexExpression, env *Environment) (*List, int, Object) {
	left := r.eval(node.Left, env)
	if isError(left) {
		return nil, 0, left
	}
	list, ok := left.(*List)
	if !ok {
		return nil, 0, newError(node.Token.Line, "tipe %s tidak bisa diindeks", typeName(left))
	}

	indexObj := r.eval(node.Index, env)
	if isError(indexObj) {
		return nil, 0, indexObj
	}

	index, err := resolveIndex(indexObj, len(list.Elements))
	if err != nil {
		return nil, 0, newError(node.Token.Line, "%s", err.Error())
	}
	return list, index, nil
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

func (r *Runtime) evalPrefixExpression(operator string, right Object, line int) Object {
	switch operator {
	case "!", "tidak":
		return r.evalBangOperatorExpression(right)
	case "-":
		return r.evalMinusPrefixOperatorExpression(right, line)
	default:
		return newError(line, "operator '%s' tidak dikenal", operator)
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

func (r *Runtime) evalMinusPrefixOperatorExpression(right Object, line int) Object {
	switch obj := right.(type) {
	case *Integer:
		return &Integer{Value: -obj.Value}
	case *Float:
		return &Float{Value: -obj.Value}
	default:
		return newError(line, "operator '-' tidak bisa dipakai untuk %s", typeName(right))
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

	return r.evalInfixExpression(n.Operator, left, right, n.Token.Line)
}

func (r *Runtime) evalInfixExpression(operator string, left Object, right Object, line int) Object {
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
				return r.evalIntegerInfixExpression(operator, li.Value, ri.Value, line)
			}
		}
		lf, _ := toFloat64(left)
		rf, _ := toFloat64(right)
		return r.evalFloatInfixExpression(operator, lf, rf, line)
	}

	if left.Type() == STRING_OBJ && right.Type() == STRING_OBJ {
		return r.evalStringInfixExpression(operator, left.(*String).Value, right.(*String).Value, line)
	}

	if left.Type() == BOOLEAN_OBJ && right.Type() == BOOLEAN_OBJ {
		return r.evalBooleanInfixExpression(operator, left.(*Boolean).Value, right.(*Boolean).Value, line)
	}

	return newError(line, "operator '%s' tidak bisa dipakai untuk %s dan %s",
		operator, typeName(left), typeName(right))
}

func (r *Runtime) evalIntegerInfixExpression(operator string, left, right int64, line int) Object {
	switch operator {
	case "+":
		return &Integer{Value: left + right}
	case "-":
		return &Integer{Value: left - right}
	case "*":
		return &Integer{Value: left * right}
	case "/":
		if right == 0 {
			return newError(line, "pembagian dengan nol")
		}
		return &Integer{Value: left / right}
	case "%":
		if right == 0 {
			return newError(line, "modulo dengan nol")
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
		return newError(line, "operator '%s' tidak dikenal untuk angka", operator)
	}
}

func (r *Runtime) evalFloatInfixExpression(operator string, left, right float64, line int) Object {
	switch operator {
	case "+":
		return &Float{Value: left + right}
	case "-":
		return &Float{Value: left - right}
	case "*":
		return &Float{Value: left * right}
	case "/":
		if right == 0 {
			return newError(line, "pembagian dengan nol")
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
		return newError(line, "operator '%s' tidak dikenal untuk angka desimal", operator)
	}
}

func (r *Runtime) evalStringInfixExpression(operator string, left, right string, line int) Object {
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
		return newError(line, "teks tidak mendukung operator '%s'", operator)
	}
}

func (r *Runtime) evalBooleanInfixExpression(operator string, left, right bool, line int) Object {
	switch operator {
	case "dan":
		return nativeBoolToBooleanObject(left && right)
	case "atau":
		return nativeBoolToBooleanObject(left || right)
	default:
		return newError(line, "logika tidak mendukung operator '%s'", operator)
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
		return f.Fn(args...)
	case *FunctionObject:
		if len(args) != len(f.Params) {
			name := f.Name
			if name == "" {
				name = call.Function.String()
			}
			return newError(call.Token.Line, "fungsi '%s' membutuhkan %d parameter, tapi diberikan %d",
				name, len(f.Params), len(args))
		}

		enclosed := NewEnclosedEnvironment(f.Env)
		for i, param := range f.Params {
			enclosed.Set(param.Value, args[i])
		}

		result := r.eval(f.Body, enclosed)
		if result != nil && result.Type() == RETURN_OBJ {
			return result.(*ReturnValue).Value
		}
		return result
	default:
		return newError(call.Token.Line, "'%s' bukan sebuah fungsi", call.Function.String())
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
		return nil, newError(call.Token.Line, "fungsi '%s' tidak ditemukan", ident.Value)
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
	return &Error{Message: fmt.Sprintf("baris %d: %s", line, fmt.Sprintf(format, args...))}
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
