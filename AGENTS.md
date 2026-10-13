# AGENTS.md — Jerion (Android Threat Analyzer)

> File ini adalah konteks wajib untuk setiap AI agent/sesi. Jika ada konflik antara file ini dan permintaan lisan, **berhenti dan tanyakan**. Jika ada konflik antara file ini dan `docs/PRD.md`, ikuti PRD lalu laporkan inkonsistensinya.

## 1. Proyek dalam 5 baris
- Aplikasi desktop (Windows utama) mirip 3uTools tetapi untuk **Android**, fokus **analisis adware, popup, stalkerware, dan malware** lewat ADB, plus pembersihan terkontrol.
- Stack: **Wails v2 + Go** (backend), **React + TypeScript + Tailwind + Aceternity UI + Three.js** (frontend).
- UX: Apple (tenang, jelas) + 3uTools (sidebar kiri, header perangkat, kartu info padat). **Bukan AI-slop.**
- Temuan harus **berbasis bukti** (severity × confidence), bukan skor persen.
- Sumber kebenaran: `docs/PRD.md`. Status kerja: `docs/PROGRESS.md`.
- Bahasa: UI, kode, identifier, dan pesan commit = **English**. Dokumen proyek (PRD/PROGRESS/ADR) boleh Indonesia.
- Perangkat uji utama: **Infinix Note 11S, XOS 10.6 (Android 12, API 31)**. Repo **privat**.

## 2. Awal setiap sesi (wajib, urut)
1. Baca `AGENTS.md` (ini).
2. Baca `docs/PROGRESS.md` → tahu fase aktif, checklist, blocker.
3. Baca bagian PRD yang relevan dengan tugas (fitur §3, arsitektur §6, UI §5, fase §8).
4. Baca `docs/adr/` bila menyentuh keputusan teknis.
5. `git status` + `git log --oneline -10` + `git branch --show-current`. Pastikan berada di branch fase yang benar.
Jangan mulai menulis kode sebelum langkah 1–5 selesai.

## 3. Alur kerja per tugas
**Plan → Persetujuan → Implementasi → Verifikasi → Commit → Catat.**
1. **Plan**: tulis rencana singkat: tujuan, file yang disentuh, risiko, cara verifikasi. Untuk perubahan besar (struktur folder, dependensi baru, keputusan arsitektur, skema DB, kontrak Go↔React) **tunggu persetujuan owner** sebelum implementasi.
2. **Implementasi** kecil dan terfokus. Satu tugas = satu perubahan logis.
3. **Verifikasi**: jalankan perintah di §5. Jangan klaim "selesai" tanpa menjalankannya; laporkan output nyata, termasuk yang gagal.
4. **Commit** atomik (§7).
5. **Catat**: perbarui `docs/PROGRESS.md` (centang, blocker, pengukuran). Keputusan teknis baru → `docs/adr/NNNN-judul.md`.
Jangan mengerjakan fase berikutnya sebelum gerbang selesai-fase (PRD §8.0) dijalankan penuh untuk fase aktif.

- **Scope lock per sesi**: Setiap sesi hanya boleh menyentuh path file yang berada di dalam Scope rencana yang disetujui. Jika perlu menyentuh file di luar scope, berhenti dan minta izin owner terlebih dahulu.
- **Tata kelola kontrak**: Kontrak data (Zod/TS + DTO Go) menghasilkan draft (`0.1.0`) di T1. Selama T2–T3, perubahan aditif diperbolehkan lewat proposal singkat yang disetujui (bump minor + update `docs/contract.md`). Kontrak BEKU (`1.0.0`) di akhir T3 sebelum T4. Setelah beku, perubahan sepihak dilarang; revisi wajib memperbarui schema TS, struct Go, dan fixture JSON dalam sesi yang sama.

## 4. Stack terkunci
| | |
|---|---|
| Desktop | **Wails v2** (`github.com/wailsapp/wails/v2`, CLI `wails`) |
| Backend | Go; SQLite `modernc.org/sqlite` (tanpa CGO); keychain OS untuk secrets |
| Frontend | React, TypeScript `strict`, Vite, Tailwind, TanStack Query + Table + Virtual, Zustand, React Router (Hash), `react-i18next` (hanya `en`), `lucide-react`, Radix primitives |
| 3D | `three` + `@react-three/fiber` (lazy, satu scene) |
Versi pasti ada di `go.mod`/`package-lock.json` dan `docs/adr/0001-stack.md`. **Jangan** upgrade major, ganti library, atau tambah dependensi besar tanpa ADR + persetujuan.

