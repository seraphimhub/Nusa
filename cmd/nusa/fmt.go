package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/seraphimhub/Nusa/internal/format"
)

// fmtOptions adalah pilihan perintah fmt.
type fmtOptions struct {
	// tulis menyimpan hasil ke berkas aslinya.
	tulis bool
	// cek hanya memeriksa apakah berkas sudah rapi; berkas tidak diubah
	// dan exit code 1 bila ada berkas yang belum rapi.
	cek bool
	// jalur berisi berkas atau direktori yang akan diproses.
	jalur []string
}

// runFmt mengurai pilihan perintah fmt lalu menjalankannya.
func runFmt(args []string) int {
	opts := fmtOptions{}

	for _, arg := range args {
		switch arg {
		case "-w", "--tulis":
			opts.tulis = true
		case "--cek", "--check":
			opts.cek = true
		case "-h", "--help":
			printFmtHelp()
			return 0
		default:
			if strings.HasPrefix(arg, "-") {
				fmt.Fprintf(os.Stderr, "error: opsi tidak dikenal: %s\n", arg)
				return 1
			}
			opts.jalur = append(opts.jalur, arg)
		}
	}

	if len(opts.jalur) == 0 {
		fmt.Fprintln(os.Stderr, "pakai: nusa fmt [-w] [--cek] <file.nusa|direktori>")
		return 1
	}
	if opts.cek && opts.tulis {
		fmt.Fprintln(os.Stderr, "error: --cek dan -w tidak bisa dipakai bersamaan")
		return 1
	}

	return fmtFiles(opts)
}

// printFmtHelp menampilkan bantuan perintah fmt.
func printFmtHelp() {
	fmt.Println(`Pakai: nusa fmt [-w] [--cek] <file.nusa|direktori>

Opsi:
  -w, --tulis    Simpan hasil ke berkas aslinya
  --cek          Hanya periksa; keluar dengan kode 1 bila berkas belum rapi
  -h, --help     Tampilkan bantuan ini

Tanpa opsi, hasil format dicetak ke layar.`)
}

// fmtFiles memformat berkas atau direktori yang diberikan.
func fmtFiles(opts fmtOptions) int {
	berkas, err := kumpulkanBerkas(opts.jalur)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	if len(berkas) == 0 {
		fmt.Fprintln(os.Stderr, "error: tidak ada berkas .nusa yang ditemukan")
		return 1
	}

	adaGalat := false
	adaBelumRapi := false

	for _, nama := range berkas {
		data, err := os.ReadFile(nama)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: tidak bisa membaca file '%s': %v\n", nama, err)
			adaGalat = true
			continue
		}

		sumber := string(data)
		hasil := format.Format(sumber)

		if opts.cek {
			if hasil != sumber {
				fmt.Fprintf(os.Stderr, "%s: berkas belum rapi\n", nama)
				adaBelumRapi = true
			}
			continue
		}

		if opts.tulis {
			if hasil == sumber {
				continue
			}
			if err := os.WriteFile(nama, []byte(hasil), 0o644); err != nil {
				fmt.Fprintf(os.Stderr, "error: tidak bisa menulis file '%s': %v\n", nama, err)
				adaGalat = true
				continue
			}
			fmt.Printf("%s: sudah dirapikan\n", nama)
			continue
		}

		os.Stdout.WriteString(hasil)
	}

	if adaGalat {
		return 1
	}
	if adaBelumRapi {
		return 1
	}
	return 0
}

// kumpulkanBerkas collects .nusa files from the given paths.
func kumpulkanBerkas(jalur []string) ([]string, error) {
	var keluar []string

	for _, path := range jalur {
		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("tidak bisa membaca '%s': %v", path, err)
		}

		if !info.IsDir() {
			keluar = append(keluar, path)
			continue
		}

		entries, err := os.ReadDir(path)
		if err != nil {
			return nil, fmt.Errorf("tidak bisa membaca direktori '%s': %v", path, err)
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".nusa") {
				continue
			}
			keluar = append(keluar, filepath.Join(path, entry.Name()))
		}
	}

	sort.Strings(keluar)
	return keluar, nil
}
