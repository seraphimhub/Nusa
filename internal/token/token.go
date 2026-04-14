package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
}

const (
	ILLEGAL TokenType = "ILLEGAL"
	EOF     TokenType = "EOF"

	BUAT    TokenType = "BUAT"
	TULIS   TokenType = "TULIS"
	JIKA    TokenType = "JIKA"
	ULANG   TokenType = "ULANG"
	FUNGSI  TokenType = "FUNGSI"
	PANGGIL TokenType = "PANGGIL"

	IDENT   TokenType = "IDENT"
	STRING  TokenType = "STRING"
	NUMBER  TokenType = "NUMBER"

	EQUAL   TokenType = "EQUAL"
	GREATER TokenType = "GREATER"
	LESS    TokenType = "LESS"

	PLUS     TokenType = "PLUS"
	MINUS    TokenType = "MINUS"
	MULTIPLY TokenType = "MULTIPLY"
	DIVIDE   TokenType = "DIVIDE"

	GTE TokenType = "GTE"
	LTE TokenType = "LTE"
	EQ  TokenType = "EQ"
	NEQ TokenType = "NEQ"
)

var Keywords = map[string]TokenType{
	"buat":    BUAT,
	"tulis":   TULIS,
	"jika":    JIKA,
	"ulang":   ULANG,
	"fungsi":  FUNGSI,
	"panggil": PANGGIL,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := Keywords[ident]; ok {
		return tok
	}
	return IDENT
}