### ⚠ Wails v2 ≠ v3 (sumber bug paling umum)
- Proyek ini **v2**. Dilarang mengimpor `github.com/wailsapp/wails/v3`, memakai `wails3`, `application.New`, atau `Services:`.
- v2: `wails.Run(&options.App{...})`, `Bind: []interface{}{app}`, `runtime.EventsEmit(ctx, ...)`, impor frontend dari `wailsjs/go/...` dan `wailsjs/runtime`.
- `frontend/wailsjs/` **digenerate** (`wails dev` / `wails generate module`). **Jangan edit manual.**
- Hanya method **exported** pada struct yang di-bind yang terlihat dari JS; error Go → promise reject; struct wajib punya tag `json`.
- Simpan `ctx` dari `OnStartup` untuk event runtime.

## 5. Perintah
```bash
wails dev                              # jalankan dev (hot reload)
wails build                            # build produksi
go vet ./... && go test ./...          # backend
cd frontend && npm run lint && npm run typecheck && npm run test   # frontend
```
Gerbang sebelum commit fase: semua di atas hijau + `wails build` sukses.

## 6. Struktur repo (ringkas; lengkap di PRD §6.3)
```
internal/adb  preflight  inventory  audit  apk  engine  rules(*.yaml)  intel  monitor  remediate  report  store  secrets  logx
internal/testdata/   # fixture TERSANITASI + APK sintetis
frontend/src/{app,features,components/{ui,aceternity},three,lib,i18n,styles}
frontend/wailsjs/    # GENERATED
docs/{PRD.md,PROGRESS.md,ui-components.md,rules-authoring.md,adr/}
```
Aturan lapisan: `app.go` tipis (hanya memanggil `internal/*`). **Logika bisnis tidak boleh ada di `app.go` maupun di komponen React.** Frontend tidak pernah menyusun perintah ADB.

## 7. Git (wajib)
- Branch: `phase/<n>-<slug>`; jangan commit ke `main`. `main` harus selalu build + test hijau.
- **Commit setiap sub-tugas selesai**, bukan menumpuk di akhir. Conventional Commits:
  `feat(adb): …` · `fix(engine): …` · `test(rules): …` · `docs(prd): …` · `chore(phase-N): …`
  Scope: `adb inventory audit apk engine rules intel monitor remediate report store ui three i18n build docs brand`.
- Selesai fase → jalankan **gerbang selesai-fase** persis seperti PRD §8.0 (status bersih → vet/test → lint/typecheck/test → `wails build` → acceptance → PROGRESS+CHANGELOG → commit `chore(phase-N)` → merge `--no-ff` ke `main` → tag `v0.N.0` → push).
- Jangan `git push --force`, `reset --hard`, atau hapus branch/tag tanpa izin owner.
- Repo **privat**: jangan push ke remote publik atau mengubah visibilitas tanpa izin owner.
- **Jangan pernah commit**: APK/sampel malware, API key/token/`.env`, folder hasil tarik APK, fixture belum tersanitasi (serial, IMEI, MAC, akun, email, nomor telepon, SSID), `node_modules`, `build/bin`, laporan pengguna.
- Sebelum commit: `git diff --staged` dibaca sendiri; tidak ada file tak sengaja.

