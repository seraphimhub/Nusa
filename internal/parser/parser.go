package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/seraphimhub/Nusa/internal/ast"
	"github.com/seraphimhub/Nusa/internal/tipe"
	"github.com/seraphimhub/Nusa/internal/token"
)

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

// Precedence levels
type Precedence int

const (
	_          Precedence = iota
	ASSIGNMENT            // =
	LOWEST                // default
	LOGICAL               // dan, atau
	COMPARISON            // ==, !=, >, <, >=, <=
	SUM                   // +, -
	PRODUCT               // *, /, %
	PREFIX                // !, -
	CALL                  // fn(x)
	INDEX                 // daftar[0]
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
	token.LBRACKET: INDEX,
	token.TITIK:    9,
}

type Parser struct {
	l         *lexerHelper
	curToken  token.Token
	peekToken token.Token
	errors    []string
	// pulih menandai bahwa kursor sudah diposisikan ulang oleh
	// skipToNextStatement sehingga caller tidak boleh menggesernya lagi.
	pulih bool

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
	p.registerPrefix(token.LBRACKET, p.parseListLiteral)
	p.registerPrefix(token.PETA, p.parsePetaLiteral)

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
	p.registerInfix(token.LPAREN, p.parseCallFromExpression)
	p.registerInfix(token.LBRACKET, p.parseIndexExpression)
	p.registerInfix(token.TITIK, p.parsePropertyExpression)

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
	p.pulih = true
	for p.curToken.Type != token.EOF {
		if p.curToken.Type == token.BUAT ||
			p.curToken.Type == token.TETAP ||
			p.curToken.Type == token.TULIS ||
			p.curToken.Type == token.JIKA ||
			p.curToken.Type == token.ULANG ||
			p.curToken.Type == token.SELAMA ||
			p.curToken.Type == token.UNTUK ||
			p.curToken.Type == token.PILIH ||
			p.curToken.Type == token.KASUS ||
			p.curToken.Type == token.BAWAAN ||
			p.curToken.Type == token.COBA ||
			p.curToken.Type == token.LEMPAR ||
			p.curToken.Type == token.IMPOR ||
			p.curToken.Type == token.UJI ||
			p.curToken.Type == token.CEK ||
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
		p.pulih = false
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		// Pemulihan galat sudah memposisikan kursor pada awal pernyataan
		// berikutnya, sehingga kursor tidak boleh digeser lagi.
		if !p.pulih {
			p.nextToken()
		}
	}

	return program
}

// parseStatement mengembalikan nil bila pernyataan gagal di-parse. Setiap
// hasil bertipe konkret harus diperiksa sebelum dikembalikan sebagai interface
// agar tidak terbentuk interface berisi pointer nil (typed nil) yang lolos
// pemeriksaan `stmt != nil` dan memicu panic saat AST dicetak.
func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.BUAT, token.TETAP:
		if stmt := p.parseLetStatement(); stmt != nil {
			return stmt
		}
	case token.TULIS:
		if stmt := p.parsePrintStatement(); stmt != nil {
			return stmt
		}
	case token.JIKA:
		if stmt := p.parseIfStatement(); stmt != nil {
			return stmt
		}
	case token.ULANG:
		if stmt := p.parseForStatement(); stmt != nil {
			return stmt
		}
	case token.SELAMA:
		if stmt := p.parseWhileStatement(); stmt != nil {
			return stmt
		}
	case token.UNTUK:
		if stmt := p.parseUntukStatement(); stmt != nil {
			return stmt
		}
	case token.PILIH:
		if stmt := p.parsePilihStatement(); stmt != nil {
			return stmt
		}
	case token.COBA:
		if stmt := p.parseCobaStatement(); stmt != nil {
			return stmt
		}
	case token.LEMPAR:
		if stmt := p.parseLemparkanStatement(); stmt != nil {
			return stmt
		}
	case token.FUNGSI:
		if stmt := p.parseFunctionStatement(); stmt != nil {
			return stmt
		}
	case token.PANGGIL:
		if stmt := p.parseCallStatement(); stmt != nil {
			return stmt
		}
	case token.KEMBALI:
		if stmt := p.parseReturnStatement(); stmt != nil {
			return stmt
		}
	case token.BERHENTI:
		if stmt := p.parseBreakStatement(); stmt != nil {
			return stmt
		}
	case token.LANJUTKAN:
		if stmt := p.parseContinueStatement(); stmt != nil {
			return stmt
		}
	case token.MEMBACA:
		if stmt := p.parseInputStatement(); stmt != nil {
			return stmt
		}
	case token.IMPOR:
		if stmt := p.parseImporStatement(); stmt != nil {
			return stmt
		}
	case token.UJI:
		if stmt := p.parseUjiStatement(); stmt != nil {
			return stmt
		}
	case token.CEK:
		if stmt := p.parseCekStatement(); stmt != nil {
			return stmt
		}
	case token.RBRACE:
		return nil
	default:
		if stmt := p.parseExpressionStatement(); stmt != nil {
			return stmt
		}
	}
	return nil
}

