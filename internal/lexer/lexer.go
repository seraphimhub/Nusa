package lexer

import (
	"strings"

	"github.com/seraphimhub/Nusa/internal/token"
)

type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           byte
	line         int
	lineStart    int
}

func New(input string) *Lexer {
	l := &Lexer{input: input, line: 1}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
		// Pertahankan position di akhir input agar perhitungan kolom
		// dan pengecekan baris tidak melewati batas slice.
		l.position = len(l.input)
	} else {
		l.ch = l.input[l.readPosition]
		l.position = l.readPosition
		l.readPosition++
	}

	// Catat awal baris agar nomor kolom bisa dihitung.
	if l.position > 0 && l.position <= len(l.input) && l.input[l.position-1] == '\n' {
		l.lineStart = l.position
	}
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) column() int {
	col := l.position - l.lineStart + 1
	if col < 1 {
		return 1
	}
	return col
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) skipLineComment() {
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
}

func (l *Lexer) skipBlockComment() {
	l.readChar() // '/'
	l.readChar() // '*'
	for l.ch != 0 {
		if l.ch == '*' && l.peekChar() == '/' {
			l.readChar()
			l.readChar()
			return
		}
		if l.ch == '\n' {
			l.line++
		}
		l.readChar()
	}
}

func (l *Lexer) skipIgnored() {
	for {
		l.skipWhitespace()

		if l.ch == '#' || (l.ch == '/' && l.peekChar() == '/') {
			l.skipLineComment()
			continue
		}
		if l.ch == '/' && l.peekChar() == '*' {
			l.skipBlockComment()
			continue
		}
		return
	}
}

func (l *Lexer) NextToken() token.Token {
	l.skipIgnored()

	line := l.line
	col := l.column()

	makeToken := func(typ token.TokenType, lit string) token.Token {
		return token.Token{Type: typ, Literal: lit, Line: line, Column: col}
	}

	// newline sengaja ditangani sendiri agar nomor baris tetap akurat.
	if l.ch == '\n' {
		l.readChar()
		l.line++
		return token.Token{Type: token.NEWLINE, Literal: "\n", Line: line, Column: col}
	}
	if l.ch == 0 {
		return makeToken(token.EOF, "")
	}

	if isLetter(l.ch) {
		lit := l.readIdentifier()
		return makeToken(token.LookupIdent(lit), lit)
	}
	if isDigit(l.ch) {
		typ, lit := l.readNumber()
		return makeToken(typ, lit)
	}

	var tok token.Token

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			l.readChar()
			tok = makeToken(token.EQ, "==")
		} else {
			tok = makeToken(token.ASSIGN, "=")
		}
	case '!':
		if l.peekChar() == '=' {
			l.readChar()
			tok = makeToken(token.NEQ, "!=")
		} else {
			tok = makeToken(token.BANG, "!")
		}
	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			tok = makeToken(token.GTE, ">=")
		} else {
			tok = makeToken(token.GREATER, ">")
		}
	case '<':
		if l.peekChar() == '=' {
			l.readChar()
			tok = makeToken(token.LTE, "<=")
		} else {
			tok = makeToken(token.LESS, "<")
		}
	case '+':
		if l.peekChar() == '=' {
			l.readChar()
			tok = makeToken(token.PLUS_ASSIGN, "+=")
		} else {
			tok = makeToken(token.PLUS, "+")
		}
	case '-':
		if l.peekChar() == '=' {
			l.readChar()
			tok = makeToken(token.MINUS_ASSIGN, "-=")
		} else {
			tok = makeToken(token.MINUS, "-")
		}
	case '*':
		if l.peekChar() == '=' {
			l.readChar()
			tok = makeToken(token.MULTIPLY_ASSIGN, "*=")
		} else {
			tok = makeToken(token.MULTIPLY, "*")
		}
	case '/':
		if l.peekChar() == '=' {
			l.readChar()
			tok = makeToken(token.DIVIDE_ASSIGN, "/=")
		} else {
			tok = makeToken(token.DIVIDE, "/")
		}
	case '%':
		if l.peekChar() == '=' {
			l.readChar()
			tok = makeToken(token.MODULO_ASSIGN, "%=")
		} else {
			tok = makeToken(token.MODULO, "%")
		}
	case '(':
		tok = makeToken(token.LPAREN, "(")
	case ')':
		tok = makeToken(token.RPAREN, ")")
	case '{':
		tok = makeToken(token.LBRACE, "{")
	case '}':
		tok = makeToken(token.RBRACE, "}")
	case '[':
		tok = makeToken(token.LBRACKET, "[")
	case ']':
		tok = makeToken(token.RBRACKET, "]")
	case ',':
		tok = makeToken(token.COMMA, ",")
	case ':':
		tok = makeToken(token.TITIK_DUA, ":")
	case '"', '\'':
		quote := l.ch
		lit := l.readString(quote)
		tok = makeToken(token.STRING, lit)
	default:
		ch := l.ch
		tok = makeToken(token.ILLEGAL, string(ch))
	}

	l.readChar()
	return tok
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readNumber() (token.TokenType, string) {
	position := l.position
	isFloat := false
	for isDigit(l.ch) {
		l.readChar()
	}
	if l.ch == '.' && isDigit(l.peekChar()) {
		isFloat = true
		l.readChar()
		for isDigit(l.ch) {
			l.readChar()
		}
	}
	if isFloat {
		return token.FLOAT, l.input[position:l.position]
	}
	return token.NUMBER, l.input[position:l.position]
}

func (l *Lexer) readString(quote byte) string {
	var out strings.Builder
	l.readChar() // lewati kutip pembuka

	for l.ch != quote && l.ch != 0 {
		if l.ch == '\\' {
			l.readChar()
			switch l.ch {
			case 'n':
				out.WriteByte('\n')
			case 't':
				out.WriteByte('\t')
			case 'r':
				out.WriteByte('\r')
			case '\\':
				out.WriteByte('\\')
			case '"':
				out.WriteByte('"')
			case '\'':
				out.WriteByte('\'')
			case '0':
				out.WriteByte(0)
			case 0:
				return out.String()
			default:
				out.WriteByte('\\')
				out.WriteByte(l.ch)
			}
			l.readChar()
			continue
		}

		if l.ch == '\n' {
			l.line++
		}
		out.WriteByte(l.ch)
		l.readChar()
	}

	return out.String()
}

func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

// Tokenize returns all tokens at once.
func Tokenize(input string) []token.Token {
	l := New(input)
	var tokens []token.Token
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == token.EOF {
			break
		}
	}
	return tokens
}