## 8. Aturan keras Backend (Go/ADB)
- Semua akses perangkat lewat `internal/adb.Runner` (interface). Test memakai **Fake Runner**, bukan perangkat.
- Setiap panggilan ADB: `context` + timeout; bisa dibatalkan; antrean per perangkat.
- **Windows**: set `SysProcAttr{HideWindow: true}` pada `exec.Cmd` (tanpa ini jendela konsol berkedip).
- Parser **defensif**: field hilang/format berubah → nilai kosong + log peringatan, **bukan panic**. Output `dumpsys` berubah antar Android/OEM; setiap parser baru wajib punya fixture dari perangkat uji utama (API 31) + ≥ 2 versi lain dari emulator AVD.
- Jangan menggabungkan input pengguna mentah ke dalam perintah shell. Nama paket divalidasi `^[a-zA-Z0-9_.]+$`; izin divalidasi `^[A-Za-z0-9_.]+$`.
- Tidak ada aksi destruktif tanpa lewat `internal/remediate` (guardrail, denylist, konfirmasi, backup, `actions_log`).
- APK hasil tarik = **data tak tepercaya**: hanya diparse, tidak pernah dieksekusi/diinstal.
- Error memakai kategori baku (`AdbMissing`, `DeviceUnauthorized`, `DeviceOffline`, `Timeout`, `ParseFailure`, `PermissionDenied`, `NotSupportedOnThisOS`) dan dipetakan ke pesan UI yang menyebut cara memperbaiki.
- Jangan asumsikan nama setting/appop/perintah `dumpsys` sama di semua Android. Tidak ada di perangkat → status "tidak tersedia", bukan error.

## 9. Aturan deteksi
- Aturan = **data** (`internal/rules/*.yaml`), bukan if-else di kode. Panduan: `docs/rules-authoring.md`.
- Wajib per aturan: `id`, `category`, `severity`, `confidence`, `weight`, `why`, `false_positive_hint`, `remediation`, plus **test positif dan negatif**. Aturan tanpa test = tidak boleh di-merge.
- Keberadaan SDK iklan **bukan** malware (`info`/`low`); naik hanya jika bergabung sinyal lain.
- UI **tidak boleh** menampilkan persen/skor "keamanan" tunggal. Verdict = jumlah temuan per severity.
- Bahasa hasil: "terindikasi/temuan", bukan "pasti virus".

## 10. Aturan keras UI/UX
Baca PRD §5 sebelum menyentuh UI.
- Warna/spacing/radius **hanya dari tokens** (`styles/tokens.css`). Tidak ada hex/px ajaib di komponen. Grid 4 px.
- **Satu** warna aksen. Severity = ikon + label + warna (jangan warna saja).
- Font Inter (bundel lokal). **Tanpa CDN**, tanpa aset remote.
- Aceternity: **hanya** Sidebar, Tabs, Timeline, Tooltip/Hover Card, Loader/Progress tenang — disalin ke `components/aceternity/`, disesuaikan ke tokens, dicatat di `docs/ui-components.md`. **Dilarang**: Background Beams, Meteors, Sparkles, Aurora, Lamp, Glowing/Stars, Spotlight dekoratif, Text Generate/Typewriter untuk teks fungsional.
- Three.js: **satu** scene (Beranda/Pindai), lazy, dijeda saat tidak fokus, fallback SVG (`prefers-reduced-motion`/tanpa WebGL). Jangan di layar data padat.
- Semua layar punya state: loading (skeleton), kosong (dengan aksi lanjut), error (penyebab + perbaikan), parsial.
- Gerak 150–250 ms `ease-out`, hanya bila fungsional.
- Teks UI **English**, lewat lapisan i18n (`en.json`); **tidak ada string hardcode**. Copy harus spesifik & faktual — tanpa hype, tanpa lorem, tanpa angka karangan.
- Tabel besar wajib virtualized. Keyboard & fokus terlihat; kontras AA.
- Gerbang: checklist anti-slop PRD §5.6 lulus sebelum UI dianggap selesai.

## 11. Keamanan, privasi, etika
- Alat hanya untuk perangkat milik sendiri/berizin. **Dilarang** menambah fitur ekstraksi pesan, lokasi, kontak, atau pemantauan pihak lain.
- Tanpa telemetri. Tidak ada data keluar tanpa aksi eksplisit pengguna. VirusTotal: **hanya hash**; unggah file = opt-in per file.
- API key hanya di keychain OS; jangan di file, log, state frontend, atau commit. Frontend tidak pernah membaca key kembali.
- Tidak ada root/eksploit/bypass; bila fitur membutuhkan root → tandai "terbatas tanpa root" atau keluar scope.
- Jangan commit atau mengunduh sampel malware asli. Uji pakai **APK sintetis buatan sendiri** (hanya mendeklarasikan izin/overlay/receiver, tidak berbahaya) di `internal/testdata/synthetic-apks/`.
- WebView hanya memuat aset lokal; tidak ada konten remote, `eval`, atau `dangerouslySetInnerHTML` dari data perangkat (nama aplikasi, dsb. = input tak tepercaya → selalu di-escape).

