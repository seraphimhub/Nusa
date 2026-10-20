# Referensi Bahasa Nusa

Dokumen ini merangkum seluruh sintaks dan perilaku bahasa Nusa versi 1.2.0.

## Daftar Isi

1. [Komentar](#komentar)
2. [Variabel](#variabel)
3. [Konstanta](#konstanta)
4. [Tipe Data](#tipe-data)
5. [Anotasi Tipe](#anotasi-tipe)
6. [Operator](#operator)
7. [Percabangan](#percabangan)
8. [Perulangan](#perulangan)
9. [Pernyataan Pilih](#pernyataan-pilih)
10. [Penanganan Galat](#penanganan-galat)
11. [Fungsi](#fungsi)
12. [Daftar](#daftar)
13. [Peta](#peta)
14. [Modul](#modul)
15. [Input](#input)
16. [Fungsi Bawaan](#fungsi-bawaan)
17. [Pengujian](#pengujian)
18. [Perintah CLI](#perintah-cli)
19. [Pesan Galat](#pesan-galat)

## Komentar

```nusa
# komentar satu baris
// komentar satu baris juga
/* komentar blok
   sampai beberapa baris */
```

## Variabel

Deklarasi memakai `buat`, penugasan ulang memakai `=` atau operator gabungan.

```nusa
buat umur = 25
umur = 26
umur += 1      # 27
umur *= 2      # 54
```

- `buat` mendeklarasikan variabel pada scope saat ini (global atau fungsi).
- Penugasan `=` memperbarui variabel tempat ia dideklarasikan, sehingga fungsi
  dapat mengubah variabel di scope luar.
- Blok `{ ... }` (if/loop) **tidak** membuat scope baru; hanya pemanggilan
  fungsi yang membuat scope baru.

## Konstanta

`tetap` mendeklarasikan konstanta. Nilai konstanta tidak boleh diubah
setelah dideklarasikan.

```nusa
tetap PI = 3.14
tetap VERSI = "1.2.0"

tulis PI * 2
```

- Penugasan, penugasan gabungan, dan deklarasi ulang pada scope yang sama
  ditolak dengan pesan galat yang jelas.
- Konstanta boleh "dibayangi" (shadowing) oleh deklarasi lokal di dalam
  fungsi, dan nilai konstanta tersedia di lingkup dalam fungsi tersebut.

## Tipe Data

| Tipe | Contoh | Keterangan |
| --- | --- | --- |
| `angka` | `42`, `-7` | Bilangan bulat 64-bit |
| `desimal` | `3.14`, `-0.5` | Bilangan pecahan |
| `teks` | `"halo"`, `'halo'` | String UTF-8 |
| `logika` | `benar`, `salah` | Boolean |
| `kosong` | `kosong` | Nilai kosong |
| `daftar` | `[1, "dua", benar]` | Daftar dinamis (mutable) |
| `peta` | `peta { "a": 1 }` | Peta kunci-nilai (mutable) |
| `fungsi` | `fungsi f() {}` | Fungsi atau fungsi bawaan |

Escape string yang dikenali: `\n`, `\t`, `\r`, `\\`, `\"`, `\'`, dan `\0`.

```nusa
tulis "baris pertama\nbaris kedua\tberakhir"
tulis 'teks dengan "kutip ganda"'
```

## Anotasi Tipe

Anotasi tipe bersifat opsional. Bila ditulis, nilainya diperiksa saat
program berjalan dan galat menunjuk ke baris yang salah.

```nusa
buat nama: teks = "Budi"
buat umur: angka = 25

fungsi luas_persegi(sisi: angka): angka {
    kembali sisi * sisi
}
```

- Anotasi variabel: `buat nama: tipe = nilai`.
- Anotasi parameter: `fungsi f(x: tipe)`.
- Anotasi nilai kembalian: `fungsi f(x): tipe`.
- Tipe yang dikenal: `angka`, `desimal`, `teks`, `logika`, `kosong`, `daftar`,
  `peta`, `fungsi`, dan `apa_saja`.
- `apa_saja` menerima nilai bertipe apa pun.
- Nilai yang tidak diberi anotasi tidak diperiksa, sehingga kode lama tetap
  berjalan apa adanya.

Galat anotasi tipe muncul dalam dua bentuk:

```
ERROR: baris 3: nilai untuk 'umur' harus bertipe angka, mendapat teks
ERROR: baris 3: fungsi 'luas_persegi' harus mengembalikan nilai bertipe angka, mendapat teks
```

## Operator

Dari prioritas tertinggi ke terendah:

| Prioritas | Operator | Keterangan |
| --- | --- | --- |
| 1 | `( )`, `f(x)`, `d[i]` | Grup, pemanggilan, indeks |
| 2 | `tidak`, `!`, `-` | Operator awalan (unary) |
| 3 | `*`, `/`, `%` | Aritmatika |
| 4 | `+`, `-` | Aritmatika |
| 5 | `>`, `<`, `>=`, `<=` | Perbandingan |
| 6 | `==`, `!=` | Kesetaraan |
| 7 | `dan` | Logika AND |
| 8 | `atau` | Logika OR |
| 9 | `=`, `+=`, `-=`, `*=`, `/=`, `%=` | Penugasan (level statement) |

Aturan penting:

- Pembagian dua `angka` menghasilkan `angka` (pembagian bulat): `10 / 3` → `3`.
  Jika salah satu sisi `desimal`, hasilnya `desimal`: `10 / 4.0` → `2.5`.
- `+` dengan `teks` akan menggabungkan dan mengubah sisi lain menjadi teks:
  `"Umur: " + 25` → `"Umur: 25"`.
- `dan`/`atau` memakai evaluasi short-circuit dan selalu menghasilkan `benar`
  atau `salah`.
- `==` dan `!=` aman untuk semua tipe: tipe yang berbeda menghasilkan `salah`
  tanpa galat, sementara daftar dibandingkan elemen per elemen. `angka` dan
  `desimal` dibandingkan secara numerik.
- Nilai dianggap salah (falsy) jika `salah`, `kosong`, `0`, `0.0`, atau `""`.

## Percabangan

```nusa
jika nilai >= 90 {
    tulis "A"
} kalau_tidak jika nilai >= 80 {
    tulis "B"
} kalau_tidak {
    tulis "C"
}
```

Kurung kurawal opsional untuk satu statement:

```nusa
jika umur >= 18 tulis "dewasa"
kalau_tidak tulis "anak"
```

## Perulangan

```nusa
# Mengulang sebanyak n kali
ulang 5 {
    tulis "halo"
}

# While
buat i = 0
selama i < 3 {
    tulis i
    i += 1
}

# Rentang angka (inklusif), langkah opsional
untuk i dari 1 sampai 10 langkah 2 {
    tulis i
}

# Mundur
untuk i dari 10 sampai 1 langkah -1 {
    tulis i
}

# Iterasi daftar atau teks
untuk buah dalam ["apel", "jeruk"] {
    tulis buah
}
```

Di dalam perulangan tersedia:

- `berhenti` — keluar dari perulangan terdekat.
- `lanjutkan` — lewati sisa body dan lanjut ke iterasi berikutnya.
- `kembali` — keluar dari fungsi walaupun berada di dalam perulangan.

Nilai `langkah` tidak boleh nol. Jika nilai awal lebih besar dari nilai akhir
dengan langkah positif, body tidak dijalankan.

## Pernyataan Pilih

`pilih` mencocokkan sebuah nilai dengan beberapa `kasus`. Cabang pertama yang
cocok dijalankan, dan `bawaan` dijalankan bila tidak ada yang cocok.

```nusa
pilih hari {
    kasus 1 { tulis "Senin" }
    kasus 2 { tulis "Selasa" }
    kasus 3 { tulis "Rabu" }
    bawaan { tulis "Bukan hari kerja" }
}
```

Keterangan:

- Pencocokan memakai aturan `==` sehingga angka, desimal, teks, logika,
  `kosong`, dan daftar semuanya bisa dipakai sebagai nilai kasus.
- Tidak ada fallthrough: hanya satu cabang yang dijalankan.
- `bawaan` bersifat opsional dan hanya boleh muncul sekali.
- `kembali` di dalam cabang akan keluar dari fungsi, dan `berhenti` /
  `lanjutkan` berlaku untuk perulangan terdekat yang membungkus `pilih`.

```nusa
fungsi namaBulan(bulan) {
    pilih bulan {
        kasus 1 { kembali "Januari" }
        kasus 2 { kembali "Februari" }
        bawaan { kembali "Bulan lain" }
    }
}
```

## Penanganan Galat

`coba` menjalankan blok kode dan `tangkap` menangani galat yang muncul di
dalamnya. `lempar` membuat galat sendiri.

```nusa
fungsi bagi(a, b) {
    jika b == 0 {
        lemparkan "pembagian dengan nol"
    }
    kembali a / b
}

coba {
    tulis bagi(10, 0)
} tangkap galat {
    tulis "tertangkap: " + galat["pesan"]
    tulis "baris: " + galat["baris"]
}
```

- `tangkap` tanpa nama galat tetap valid: `} tangkap { ... }`.
- Nilai galat yang tertangkap dapat diindeks dengan kunci `pesan` (teks),
  `baris` (angka), dan `jejak` (daftar riwayat pemanggilan fungsi). Kunci
  lain ditolak dengan pesan galat yang jelas.
- Galat dapat dilempar ulang apa adanya dengan `lempar galat`.
- Galat yang tidak tertangkap menghentikan program dan dicetak CLI lengkap
  dengan cuplikan sumber serta penanda kolom.

## Fungsi

```nusa
fungsi faktorial(n) {
    jika n <= 1 {
        kembali 1
    }
    kembali n * faktorial(n - 1)
}

tulis faktorial(5)          # pemanggilan langsung
panggil faktorial(5)        # bentuk dengan kata kunci, nilai dibuang
```

- Jumlah argumen harus sama dengan jumlah parameter.
- Fungsi tanpa `kembali` mengembalikan `kosong`.
- Fungsi dapat memanggil dirinya sendiri (rekursi) dan mengakses variabel
  scope luar (closure sederhana).
- Argumen dipanggil dengan **salinan referensi** untuk daftar: mengubah isi
  daftar di dalam fungsi akan terlihat oleh pemanggil.

## Daftar

```nusa
buat buah = ["apel", "jeruk"]
buah[0] = "mangga"      # ubah elemen
buah[1] += " manis"     # penugasan gabungan pada elemen daftar
tulis buah[-1]          # indeks negatif dari akhir
tulis panjang(buah)

tambahkan(buah, "pisang")   # menambah elemen (mengubah daftar)
hapus(buah, 0)              # menghapus elemen berdasarkan indeks
```

Catatan:

- Indeks di luar batas dan indeks bukan `angka` menghasilkan galat runtime.
- Indeks negatif dihitung dari akhir: `-1` adalah elemen terakhir.
- `tambahkan`, `hapus`, dan `urutkan` mengubah daftar asli; `balik` dan
  `rentang` mengembalikan daftar baru.
- Teks juga bisa diindeks dan menghasilkan satu karakter: `"nusa"[0]` → `"n"`.

## Peta

Peta menyimpan pasangan kunci-nilai. Kunci dapat berupa teks, angka, atau
logika.

```nusa
buat pengguna = peta {
    "nama": "Budi",
    "umur": 25,
}

tulis pengguna["nama"]
pengguna["umur"] = 26
tulis pengguna["umur"]
```

- Peta dapat diiterasi dengan dua variabel, kunci lalu nilai:

  ```nusa
  untuk kunci, nilai dalam pengguna {
      tulis kunci + " = " + nilai
  }
  ```

- Peta yang dipakai dengan satu variabel diiterasi untuk menghasilkan kunci.
- Mengambil kunci yang tidak ada menghasilkan galat. Gunakan `ada(peta, kunci)`
  untuk mengecek atau `ambil(peta, kunci, bawaan)` untuk nilai cadangan.
- Fungsi bawaan peta: `ada`, `kunci`, `nilai`, `ambil`. `panjang` menerima
  peta dan menghasilkan jumlah pasangannya.

## Modul

Modul adalah berkas `.nusa` lain yang dimuat dengan `impor`.

```nusa
# util.nusa
tetap VERSI = "1.0"

fungsi sapa(nama) {
    kembali "Halo, " + nama
}
```

```nusa
# utama.nusa
impor "util.nusa"

tulis sapa("Dunia")
tulis VERSI
```

- Path modul diselesaikan relatif terhadap direktori berkas pemanggil.
- Semua fungsi dan konstanta yang dideklarasikan di tingkat atas modul ikut
  diekspor ke lingkup berkas pemanggil.
- Modul yang sama hanya dieksekusi satu kali, dan impor melingkar ditolak
  dengan pesan galat yang jelas.
- Modul dapat mengimpor modul lain.

## Input

```nusa
membaca nama
tulis "Halo " + nama
```

`membaca` menampilkan prompt `> ` lalu menyimpan satu baris masukan ke variabel
yang disebutkan.

## Fungsi Bawaan

| Fungsi | Argumen | Hasil |
| --- | --- | --- |
| `panjang(nilai)` | teks, daftar, atau peta | jumlah karakter (rune), elemen, atau pasangan |
| `tipe(nilai)` | semua tipe | nama tipe dalam bahasa Indonesia |
| `pesan(galat)` | nilai galat | pesan galat sebagai teks |
| `jejak(galat)` | nilai galat | daftar riwayat pemanggilan fungsi |
| `ada(peta, kunci)` | peta, kunci | `benar` bila kunci ada |
| `kunci(peta)` | peta | daftar seluruh kunci |
| `nilai(peta)` | peta | daftar seluruh nilai |
| `ambil(peta, kunci, bawaan)` | peta, kunci, semua tipe | nilai kunci atau nilai cadangan |
| `ke_teks(nilai)` | semua tipe | representasi teks |
| `ke_angka(nilai)` | angka, desimal, teks | `angka` (desimal dipotong, teks di-parse) |
| `ke_desimal(nilai)` | angka, desimal, teks | `desimal` |
| `huruf_besar(teks)` | teks | teks huruf besar |
| `huruf_kecil(teks)` | teks | teks huruf kecil |
| `potong(teks, mulai[, akhir])` | teks, angka | potongan teks berdasarkan karakter |
| `cari(teks, sub)` | teks | indeks karakter pertama, `-1` jika tidak ada |
| `pisah(teks, pemisah)` | teks | daftar teks |
| `gabung(daftar[, pemisah])` | daftar, teks | teks gabungan |
| `tambahkan(daftar, nilai)` | daftar, semua tipe | daftar setelah ditambah |
| `hapus(daftar, indeks)` | daftar, angka | daftar setelah dihapus |
| `hapus(peta, kunci)` | peta, kunci | peta tanpa kunci tersebut |
| `urutkan(daftar)` | daftar angka/teks | daftar terurut (mengubah asli) |
| `balik(daftar)` | daftar | daftar baru dengan urutan terbalik |
| `jumlah(daftar)` | daftar angka | total, `angka` jika semua bulat |
| `rentang(stop)` / `rentang(mulai, stop[, langkah])` | angka | daftar angka (stop eksklusif) |
| `akar(x)` | angka ≥ 0 | akar kuadrat (`desimal`) |
| `pangkat(a, b)` | angka | hasil pemangkatan |
| `bulat(x)` | angka | pembulatan ke bilangan bulat terdekat |
| `lantai(x)` | angka | pembulatan ke bawah |
| `langit(x)` | angka | pembulatan ke atas |
| `mutlak(x)` | angka | nilai absolut |
| `kudiag()` | — | ringkasan status runtime (langkah, kedalaman, variabel aktif) |

Semua pemanggilan fungsi bawaan memeriksa jumlah dan tipe argumen serta
menghasilkan galat yang jelas bila tidak sesuai.

## Pengujian

Bahasa Nusa menyediakan blok `uji` berisi pemeriksaan `cek` untuk menulis
test langsung di dalam berkas sumber.

```nusa
fungsi tambah(a, b) {
    kembali a + b
}

uji "penjumlahan" {
    cek tambah(1, 2) == 3
    cek tambah(0, 0) == 0
}

uji "penanganan galat" {
    coba {
        tambah(1, "a")
        cek salah, "harus gagal"
    } tangkap galat {
        cek galat["pesan"] != ""
    }
}
```

- `cek ekspresi` gagal bila ekspresi bernilai `salah` atau `kosong`.
- `cek ekspresi, "pesan"` gagal dengan pesan kustom bila ekspresi bernilai
  `salah` atau `kosong`.
- Setiap blok `uji` boleh berisi banyak `cek`. Bila satu `cek` gagal, blok
  tersebut dicatat sebagai gagal dan pemeriksaan berikutnya tetap berjalan.
- Perintah `nusa uji <file.nusa>` menjalankan seluruh blok `uji` pada berkas
  tersebut dan menampilkan ringkasan. Exit code `1` bila ada blok gagal.

## Perintah CLI

| Perintah | Kegunaan |
| --- | --- |
| `nusa run <file>` | Menjalankan berkas `.nusa` |
| `nusa <file>` | Singkatan dari `nusa run <file>` |
| `nusa repl` | Mode interaktif |
| `nusa cek <file ...>` | Pemeriksaan statis: tipe, variabel tak terdefinisi, variabel tak terpakai |
| `nusa fmt <file ...>` | Memformat kode dan menjaga komentar |
| `nusa uji [<file ...>]` | Menjalankan blok `uji` pada berkas |
| `nusa ast <file>` | Menampilkan pohon sintaks |
| `nusa token <file>` | Menampilkan daftar token |
| `nusa version` | Menampilkan versi |
| `nusa help` | Menampilkan bantuan |

`nusa cek` melaporkan temuan berupa galat dan peringatan. Galat membuat exit
code `1`, sedangkan peringatan saja tetap exit code `0`.

### Opsi nusa fmt

| Opsi | Kegunaan |
| --- | --- |
| (tanpa opsi) | Hasil format dicetak ke layar |
| `-w`, `--tulis` | Menyimpan hasil format ke berkas aslinya |
| `--cek` | Hanya memeriksa; exit code `1` bila ada berkas yang belum rapi |
| `-h`, `--help` | Menampilkan bantuan perintah |

`nusa fmt` menerima berkas maupun direktori. Untuk direktori, seluruh berkas
`.nusa` di dalamnya diproses. Opsi `--cek` dan `-w` tidak bisa dipakai
bersamaan.

### Penemuan berkas uji

`nusa uji` tanpa argumen mencari berkas uji pada direktori sekarang. Pola
nama yang dikenali adalah sufiks `_uji.nusa` dan `_test.nusa`.

## Pesan Galat

Galat runtime memuat nomor baris:

```
ERROR: baris 3: pembagian dengan nol
```

Galat leksikal dan parsing ditampilkan terpisah, misalnya:

```
Error parsing:
   baris 2: tidak diduga '=', mengharapkan 'IDENT'
```

CLI menampilkan cuplikan sumber di sekitar baris galat beserta penanda kolom:

```
ERROR: baris 3: pembagian dengan nol
2 |     tulis a / b
3 |     tulis bagi(10, 0)
  |              ^
```

Perintah `nusa run` mengembalikan exit code `0` bila program berhasil dan `1`
bila terjadi galat leksikal, parsing, atau runtime.
