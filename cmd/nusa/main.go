package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/seraphimhub/Nusa/internal/lexer"
	"github.com/seraphimhub/Nusa/internal/modul"
	"github.com/seraphimhub/Nusa/internal/parser"
	"github.com/seraphimhub/Nusa/internal/runtime"
	"github.com/seraphimhub/Nusa/internal/token"
)

const VERSION = "1.2.0"

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		printHelp()
		return 0
	}

	command := args[0]

	switch command {
	case "run":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "pakai: nusa run <file.nusa>")
			return 1
		}
		return runFile(args[1])
	case "repl":
		return runREPL()
	case "ast":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "pakai: nusa ast <file.nusa>")
			return 1
		}
		return dumpAST(args[1])
	case "token", "tokens":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "pakai: nusa token <file.nusa>")
			return 1
		}
		return dumpTokens(args[1])
	case "cek", "check":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "pakai: nusa cek <file.nusa> [file.nusa ...]")
			return 1
		}
		return cekFiles(args[1:])
	case "fmt", "format":
		return runFmt(args[1:])
	case "uji", "test":
		return runUji(args[1:])
	case "help", "--help", "-h":
		printHelp()
		return 0
	case "version", "--version", "-v":
		fmt.Printf("Nusa v%s\n", VERSION)
		return 0
	default:
		// Dukungan singkat: nusa program.nusa
		if strings.HasSuffix(command, ".nusa") {
			return runFile(command)
		}
		fmt.Fprintf(os.Stderr, "command tidak dikenal: %s\n", command)
		fmt.Fprintln(os.Stderr, "Gunakan 'nusa help' untuk bantuan")
		return 1
	}
}

func readSource(filename string) (string, bool) {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: tidak bisa membaca file '%s': %v\n", filename, err)
		return "", false
	}
	return string(data), true
}

func runFile(filename string) int {
	source, ok := readSource(filename)
	if !ok {
		return 1
	}

	tokens := lexer.Tokenize(source)
	if reportLexErrors(tokens, source) {
		return 1
	}

	p := parser.New(tokens)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		fmt.Fprintln(os.Stderr, "Error parsing:")
		for _, err := range p.Errors() {
			fmt.Fprint(os.Stderr, "  "+formatGalatTeks(err, source))
		}
		return 1
	}

	rt := newRuntime(filename)
	result := rt.Execute(program)
	if result != nil && result.Type() == runtime.ERROR_OBJ {
		fmt.Fprint(os.Stderr, formatGalat(result.(*runtime.Error), source))
		return 1
	}

	return 0
}

// newRuntime membuat runtime dengan pemuat modul yang mencari berkas modul
// relatif terhadap direktori berkas utama.
func newRuntime(mainFile string) *runtime.Runtime {
	rt := runtime.New()

	basis := "."
	if mainFile != "" {
		basis = filepath.Dir(mainFile)
	}
	loader := modul.New(basis, os.Stdout, os.Stdin)
	rt.SetModuleLoader(loader.Import)

	return rt
}

func dumpAST(filename string) int {
	source, ok := readSource(filename)
	if !ok {
		return 1
	}

	tokens := lexer.Tokenize(source)
	if reportLexErrors(tokens, source) {
		return 1
	}

	p := parser.New(tokens)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		for _, err := range p.Errors() {
			fmt.Fprint(os.Stderr, formatGalatTeks(err, source))
		}
		return 1
	}

	fmt.Print(program.String())
	return 0
}

func dumpTokens(filename string) int {
	source, ok := readSource(filename)
	if !ok {
		return 1
	}

	tokens := lexer.Tokenize(source)
	illegal := false
	for _, tok := range tokens {
		if tok.Type == token.NEWLINE {
			continue
		}
		fmt.Printf("%3d:%-3d %-14s %q\n", tok.Line, tok.Column, tok.Type, tok.Literal)
		if tok.Type == token.ILLEGAL {
			illegal = true
		}
	}
	if illegal {
		return 1
	}
	return 0
}

func reportLexErrors(tokens []token.Token, source string) bool {
	found := false
	for _, tok := range tokens {
		if tok.Type == token.ILLEGAL {
			if !found {
				fmt.Fprintln(os.Stderr, "Error leksikal:")
			}
			found = true
			pesan := fmt.Sprintf("baris %d:%d: karakter tidak dikenal %q", tok.Line, tok.Column, tok.Literal)
			fmt.Fprint(os.Stderr, "   "+formatGalatTeks(pesan, source))
		}
	}
	return found
}

func runREPL() int {
	fmt.Println("Nusa REPL v" + VERSION)
	fmt.Println("Ketik 'bantuan' untuk daftar perintah, 'keluar' untuk exit")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	rt := runtime.NewWithIO(os.Stdout, reader)
	loader := modul.New(".", os.Stdout, reader)
	rt.SetModuleLoader(loader.Import)

	for {
		fmt.Print("nusa> ")
		source, ok := readLine(reader)
		if !ok {
			break
		}

		for needsMoreInput(source) {
			fmt.Print("...> ")
			next, ok := readLine(reader)
			if !ok {
				break
			}
			source += "\n" + next
		}

		trimmed := strings.TrimSpace(source)
		if trimmed == "" {
			continue
		}

		switch trimmed {
		case "keluar", "exit", "quit":
			fmt.Println("Sampai jumpa!")
			return 0
		case "bantuan", "help":
			printREPLHelp()
			continue
		}

		tokens := lexer.Tokenize(source)
		if reportLexErrors(tokens, source) {
			continue
		}

		p := parser.New(tokens)
		program := p.ParseProgram()
		if len(p.Errors()) > 0 {
			for _, err := range p.Errors() {
				fmt.Fprint(os.Stderr, formatGalatTeks(err, source))
			}
			continue
		}

		result := rt.Execute(program)
		if result == nil {
			continue
		}
		if result.Type() == runtime.ERROR_OBJ {
			fmt.Fprint(os.Stderr, formatGalat(result.(*runtime.Error), source))
			continue
		}
		// Tampilkan nilai ekspresi yang dievaluasi di REPL.
		if result.Type() != runtime.NIL_OBJ {
			fmt.Println(result.Inspect())
		}
	}

	return 0
}

