# PROGRESS

## Status sekarang
- Fase aktif: 2 — Design System & App Shell (Tahap T2 — Selesai)
- Branch: phase/2-shell
- Terakhir selesai: Tahap 1 — Kontrak + Mock Bridge (tag v0.1.0)

## Fase 2 (Tahap T2) — checklist
- [x] Sub-tugas 2.1: Local variable font `@fontsource-variable/inter`, CSS tokens & Zustand theme store
- [x] Sub-tugas 2.2: Konfigurasi `react-i18next` & catalog English (`en.json`) tanpa string hardcode
- [x] Sub-tugas 2.3: Implementasi Core Design System (`Button`, `Card`, `Badge` [icon+label+color], `Skeleton`, `EmptyState`, `ErrorState`, `Table`, `Dialog`, `Tooltip`, `Tabs`)
- [x] Sub-tugas 2.4: Implementasi Aceternity Sidebar (220px/64px, pure CSS 200ms ease-out, zero framer-motion) & registrasi di `docs/ui-components.md`
- [x] Sub-tugas 2.5: Implementasi layout shell, DeviceHeader, StatusBar, dan HashRouter 8 rute
- [x] Sub-tugas 2.6: Implementasi PreflightView dengan alur verifikasi P1–P7, aksi unduh/pilih, dan transisi ke app utama
- [x] Sub-tugas 2.7: Populasi seluruh 8 modul placeholder dengan fixture kontrak dan state handling (loading/empty/error)
- [x] Sub-tugas 2.8: Three.js Scene v0 (`DeviceScene`, `DeviceModel`), fallback SVG (`StaticFallback`), pause on blur/hidden, bounded DPR `[1, 2]`, dan lazy code-split
- [x] Sub-tugas 2.9: Audit checklist PRD §5.6 & gerbang selesai-tahap T2

## Fase 1 (Tahap T1) — riwayat selesai
- [x] Sub-tugas 1.1: Pembaruan aturan proyek & solusi build gitkeep
- [x] Sub-tugas 1.2: Pin Zod & runtime schemas (`schema.ts`)
- [x] Sub-tugas 1.3: Interface bridge & peta event (`bridge.ts`, `events.ts`)
- [x] Sub-tugas 1.4: Fixture JSON bersama (`internal/testdata/contract/*.json`)
- [x] Sub-tugas 1.5: Implementasi mock bridge (`mock.ts`)
- [x] Sub-tugas 1.6: Bridge selector, script `check:no-mock`, & dead-code elimination
- [x] Sub-tugas 1.7: Vitest suite untuk mock lifecycle & events
- [x] Sub-tugas 1.8: Dokumen tata kelola (`docs/contract.md`) & gerbang selesai-tahap T1

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

## Pengukuran & Kualitas (Tahap T2)
- **Vitest Suite**: 92 pengujian lolos di 12 file tes (`src/three/scene.test.tsx`, `src/features/preflight/preflight.test.tsx`, `src/components/aceternity/Sidebar.test.tsx`, `src/components/ui/components.test.tsx`, `src/app/shell.test.tsx`, `src/App.test.tsx`, `src/features/placeholders.test.tsx`, `src/lib/bridge/mock.test.ts`, `src/i18n/i18n.test.ts`, `src/app/theme.test.ts`, `src/lib/bridge/schema.test.ts`, `src/lib/bridge/fixtures.test.ts`).
- **ESLint & TypeScript**: 0 error, 0 warning pada `npm run lint` & `npm run typecheck`.
- **Mock Elimination (`npm run check:no-mock`)**: 0 marker atau skema runtime mock terdeteksi pada bundle produksi Vite.
- **Bundle Splitting**: `DeviceScene` berhasil dipisah secara dinamis (chunk ~923 kB / gzip ~248 kB), bundle aplikasi inti tetap ramping (~481 kB / gzip ~151 kB).
- **Anti-AI-Slop Checklist (PRD §5.6)**: 100% dipatuhi (1 warna aksen, severity berbasis icon+label+warna, font lokal tanpa CDN, zero framer-motion, 1 scene 3D lazy dengan SVG fallback dan DPR `[1, 2]`, teks faktual tanpa lorem).

## Usulan
- —
