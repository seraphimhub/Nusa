package parser

import (
	"fmt"
	"strconv"

	"github.com/seraphimhub/Nusa/internal/ast"
	"github.com/seraphimhub/Nusa/internal/token"
)

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

// Precedence levels
type Precedence int

const (
	_ Precedence = iota
	ASSIGNMENT   // =
	LOWEST       // default
	LOGICAL      // dan, atau
	COMPARISON   // ==, !=, >, <, >=, <=
	SUM          // +, -
	PRODUCT      // *, /, %
	PREFIX       // !, -
	CALL         // panggil fn(x)
)

var precedences = map[token.TokenType]Precedence{
	token.ASSIGN:   ASSIGNMENT,
	token.DAN:      LOGICAL,
	token.ATAU:     LOGICAL,
	token.EQ:       COMPARISON,
	token.NEQ:      COMPARISON,
	token.GREATER:  COMPARISON,
	token.LESS:     COMPARISON,
	token.GTE:      COMPARISON,
	token.LTE:      COMPARISON,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.MULTIPLY: PRODUCT,
	token.DIVIDE:   PRODUCT,
	token.MODULO:   PRODUCT,
	token.LPAREN:   CALL,
}

type Parser struct {
	l         *lexerHelper
	curToken  token.Token
	peekToken token.Token
	errors    []string

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

// lexerHelper wraps the lexer to get tokens lexer-style
type lexerHelper struct {
	tokens []token.Token
	pos    int
}

func newLexerHelper(tokens []token.Token) *lexerHelper {
	return &lexerHelper{tokens: tokens, pos: 0}
}

func (lh *lexerHelper) nextToken() token.Token {
	if lh.pos >= len(lh.tokens) {
		return token.Token{Type: token.EOF}
	}
	tok := lh.tokens[lh.pos]
	lh.pos++
	return tok
}

func New(tokens []token.Token) *Parser {
	p := &Parser{
		l:      newLexerHelper(tokens),
		errors: []string{},
	}

	p.nextToken()
	p.nextToken()

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.NUMBER, p.parseIntegerLiteral)
	p.registerPrefix(token.FLOAT, p.parseFloatLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.BENAR, p.parseBooleanLiteral)
	p.registerPrefix(token.SALAH, p.parseBooleanLiteral)
	p.registerPrefix(token.KOSONG, p.parseNilLiteral)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.TIDAK, p.parsePrefixExpression)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.PANGGIL, p.parseCallExpression)
	p.registerPrefix(token.MEMBACA, p.parseInputExpression)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.MULTIPLY, p.parseInfixExpression)
	p.registerInfix(token.DIVIDE, p.parseInfixExpression)
	p.registerInfix(token.MODULO, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NEQ, p.parseInfixExpression)
	p.registerInfix(token.GREATER, p.parseInfixExpression)
	p.registerInfix(token.LESS, p.parseInfixExpression)
	p.registerInfix(token.GTE, p.parseInfixExpression)
	p.registerInfix(token.LTE, p.parseInfixExpression)
	p.registerInfix(token.DAN, p.parseInfixExpression)
	p.registerInfix(token.ATAU, p.parseInfixExpression)

	return p
}

func (p *Parser) registerPrefix(tokType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokType] = fn
}

func (p *Parser) registerInfix(tokType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokType] = fn
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.nextMeaningfulToken()
}

