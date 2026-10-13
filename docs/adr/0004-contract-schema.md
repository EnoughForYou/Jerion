# ADR-0004: Zod Runtime Schema as Single Source of Truth for Contract

## Status
Accepted

## Date
2026-10-13

## Context
In Stage T1 (Contract + Mock Bridge), the application contract between the Go backend and React frontend is defined prior to backend implementation. Pure TypeScript types are erased at runtime during compilation, making automated runtime validation of shared JSON fixtures (`internal/testdata/contract/*.json`) impossible using TS types alone.

To ensure bidirectional consistency between frontend expectations and future Go DTO unmarshaling, a formal runtime schema is required. The schema must serve as the single source of truth for all domain models, allow strict validation of fixtures by rejecting unknown fields, and produce zero bundle overhead in production builds.

## Decisions

### 1. Zod as Schema Definition Tool
- **Package**: `zod@4.6.5` installed as a direct development dependency in `frontend/package.json`.
- **Location**: Defined strictly in `frontend/src/lib/bridge/schema.ts`.
- **API Pattern**:
  - Schemas defined using Zod v4 syntax: `export const AppInfoSchema = z.object({ ... }).strict();`
  - TypeScript types inferred directly from schemas: `export type AppInfo = z.infer<typeof AppInfoSchema>;`
  - No manual duplicate TypeScript interfaces.

### 2. Strict Parsing Enforcement
- All contract schemas enforce `.strict()`.
- Fixture test suites in Vitest parse JSON fixtures with `Schema.parse(fixture)`. Any extraneous, misspelled, or missing fields fail the test suite immediately.

### 3. Production Dead-Code Elimination
- Zod is treated as a dev/build-time and testing dependency.
- In production builds (`VITE_BRIDGE=wails`), `mock.ts` and `schema.ts` are eliminated via Rollup/Vite tree-shaking.
- Verified via `npm run check:no-mock` in CI to ensure zero Zod or mock artifacts are present in `frontend/dist/`.

### 4. Contract Freeze Governance
- Stage T1 establishes contract version `0.1.0` (DRAFT).
- During Stages T2 and T3 (frontend development), additive field additions are permitted via approved proposals with minor version bumps documented in `docs/contract.md`.
- The contract strictly **freezes** as version `1.0.0` at the conclusion of Stage T3 before Stage T4 (Backend Go) commences.
- Post-freeze changes require synchronized updates to TypeScript schemas, Go struct DTOs, and shared JSON fixtures in a single atomic task.

## Consequences
- Single source of truth prevents drift between TS type declarations and test fixtures.
- Strict validation guarantees all fixtures accurately represent contract expectations.
- Zero production bundle bloat.