func readLine(reader *bufio.Reader) (string, bool) {
	line, err := reader.ReadString('\n')
	if err != nil && line == "" {
		return "", false
	}
	return strings.TrimRight(line, "\r\n"), true
}

// needsMoreInput mendeteksi blok/kurung/string yang belum ditutup agar
// REPL bisa meminta baris lanjutan.
func needsMoreInput(source string) bool {
	depth := 0
	var quote byte
	inBlockComment := false

	for i := 0; i < len(source); i++ {
		ch := source[i]

		if inBlockComment {
			if ch == '*' && i+1 < len(source) && source[i+1] == '/' {
				inBlockComment = false
				i++
			}
			continue
		}

		if quote != 0 {
			if ch == '\\' {
				i++
				continue
			}
			if ch == quote {
				quote = 0
			}
			continue
		}

		switch ch {
		case '#':
			for i < len(source) && source[i] != '\n' {
				i++
			}
		case '/':
			if i+1 < len(source) && source[i+1] == '/' {
				for i < len(source) && source[i] != '\n' {
					i++
				}
			} else if i+1 < len(source) && source[i+1] == '*' {
				inBlockComment = true
				i++
			}
		case '"', '\'':
			quote = ch
		case '{', '[':
			depth++
		case '}', ']':
			depth--
		}
	}

	return depth > 0 || quote != 0 || inBlockComment
}

func printREPLHelp() {
	fmt.Println(`PERINTAH REPL:
  keluar, exit, quit   Keluar dari REPL
  bantuan, help        Tampilkan bantuan ini

Contoh:
  buat nama = "Budi"
  tulis "Halo " + nama
  untuk i dari 1 sampai 3 { tulis i }
  panjang([1, 2, 3])`)
}

func printHelp() {
	fmt.Println(`Nusa - Bahasa Pemrograman Modern dengan Sintaks Indonesia

PENGGUNAAN:
  nusa run <file>     Jalankan file .nusa
  nusa <file>         Singkatan untuk 'nusa run <file>'
  nusa repl           Mode interaktif REPL (mendukung input multi-baris)
  nusa cek <file>     Pemeriksaan statis (tipe, variabel tak terdefinisi)
  nusa fmt <file>     Format kode (komentar tetap terjaga)
  nusa uji <file>     Jalankan blok uji pada file
  nusa ast <file>     Tampilkan pohon sintaks (AST)
  nusa token <file>   Tampilkan daftar token
  nusa help           Tampilkan bantuan ini
  nusa version        Tampilkan versi

CONTOH:
  nusa run examples/fitur.nusa
  nusa examples/daftar.nusa
  nusa cek examples/tipe.nusa
  nusa fmt examples/halo.nusa
  nusa uji examples/matematika_uji.nusa
  nusa repl

FITUR BAHASA:
  buat        - Deklarasi variabel      (buat x = 10)
  tetap       - Deklarasi konstanta      (tetap PI = 3.14)
  tulis       - Cetak output            (tulis "halo")
  jika        - Percabangan             (jika x > 5 { ... })
  kalau_tidak - Else / else-if          (kalau_tidak { ... })
  ulang       - Ulang sebanyak n kali   (ulang 5 { ... })
  selama      - While loop              (selama x < 10 { ... })
  untuk       - For range               (untuk i dari 1 sampai 5 { ... })
  untuk ... dalam - Iterasi daftar      (untuk x dalam daftar { ... })
  pilih       - Pencocokan nilai        (pilih x { kasus 1 { ... } bawaan { ... } })
  coba/tangkap - Tangani galat          (coba { ... } tangkap e { ... })
  lemparkan   - Buat galat              (lempar "pesan")
  impor       - Muat modul              (impor "util.nusa")
  uji/cek     - Tulis pengujian         (uji "nama" { cek x == 1 })
  fungsi      - Definisi fungsi         (fungsi nama(a, b) { ... })
  panggil     - Panggil fungsi          (panggil nama(1, 2) atau nama(1, 2))
  kembali     - Return value            (kembali x)
  membaca     - Input dari user         (membaca nama)
  berhenti    - Break dari loop         (berhenti)
  lanjutkan   - Continue loop            (lanjutkan)

TIPE DATA:
  angka, desimal, teks, benar/salah, kosong, daftar ([1, 2, 3]), peta

ANOTASI TIPE (opsional):
  buat umur: angka = 25
  fungsi luas(sisi: angka): angka { kembali sisi * sisi }

OPERATOR:
  +, -, *, /, %, >, <, >=, <=, ==, !=, tidak, dan, atau
  +=, -=, *=, /=, %=

FUNGSI BAWAAN:
  panjang, tipe, pesan, jejak, ada, kunci, nilai, ambil,
  ke_teks, ke_angka, ke_desimal,
  huruf_besar, huruf_kecil, potong, cari, pisah, gabung,
  tambahkan, hapus, urutkan, balik, jumlah, rentang,
  akar, pangkat, bulat, lantai, langit, mutlak, kudiag

KOMENTAR:
  // komentar satu baris
  # komentar satu baris
  /* komentar blok */`)
}