// nextMeaningfulToken returns the next non-NEWLINE token.
func (p *Parser) nextMeaningfulToken() token.Token {
	for {
		tok := p.l.nextToken()
		if tok.Type != token.NEWLINE {
			return tok
		}
	}
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("baris %d: tidak diduga '%s', mengharapkan '%s'",
		p.peekToken.Line, p.peekToken.Literal, t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("baris %d: tidak ada ekspresi yang dikenal untuk '%s'",
		p.curToken.Line, p.curToken.Literal)
	p.errors = append(p.errors, msg)
}

func (p *Parser) peekPrecedence() Precedence {
	if pre, ok := precedences[p.peekToken.Type]; ok {
		return pre
	}
	return LOWEST
}

func (p *Parser) curPrecedence() Precedence {
	if pre, ok := precedences[p.curToken.Type]; ok {
		return pre
	}
	return LOWEST
}

// skipToNextStatement consumes tokens until we find a statement boundary
func (p *Parser) skipToNextStatement() {
	for p.curToken.Type != token.EOF {
		if p.curToken.Type == token.BUAT ||
			p.curToken.Type == token.TULIS ||
			p.curToken.Type == token.JIKA ||
			p.curToken.Type == token.ULANG ||
			p.curToken.Type == token.SELAMA ||
			p.curToken.Type == token.FUNGSI ||
			p.curToken.Type == token.PANGGIL ||
			p.curToken.Type == token.KEMBALI ||
			p.curToken.Type == token.BERHENTI ||
			p.curToken.Type == token.LANJUTKAN ||
			p.curToken.Type == token.MEMBACA ||
			p.curToken.Type == token.KALAU_TIDAK ||
			p.curToken.Type == token.RBRACE ||
			p.curToken.Type == token.EOF {
			return
		}
		p.nextToken()
	}
}

// --- Parse Program ---

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.BUAT:
		return p.parseLetStatement()
	case token.TULIS:
		return p.parsePrintStatement()
	case token.JIKA:
		return p.parseIfStatement()
	case token.ULANG:
		return p.parseForStatement()
	case token.SELAMA:
		return p.parseWhileStatement()
	case token.FUNGSI:
		return p.parseFunctionStatement()
	case token.PANGGIL:
		return p.parseCallStatement()
	case token.KEMBALI:
		return p.parseReturnStatement()
	case token.BERHENTI:
		return p.parseBreakStatement()
	case token.LANJUTKAN:
		return p.parseContinueStatement()
	case token.MEMBACA:
		return p.parseInputStatement()
	case token.RBRACE:
		return nil
	default:
		// Check for assignment statement: identifier = expr
		if p.curToken.Type == token.IDENT && p.peekTokenIs(token.ASSIGN) {
			return p.parseAssignStatement()
		}
		return p.parseExpressionStatement()
	}
}

// --- Let Statement ---

func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}

	p.nextToken()
	if p.curToken.Type != token.IDENT {
		msg := fmt.Sprintf("baris %d: mengharapkan identifier setelah 'buat'", p.curToken.Line)
		p.errors = append(p.errors, msg)
		p.skipToNextStatement()
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	p.nextToken()
	if p.curToken.Type != token.ASSIGN {
		msg := fmt.Sprintf("baris %d: mengharapkan '=' setelah identifier, mendapat '%s'",
			p.curToken.Line, p.curToken.Literal)
		p.errors = append(p.errors, msg)
		p.skipToNextStatement()
		return nil
	}

	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	return stmt
}

// --- Print Statement ---

func (p *Parser) parseAssignStatement() *ast.ExpressionStatement {
	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	p.nextToken() // consume IDENT
	assignTok := p.curToken

	p.nextToken() // consume =
	value := p.parseExpression(LOWEST)

	expr := &ast.InfixExpression{
		Token:    assignTok,
		Operator: "=",
		Left:     ident,
		Right:    value,
	}

	return &ast.ExpressionStatement{Token: ident.Token, Expression: expr}
}

func (p *Parser) parsePrintStatement() *ast.PrintStatement {
	stmt := &ast.PrintStatement{Token: p.curToken}

	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	return stmt
}

// --- If Statement ---

