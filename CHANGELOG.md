# Riwayat Perubahan

Semua perubahan penting proyek Nusa dicatat di berkas ini.
Format mengikuti [Keep a Changelog](https://keepachangelog.com/id/1.1.0/) dan
[Semantic Versioning](https://semver.org/lang/id/). Setiap versi rilis ditandai
annotated tag `vX.Y.Z` serta GitHub Release. Prosedur rilis ada di
[docs/rilis.md](docs/rilis.md).

## [1.2.0] - 2026-10-20

Rilis ini menambah tipe peta, sistem modul, pengetikan opsional, penanganan
galat, dan perkakas pengembangan (pemeriksa statis, pemformat, penjalur uji).

### Ditambahkan

- Tipe `peta` dengan literal `peta { "kunci": nilai }`, indeks, penugasan, dan
  iterasi satu variabel (kunci) maupun dua variabel (kunci dan nilai).
- Fungsi bawaan peta: `ada`, `kunci`, `nilai`, dan `ambil`. `panjang` kini
  menerima peta, dan `hapus` dapat menghapus kunci pada peta.
- Pernyataan `impor "modul.nusa"` untuk memuat modul dari berkas lain. Path
  diselesaikan relatif terhadap direktori berkas pemanggil, modul yang sama
  hanya dieksekusi satu kali, dan impor melingkar ditolak.
- Anotasi tipe opsional pada variabel (`buat x: angka = 1`), parameter
  (`fungsi f(x: angka)`), dan nilai kembalian (`fungsi f(x): angka`). Tipe
  yang dikenal: `angka`, `desimal`, `teks`, `logika`, `kosong`, `daftar`,
  `peta`, `fungsi`, dan `apa_saja`.
- Pernyataan `coba`/`tangkap` untuk menangani galat, beserta `lempar` untuk
  membuat galat sendiri. Nilai galat tertangkap adalah peta dengan kunci
  `pesan`, `baris`, `tipe`, dan `jejak`.
- Fungsi bawaan `pesan` dan `jejak` untuk membaca detail galat, serta
  `kudiag` untuk melihat status runtime.
- Blok `uji` berisi pemeriksaan `cek` untuk menulis pengujian di dalam berkas
  sumber, beserta perintah CLI `nusa uji` sebagai penjalur uji.
- Perintah CLI `nusa cek` sebagai pemeriksa statis: pemeriksaan tipe,
  variabel tak terdefinisi, dan variabel tak terpakai.
- Perintah CLI `nusa fmt` sebagai pemformat kode berbasis token yang menjaga
  komentar baris maupun blok.
- Paket `internal/tipe` sebagai satu sumber kebenaran nama tipe, dipakai
  parser, runtime, dan pemeriksa statis.
- Galat runtime kini menampilkan cuplikan sumber di sekitar baris penyebab
  beserta penanda kolom.
- Fuzz test untuk lexer, parser, dan runtime beserta seed awal.
- Contoh baru: `peta.nusa`, `tipe.nusa`, `galat.nusa`, `galat_tipe.nusa`,
  `matematika_uji.nusa`, dan direktori `modul/`.

### Diubah

- CLI kini membuat runtime dengan pemuat modul sehingga `impor` bekerja pada
  `run` maupun REPL.
- Pesan galat leksikal dan parsing juga disertai cuplikan sumber.
- Nomor versi pada CLI menjadi `1.2.0`.

### Diperbaiki

- Galat indeks pada peta kini menunjuk kunci yang hilang dan menyarankan
  `ada` atau `ambil` sebagai alternatif.

## [1.1.0] - 2026-10-19

Rilis ini memperluas bahasa dengan daftar, loop rentang, penugasan gabungan,
serta pustaka fungsi bawaan.

### Ditambahkan

- Tipe `daftar` (`[1, 2, 3]`) dengan indeks, indeks negatif, dan penugasan
  elemen.
- Perulangan `untuk i dari a sampai b [langkah n]` dan `untuk x dalam daftar`.
- Operator penugasan gabungan `+=`, `-=`, `*=`, `/=`, `%=`.
- Pernyataan pencocokan nilai `pilih` dengan cabang `kasus` dan `bawaan`
  opsional; cabang pertama yang cocok dijalankan tanpa fallthrough.
- Pemanggilan fungsi langsung tanpa kata kunci `panggil`.
- Escape string `\n`, `\t`, `\r`, `\\`, `\"`, `\'`, `\0` dan literal kutip
  tunggal.
- Komentar blok `/* ... */`.
- 23 fungsi bawaan: teks (`panjang`, `potong`, `cari`, `pisah`, `gabung`,
  `huruf_besar`, `huruf_kecil`), konversi (`ke_teks`, `ke_angka`,
  `ke_desimal`, `tipe`), daftar (`tambahkan`, `hapus`, `urutkan`, `balik`,
  `jumlah`, `rentang`), dan matematika (`akar`, `pangkat`, `bulat`, `lantai`,
  `langit`, `mutlak`).
- Perintah CLI `nusa ast`, `nusa token`, serta dukungan `nusa berkas.nusa`.
- REPL mendukung input multi-baris dan menampilkan nilai ekspresi.
- Nomor kolom pada token dan pesan galat yang lebih jelas.
- Test lexer, parser, runtime, dan CLI.

### Diubah

- `dan`/`atau` kini memakai evaluasi short-circuit.
- Kesetaraan `==`/`!=` berlaku lintas tipe dan membandingkan daftar secara
  rekursif, bukan menghasilkan galat.
- `benar`/`salah` ditampilkan sebagai `benar`/`salah`, dan elemen teks di
  dalam daftar diberi tanda kutip.
- `kalau_tidak jika` resmi didukung pada rantai percabangan.
- Exit code `nusa run` bernilai 1 saat terjadi galat.

### Diperbaiki

- Parser salah memakai `)` milik pemanggilan dalam pada call bersarang seperti
  `jumlah(rentang(1, 101))`.
- Lexer panic saat berkas berakhir di tengah string yang belum ditutup.
- `kosong == kosong` dan perbandingan nilai bertipe berbeda tidak lagi
  menghasilkan galat.

### Dihapus

- Binary hasil build tidak lagi disertakan di repositori.

## [1.0.0] - 2026-07-01

Rilis stabil pertama: interpreter lengkap dengan sintaks Indonesia, memakai
nomor versi `1.0.0` pada CLI.

### Ditambahkan

- Lexer, parser Pratt, dan evaluator dasar.
- Variabel (`buat`), output (`tulis`), input (`membaca`).
- Percabangan `jika`/`kalau_tidak`.
- Perulangan `ulang` dan `selama` beserta `berhenti` dan `lanjutkan`.
- Fungsi `fungsi`/`kembali` dan pemanggilan `panggil`.
- Operator aritmatika, perbandingan, dan logika.
- CLI `run`, `repl`, `help`, dan `version`.

## Tonggak pra-1.0 (2026-04-14)

Versi awal dicatat sebagai tag ringan tanpa catatan rilis. Sejak 1.0.0 semua
rilis memakai annotated tag dan GitHub Release.

| Tag | Isi |
| --- | --- |
| [v0.3.0][v030] | Pernyataan `jika` dengan operator perbandingan |
| [v0.4.0][v040] | Perulangan `ulang` dan penanganan token `NEWLINE` |
| [v0.5.0][v050] | Fungsi dengan parameter dan pemanggilan `panggil` |
| [v0.6.0][v060] | Perbaikan duplikasi token operator plus |
| [v0.7.0][v070] | Operator aritmatika lengkap (`+`, `-`, `*`, `/`, `%`) |
| [v0.7.1][v071] | Operator perbandingan lengkap (`==`, `!=`, `>=`, `<=`) |
| [v0.7.2][v072] | Cabang `kalau_tidak` untuk pernyataan `jika` |

[1.2.0]: https://github.com/seraphimhub/Nusa/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/seraphimhub/Nusa/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/seraphimhub/Nusa/releases/tag/v1.0.0
[v030]: https://github.com/seraphimhub/Nusa/releases/tag/v0.3.0
[v040]: https://github.com/seraphimhub/Nusa/releases/tag/v0.4.0
[v050]: https://github.com/seraphimhub/Nusa/releases/tag/v0.5.0
[v060]: https://github.com/seraphimhub/Nusa/releases/tag/v0.6.0
[v070]: https://github.com/seraphimhub/Nusa/releases/tag/v0.7.0
[v071]: https://github.com/seraphimhub/Nusa/releases/tag/v0.7.1
[v072]: https://github.com/seraphimhub/Nusa/releases/tag/v0.7.2
