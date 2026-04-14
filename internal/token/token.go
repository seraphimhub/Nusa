package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
}

const (
	ILLEGAL TokenType = "ILLEGAL"
	EOF     TokenType = "EOF"

	BUAT   TokenType = "BUAT"
	TULIS  TokenType = "TULIS"
	JIKA   TokenType = "JIKA"
	ULANG  TokenType = "ULANG"

	IDENT   TokenType = "IDENT"
	STRING  TokenType = "STRING"
	NUMBER  TokenType = "NUMBER"
	EQUAL   TokenType = "EQUAL"
	GREATER TokenType = "GREATER"
)

var Keywords = map[string]TokenType{
	"buat":  BUAT,
	"tulis": TULIS,
	"jika":  JIKA,
	"ulang": ULANG,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := Keywords[ident]; ok {
		return tok
	}
	return IDENT
}
