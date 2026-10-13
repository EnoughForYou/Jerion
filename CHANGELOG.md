# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
