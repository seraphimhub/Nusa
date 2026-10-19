# Referensi Bahasa Nusa

Dokumen ini merangkum seluruh sintaks dan perilaku bahasa Nusa versi 1.1.0.

## Daftar Isi

1. [Komentar](#komentar)
2. [Variabel](#variabel)
3. [Tipe Data](#tipe-data)
4. [Operator](#operator)
5. [Percabangan](#percabangan)
6. [Perulangan](#perulangan)
7. [Pernyataan Pilih](#pernyataan-pilih)
8. [Fungsi](#fungsi)
9. [Daftar](#daftar)
10. [Input](#input)
11. [Fungsi Bawaan](#fungsi-bawaan)
12. [Pesan Galat](#pesan-galat)

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

## Tipe Data

| Tipe | Contoh | Keterangan |
| --- | --- | --- |
| `angka` | `42`, `-7` | Bilangan bulat 64-bit |
| `desimal` | `3.14`, `-0.5` | Bilangan pecahan |
| `teks` | `"halo"`, `'halo'` | String UTF-8 |
| `logika` | `benar`, `salah` | Boolean |
| `kosong` | `kosong` | Nilai kosong |
| `daftar` | `[1, "dua", benar]` | Daftar dinamis (mutable) |
| `fungsi` | `fungsi f() {}` | Fungsi atau fungsi bawaan |

Escape string yang dikenali: `\n`, `\t`, `\r`, `\\`, `\"`, `\'`, dan `\0`.

```nusa
tulis "baris pertama\nbaris kedua\tberakhir"
tulis 'teks dengan "kutip ganda"'
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
| `panjang(nilai)` | teks atau daftar | jumlah karakter (rune) atau elemen |
| `tipe(nilai)` | semua tipe | nama tipe dalam bahasa Indonesia |
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

Semua pemanggilan fungsi bawaan memeriksa jumlah dan tipe argumen serta
menghasilkan galat yang jelas bila tidak sesuai.

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

Perintah `nusa run` mengembalikan exit code `0` bila program berhasil dan `1`
bila terjadi galat leksikal, parsing, atau runtime.
