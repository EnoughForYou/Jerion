# UI Components & Aceternity Registry

This document records all functional UI components used in Jerion and audits any Aceternity UI adoptions per PRD §5.4.

## Allowed Aceternity Components
Only functional components may be copied into `frontend/src/components/aceternity/` and adapted to tokens:
- **Sidebar** (collapsible left navigation)
- **Tabs** (view switching)
- **Timeline** (scan steps / Popup Hunter sequence)
- **Tooltip / Hover Card** (evidence previews)
- **Calm Progress / Loader** (deterministic scan progress)

## Prohibited Components (Anti-AI-Slop Gate)
The following decorative components are strictly prohibited:
- Background Beams, Meteors, Sparkles, Aurora, Lamp, Glowing Stars, Decorative Spotlights, Typewriter / Text Generate animations for functional data.

## Component Registry Table
| Component Name | Source / Origin | Target Location | Functional Purpose | Token Adaptation Status |
|---|---|---|---|---|
| Sidebar | Aceternity UI (adapted) | `frontend/src/components/aceternity/Sidebar.tsx` | Collapsible left navigation (220px to 64px) for 8 application modules | Adapted to design tokens, pure CSS transition (200ms ease-out), zero framer-motion dependency |
