package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/seraphimhub/Nusa/internal/runtime"
)

// konteksGalat adalah jumlah baris sebelum dan sesudah baris galat yang
// ditampilkan sebagai cuplikan sumber.
const konteksGalat = 1

// formatGalat menyusun pesan galat runtime berikut cuplikan sumber yang
// menunjuk kolom penyebab. Bila sumber tidak tersedia atau baris galat di
// luar jangkauan, hanya pesan galat yang ditampilkan.
func formatGalat(err *runtime.Error, source string) string {
	if err == nil {
		return ""
	}
	return formatPesan(err.Inspect(), err.Line, err.Column, source)
}

// formatGalatTeks menyusun pesan galat yang sudah berbentuk teks, misalnya
// galat leksikal atau galat parsing. Nomor baris diambil dari awalan pesan
// sehingga galat ini tetap bisa ditampilkan bersama cuplikan sumber.
func formatGalatTeks(pesan string, source string) string {
	return formatPesan(pesan, nomorBarisPesan(pesan), 0, source)
}

// formatPesan merakit pesan galat dengan cuplikan sumbernya.
func formatPesan(pesan string, line int, column int, source string) string {
	var out strings.Builder
	out.WriteString(decorateErrorHeader(pesan))
	out.WriteString("\n")

	if cuplikan := cuplikanSumber(source, line, column); cuplikan != "" {
		out.WriteString(cuplikan)
	}

	return out.String()
}

// nomorBarisPesan mengambil nomor baris dari pesan galat berawalan
// "baris N:". Mengembalikan 0 bila pola tidak ditemukan.
func nomorBarisPesan(pesan string) int {
	rest, ok := strings.CutPrefix(pesan, "baris ")
	if !ok {
		return 0
	}
	digits, _, _ := strings.Cut(rest, ":")
	nomor, err := strconv.Atoi(strings.TrimSpace(digits))
	if err != nil {
		return 0
	}
	return nomor
}

// cuplikanSumber merender potongan berkas di sekitar baris galat beserta
// penanda caret pada kolom penyebab.
func cuplikanSumber(source string, line int, column int) string {
	if source == "" || line < 1 {
		return ""
	}

	baris := strings.Split(source, "\n")
	// Baris kosong di ujung berkas bukan bagian dari program, sehingga
	// tidak perlu ikut ditampilkan sebagai konteks.
	for len(baris) > 0 && strings.TrimSpace(baris[len(baris)-1]) == "" {
		baris = baris[:len(baris)-1]
	}
	if line > len(baris) {
		return ""
	}

	awal := line - konteksGalat
	if awal < 1 {
		awal = 1
	}
	akhir := line + konteksGalat
	if akhir > len(baris) {
		akhir = len(baris)
	}

	lebar := len(strconv.Itoa(akhir))
	var out strings.Builder

	for nomor := awal; nomor <= akhir; nomor++ {
		isi := strings.TrimRight(baris[nomor-1], "\r")
		fmt.Fprintf(&out, "%*d | %s\n", lebar, nomor, isi)

		if nomor != line {
			continue
		}
		// Penanda hanya digambar bila kolom diketahui; bila tidak, ujung
		// baris ditunjuk dengan panah pendek.
		if column < 1 {
			fmt.Fprintf(&out, "%*s | %s\n", lebar, "", "  ^")
			continue
		}
		if column > len(isi)+1 {
			column = len(isi) + 1
		}
		fmt.Fprintf(&out, "%*s | %s^\n", lebar, "", strings.Repeat(" ", column-1))
	}

	return out.String()
}

// decorateErrorHeader memberi warna merah pada kepala pesan galat bila
// keluaran adalah terminal sungguhan.
func decorateErrorHeader(pesan string) string {
	if !stderrIsTerminal() || !strings.HasPrefix(pesan, "ERROR:") {
		return pesan
	}
	const merah = "\x1b[31;1m"
	const reset = "\x1b[0m"
	return merah + pesan + reset
}

// stderrIsTerminal memeriksa apakah keluaran galat menulis ke terminal.
func stderrIsTerminal() bool {
	info, err := os.Stderr.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}
