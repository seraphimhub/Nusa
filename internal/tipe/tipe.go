// Package tipe menyediakan daftar nama tipe Nusa untuk anotasi tipe opsional.
//
// Satu sumber kebenaran ini dipakai parser (memeriksa nama tipe yang tidak
// dikenal), runtime (memverifikasi nilai), dan pemeriksa statis sehingga
// ketiganya selalu sepakat tentang nama tipe yang sah.
package tipe

// Nama tipe yang dikenal Nusa. Nilai ini sama dengan nama tipe yang
// dihasilkan fungsi bawaan `tipe`.
const (
	Angka   = "angka"
	Desimal = "desimal"
	Teks    = "teks"
	Logika  = "logika"
	Kosong  = "kosong"
	Daftar  = "daftar"
	Peta    = "peta"
	Fungsi  = "fungsi"
	// ApaSaja menerima nilai bertipe apa pun.
	ApaSaja = "apa_saja"
)

// daftar adalah urutan baku nama tipe yang dikenal.
var daftar = []string{
	Angka,
	Desimal,
	Teks,
	Logika,
	Kosong,
	Daftar,
	Peta,
	Fungsi,
	ApaSaja,
}

// Dikenal melaporkan apakah nama tersebut adalah nama tipe yang dikenal.
func Dikenal(nama string) bool {
	for _, known := range daftar {
		if known == nama {
			return true
		}
	}
	return false
}

// Semua mengembalikan salinan daftar nama tipe yang dikenal.
func Semua() []string {
	out := make([]string, len(daftar))
	copy(out, daftar)
	return out
}
