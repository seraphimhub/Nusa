package parser

import "github.com/kamu/nusa/internal/token"

type Parser struct {
	Tokens []token.Token
	Pos    int
}

func New(tokens []token.Token) *Parser {
	return &Parser{
		Tokens: tokens,
		Pos:    0,
	}
}

func (p *Parser) Next() token.Token {
	if p.Pos >= len(p.Tokens) {
		return token.Token{Type: token.EOF}
	}
	tok := p.Tokens[p.Pos]
	p.Pos++
	return tok
}