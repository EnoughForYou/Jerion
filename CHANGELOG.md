# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.0] - 2026-10-13

### Added
- **Stage T2 (Design System & App Shell)**:
  - **Local Typography & Theming**: Locally bundled Inter variable font (`@fontsource-variable/inter`) without remote CDNs. Complete CSS custom properties token system (`frontend/src/styles/tokens.css`) mapped to `tailwind.config.js` with light/dark/system theme store (`useThemeStore`).
  - **i18n Translation Layer**: Configured `react-i18next` with comprehensive English locale catalog (`frontend/src/i18n/locales/en.json`) strictly eliminating hardcoded user strings across the application shell and views.
  - **Core Design System**: Implemented Radix-backed and semantic UI components in `frontend/src/components/ui/` (`Button`, `Card`, `Badge` with mandatory Icon+Label+Color, `Skeleton`, `EmptyState`, `ErrorState`, `Table`, `Dialog`, `Tooltip`, `Tabs`).
  - **Aceternity Sidebar**: Implemented collapsible 220px/64px left navigation in `frontend/src/components/aceternity/Sidebar.tsx` utilizing pure Tailwind CSS transitions (200ms ease-out) without `framer-motion`, audited in `docs/ui-components.md`.
  - **Application Shell**: Implemented `AppLayout`, `DeviceHeader` (serial, OEM model, battery level, connection badge, switch/disconnect actions), and `StatusBar` (daemon status, log stream count, threat summary, theme switch). Configured HashRouter with 8 module routes.
  - **Preflight Launch View**: Implemented `PreflightView` and `PreflightCheckItem` supporting P1–P7 prerequisites with actionable download/browse buttons, real-time progress, and seamless launch transition.
  - **Mock-Driven Feature Views**: Populated all 8 PRD §5.2 views (`HomeView`, `ScanView`, `AppsView`, `AuditView`, `PopupHunterView`, `MonitorView`, `RemediateView`, `SettingsView`) using real contract fixtures and complete loading/empty/error states.
  - **Three.js Scene v0**: Procedural device 3D scene (`DeviceScene`) with automatic static SVG fallback (`StaticFallback`) for environments without WebGL or with `prefers-reduced-motion: reduce`. Scene pauses on window blur and tab hiding, restricts DPR to `[1, 2]`, and is lazily code-split into a separate bundle chunk.
  - **Testing & Verification**: 92 automated tests across 12 test files with Vitest, strict ESLint with zero warnings, strict TypeScript checking, and verified production bundle dead-code elimination.

## [0.1.0] - 2026-10-12

### Added
- **Stage T1 (Contract & Mock Bridge)**:
  - Strict Zod runtime schemas and inferred TypeScript types in `frontend/src/lib/bridge/schema.ts` covering all PRD §5.2 screens.
  - Wire error protocol (`JERION_ERR:{"code":...}`) and typed `BridgeError` in `frontend/src/lib/bridge/errors.ts`.
  - Grouped `Bridge` contract interfaces and typed `BridgeEventMap` in `frontend/src/lib/bridge/bridge.ts` and `events.ts`.
  - 21 shared JSON fixtures in `internal/testdata/contract/` strictly validated against schemas in Vitest.
  - Stateful `MockBridge` implementation with simulated timers, snapshot getters (`scan.status()`, `monitor.snapshot()`), and sanitized mock data (`MOCK-0001`, `com.example.mock.*`).
  - Production `WailsBridge` stub and environment-based bridge selector (`index.ts`).
  - Bundle verification script `scripts/check-no-mock.js` and automated CI check confirming zero mock/schema artifacts in production builds.
  - Contract governance document `docs/contract.md` defining DRAFT version `0.1.0` and freeze procedures.
  - Architectural Decision Record `docs/adr/0004-contract-schema.md`.
- Project governance updates: Session scope lock, T0–T6 execution stage model in PRD §8.1 and AGENTS §3.

## [0.0.1] - 2026-10-11

### Added
- Wails v2.16.0 desktop application scaffolding with React 19 and TypeScript (`strict: true`).
- Pinned frontend dependencies: Tailwind CSS v3.4.19, PostCSS, Autoprefixer, Three.js 0.186.1, @react-three/fiber 9.8.1, Zustand 5.0.15, Lucide React 1.55.0, TanStack Query, TanStack Table, TanStack Virtual, React Router DOM, i18next.
- Design tokens system (`frontend/src/styles/tokens.css`) mapped to `tailwind.config.js`.
- Architecture directories per PRD §6.3 with `.gitkeep` trackers.
- Automated testing harness: Go test placeholder, Vitest + Testing Library + jsdom suite, ESLint flat config, Prettier with LF line endings.
- GitHub Actions CI workflow for Ubuntu test validation and Windows Wails production build.
- Documentation: `docs/PRD.md`, `docs/PROGRESS.md`, `docs/adr/0001-stack.md`, `docs/adr/0002-adb-distribution-preflight.md`, `docs/adr/0003-ui-english-only.md`, `docs/ui-components.md`, `docs/rules-authoring.md`, `docs/devices/infinix-note-11s.md`.
