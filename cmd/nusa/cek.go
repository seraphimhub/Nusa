package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/seraphimhub/Nusa/internal/cek"
	"github.com/seraphimhub/Nusa/internal/lexer"
	"github.com/seraphimhub/Nusa/internal/modul"
	"github.com/seraphimhub/Nusa/internal/parser"
)

// cekFiles menjalankan pemeriksa statis pada setiap berkas yang diberikan.
// Exit code 1 bila ditemukan galat; peringatan saja tetap exit code 0.
func cekFiles(filenames []string) int {
	adaGalat := false

	for _, filename := range filenames {
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

		temuan := cek.PeriksaDengan(program, cek.Opsi{
			Impor: resolverImpor(filename),
		})
		if len(temuan) == 0 {
			fmt.Printf("%s: tidak ada temuan\n", filename)
			continue
		}

		for _, t := range temuan {
			if t.Severity == cek.Galat {
				adaGalat = true
			}
			pesan := fmt.Sprintf("%s: %s", t.Severity, t.Pesan)
			fmt.Fprint(os.Stderr, formatPesan(pesan, t.Line, t.Column, source))
		}
	}

	if adaGalat {
		return 1
	}
	return 0
}

// resolverImpor membuat fungsi yang mencari nama yang diekspor modul
// relatif terhadap direktori berkas utama.
func resolverImpor(mainFile string) func(path string) ([]string, error) {
	basis := "."
	if mainFile != "" {
		basis = filepath.Dir(mainFile)
	}

	loader := modul.New(basis, os.Stdout, os.Stdin)
	return loader.Ekspor
}
