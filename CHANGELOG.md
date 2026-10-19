# Riwayat Perubahan

Semua perubahan penting proyek Nusa dicatat di berkas ini.
Format mengikuti [Keep a Changelog](https://keepachangelog.com/id/1.1.0/).

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

## [1.0.0] - 2026-04-14

### Ditambahkan

- Lexer, parser Pratt, dan evaluator dasar.
- Variabel (`buat`), output (`tulis`), input (`membaca`).
- Percabangan `jika`/`kalau_tidak`.
- Perulangan `ulang` dan `selama` beserta `berhenti` dan `lanjutkan`.
- Fungsi `fungsi`/`kembali` dan pemanggilan `panggil`.
- Operator aritmatika, perbandingan, dan logika.
- CLI `run`, `repl`, `help`, dan `version`.

[1.1.0]: https://github.com/seraphimhub/Nusa/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/seraphimhub/Nusa/releases/tag/v1.0.0
