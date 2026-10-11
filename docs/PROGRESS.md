# PROGRESS

## Status sekarang
- Fase aktif: 0 — Fondasi (Setup)
- Branch: phase/0-foundation
- Terakhir selesai: — (Tahap inisiasi)

## Fase 0 — checklist
- [x] Git init di branch `main` dengan `.gitattributes` (LF), `.gitignore`, `LICENSE`, `README.md`, `CHANGELOG.md`
- [x] Dokumentasi inti di `docs/` (`docs/PRD.md`, `docs/PROGRESS.md`, ADR 0001–0003)
- [x] Scaffold Wails v2 (React+TS) tanpa folder bersarang
- [ ] Struktur folder arsitektur sesuai PRD §6.3
- [ ] Backend setup: module path `github.com/EnoughForYou/Jerion`, harness test Go placeholder (sementara, dihapus di T4)
- [ ] Frontend setup: Tailwind v3.4.19, tokens CSS, dependencies terkunci (React 18.3.1, Three 0.170.0, R3F 8.18.0, Zustand, Lucide, dll.)
- [ ] Tooling: ESLint strict, Prettier (LF), Vitest setup
- [ ] Spike WebGL: scene Three.js minimal di WebView2 (sekali pakai, tidak di-commit)
- [ ] CI workflow GitHub Actions (pin runner Windows untuk wails build, Ubuntu untuk test)
- [ ] Acceptance & Gerbang Selesai-Fase 0

## Keputusan baru (ADR)
- ADR-0001: Stack terkunci (Wails v2.16.0, Go 1.27.1, React 18.3.1, Tailwind v3.4.19, Three 0.170.0, R3F 8.18.0)
- ADR-0002: Kebijakan distribusi ADB dan preflight per-peluncuran
- ADR-0003: Kebijakan UI English-only dengan lapisan i18n

## Masalah terbuka / blocker
- Remote GitHub privat belum diset (menunggu URL dari owner untuk push & tag v0.0.1)
- CI hijau: Belum terverifikasi (menunggu remote repository terhubung)
- golangci-lint lokal: Tidak terverifikasi (binary golangci-lint belum terpasang di mesin dev)

## Pengukuran & Spike
- WebGL Spike di WebView2: Belum dijalankan (akan diuji di Sub-tugas 0.8)

## Usulan
- —
