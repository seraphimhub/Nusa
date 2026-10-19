# Kebijakan Keamanan

## Versi yang Didukung

Perbaikan keamanan diberikan untuk rilis terbaru pada cabang `main`.

| Versi | Didukung |
| --- | --- |
| 1.1.x | ✅ |
| 1.0.x | ❌ |

## Melaporkan Kerentanan

Jangan membuka issue publik untuk masalah keamanan. Gunakan pelaporan privat:

1. Buka tab **Security** repositori ini.
2. Pilih **Report a vulnerability** (GitHub Security Advisory privat).
3. Sertakan langkah reproduksi, berkas `.nusa` contoh, versi Nusa
   (`nusa version`), dan dampaknya.

Kami berusaha membalas dalam 7 hari dan akan mengumumkan perbaikan setelah
tembolok tersedia.

## Praktik Keamanan Repositori

### Proteksi cabang `main`

Cabang `main` dilindungi dengan aturan berikut, berlaku juga untuk pemilik
repositori (tidak ada bypass admin):

| Aturan | Status | Arti |
| --- | --- | --- |
| Force push | Diblokir | Riwayat `main` tidak bisa ditimpa |
| Hapus cabang | Diblokir | `main` tidak bisa dihapus |
| Riwayat linear | Wajib | Tidak ada merge commit yang memecah riwayat |
| Pull request | Tidak diwajibkan | Push langsung ke `main` tetap diizinkan untuk alur kerja cepat |
| Status check CI | Tidak diwajibkan | CI tetap berjalan otomatis, tetapi tidak memblokir push langsung |
| Commit bertanda tangan | Tidak diwajibkan | - |

> Catatan: aturan "wajib status check lulus" sengaja tidak diaktifkan karena
> jika digabung dengan "tidak ada bypass admin", GitHub akan menolak push
> langsung ke `main` (check belum ada untuk commit yang belum ter-push).
> Jika nanti ingin CI menjadi gerbang wajib, aktifkan proteksi pull request:
>
> ```bash
> gh api -X PUT repos/seraphimhub/Nusa/branches/main/protection \
>   -f 'required_pull_request_reviews[required_approving_review_count]=0' \
>   -F 'required_status_checks[strict]=false' \
>   -f 'required_status_checks[contexts][]=Format, vet, test, dan build'
> ```
>
> Setelah itu seluruh perubahan harus masuk lewat pull request yang lulus CI.

### Pemeriksaan otomatis

Setiap push dan pull request ke `main` menjalankan
[`.github/workflows/ci.yml`](.github/workflows/ci.yml) yang memeriksa:

1. `gofmt` — format kode,
2. `go vet` — analisis statis,
3. `go test ./... -cover` — seluruh test,
4. build binary `nusa`,
5. eksekusi semua program pada direktori `examples/`.

### Permukaan serangan

- Nusa **tidak memiliki dependensi eksternal** (`go.mod` tanpa blok `require`),
  sehingga rantai pasok hanya bergantung pada pustaka standar Go.
- Interpreter tidak melakukan akses jaringan, membuka berkas selain skrip yang
  dijalankan, atau menjalankan proses lain. Skrip `.nusa` hanya membaca input
  standar dan menulis ke keluaran standar.
- Tidak ada kredensial, token, atau data pribadi yang disimpan di repositori.

### Untuk kontributor

Sebelum mengirim perubahan, jalankan pemeriksaan yang sama seperti CI:

```bash
make fmt vet test
```

## Pelaporan Bug Biasa

Untuk bug yang bukan masalah keamanan, gunakan issue biasa dan sertakan berkas
`.nusa` minimal yang mereproduksi masalahnya.