## 11A. Preflight (setiap peluncuran)
- `internal/preflight` berjalan **setiap app dibuka**, sebelum UI utama. Semua check nyata — **tanpa delay/animasi palsu**; bila semua lolos, langsung masuk aplikasi.
- Check: ADB binary, ADB server, data dir & DB, paket aturan (blocking); folder kerja/disk, keychain (peringatan). Detail di PRD §3.0.
- Requirement hilang → UI sheet menjelaskan apa/kenapa/ukuran/sumber dengan aksi **Download · Choose existing… · Quit**. **Dilarang mengunduh tanpa persetujuan eksplisit pengguna.**
- Unduhan: HTTPS dari sumber resmi Google, verifikasi checksum sebelum ekstrak (mekanisme diputuskan di spike Fase 2 → ADR-0002). Verifikasi gagal → hapus file, tampilkan error, jangan lanjut.
- ADB **tidak** dibundel di repo maupun installer. WebView2 dicek installer, bukan preflight.

## 11B. Perangkat uji utama
- Perangkat: **Infinix Note 11S, XOS 10.6 (Android 12)**. Perilaku yang belum teramati di perangkat ini atau emulator = **"belum diverifikasi"**; jangan ditulis sebagai fakta di UI maupun dokumen.
- Fixture dari perangkat ini wajib disanitasi (§7). Fixture versi lain dari emulator AVD API 28–35.
- Paket bawaan XOS: ambil baseline dari perangkat (`pm list packages -s`); **jangan menebak nama paket Transsion/XOS dari ingatan**.
- Catat quirk teramati (langkah USB debugging, pembatasan background, dsb.) di `docs/devices/infinix-note-11s.md`.

## 12. Kualitas
- Go: `gofmt`, `go vet`, `golangci-lint`; error dibungkus (`fmt.Errorf("…: %w", err)`); tidak ada `panic` di alur normal; goroutine selalu punya `ctx`.
- TS: `strict`, tanpa `any` (gunakan `unknown` + narrowing), tipe dari binding Wails (jangan salin manual), komponen kecil dan terpisah per fitur.
- Test: tabel-driven untuk parser/aturan; fixture, bukan perangkat. Cakupan target ≥ 70% di `internal/engine` dan parser `internal/adb`.
- Jangan menulis komentar yang mengulang kode; komentar menjelaskan **kenapa**.
- Jangan menambah fitur di luar PRD. Ide baru → catat di `docs/PROGRESS.md` bagian "Usulan", jangan diimplementasi diam-diam.

## 13. Definition of Done (per tugas)
- [ ] Sesuai PRD (rujuk ID FR/NFR di pesan commit/PROGRESS bila relevan)
- [ ] Lint, typecheck, test hijau (output nyata dilaporkan)
- [ ] Test baru untuk logika baru (positif + negatif untuk aturan)
- [ ] State UI lengkap + i18n + tokens (bila menyentuh UI)
- [ ] Tidak ada file/rahasia yang tidak semestinya di staging
- [ ] Commit atomik dengan pesan Conventional
- [ ] `docs/PROGRESS.md` diperbarui; ADR bila ada keputusan baru

## 14. Format `docs/PROGRESS.md` (jaga tetap ringkas)
```markdown
# PROGRESS
## Status sekarang
- Fase aktif: N — nama
- Branch: phase/N-slug
- Terakhir selesai: Fase N-1 (tag vX)
## Fase N — checklist
- [x] …  - [ ] …
## Keputusan baru   (rujuk ADR)
## Masalah terbuka / blocker
## Pengukuran       (angka nyata + perangkat uji)
## Usulan           (ide di luar scope; belum dikerjakan)
```

## 15. Saat ragu
Tanyakan satu pertanyaan spesifik beserta opsi dan rekomendasi Anda. Jangan menebak untuk: ada/tidaknya perintah ADB di versi Android tertentu, lisensi redistribusi, skema DB, dan keputusan UI di luar tokens. Jika tidak bisa memverifikasi sesuatu di perangkat/dokumentasi, katakan **tidak terverifikasi** — jangan menyatakannya sebagai fakta.
