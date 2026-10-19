# Nusa

[![CI](https://github.com/seraphimhub/Nusa/actions/workflows/ci.yml/badge.svg)](https://github.com/seraphimhub/Nusa/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![Lisensi: MIT](https://img.shields.io/badge/Lisensi-MIT-green.svg)](LICENSE)

Bahasa pemrograman modern dengan sintaks Indonesia: *buat*, *tulis*, *jika*,
*untuk*, *fungsi*. Berkas sumber memakai ekstensi `.nusa`.

Nusa adalah interpreter lengkap (lexer → parser → evaluator) yang ditulis
dengan Go tanpa dependensi eksternal, cocok untuk belajar cara kerja bahasa
pemrograman sekaligus dipakai untuk skrip kecil sehari-hari.

## Fitur

- **Sintaks Indonesia** yang konsisten dan mudah dibaca.
- **Tipe data**: angka, desimal, teks, logika (`benar`/`salah`), `kosong`, dan daftar.
- **Percabangan**: `jika`, `kalau_tidak`, termasuk `kalau_tidak jika`.
- **Pencocokan nilai**: `pilih`, `kasus`, dan `bawaan` (mirip switch).
- **Perulangan**: `ulang`, `selama`, `untuk ... dari ... sampai ... [langkah]`,
  dan `untuk ... dalam`.
- **Fungsi**: parameter, `kembali`, rekursi, dan penutupan (closure) sederhana.
- **Daftar**: literal `[1, 2, 3]`, indeks termasuk negatif, penugasan elemen.
- **23 fungsi bawaan** untuk teks, daftar, konversi, dan matematika.
- **Operator lengkap**: aritmatika, perbandingan, logika short-circuit,
  serta penugasan gabungan `+= -= *= /= %=`.
- **REPL interaktif** dengan input multi-baris dan penampilan nilai ekspresi.
- **Perkakas**: perintah `ast` dan `token` untuk memeriksa hasil lexer/parser.
- **Pesan galat berbahasa Indonesia** beserta nomor baris.
- **88 test** (lexer, parser, runtime, dan CLI) tanpa dependensi luar.

## Instalasi

Butuh Go 1.22 atau lebih baru.

```bash
git clone https://github.com/seraphimhub/Nusa.git
cd Nusa
go build -o nusa ./cmd/nusa
```

Letakkan hasil build di `PATH`, atau jalankan langsung tanpa build:

```bash
go run ./cmd/nusa run examples/halo.nusa
```

## Pemakaian

```bash
nusa run examples/fitur.nusa   # jalankan berkas .nusa
nusa examples/fitur.nusa       # singkatan dari perintah di atas
nusa repl                      # mode interaktif
nusa ast examples/halo.nusa    # tampilkan AST
nusa token examples/halo.nusa  # tampilkan token
nusa version                   # versi
nusa help                      # bantuan
```

## Contoh

```nusa
# Sapaan sederhana dengan percabangan
buat nama = "Budi"
buat umur = 25

fungsi sapa(nama, umur) {
    kembali "Halo " + nama + " (" + umur + " tahun)"
}

tulis sapa(nama, umur)

jika umur >= 18 {
    tulis "Status: dewasa"
} kalau_tidak {
    tulis "Status: anak"
}

# Daftar dan perulangan
buat buah = ["apel", "jeruk", "mangga"]
tambahkan(buah, "pisang")

untuk i dari 0 sampai panjang(buah) - 1 {
    tulis ke_teks(i + 1) + ". " + buah[i]
}

# Perulangan dengan rentang dan langkah
buat total = 0
untuk i dari 1 sampai 10 {
    total += i
}
tulis "Jumlah 1 sampai 10: " + total

untuk x dalam buah {
    tulis "Buah: " + x
}
```

Contoh lain ada di direktori [`examples/`](examples):

| Berkas | Isi |
| --- | --- |
| [`halo.nusa`](examples/halo.nusa) | Program terkecil dengan percabangan |
| [`fitur.nusa`](examples/fitur.nusa) | Tur lengkap fitur bahasa |
| [`daftar.nusa`](examples/daftar.nusa) | Daftar dan fungsi bawaan daftar |
| [`teks.nusa`](examples/teks.nusa) | Pengolahan teks |
| [`faktorial.nusa`](examples/faktorial.nusa) | Rekursi vs iterasi |
| [`fibonacci.nusa`](examples/fibonacci.nusa) | Fibonacci rekursif dan daftar |
| [`pilih.nusa`](examples/pilih.nusa) | Pencocokan nilai dengan `pilih` |

## Fungsi Bawaan

| Kategori | Fungsi |
| --- | --- |
| Teks | `panjang`, `potong`, `cari`, `pisah`, `gabung`, `huruf_besar`, `huruf_kecil` |
| Konversi | `ke_teks`, `ke_angka`, `ke_desimal`, `tipe` |
| Daftar | `tambahkan`, `hapus`, `urutkan`, `balik`, `jumlah`, `rentang` |
| Matematika | `akar`, `pangkat`, `bulat`, `lantai`, `langit`, `mutlak` |

Daftar fungsi dan operator selengkapnya ada di
[dokumentasi referensi](docs/referensi.md).

## Struktur Proyek

```
cmd/nusa/            CLI: run, repl, ast, token, help, version
internal/token/      Definisi token dan kata kunci
internal/lexer/      Pengubah teks menjadi token
internal/ast/        Definisi pohon sintaks abstrak
internal/parser/     Pratt parser
internal/runtime/    Evaluator, objek runtime, dan fungsi bawaan
examples/            Program contoh .nusa
docs/                Dokumentasi bahasa
```

Alur eksekusi:

```
teks .nusa → lexer → token → parser → AST → runtime → output
```

## Pengembangan

```bash
make test     # jalankan seluruh test
make build    # bangun binary nusa
make fmt      # rapikan format kode
make vet      # analisis statis
make run      # jalankan examples/fitur.nusa
```

Atau langsung dengan Go:

```bash
gofmt -l .
go vet ./...
go test ./...
```

## Roadmap

- [x] Tipe daftar dan indeks negatif
- [x] Loop `untuk ... dari ... sampai` dan `untuk ... dalam`
- [x] Penugasan gabungan
- [x] Fungsi bawaan teks, daftar, dan matematika
- [x] Pencocokan nilai `pilih`/`kasus`/`bawaan`
- [ ] Tipe peta (`peta`) dan `untuk ... di ...`
- [ ] Modul dan `impor`
- [ ] Pengetikan opsional dan pemeriksa tipe
- [ ] Pemformatan kode otomatis (`nusa fmt`)

## Keamanan

Cabang `main` dilindungi: force-push dan penghapusan cabang diblokir, riwayat
linear, serta check CI wajib lulus sebelum pull request di-merge (push langsung
masih diizinkan). Untuk melaporkan kerentanan, gunakan GitHub Security Advisory
privat — lihat [SECURITY.md](SECURITY.md).

## Lisensi

MIT — lihat [LICENSE](LICENSE).
