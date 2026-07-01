package runtime

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

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
	RETURN_OBJ   ObjectType = "RETURN"
	BREAK_OBJ    ObjectType = "BREAK"
	CONTINUE_OBJ ObjectType = "CONTINUE"
	FUNCTION_OBJ ObjectType = "FUNCTION"
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
func (b *Boolean) Inspect() string  { return strconv.FormatBool(b.Value) }

type Nil struct{}

func (n *Nil) Type() ObjectType { return NIL_OBJ }
func (n *Nil) Inspect() string  { return "kosong" }

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
	Params []*ast.Identifier
	Body   *ast.BlockStatement
	Env    *Environment
}

func (fo *FunctionObject) Type() ObjectType { return FUNCTION_OBJ }
func (fo *FunctionObject) Inspect() string {
	return fmt.Sprintf("fungsi(%d params)", len(fo.Params))
}

type Error struct{ Message string }

func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string  { return "ERROR: " + e.Message }

// --- Environment ---

type Environment struct {
	store  map[string]Object
	outer  *Environment
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

// --- Runtime (Evaluator) ---

type Runtime struct {
	env *Environment
}

func New() *Runtime {
	return &Runtime{env: NewEnvironment()}
}

func (r *Runtime) Execute(p *ast.Program) {
	result := r.eval(p, r.env)
	if result != nil && result.Type() == ERROR_OBJ {
		fmt.Fprintln(os.Stderr, result.Inspect())
	}
}

func (r *Runtime) eval(node ast.Node, env *Environment) Object {
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
		env.Set(n.Name.Value, val)

	case *ast.PrintStatement:
		val := r.eval(n.Value, env)
		if isError(val) {
			return val
		}
		fmt.Println(val.Inspect())

	case *ast.IfStatement:
		return r.evalIfExpression(n, env)

	case *ast.ForStatement:
		return r.evalForStatement(n, env)

	case *ast.WhileStatement:
		return r.evalWhileStatement(n, env)

	case *ast.FunctionStatement:
		fn := &FunctionObject{
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
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimRight(input, "\n")
		env.Set(n.Name.Value, &String{Value: input})

	// --- Expressions ---
	case *ast.InfixExpression:
		// Handle assignment ("=")
		if n.Operator == "=" {
			ident, ok := n.Left.(*ast.Identifier)
			if !ok {
				return &Error{Message: fmt.Sprintf("baris %d: hanya variabel yang bisa di-assign", n.Token.Line)}
			}
			val := r.eval(n.Right, env)
			if isError(val) {
				return val
			}
			env.Set(ident.Value, val)
			return val
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

// --- For Statement ---

func (r *Runtime) evalForStatement(fs *ast.ForStatement, env *Environment) Object {
	countObj := r.eval(fs.Count, env)
	if isError(countObj) {
		return countObj
	}

	count := toInt64(countObj)

	for i := int64(0); i < count; i++ {
		result := r.eval(fs.Body, env)
		if result != nil {
			switch result.Type() {
			case BREAK_OBJ:
				return nil
			case CONTINUE_OBJ:
				continue
			case RETURN_OBJ:
				return result
			case ERROR_OBJ:
				return result
			}
		}
	}

	return nil
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
		if result != nil {
			switch result.Type() {
			case BREAK_OBJ:
				return nil
			case CONTINUE_OBJ:
				continue
			case RETURN_OBJ:
				return result
			case ERROR_OBJ:
				return result
			}
		}
	}

	return nil
}

// --- Identifier ---

func (r *Runtime) evalIdentifier(node *ast.Identifier, env *Environment) Object {
	val, ok := env.Get(node.Value)
	if !ok {
		return &Error{Message: fmt.Sprintf("baris %d: variabel '%s' tidak ditemukan", node.Token.Line, node.Value)}
	}
	return val
}

// --- Prefix Expression ---

func (r *Runtime) evalPrefixExpression(operator string, right Object, line int) Object {
	switch operator {
	case "!", "tidak":
		return r.evalBangOperatorExpression(right)
	case "-":
		return r.evalMinusPrefixOperatorExpression(right, line)
	default:
		return &Error{Message: fmt.Sprintf("baris %d: operator '%s' tidak dikenal", line, operator)}
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
		return &Error{Message: fmt.Sprintf("baris %d: operator '-' tidak bisa dipakai untuk %s", line, right.Type())}
	}
}

// --- Infix Expression ---

func (r *Runtime) evalInfixExpression(operator string, left Object, right Object, line int) Object {
	// Handle mixed-type string concatenation: if one operand is a string, convert the other
	if operator == "+" && (left.Type() == STRING_OBJ || right.Type() == STRING_OBJ) {
		leftStr := left.Inspect()
		rightStr := right.Inspect()
		return &String{Value: leftStr + rightStr}
	}

	switch {
	case left.Type() == INTEGER_OBJ && right.Type() == INTEGER_OBJ:
		return r.evalIntegerInfixExpression(operator, left, right, line)
	case left.Type() == FLOAT_OBJ || right.Type() == FLOAT_OBJ:
		return r.evalFloatInfixExpression(operator, left, right, line)
	case left.Type() == STRING_OBJ && right.Type() == STRING_OBJ:
		return r.evalStringInfixExpression(operator, left, right, line)
	case left.Type() == BOOLEAN_OBJ && right.Type() == BOOLEAN_OBJ:
		return r.evalBooleanInfixExpression(operator, left, right, line)
	case operator == "==":
		return nativeBoolToBooleanObject(left == right)
	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)
	default:
		return &Error{Message: fmt.Sprintf("baris %d: operator '%s' tidak bisa dipakai untuk %s dan %s",
			line, operator, left.Type(), right.Type())}
	}
}

func (r *Runtime) evalIntegerInfixExpression(operator string, left, right Object, line int) Object {
	leftVal := left.(*Integer).Value
	rightVal := right.(*Integer).Value

	switch operator {
	case "+":
		return &Integer{Value: leftVal + rightVal}
	case "-":
		return &Integer{Value: leftVal - rightVal}
	case "*":
		return &Integer{Value: leftVal * rightVal}
	case "/":
		if rightVal == 0 {
			return &Error{Message: fmt.Sprintf("baris %d: pembagian dengan nol", line)}
		}
		return &Integer{Value: leftVal / rightVal}
	case "%":
		if rightVal == 0 {
			return &Error{Message: fmt.Sprintf("baris %d: modulo dengan nol", line)}
		}
		return &Integer{Value: leftVal % rightVal}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "<=":
		return nativeBoolToBooleanObject(leftVal <= rightVal)
	case ">=":
		return nativeBoolToBooleanObject(leftVal >= rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	case "dan":
		return nativeBoolToBooleanObject(isTruthy(left) && isTruthy(right))
	case "atau":
		return nativeBoolToBooleanObject(isTruthy(left) || isTruthy(right))
	default:
		return &Error{Message: fmt.Sprintf("baris %d: operator '%s' tidak dikenal untuk integer", line, operator)}
	}
}

func (r *Runtime) evalFloatInfixExpression(operator string, left, right Object, line int) Object {
	leftVal := toFloat64(left)
	rightVal := toFloat64(right)

	switch operator {
	case "+":
		return &Float{Value: leftVal + rightVal}
	case "-":
		return &Float{Value: leftVal - rightVal}
	case "*":
		return &Float{Value: leftVal * rightVal}
	case "/":
		if rightVal == 0 {
			return &Error{Message: fmt.Sprintf("baris %d: pembagian dengan nol", line)}
		}
		return &Float{Value: leftVal / rightVal}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "<=":
		return nativeBoolToBooleanObject(leftVal <= rightVal)
	case ">=":
		return nativeBoolToBooleanObject(leftVal >= rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return &Error{Message: fmt.Sprintf("baris %d: operator '%s' tidak dikenal", line, operator)}
	}
}

func (r *Runtime) evalStringInfixExpression(operator string, left, right Object, line int) Object {
	leftVal := left.(*String).Value
	rightVal := right.(*String).Value

	switch operator {
	case "+":
		return &String{Value: leftVal + rightVal}
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return &Error{Message: fmt.Sprintf("baris %d: string tidak mendukung operator '%s'", line, operator)}
	}
}

func (r *Runtime) evalBooleanInfixExpression(operator string, left, right Object, line int) Object {
	leftVal := left.(*Boolean).Value
	rightVal := right.(*Boolean).Value

	switch operator {
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	case "dan":
		return nativeBoolToBooleanObject(leftVal && rightVal)
	case "atau":
		return nativeBoolToBooleanObject(leftVal || rightVal)
	default:
		return &Error{Message: fmt.Sprintf("baris %d: boolean tidak mendukung operator '%s'", line, operator)}
	}
}

// --- Call Expression ---

func (r *Runtime) evalCallExpression(call *ast.CallExpression, env *Environment) Object {
	// Evaluate the function identifier
	fn, ok := env.Get(call.Function.String())
	if !ok {
		return &Error{Message: fmt.Sprintf("baris %d: fungsi '%s' tidak ditemukan",
			call.Token.Line, call.Function.String())}
	}

	functionObj, ok := fn.(*FunctionObject)
	if !ok {
		return &Error{Message: fmt.Sprintf("baris %d: '%s' bukan sebuah fungsi",
			call.Token.Line, call.Function.String())}
	}

	// Evaluate arguments
	args := r.evalExpressions(call.Arguments, env)
	if len(args) == 1 && isError(args[0]) {
		return args[0]
	}

	// Create new environment for function
	if len(args) != len(functionObj.Params) {
		return &Error{Message: fmt.Sprintf("baris %d: fungsi '%s' membutuhkan %d parameter, tapi diberikan %d",
			call.Token.Line, call.Function.String(), len(functionObj.Params), len(args))}
	}

	enclosed := NewEnclosedEnvironment(functionObj.Env)
	for i, param := range functionObj.Params {
		enclosed.Set(param.Value, args[i])
	}

	result := r.eval(functionObj.Body, enclosed)
	if result != nil && result.Type() == RETURN_OBJ {
		return result.(*ReturnValue).Value
	}
	return result
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

// --- Helpers ---

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

func nativeBoolToBooleanObject(input bool) *Boolean {
	return &Boolean{Value: input}
}

func toInt64(obj Object) int64 {
	switch o := obj.(type) {
	case *Integer:
		return o.Value
	case *Float:
		return int64(o.Value)
	}
	return 0
}

func toFloat64(obj Object) float64 {
	switch o := obj.(type) {
	case *Integer:
		return float64(o.Value)
	case *Float:
		return o.Value
	}
	return 0
}
