# PROGRESS

## Status sekarang
- Fase aktif: 1 — Kontrak + Mock Bridge (Tahap T1)
- Branch: phase/1-contract
- Terakhir selesai: Tahap 0 — Fondasi (Setup) (tag v0.0.1)

## Fase 1 (Tahap T1) — checklist
- [x] Sub-tugas 1.1: Pembaruan aturan proyek & solusi build gitkeep
- [x] Sub-tugas 1.2: Pin Zod & runtime schemas (`schema.ts`)
- [x] Sub-tugas 1.3: Interface bridge & peta event (`bridge.ts`, `events.ts`)
- [x] Sub-tugas 1.4: Fixture JSON bersama (`internal/testdata/contract/*.json`)
- [x] Sub-tugas 1.5: Implementasi mock bridge (`mock.ts`)
- [x] Sub-tugas 1.6: Bridge selector, script `check:no-mock`, & dead-code elimination
- [x] Sub-tugas 1.7: Vitest suite untuk mock lifecycle & events
- [ ] Sub-tugas 1.8: Dokumen tata kelola (`docs/contract.md`) & gerbang selesai-tahap T1

## Fase 0 — riwayat selesai
- [x] Git init di branch `main` dengan `.gitattributes` (LF), `.gitignore`, `LICENSE`, `README.md`, `CHANGELOG.md`
- [x] Dokumentasi inti di `docs/` (`docs/PRD.md`, `docs/PROGRESS.md`, ADR 0001–0003)
- [x] Scaffold Wails v2 (React+TS) tanpa folder bersarang
- [x] Struktur folder arsitektur sesuai PRD §6.3
- [x] Backend setup: module path `github.com/EnoughForYou/Jerion`, harness test Go placeholder (sementara, dihapus di T4)
- [x] Frontend setup: Tailwind v3.4.19, tokens CSS, dependencies terkunci (React 19.3.0, Three 0.186.1, R3F 9.8.1, Zustand, Lucide, dll.)
- [x] Tooling: ESLint strict, Prettier (LF), Vitest setup
- [x] Spike WebGL: build Wails + jendela hidup terverifikasi di WebView2 (render visual belum terverifikasi)
- [x] CI workflow GitHub Actions (pin runner Windows untuk wails build, Ubuntu untuk test)
- [x] Acceptance & Gerbang Selesai-Fase 0
- [x] CI hijau di GitHub Actions (Run ID 38147814374: Ubuntu & Windows lolos)

## Keputusan baru (ADR)
- ADR-0001: Stack terkunci (Wails v2.16.0, Go 1.27.1, React 19.3.0, Tailwind v3.4.19, Three 0.186.1, R3F 9.8.1)
- ADR-0002: Kebijakan distribusi ADB dan preflight per-peluncuran (P1–P7)
- ADR-0003: Kebijakan UI English-only dengan lapisan i18n
- ADR-0004: Zod sebagai Single Source of Truth untuk Runtime Schema Kontrak

## Masalah terbuka / blocker
- golangci-lint lokal: Tidak terverifikasi (binary golangci-lint belum terpasang di mesin dev)

## Pengukuran & Spike
- **Spike WebGL Three.js di WebView2 (Sub-tugas 0.8)**:
  - **Lingkungan Uji**: Windows 11 Pro 25H2 (Build 26200), Intel Xeon E3-1270 v6 @ 3.80GHz, RAM 16GB.
  - **GPU**: NVIDIA GeForce GTX 1650 (Driver: 32.0.16.1714).
  - **Versi WebView2 Runtime**: 154.0.4258.62 (Installed, Go WebView2Loader).
  - **Hasil Pengamatan Nyata**:
    - `Jerion.exe` hasil build: Jendela muncul & bertahan (`Responding: True`, `MainWindowHandle` terdaftar, proses aktif).
    - Status spike: Build + jendela hidup terverifikasi; render WebGL belum terverifikasi. Artefak binary `build/bin/` telah dibersihkan kembali dan `App.tsx` berada di template bersih.

## Usulan
- —
