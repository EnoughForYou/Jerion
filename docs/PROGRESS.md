# PROGRESS

## Status sekarang
- Fase aktif: 0 — Fondasi (Setup)
- Branch: phase/0-foundation
- Terakhir selesai: Fase 0 — Fondasi (Setup) (menunggu tag v0.0.1 pasca-remote)

## Fase 0 — checklist
- [x] Git init di branch `main` dengan `.gitattributes` (LF), `.gitignore`, `LICENSE`, `README.md`, `CHANGELOG.md`
- [x] Dokumentasi inti di `docs/` (`docs/PRD.md`, `docs/PROGRESS.md`, ADR 0001–0003)
- [x] Scaffold Wails v2 (React+TS) tanpa folder bersarang
- [x] Struktur folder arsitektur sesuai PRD §6.3
- [x] Backend setup: module path `github.com/EnoughForYou/Jerion`, harness test Go placeholder (sementara, dihapus di T4)
- [x] Frontend setup: Tailwind v3.4.19, tokens CSS, dependencies terkunci (React 19.1.0, Three 0.186.1, R3F 9.8.1, Zustand, Lucide, dll.)
- [x] Tooling: ESLint strict, Prettier (LF), Vitest setup
- [x] Spike WebGL: scene Three.js minimal di WebView2 (sekali pakai, tidak di-commit)
- [x] CI workflow GitHub Actions (pin runner Windows untuk wails build, Ubuntu untuk test)
- [x] Acceptance & Gerbang Selesai-Fase 0

## Keputusan baru (ADR)
- ADR-0001: Stack terkunci (Wails v2.16.0, Go 1.27.1, React 19.1.0, Tailwind v3.4.19, Three 0.186.1, R3F 9.8.1)
- ADR-0002: Kebijakan distribusi ADB dan preflight per-peluncuran
- ADR-0003: Kebijakan UI English-only dengan lapisan i18n

## Masalah terbuka / blocker
- Remote GitHub privat belum diset (menunggu URL dari owner untuk push & tag v0.0.1)
- CI hijau: Belum terverifikasi (menunggu remote repository terhubung)
- golangci-lint lokal: Tidak terverifikasi (binary golangci-lint belum terpasang di mesin dev)

## Pengukuran & Spike
- **Spike WebGL Three.js di WebView2 (Sub-tugas 0.8)**:
  - **Lingkungan Uji**: Windows 11 Pro 25H2 (Build 26200), Intel Xeon E3-1270 v6 @ 3.80GHz, RAM 16GB.
  - **GPU**: NVIDIA GeForce GTX 1650 (Driver: 32.0.16.1714).
  - **Versi WebView2 Runtime**: 154.0.4258.62 (Installed, Go WebView2Loader).
  - **Scene yang Diuji**: R3F v9.8.1 `<Canvas>` dengan BoxGeometry prosedural `[2, 2, 2]`, pencahayaan AmbientLight + DirectionalLight, dan MeshStandardMaterial dengan warna token `--accent` (`#0A84FF`).
  - **Hasil Render & Eksekusi**:
    - Wails compiler bundling aset frontend Vite dan Three.js sukses (`Built 'Jerion.exe' in 13.04s`).
    - Inisialisasi WebView2 berjalan sukses (`Executable launched successfully: True`).
    - Konteks WebGL terakselerasi hardware terdeteksi aktif pada GPU diskret (NVIDIA GTX 1650) tanpa error kompilasi shader maupun crash WebView2.
    - Seluruh kode spike sekali pakai telah dibersihkan kembali dari `frontend/src/App.tsx` dan artefak binary `build/bin/` telah dihapus.

## Usulan
- —
