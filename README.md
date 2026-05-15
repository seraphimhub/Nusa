# Nusa

Nusa adalah bahasa pemrograman modern dengan sintaks Indonesia yang fokus pada kesederhanaan dan keterbacaan.

## Menjalankan

```bash
go run ./cmd/nusa run examples/halo.nusa
```

## Sintaks yang didukung

- `buat` untuk deklarasi variabel
- `tulis` untuk output
- `jika` / `kalau_tidak` untuk percabangan
- `ulang` untuk perulangan
- `fungsi` / `panggil` untuk fungsi sederhana
- Operasi aritmetika: `+`, `-`, `*`, `/`
- Operasi perbandingan: `>`, `<`, `>=`, `<=`, `==`, `!=`

## Catatan parser/lexer

- String mendukung spasi, contoh: `tulis "halo dunia"`
- Komentar baris didukung dengan awalan `#` atau `//`
