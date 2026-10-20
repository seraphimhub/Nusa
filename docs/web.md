# Kesiapan Bahasa Nusa untuk Pengembangan Web

Dokumen ini merangkum hasil audit kemampuan bahasa Nusa untuk dipakai
menulis aplikasi web. Semua poin di bawah diverifikasi dengan menjalankan
program `.nusa` sungguhan, bukan dengan membaca kode saja.

Ringkas: **Nusa belum siap untuk menulis aplikasi web dari nol.** Bahasa ini
sudah cukup matang sebagai bahasa skrip, template engine, dan logika bisnis,
tetapi belum punya operasi I/O maupun jaringan yang menjadi syarat mutlak
sebuah web.

## Yang Sudah Bisa Digunakan

| Kebutuhan web | Status | Bukti |
| --- | --- | --- |
| Render HTML | Siap | `gabung(daftar, "")` setelah loop menghasilkan markup utuh |
| Serialisasi JSON | Siap | Peta + `untuk k, v dalam peta` |
| Handler sebagai nilai | Siap | Fungsi bisa menjadi argumen (higher-order) |
| Penanganan error per-request | Siap | `coba`/`tangkap` mengembalikan peta `pesan` |
| Partial/template terpisah | Siap | `impor "bagian.nusa"`, modul di-cache sekali jalan |
| Embed dari Go | Siap | `runtime.NewWithIO` + `Execute` tersedia |
| Performa render | Memadai | 20.000 iterasi loop selesai dalam 0,075 detik |

Contoh render HTML yang berhasil:

```nusa
buat baris = []
untuk i dari 1 sampai 3 {
    tambahkan(baris, "<li>Item " + i + "</li>")
}
tulis gabung(baris, "")
```

## Yang Belum Ada (Blocker)

1. **Operasi berkas.** Tidak ada `baca_file` atau `tulis_file`. Satu-satunya
   jalan masuk ke berkas adalah `impor`, yang khusus memuat modul `.nusa`.
2. **Jaringan.** Tidak ada `server`, `dengar`, atau klien HTTP. Tidak ada
   cara membuka socket, menentukan port, atau menerima request.
3. **FFI / embedding dua arah.** Go bisa menjalankan Nusa, tetapi Nusa tidak
   bisa memanggil Go. Tanpa ini, Nusa tidak bisa menjalankan server HTTP
   sendiri; hanya bisa jadi template yang dijalankan oleh host Go.
4. **Literal fungsi anonim.** `fungsi(req) { ... }` gagal di-parse. Pola
   `rute["/"] = fungsi(req) { ... }` tidak bisa ditulis.
5. **Akses properti objek.** `obj.nama` menghasilkan galat leksikal. Objek
   hanya bisa diakses lewat `obj["nama"]`.
6. **Literal dict.** `{ "a": 1 }` gagal di-parse. Peta harus ditulis
   `peta { "a": 1 }`.
7. **Pemisah statement `;`.** Setiap statement harus dipisah baris baru.
8. **Fungsi teks.** Tidak ada `ganti` (replace), `potong_kiri`/`potong_kanan`
   (trim), atau `terdalam` (contains). Teks harus diolah dengan `cari`,
   `potong`, `pisah`, dan `gabung`.
9. **Batas eksekusi tak bisa diatur.** Default 5.000.000 langkah dan
   kedalaman 1000. API Go punya `SetLimits`, tetapi CLI tidak mengekspos
   opsi untuk mengubahnya. Program rekursif seperti `fib(30)` terhenti.
10. **Concurrency.** Tidak ada thread, channel, atau select. Program
    sekuensial.

## Jalur yang Direkomendasikan

### Tahap 1 — Nusa sebagai bahasa render (mulai sekarang, tanpa tunggu)

Pola: Go melayani HTTP, Nusa menulis logika dan markup.

- Go memakai `net/http` sebagai host.
- Halaman dan partial ditulis sebagai berkas `.nusa`.
- Go mem-parse sekali lalu mengeksekusi ulang AST per request dengan
  environment baru, sehingga variabel per-request tidak bocor.
- Data dari request (query string, form) diteruskan lewat environment.

Kelebihan: seluruh logika aplikasi tetap dalam Nusa, dan bottleneck
jaringan ditangani Go yang memang sudah matang untuk itu.

Kekurangan: validasi input dan routing tetap di Go, sehingga logika
terbelah dua bahasa.

### Tahap 2 — Tutup gap bahasa yang murah

Baru-baru ini yang paling besar hasilnya per usaha:

- Fungsi berkonteks runtime: `baca_file`, `tulis_file`, `waktu_sekarang`,
  `acak`, `baca_env`.
- Opsi CLI `--batas-langkah` dan `--batas-kedalaman`.
- Fungsi teks: `ganti`, `terdalam`, `potong_kiri`, `potong_kanan`.
- Literal fungsi anonim dan akses properti `obj.nama`.

### Tahap 3 — Server bawaan (besar, tapi membuka jalan mandiri)

- Fungsi berkonteks: `dengar(port, handler)` yang menjalankan server HTTP
  dengan router sederhana berbasis peta rute.
- Objek `Request` dan `Response` dengan akses properti.
- Fungsi `ke_json` dan `dari_json`.

Pada tahap ini Nusa bisa menjalankan web server sendiri:

```nusa
rute = peta {}
rute["/"] = fungsi(req) {
    kembali "<h1>Halo " + req.kueri["nama"] + "!</h1>"
}

dengar(8080, rute)
```

## Urutan Kerja yang Disarankan

1. Bangun satu contoh web nyata dengan pola Tahap 1 memakai Go host, lalu
   ukur bagian mana yang terasa ganjil. Pengalaman nyata lebih informatif
   daripada perkiraan.
2. Tutup gap Tahap 2 yang memang terasa saat menulis web.
3. Pertimbangkan Tahap 3 setelah server bawaan benar-benar ada.
