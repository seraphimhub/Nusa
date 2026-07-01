package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/seraphimhub/Nusa/internal/lexer"
	"github.com/seraphimhub/Nusa/internal/parser"
	"github.com/seraphimhub/Nusa/internal/runtime"
)

const VERSION = "1.0.0"

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	command := os.Args[1]

	switch command {
	case "run":
		runFile()
	case "repl":
		runREPL()
	case "help", "--help", "-h":
		printHelp()
	case "version", "--version", "-v":
		fmt.Printf("Nusa v%s\n", VERSION)
	default:
		fmt.Printf("command tidak dikenal: %s\n", command)
		fmt.Println("Gunakan 'nusa help' untuk bantuan")
	}
}

func runFile() {
	if len(os.Args) < 3 {
		fmt.Println("pakai: nusa run <file.nusa>")
		return
	}

	filename := os.Args[2]

	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: tidak bisa membaca file '%s': %v\n", filename, err)
		os.Exit(1)
	}

	tokens := lexer.Tokenize(string(data))
	p := parser.New(tokens)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		fmt.Fprintln(os.Stderr, "Error parsing:")
		for _, err := range p.Errors() {
			fmt.Fprintln(os.Stderr, "  ", err)
		}
		os.Exit(1)
	}

	rt := runtime.New()
	rt.Execute(program)
}

func runREPL() {
	fmt.Println("Nusa REPL v" + VERSION)
	fmt.Println("Ketik 'keluar' untuk exit")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	rt := runtime.New()

	for {
		fmt.Print("nusa> ")
		if !scanner.Scan() {
			break
		}

		line := scanner.Text()
		if line == "keluar" {
			break
		}

		tokens := lexer.Tokenize(line)
		p := parser.New(tokens)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			for _, err := range p.Errors() {
				fmt.Fprintln(os.Stderr, err)
			}
			continue
		}

		rt.Execute(program)
	}
}

func printHelp() {
	fmt.Println(`Nusa - Bahasa Pemrograman Modern dengan Sintaks Indonesia

PENGGUNAAN:
  nusa run <file>     Jalankan file .nusa
  nusa repl           Mode interaktif REPL
  nusa help           Tampilkan bantuan ini
  nusa version        Tampilkan versi

CONTOH:
  nusa run contoh.nusa
  nusa repl

FITUR BAHASA:
  buat     - Deklarasi variabel    (buat x = 10)
  tulis    - Cetak output          (tulis "halo")
  jika     - Percabangan           (jika x > 5 { ... })
  kalau_tidak - Else               (kalau_tidak { ... })
  ulang    - Perulangan counter    (ulang 5 { ... })
  selama   - While loop            (selama x < 10 { ... })
  fungsi   - Definisi fungsi       (fungsi nama(a, b) { ... })
  panggil  - Panggil fungsi        (panggil nama(1, 2))
  kembali  - Return value          (kembali x)
  membaca  - Input dari user       (membaca nama)
  berhenti - Break dari loop       (berhenti)
  lanjutkan- Continue loop         (lanjutkan)

TIPE DATA:
  angka, desimal, teks, benar/salah, kosong

OPERATOR:
  +, -, *, /, >, <, >=, <=, ==, !=, tidak, dan, atau

KOMENTAR:
  // komentar satu baris
  # komentar juga
`)
}

