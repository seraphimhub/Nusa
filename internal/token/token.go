package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

const (
	ILLEGAL TokenType = "ILLEGAL"
	EOF     TokenType = "EOF"

	// Keywords
	BUAT        TokenType = "BUAT"
	TULIS       TokenType = "TULIS"
	JIKA        TokenType = "JIKA"
	KALAU_TIDAK TokenType = "KALAU_TIDAK"
	ULANG       TokenType = "ULANG"
	SELAMA      TokenType = "SELAMA"
	UNTUK       TokenType = "UNTUK"
	DARI        TokenType = "DARI"
	SAMPAI      TokenType = "SAMPAI"
	LANGKAH     TokenType = "LANGKAH"
	DALAM       TokenType = "DALAM"
	TETAP       TokenType = "TETAP"
	PETA        TokenType = "PETA"
	COBA        TokenType = "COBA"
	TANGKAP     TokenType = "TANGKAP"
	LEMPAR      TokenType = "LEMPAR"
	PILIH       TokenType = "PILIH"
	KASUS       TokenType = "KASUS"
	BAWAAN      TokenType = "BAWAAN"
	FUNGSI      TokenType = "FUNGSI"
	PANGGIL     TokenType = "PANGGIL"
	KEMBALI     TokenType = "KEMBALI"
	MEMBACA     TokenType = "MEMBACA"
	BERHENTI    TokenType = "BERHENTI"
	LANJUTKAN   TokenType = "LANJUTKAN"
	DAN         TokenType = "DAN"
	ATAU        TokenType = "ATAU"
	TIDAK       TokenType = "TIDAK"
	BENAR       TokenType = "BENAR"
	SALAH       TokenType = "SALAH"
	KOSONG      TokenType = "KOSONG"

	// Identifiers & literals
	IDENT  TokenType = "IDENT"
	STRING TokenType = "STRING"
	NUMBER TokenType = "NUMBER"
	FLOAT  TokenType = "FLOAT"

	// Operators
	ASSIGN          TokenType = "ASSIGN"
	PLUS_ASSIGN     TokenType = "PLUS_ASSIGN"
	MINUS_ASSIGN    TokenType = "MINUS_ASSIGN"
	MULTIPLY_ASSIGN TokenType = "MULTIPLY_ASSIGN"
	DIVIDE_ASSIGN   TokenType = "DIVIDE_ASSIGN"
	MODULO_ASSIGN   TokenType = "MODULO_ASSIGN"
	GREATER         TokenType = "GREATER"
	LESS            TokenType = "LESS"
	GTE             TokenType = "GTE"
	LTE             TokenType = "LTE"
	EQ              TokenType = "EQ"
	NEQ             TokenType = "NEQ"
	PLUS            TokenType = "PLUS"
	MINUS           TokenType = "MINUS"
	MULTIPLY        TokenType = "MULTIPLY"
	DIVIDE          TokenType = "DIVIDE"
	BANG            TokenType = "BANG"
	MODULO          TokenType = "MODULO"

	// Delimiters
	LPAREN    TokenType = "LPAREN"
	RPAREN    TokenType = "RPAREN"
	LBRACE    TokenType = "LBRACE"
	RBRACE    TokenType = "RBRACE"
	LBRACKET  TokenType = "LBRACKET"
	RBRACKET  TokenType = "RBRACKET"
	COMMA     TokenType = "COMMA"
	TITIK_DUA TokenType = "TITIK_DUA"
	NEWLINE   TokenType = "NEWLINE"
)

var Keywords = map[string]TokenType{
	"buat":        BUAT,
	"tulis":       TULIS,
	"jika":        JIKA,
	"kalau_tidak": KALAU_TIDAK,
	"ulang":       ULANG,
	"selama":      SELAMA,
	"untuk":       UNTUK,
	"dari":        DARI,
	"sampai":      SAMPAI,
	"langkah":     LANGKAH,
	"dalam":       DALAM,
	"tetap":       TETAP,
	"peta":        PETA,
	"coba":        COBA,
	"tangkap":     TANGKAP,
	"lemparkan":   LEMPAR,
	"pilih":       PILIH,
	"kasus":       KASUS,
	"bawaan":      BAWAAN,
	"fungsi":      FUNGSI,
	"panggil":     PANGGIL,
	"kembali":     KEMBALI,
	"membaca":     MEMBACA,
	"berhenti":    BERHENTI,
	"lanjutkan":   LANJUTKAN,
	"dan":         DAN,
	"atau":        ATAU,
	"tidak":       TIDAK,
	"benar":       BENAR,
	"salah":       SALAH,
	"kosong":      KOSONG,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := Keywords[ident]; ok {
		return tok
	}
	return IDENT
}
