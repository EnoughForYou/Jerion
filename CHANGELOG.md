# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.0.1] - 2026-10-11

### Added
- Wails v2.16.0 desktop application scaffolding with React 19 and TypeScript (`strict: true`).
- Pinned frontend dependencies: Tailwind CSS v3.4.19, PostCSS, Autoprefixer, Three.js 0.186.1, @react-three/fiber 9.8.1, Zustand 5.0.15, Lucide React 1.55.0, TanStack Query, TanStack Table, TanStack Virtual, React Router DOM, i18next.
- Design tokens system (`frontend/src/styles/tokens.css`) mapped to `tailwind.config.js`.
- Architecture directories per PRD §6.3 with `.gitkeep` trackers.
- Automated testing harness: Go test placeholder, Vitest + Testing Library + jsdom suite, ESLint flat config, Prettier with LF line endings.
- GitHub Actions CI workflow for Ubuntu test validation and Windows Wails production build.
- Documentation: `docs/PRD.md`, `docs/PROGRESS.md`, `docs/adr/0001-stack.md`, `docs/adr/0002-adb-distribution-preflight.md`, `docs/adr/0003-ui-english-only.md`, `docs/ui-components.md`, `docs/rules-authoring.md`, `docs/devices/infinix-note-11s.md`.