func (p *Parser) parseIfStatement() *ast.IfStatement {
	stmt := &ast.IfStatement{Token: p.curToken}

	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)

	// Expect LBRACE or directly a single statement
	p.nextToken()

	if p.curToken.Type == token.LBRACE {
		stmt.Consequence = p.parseBlockStatement()
	} else {
		// Single statement
		stmt.Consequence = &ast.BlockStatement{
			Token: p.curToken,
			Statements: []ast.Statement{
				p.parseStatement(),
			},
		}
	}

	// Look for KALAU_TIDAK (else) or KALAU_TIDAK JIKA (else if)
	// Use peek to avoid consuming the next statement's token
	if p.peekTokenIs(token.KALAU_TIDAK) {
		p.nextToken() // consume KALAU_TIDAK
		p.nextToken() // advance past KALAU_TIDAK

		// Check for else-if (kalau_tidak jika)
		if p.curToken.Type == token.JIKA {
			stmt.Alternative = &ast.BlockStatement{
				Token: p.curToken,
				Statements: []ast.Statement{
					p.parseIfStatement(),
				},
			}
		} else if p.curToken.Type == token.LBRACE {
			stmt.Alternative = p.parseBlockStatement()
		} else {
			// Single statement
			stmt.Alternative = &ast.BlockStatement{
				Token: p.curToken,
				Statements: []ast.Statement{
					p.parseStatement(),
				},
			}
		}
	}

	return stmt
}

// --- Block Statement ---

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken()
	for p.curToken.Type != token.RBRACE && p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

// --- For Statement ---

func (p *Parser) parseForStatement() *ast.ForStatement {
	stmt := &ast.ForStatement{Token: p.curToken}

	p.nextToken()
	stmt.Count = p.parseExpression(LOWEST)

	p.nextToken()
	if p.curToken.Type == token.LBRACE {
		stmt.Body = p.parseBlockStatement()
	} else {
		stmt.Body = &ast.BlockStatement{
			Token: p.curToken,
			Statements: []ast.Statement{
				p.parseStatement(),
			},
		}
	}

	return stmt
}

// --- While Statement ---

func (p *Parser) parseWhileStatement() *ast.WhileStatement {
	stmt := &ast.WhileStatement{Token: p.curToken}

	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)

	p.nextToken()
	if p.curToken.Type == token.LBRACE {
		stmt.Body = p.parseBlockStatement()
	} else {
		stmt.Body = &ast.BlockStatement{
			Token: p.curToken,
			Statements: []ast.Statement{
				p.parseStatement(),
			},
		}
	}

	return stmt
}

// --- Function Statement ---

func (p *Parser) parseFunctionStatement() *ast.FunctionStatement {
	stmt := &ast.FunctionStatement{Token: p.curToken}

	p.nextToken()
	if p.curToken.Type != token.IDENT {
		msg := fmt.Sprintf("baris %d: mengharapkan nama fungsi", p.curToken.Line)
		p.errors = append(p.errors, msg)
		p.skipToNextStatement()
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	p.nextToken()
	stmt.Params = p.parseFunctionParams()

	p.nextToken()
	if p.curToken.Type == token.LBRACE {
		stmt.Body = p.parseBlockStatement()
	} else {
		stmt.Body = &ast.BlockStatement{
			Token: p.curToken,
			Statements: []ast.Statement{
				p.parseStatement(),
			},
		}
	}

	return stmt
}

func (p *Parser) parseFunctionParams() []*ast.Identifier {
	var params []*ast.Identifier

	if p.curToken.Type != token.LPAREN {
		return params
	}

	p.nextToken() // skip LPAREN
	if p.curToken.Type == token.RPAREN {
		return params
	}

	// First parameter
	if p.curToken.Type == token.IDENT {
		params = append(params, &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal})
		p.nextToken()
	}

	// Additional parameters separated by commas
	for p.curToken.Type == token.COMMA {
		p.nextToken() // skip comma
		if p.curToken.Type == token.IDENT {
			params = append(params, &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal})
			p.nextToken()
		}
	}

	return params
}

// --- Call Statement ---

func (p *Parser) parseCallStatement() *ast.CallStatement {
	stmt := &ast.CallStatement{Token: p.curToken}

	// Do NOT advance — parseCallExpression expects curToken = PANGGIL
	stmt.CallExpr = p.parseCallExpression().(*ast.CallExpression)

	return stmt
}

