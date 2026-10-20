# Nusa

[![CI](https://github.com/seraphimhub/Nusa/actions/workflows/ci.yml/badge.svg)](https://github.com/seraphimhub/Nusa/actions/workflows/ci.yml)
[![Rilis terbaru](https://img.shields.io/github/v/release/seraphimhub/Nusa?label=rilis)](https://github.com/seraphimhub/Nusa/releases/latest)
[![Go](https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![Lisensi: MIT](https://img.shields.io/badge/Lisensi-MIT-green.svg)](LICENSE)

Bahasa pemrograman modern dengan sintaks Indonesia: *buat*, *tulis*, *jika*,
*untuk*, *fungsi*. Berkas sumber memakai ekstensi `.nusa`.

Nusa adalah interpreter lengkap (lexer → parser → evaluator) yang ditulis
dengan Go tanpa dependensi eksternal, cocok untuk belajar cara kerja bahasa
pemrograman sekaligus dipakai untuk skrip kecil sehari-hari.

## Fitur

- **Sintaks Indonesia** yang konsisten dan mudah dibaca.
- **Tipe data**: angka, desimal, teks, logika (`benar`/`salah`), `kosong`,
  daftar, dan peta.
- **Percabangan**: `jika`, `kalau_tidak`, termasuk `kalau_tidak jika`.
- **Pencocokan nilai**: `pilih`, `kasus`, dan `bawaan` (mirip switch).
- **Perulangan**: `ulang`, `selama`, `untuk ... dari ... sampai ... [langkah]`,
  dan `untuk ... dalam`.
- **Penanganan galat**: `coba`/`tangkap`, `lempar`, serta jejak pemanggilan.
- **Fungsi**: parameter, `kembali`, rekursi, dan penutupan (closure) sederhana.
- **Konstanta**: `tetap` untuk nilai yang tidak boleh diubah.
- **Anotasi tipe opsional** pada variabel, parameter, dan nilai kembalian.
- **Modul**: `impor "modul.nusa"` dengan path relatif dan deteksi impor melingkar.
- **Daftar**: literal `[1, 2, 3]`, indeks termasuk negatif, penugasan elemen.
- **Peta**: literal `peta { "kunci": nilai }`, iterasi kunci-nilai, dan
  fungsi bawaan `ada`, `kunci`, `nilai`, `ambil`.
- **30 fungsi bawaan** untuk teks, daftar, peta, galat, konversi, dan matematika.
- **Operator lengkap**: aritmatika, perbandingan, logika short-circuit,
  serta penugasan gabungan `+= -= *= /= %=`.
- **REPL interaktif** dengan input multi-baris dan penampilan nilai ekspresi.
- **Perkakas**: `nusa cek` (pemeriksaan statis), `nusa fmt` (format kode),
  `nusa uji` (penjalur uji), `nusa ast`, dan `nusa token`.
- **Pesan galat berbahasa Indonesia** dengan nomor baris dan cuplikan sumber.
- **127 test** dengan 246 kasus (lexer, parser, runtime, CLI, dan fuzz) tanpa
  dependensi luar.

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
nusa cek examples/tipe.nusa    # pemeriksaan statis
nusa fmt examples/halo.nusa    # format kode (komentar terjaga)
nusa uji examples/matematika_uji.nusa  # jalankan blok uji
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
| [`peta.nusa`](examples/peta.nusa) | Peta dan iterasi kunci-nilai |
| [`tipe.nusa`](examples/tipe.nusa) | Anotasi tipe pada variabel dan fungsi |
| [`galat.nusa`](examples/galat.nusa) | Penanganan galat dengan `coba`/`tangkap` |
| [`galat_tipe.nusa`](examples/galat_tipe.nusa) | Contoh penolakan nilai salah tipe |
| [`matematika_uji.nusa`](examples/matematika_uji.nusa) | Penulisan pengujian dengan `uji`/`cek` |
| [`modul/`](examples/modul) | Pemakaian modul dengan `impor` |

## Fungsi Bawaan

| Kategori | Fungsi |
| --- | --- |
| Teks | `panjang`, `potong`, `cari`, `pisah`, `gabung`, `huruf_besar`, `huruf_kecil` |
| Konversi | `ke_teks`, `ke_angka`, `ke_desimal`, `tipe` |
| Daftar | `tambahkan`, `hapus`, `urutkan`, `balik`, `jumlah`, `rentang` |
| Peta | `ada`, `kunci`, `nilai`, `ambil` |
| Galat | `pesan`, `jejak` |
| Matematika | `akar`, `pangkat`, `bulat`, `lantai`, `langit`, `mutlak` |
| Diagnostik | `kudiag` |

Daftar fungsi dan operator selengkapnya ada di
[dokumentasi referensi](docs/referensi.md).

## Perkakas

Nusa menyediakan beberapa perintah pendukung di luar `run`:

```bash
nusa cek examples/tipe.nusa   # tipe salah, variabel tak terdefinisi, tak terpakai
nusa fmt examples/halo.nusa   # format kode, komentar tetap terjaga
nusa fmt -w examples/         # rapikan seluruh berkas .nusa di direktori
nusa fmt --cek examples/      # hanya periksa, exit code 1 bila belum rapi
nusa uji examples/matematika_uji.nusa  # jalankan blok uji dan cek
nusa uji                       # cari berkas *_uji.nusa / *_test.nusa
```

`nusa cek` melaporkan temuan sebagai galat atau peringatan. Galat membuat exit
code `1`, peringatan saja tetap exit code `0`.

## Pengembangan Web

Nusa sudah bisa menulis markup, serialisasi JSON, dan menangani error
per-request, tetapi belum punya operasi berkas maupun jaringan. Untuk memakai
Nusa dalam web, pola yang paling masuk akal saat ini adalah Go sebagai host
HTTP dan `.nusa` sebagai bahasa render. Hasil audit lengkap, termasuk daftar
gap bahasa yang harus ditutup, ada di [docs/web.md](docs/web.md).

## Struktur Proyek

```
cmd/nusa/            CLI: run, repl, cek, fmt, uji, ast, token, help, version
internal/token/      Definisi token dan kata kunci
internal/lexer/      Pengubah teks menjadi token
internal/ast/        Definisi pohon sintaks abstrak
internal/parser/     Pratt parser
internal/runtime/    Evaluator, objek runtime, dan fungsi bawaan
internal/tipe/       Satu sumber kebenaran nama tipe Nusa
internal/cek/        Pemeriksa statis (tipe dan variabel)
internal/format/     Pemformat kode berbasis token
internal/modul/      Pemuat modul .nusa
examples/            Program contoh .nusa
docs/                Dokumentasi bahasa, kesiapan web, dan kebijakan rilis
.github/             Workflow CI dan konfigurasi catatan rilis
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
- [x] Tipe peta (`peta`) dan iterasi dua variabel
- [x] Modul dan `impor`
- [x] Pengetikan opsional dan pemeriksa tipe
- [x] Pemformatan kode otomatis (`nusa fmt`)
- [x] Penanganan galat `coba`/`tangkap` dan `lempar`
- [x] Penulisan pengujian di dalam bahasa (`uji`/`cek`)
- [ ] Tipe data generik dan tipe union
- [ ] Pemuat paket dari repositori daring

## Rilis

Setiap rilis memakai annotated tag `vX.Y.Z` beserta GitHub Release dengan
catatan dari [CHANGELOG.md](CHANGELOG.md). Versi terbaru: **v1.2.0**
(2026-10-20). Kebijakan versi dan prosedur rilis ada di [docs/rilis.md](docs/rilis.md).

| Versi | Fokus |
| --- | --- |
| [v1.2.0](https://github.com/seraphimhub/Nusa/releases/tag/v1.2.0) | Peta, modul, anotasi tipe, `coba`/`tangkap`, `nusa cek`/`fmt`/`uji`, 30 fungsi bawaan, 127 test |
| [v1.1.0](https://github.com/seraphimhub/Nusa/releases/tag/v1.1.0) | Daftar, `untuk`, `pilih`, 23 fungsi bawaan, 88 test |
| [v1.0.0](https://github.com/seraphimhub/Nusa/releases/tag/v1.0.0) | Rilis stabil pertama: lexer, parser, evaluator, CLI |
| v0.3.0 – v0.7.2 | Tonggak pengembangan pra-1.0 |

Daftar lengkap: <https://github.com/seraphimhub/Nusa/releases>

## Keamanan

Cabang `main` dilindungi: force-push dan penghapusan cabang diblokir, riwayat
linear, serta check CI wajib lulus sebelum pull request di-merge (push langsung
masih diizinkan). Untuk melaporkan kerentanan, gunakan GitHub Security Advisory
privat — lihat [SECURITY.md](SECURITY.md).

## Lisensi

MIT — lihat [LICENSE](LICENSE).
