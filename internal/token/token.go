package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
}

const (
	ILLEGAL TokenType = "ILLEGAL"
	EOF     TokenType = "EOF"

	BUAT  TokenType = "BUAT"
	TULIS TokenType = "TULIS"

	IDENT  TokenType = "IDENT"
	STRING TokenType = "STRING"
	EQUAL  TokenType = "EQUAL"
)

var Keywords = map[string]TokenType{
	"buat":  BUAT,
	"tulis": TULIS,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := Keywords[ident]; ok {
		return tok
	}
	return IDENT
}