// --- Let Statement ---

func (p *Parser) parseLetStatement() *ast.LetStatement {
	keyword := p.curToken
	stmt := &ast.LetStatement{Token: keyword, Konst: keyword.Type == token.TETAP}

	p.nextToken()
	if p.curToken.Type != token.IDENT {
		msg := fmt.Sprintf("baris %d: mengharapkan identifier setelah '%s'", p.curToken.Line, keyword.Literal)
		p.errors = append(p.errors, msg)
		p.skipToNextStatement()
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Anotasi tipe opsional: buat nama: tipe = nilai
	if p.peekTokenIs(token.TITIK_DUA) {
		p.nextToken() // ke ':'
		typeHint, ok := p.parseTypeHint()
		if !ok {
			return nil
		}
		stmt.TypeHint = typeHint
		p.nextToken() // ke '='
	} else {
		p.nextToken() // ke '='
	}

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

// parseTypeHint membaca nama tipe. Token kursor harus berada di tanda ':'
// sebelum fungsi ini dipanggil, dan setelah selesai kursor berada di nama
// tipe. Sebagian nama tipe ('peta', 'kosong', 'fungsi') sekaligus merupakan
// kata kunci bahasa, sehingga tokennya perlu diterima secara khusus.
func (p *Parser) parseTypeHint() (string, bool) {
	p.nextToken() // ke nama tipe

	switch p.curToken.Type {
	case token.IDENT, token.FUNGSI, token.PETA, token.KOSONG:
	default:
		msg := fmt.Sprintf("baris %d: ekspresi tipe tidak dikenal: %q", p.curToken.Line, p.curToken.Literal)
		p.errors = append(p.errors, msg)
		p.skipToNextStatement()
		return "", false
	}

	name := p.curToken.Literal
	if !tipe.Dikenal(name) {
		msg := fmt.Sprintf("baris %d: tipe '%s' tidak dikenal; tipe yang tersedia: %s",
			p.curToken.Line, name, strings.Join(tipe.Semua(), ", "))
		p.errors = append(p.errors, msg)
		p.skipToNextStatement()
		return "", false
	}

	return name, true
}

// --- Print Statement ---

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
		stmt.Consequence = p.singleStatementBlock()
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
			stmt.Alternative = p.singleStatementBlock()
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
		p.pulih = false
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		if !p.pulih {
			p.nextToken()
		}
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
		stmt.Body = p.singleStatementBlock()
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
		stmt.Body = p.singleStatementBlock()
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

	// Tipe kembalian opsional: ): tipe {
	if p.curToken.Type == token.TITIK_DUA {
		returnType, ok := p.parseTypeHint()
		if !ok {
			return nil
		}
		stmt.ReturnType = returnType
		p.nextToken() // ke '{'
	}

	if p.curToken.Type == token.LBRACE {
		stmt.Body = p.parseBlockStatement()
	} else {
		stmt.Body = p.singleStatementBlock()
	}

	return stmt
}

func (p *Parser) parseFunctionParams() []*ast.Param {
	var params []*ast.Param

	if p.curToken.Type != token.LPAREN {
		return params
	}

	p.nextToken() // skip LPAREN
	if p.curToken.Type == token.RPAREN {
		return params
	}

	// First parameter
	if p.curToken.Type == token.IDENT {
		params = append(params, p.parseParam())
		p.nextToken()
	}

	// Additional parameters separated by commas
	for p.curToken.Type == token.COMMA {
		p.nextToken() // skip comma
		if p.curToken.Type == token.IDENT {
			params = append(params, p.parseParam())
			p.nextToken()
		}
	}

	return params
}

// parseParam membaca satu parameter beserta anotasi tipenya bila ada.
// Token kursor tetap berada di nama parameter saat fungsi ini selesai.
func (p *Parser) parseParam() *ast.Param {
	param := &ast.Param{
		Name: &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal},
	}

	if p.peekTokenIs(token.TITIK_DUA) {
		p.nextToken() // ke ':'
		typeHint, ok := p.parseTypeHint()
		if !ok {
			return param
		}
		param.TypeHint = typeHint
	}

	return param
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
		if p.curTokenIs(token.RPAREN) {
			return expr
		}
		expr.Arguments = p.parseExpressionList(token.RPAREN)
		if !p.expectPeek(token.RPAREN) {
			return nil
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

	// Handle comma-separated arguments. Token penutup sengaja tidak
	// dikonsumsi di sini agar pemanggilan bersarang seperti
	// jumlah(rentang(1, 5)) tidak memakai ')' milik call dalam.
	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // skip comma
		p.nextToken() // move to next expression
		if p.curToken.Type == end || p.curToken.Type == token.EOF {
			break
		}
		args = append(args, p.parseExpression(LOWEST))
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

// --- Impor ---

// parseImporStatement membaca pernyataan impor modul. Path modul ditulis
// sebagai teks, misalnya: impor "util.nusa"
func (p *Parser) parseImporStatement() *ast.ImporStatement {
	stmt := &ast.ImporStatement{Token: p.curToken}

	p.nextToken()
	if p.curToken.Type != token.STRING {
		msg := fmt.Sprintf("baris %d: nama berkas modul harus berupa teks setelah 'impor', mendapat '%s'",
			p.curToken.Line, p.curToken.Literal)
		p.errors = append(p.errors, msg)
		p.skipToNextStatement()
		return nil
	}
	stmt.Path = p.curToken.Literal

	return stmt
}

// --- Uji / Cek ---

// parseUjiStatement membaca blok uji: uji "nama" { ... }. Tubuh blok boleh
// berisi pernyataan cek dan pernyataan biasa lainnya.
func (p *Parser) parseUjiStatement() *ast.UjiStatement {
	stmt := &ast.UjiStatement{Token: p.curToken}

	p.nextToken()
	if p.curToken.Type != token.STRING {
		msg := fmt.Sprintf("baris %d: nama uji harus berupa teks setelah 'uji', mendapat '%s'",
			p.curToken.Line, p.curToken.Literal)
		p.errors = append(p.errors, msg)
		p.skipToNextStatement()
		return nil
	}
	stmt.Nama = p.curToken.Literal

	p.nextToken()
	if p.curToken.Type != token.LBRACE {
		msg := fmt.Sprintf("baris %d: blok uji harus dimulai dengan '{', mendapat '%s'",
			p.curToken.Line, p.curToken.Literal)
		p.errors = append(p.errors, msg)
		p.skipToNextStatement()
		return nil
	}
	stmt.Tubuh = p.parseBlockStatement()

	return stmt
}

// parseCekStatement membaca pernyataan cek: cek ekspresi [, "pesan"].
func (p *Parser) parseCekStatement() *ast.CekStatement {
	stmt := &ast.CekStatement{Token: p.curToken}

	p.nextToken()
	stmt.Ekspresi = p.parseExpression(LOWEST)
	if stmt.Ekspresi == nil {
		return nil
	}

	// Pesan kegagalan bersifat opsional dan ditulis setelah koma.
	if p.peekTokenIs(token.COMMA) {
		p.nextToken() // ke koma
		p.nextToken() // ke teks pesan
		if p.curToken.Type != token.STRING {
			msg := fmt.Sprintf("baris %d: pesan pada 'cek' harus berupa teks, mendapat '%s'",
				p.curToken.Line, p.curToken.Literal)
			p.errors = append(p.errors, msg)
			p.skipToNextStatement()
			return nil
		}
		stmt.Pesan = p.curToken.Literal
	}

	return stmt
}

// --- Expression Statement ---

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	expr := p.parseExpression(LOWEST)
	if expr == nil {
		return nil
	}

	// Penugasan ditangani di level statement agar target bisa berupa
	// identifier maupun indeks daftar: x = 1, daftar[0] += 2.
	if isAssignOperator(p.peekToken.Type) {
		p.nextToken()
		opTok := p.curToken

		p.nextToken()
		value := p.parseExpression(LOWEST)

		stmt.Expression = &ast.AssignExpression{
			Token:    opTok,
			Target:   expr,
			Operator: opTok.Literal,
			Value:    value,
		}
		return stmt
	}

	stmt.Expression = expr
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

// --- Daftar, Indeks, dan Pemanggilan ---

func (p *Parser) parseListLiteral() ast.Expression {
	list := &ast.ListLiteral{Token: p.curToken}
	list.Elements = []ast.Expression{}

	if p.peekTokenIs(token.RBRACKET) {
		p.nextToken()
		return list
	}

	p.nextToken()
	list.Elements = append(list.Elements, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // koma
		p.nextToken() // elemen berikutnya
		if p.curTokenIs(token.RBRACKET) {
			break
		}
		list.Elements = append(list.Elements, p.parseExpression(LOWEST))
	}

	if p.curTokenIs(token.RBRACKET) {
		return list
	}
	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return list
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	expr := &ast.IndexExpression{Token: p.curToken, Left: left}

	p.nextToken()
	expr.Index = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return expr
}

// parsePropertyExpression menghilangkan properti objek, misalnya obj.nama.
func (p *Parser) parsePropertyExpression(left ast.Expression) ast.Expression {
	expr := &ast.PropertyExpression{Token: p.curToken, Left: left}

	p.nextToken() // skip TITIK
	expr.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	return expr
}

func (p *Parser) parseCallFromExpression(fn ast.Expression) ast.Expression {
	expr := &ast.CallExpression{Token: p.curToken, Function: fn}

	p.nextToken()
	if p.curTokenIs(token.RPAREN) {
		return expr
	}

	expr.Arguments = p.parseExpressionList(token.RPAREN)
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return expr
}

// --- Untuk Statement ---

func (p *Parser) parseUntukStatement() ast.Statement {
	tok := p.curToken

	p.nextToken()
	if p.curToken.Type != token.IDENT {
		msg := fmt.Sprintf("baris %d: mengharapkan nama variabel setelah 'untuk'", p.curToken.Line)
		p.errors = append(p.errors, msg)
		p.skipToNextStatement()
		return nil
	}
	names := []*ast.Identifier{{Token: p.curToken, Value: p.curToken.Literal}}

	// Bentuk dua variabel: untuk k, v dalam peta { ... }
	if p.peekTokenIs(token.COMMA) {
		p.nextToken() // koma
		p.nextToken()
		if p.curToken.Type != token.IDENT {
			msg := fmt.Sprintf("baris %d: mengharapkan nama variabel kedua setelah ','", p.curToken.Line)
			p.errors = append(p.errors, msg)
			p.skipToNextStatement()
			return nil
		}
		names = append(names, &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal})
	}

	p.nextToken()
	switch p.curToken.Type {
	case token.DARI:
		if len(names) != 1 {
			msg := fmt.Sprintf("baris %d: 'untuk ... dari ...' hanya menerima satu variabel", p.curToken.Line)
			p.errors = append(p.errors, msg)
			p.skipToNextStatement()
			return nil
		}
		p.nextToken()
		start := p.parseExpression(LOWEST)

		if !p.expectPeek(token.SAMPAI) {
			return nil
		}
		p.nextToken()
		end := p.parseExpression(LOWEST)

		var step ast.Expression
		if p.peekTokenIs(token.LANGKAH) {
			p.nextToken()
			p.nextToken()
			step = p.parseExpression(LOWEST)
		}

		p.nextToken()
		return &ast.UntukStatement{
			Token: tok,
			Name:  names[0],
			Start: start,
			End:   end,
			Step:  step,
			Body:  p.parseBody(),
		}

	case token.DALAM:
		p.nextToken()
		iterable := p.parseExpression(LOWEST)

		p.nextToken()
		return &ast.UntukDalamStatement{
			Token:    tok,
			Names:    names,
			Iterable: iterable,
			Body:     p.parseBody(),
		}

	default:
		msg := fmt.Sprintf("baris %d: setelah nama variabel 'untuk' harus ada 'dari' atau 'dalam'", p.curToken.Line)
		p.errors = append(p.errors, msg)
		p.skipToNextStatement()
		return nil
	}
}

// --- Peta Literal ---

func (p *Parser) parsePetaLiteral() ast.Expression {
	lit := &ast.PetaLiteral{Token: p.curToken}

	if !p.peekTokenIs(token.LBRACE) {
		msg := fmt.Sprintf("baris %d: mengharapkan '{' setelah 'peta', mendapat '%s'",
			p.peekToken.Line, p.peekToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}
	p.nextToken() // kurung kurawal buka

	if p.peekTokenIs(token.RBRACE) {
		p.nextToken() // peta kosong
		return lit
	}

	p.nextToken()
	for {
		kunci := p.parseExpression(LOWEST)
		if !p.expectPeek(token.TITIK_DUA) {
			return nil
		}

		p.nextToken()
		nilai := p.parseExpression(LOWEST)
		lit.Entri = append(lit.Entri, ast.PetaEntri{Kunci: kunci, Nilai: nilai})

		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
			p.nextToken()
			if p.curTokenIs(token.RBRACE) {
				break
			}
			continue
		}
		break
	}

	if p.curTokenIs(token.RBRACE) {
		return lit
	}
	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	return lit
}

// --- Coba / Tangkap ---

func (p *Parser) parseCobaStatement() ast.Statement {
	stmt := &ast.CobaStatement{Token: p.curToken}

	p.nextToken()
	if p.curToken.Type != token.LBRACE {
		msg := fmt.Sprintf("baris %d: mengharapkan '{' setelah 'coba', mendapat '%s'",
			p.curToken.Line, p.curToken.Literal)
		p.errors = append(p.errors, msg)
		p.skipToNextStatement()
		return nil
	}
	stmt.Try = p.parseBlockStatement()

	if !p.peekTokenIs(token.TANGKAP) {
		msg := fmt.Sprintf("baris %d: blok 'coba' harus diikuti 'tangkap'", p.peekToken.Line)
		p.errors = append(p.errors, msg)
		p.skipToNextStatement()
		return nil
	}

	p.nextToken() // 'tangkap'
	p.nextToken()

	if p.curToken.Type == token.IDENT {
		stmt.ErrName = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		p.nextToken()
	}

	if p.curToken.Type != token.LBRACE {
		msg := fmt.Sprintf("baris %d: mengharapkan '{' setelah 'tangkap', mendapat '%s'",
			p.curToken.Line, p.curToken.Literal)
		p.errors = append(p.errors, msg)
		p.skipToNextStatement()
		return nil
	}
	stmt.Catch = p.parseBlockStatement()

	return stmt
}

func (p *Parser) parseLemparkanStatement() *ast.LemparkanStatement {
	stmt := &ast.LemparkanStatement{Token: p.curToken}

	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	return stmt
}

// --- Pilih Statement ---

func (p *Parser) parsePilihStatement() ast.Statement {
	stmt := &ast.PilihStatement{Token: p.curToken}

	p.nextToken()
	stmt.Subjek = p.parseExpression(LOWEST)

	p.nextToken()
	if p.curToken.Type != token.LBRACE {
		msg := fmt.Sprintf("baris %d: mengharapkan '{' setelah ekspresi 'pilih', mendapat '%s'",
			p.curToken.Line, p.curToken.Literal)
		p.errors = append(p.errors, msg)
		p.skipToNextStatement()
		return nil
	}

	p.nextToken()
	for p.curToken.Type != token.RBRACE && p.curToken.Type != token.EOF {
		switch p.curToken.Type {
		case token.KASUS:
			kasus := ast.KasusBlok{Token: p.curToken}

			p.nextToken()
			kasus.Nilai = p.parseExpression(LOWEST)

			p.nextToken()
			if p.curToken.Type != token.LBRACE {
				msg := fmt.Sprintf("baris %d: mengharapkan '{' setelah nilai 'kasus', mendapat '%s'",
					p.curToken.Line, p.curToken.Literal)
				p.errors = append(p.errors, msg)
				p.skipToNextStatement()
				return nil
			}
			kasus.Body = p.parseBlockStatement()
			stmt.Kasus = append(stmt.Kasus, kasus)

		case token.BAWAAN:
			p.nextToken()
			if p.curToken.Type != token.LBRACE {
				msg := fmt.Sprintf("baris %d: mengharapkan '{' setelah 'bawaan', mendapat '%s'",
					p.curToken.Line, p.curToken.Literal)
				p.errors = append(p.errors, msg)
				p.skipToNextStatement()
				return nil
			}
			stmt.Bawaan = p.parseBlockStatement()

		default:
			msg := fmt.Sprintf("baris %d: hanya 'kasus' atau 'bawaan' yang boleh berada di dalam 'pilih', mendapat '%s'",
				p.curToken.Line, p.curToken.Literal)
			p.errors = append(p.errors, msg)
			p.skipToNextStatement()
		}

		p.nextToken()
	}

	return stmt
}

// parseBody menerima blok { ... } atau satu statement tanpa kurung kurawal.
func (p *Parser) parseBody() *ast.BlockStatement {
	if p.curToken.Type == token.LBRACE {
		return p.parseBlockStatement()
	}

	return p.singleStatementBlock()
}

// singleStatementBlock membentuk blok berisi satu pernyataan. Pernyataan
// yang gagal di-parse menghasilkan blok kosong supaya tidak ada elemen
// kosong pada pohon sintaks.
func (p *Parser) singleStatementBlock() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken, Statements: []ast.Statement{}}
	if stmt := p.parseStatement(); stmt != nil {
		block.Statements = append(block.Statements, stmt)
	}
	return block
}

// --- Helpers ---

func isAssignOperator(t token.TokenType) bool {
	switch t {
	case token.ASSIGN, token.PLUS_ASSIGN, token.MINUS_ASSIGN,
		token.MULTIPLY_ASSIGN, token.DIVIDE_ASSIGN, token.MODULO_ASSIGN:
		return true
	}
	return false
}

func isStatementStart(t token.TokenType) bool {
	return t == token.BUAT || t == token.TETAP || t == token.TULIS || t == token.JIKA ||
		t == token.ULANG || t == token.SELAMA || t == token.UNTUK ||
		t == token.PILIH || t == token.KASUS || t == token.BAWAAN ||
		t == token.COBA || t == token.TANGKAP || t == token.LEMPAR || t == token.FUNGSI ||
		t == token.PANGGIL || t == token.KEMBALI || t == token.BERHENTI ||
		t == token.LANJUTKAN || t == token.MEMBACA || t == token.IMPOR ||
		t == token.KALAU_TIDAK ||
		t == token.RBRACE || t == token.EOF
}

// Next returns the next token (backward compatibility).
func (p *Parser) Next() token.Token {
	return p.l.nextToken()
}
