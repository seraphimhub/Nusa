# Kebijakan Versi dan Prosedur Rilis

Dokumen ini menjelaskan bagaimana nomor versi Nusa ditentukan dan bagaimana
rilis dipublikasikan.

## Skema Versi

Nusa memakai [Semantic Versioning](https://semver.org/lang/id/): `MAJOR.MINOR.PATCH`.

| Perubahan | Kenaikan | Contoh |
| --- | --- | --- |
| Perubahan sintaks/evaluasi yang tidak kompatibel | `MAJOR` | `1.4.0` → `2.0.0` |
| Fitur baru yang tetap kompatibel | `MINOR` | `1.1.0` → `1.2.0` |
| Perbaikan bug tanpa perubahan fitur | `PATCH` | `1.1.0` → `1.1.1` |

Selama versi masih `0.y.z`, API dan sintaks bahasa dianggap belum stabil.

## Format Tag

- Setiap rilis memakai **annotated tag** bernama `vX.Y.Z` (contoh `v1.1.0`),
  bukan lightweight tag, agar membawa tanggal, pembuat, dan pesan rilis.
- Satu tag hanya untuk satu rilis, dan tag yang sudah dipublikasikan **tidak
  pernah** dipindahkan atau diubah isinya. Kesalahan rilis diperbaiki dengan
  versi baru.
- Tag `v*` dilindungi oleh ruleset repositori: penghapusan dan pemindahan
  tag rilis diblokir.
- Tag pra-1.0 (`v0.3.0`–`v0.7.2`) dipertahankan sebagai tonggak pengembangan
  dan tidak diubah lagi.

## Prosedur Rilis

1. Pastikan cabang `main` hijau: `make fmt vet test`.
2. Perbarui `CHANGELOG.md`: tambahkan bagian `## [X.Y.Z] - YYYY-MM-DD` dengan
   kelompok **Ditambahkan**, **Diubah**, **Diperbaiki**, dan **Dihapus**.
3. Naikkan konstanta `VERSION` di `cmd/nusa/main.go`.
4. Commit dan push perubahan tersebut ke `main`.
5. Buat annotated tag dan push:

   ```bash
   git tag -a vX.Y.Z -m "Nusa X.Y.Z: ringkasan singkat rilis"
   git push origin vX.Y.Z
   ```

6. Publikasikan GitHub Release dengan catatan dari CHANGELOG:

   ```bash
   gh release create vX.Y.Z \
     --title "Nusa vX.Y.Z - judul singkat" \
     --notes-file catatan.md \
     --verify-tag --latest
   ```

   Atau pakai catatan otomatis:

   ```bash
   gh release create vX.Y.Z --generate-notes --verify-tag
   ```

   Kategori catatan otomatis diatur pada [`.github/release.yml`](../.github/release.yml).

7. Verifikasi hasilnya:

   ```bash
   gh release list
   gh release view vX.Y.Z
   ```

## Riwayat Versi

| Versi | Tanggal | Catatan |
| --- | --- | --- |
| 1.2.0 | 2026-10-20 | Peta, modul, anotasi tipe, `coba`/`tangkap`, `nusa cek`/`fmt`/`uji` |
| 1.1.0 | 2026-10-19 | Daftar, `untuk`, `pilih`, 23 fungsi bawaan, 88 test |
| 1.0.0 | 2026-07-01 | Rilis stabil pertama |
| 0.7.2 | 2026-04-14 | `kalau_tidak` |
| 0.7.1 | 2026-04-14 | Operator perbandingan lengkap |
| 0.7.0 | 2026-04-14 | Operator aritmatika lengkap |
| 0.6.0 | 2026-04-14 | Perbaikan token plus |
| 0.5.0 | 2026-04-14 | Fungsi dan `panggil` |
| 0.4.0 | 2026-04-14 | Perulangan `ulang` |
| 0.3.0 | 2026-04-14 | Pernyataan `jika` |

Daftar lengkap beserta tombol unduh: <https://github.com/seraphimhub/Nusa/releases>