func (p *Parser) parseCallExpression() ast.Expression {
	expr := &ast.CallExpression{Token: p.curToken}

	// Parse function name
	p.nextToken()
	if p.curToken.Type != token.IDENT {
		msg := fmt.Sprintf("baris %d: mengharapkan nama fungsi setelah 'panggil'", p.curToken.Line)
		p.errors = append(p.errors, msg)
		return expr
	}
	expr.Function = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Parse arguments
	p.nextToken()
	if p.curToken.Type == token.LPAREN {
		p.nextToken() // skip past LPAREN
		expr.Arguments = p.parseExpressionList(token.RPAREN)
		// consume RPAREN
		if p.curToken.Type == token.RPAREN {
			// already consumed by parseExpressionList
		} else if p.peekTokenIs(token.RPAREN) {
			p.nextToken()
		}
	} else {
		// Single argument without parens (backward compat)
		if p.curToken.Type != token.RBRACE &&
			p.curToken.Type != token.NEWLINE &&
			p.curToken.Type != token.EOF &&
			!isStatementStart(p.curToken.Type) {
			expr.Arguments = []ast.Expression{p.parseExpression(LOWEST)}
		}
	}

	return expr
}

func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	var args []ast.Expression

	// curToken is already past LPAREN, at first argument or end token
	if p.curToken.Type == end {
		return args
	}

	// Parse first argument
	args = append(args, p.parseExpression(LOWEST))

	// Handle comma-separated arguments
	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // skip comma
		p.nextToken() // move to next expression
		if p.curToken.Type == end || p.curToken.Type == token.EOF {
			break
		}
		args = append(args, p.parseExpression(LOWEST))
	}

	// Consume the end token if curToken isn't already at it
	if p.curToken.Type != end && p.peekTokenIs(end) {
		p.nextToken()
	}

	return args
}

// --- Return Statement ---

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()
	stmt.ReturnValue = p.parseExpression(LOWEST)

	return stmt
}

// --- Break / Continue ---

func (p *Parser) parseBreakStatement() *ast.BreakStatement {
	return &ast.BreakStatement{Token: p.curToken}
}

func (p *Parser) parseContinueStatement() *ast.ContinueStatement {
	return &ast.ContinueStatement{Token: p.curToken}
}

// --- Input ---

func (p *Parser) parseInputStatement() *ast.InputStatement {
	stmt := &ast.InputStatement{Token: p.curToken}

	p.nextToken()
	if p.curToken.Type == token.IDENT {
		stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	}

	return stmt
}

func (p *Parser) parseInputExpression() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: "__input__"}
}

// --- Expression Statement ---

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	stmt.Expression = p.parseExpression(LOWEST)

	return stmt
}

// --- Expression Parsing (Pratt) ---

func (p *Parser) parseExpression(precedence Precedence) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(token.EOF) && !p.peekTokenIs(token.RBRACE) &&
		!p.peekTokenIs(token.RPAREN) &&
		!isStatementStart(p.peekToken.Type) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

// --- Prefix Parse Functions ---

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("baris %d: tidak bisa parse '%s' sebagai integer",
			p.curToken.Line, p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parseFloatLiteral() ast.Expression {
	lit := &ast.FloatLiteral{Token: p.curToken}

	value, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		msg := fmt.Sprintf("baris %d: tidak bisa parse '%s' sebagai float",
			p.curToken.Line, p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseBooleanLiteral() ast.Expression {
	return &ast.BooleanLiteral{
		Token: p.curToken,
		Value: p.curToken.Type == token.BENAR,
	}
}

func (p *Parser) parseNilLiteral() ast.Expression {
	return &ast.NilLiteral{Token: p.curToken}
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expr := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()
	expr.Right = p.parseExpression(PREFIX)

	return expr
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	expr := p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return expr
}

// --- Infix Parse Functions ---

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expr := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expr.Right = p.parseExpression(precedence)

	return expr
}

// --- Helpers ---

func isStatementStart(t token.TokenType) bool {
	return t == token.BUAT || t == token.TULIS || t == token.JIKA ||
		t == token.ULANG || t == token.SELAMA || t == token.FUNGSI ||
		t == token.PANGGIL || t == token.KEMBALI || t == token.BERHENTI ||
		t == token.LANJUTKAN || t == token.MEMBACA || t == token.KALAU_TIDAK ||
		t == token.RBRACE || t == token.EOF
}

// Next returns the next token (backward compatibility).
func (p *Parser) Next() token.Token {
	return p.l.nextToken()
}

