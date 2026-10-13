# UI Components & Aceternity Registry

This document records all functional UI components used in Jerion and audits any Aceternity UI adoptions per PRD §5.4.

## Allowed Aceternity Components
Only functional components may be copied into `frontend/src/components/aceternity/` and adapted to tokens:
- **Sidebar** (collapsible left navigation)
- **Tabs** (view switching)
- **Timeline** (scan steps / Popup Hunter sequence)
- **Tooltip / Hover Card** (evidence previews)
- **Calm Progress / Loader** (deterministic scan progress)

## Prohibited Components (Anti-AI-Slop Gate PRD §5.6)
The following decorative components are strictly prohibited:
- Background Beams, Meteors, Sparkles, Aurora, Lamp, Glowing Stars, Decorative Spotlights, Typewriter / Text Generate animations for functional data.

## Component Registry Table

### 1. Aceternity UI (Adapted)
| Component Name | Source / Origin | Target Location | Functional Purpose | Token Adaptation Status |
|---|---|---|---|---|
| Sidebar | Aceternity UI (adapted) | `frontend/src/components/aceternity/Sidebar.tsx` | Collapsible left navigation (220px to 64px) for 8 application modules | Adapted to design tokens, pure CSS transition (200ms ease-out), zero framer-motion dependency |

### 2. Core Design System Components (Radix UI & Tokens)
| Component Name | Base Primitive | Target Location | Functional Purpose | Token Adaptation Status |
|---|---|---|---|---|
| Button | Radix Slot | `frontend/src/components/ui/Button.tsx` | Standard interactive button with primary, secondary, ghost, destructive, outline variants | Full CSS token mapping, focus-visible rings |
| Card | Semantic HTML | `frontend/src/components/ui/Card.tsx` | Structured containers (`CardHeader`, `CardTitle`, `CardDescription`, `CardContent`, `CardFooter`) | Border, radius, and card background tokens |
| Badge | Semantic HTML | `frontend/src/components/ui/Badge.tsx` | Multi-signal status and severity badges (Icon + Label + Color) | Critical, High, Medium, Low, Info, Success, Warning tokens |
| Skeleton | Semantic HTML | `frontend/src/components/ui/Skeleton.tsx` | Calm shimmer loading placeholder | Muted surface tokens with CSS pulse animation |
| EmptyState | Semantic HTML | `frontend/src/components/ui/EmptyState.tsx` | Informative empty screen state with icon, title, description, and primary CTA | Full i18n copy and muted foreground tokens |
| ErrorState | Semantic HTML | `frontend/src/components/ui/ErrorState.tsx` | Error screen state with diagnosis, explanation, and retry CTA | Destructive alert tokens and retry button |
| Table | Semantic HTML | `frontend/src/components/ui/Table.tsx` | Accessible table wrapper with striped hover and header tokens | Compatible with TanStack Table and Virtual |
| Dialog | Radix Dialog | `frontend/src/components/ui/Dialog.tsx` | Modal dialogs for confirmations, remediation, and details | Backdrop blur tokens, keyboard trap, and Esc dismiss |
| Tooltip | Radix Tooltip | `frontend/src/components/ui/Tooltip.tsx` | Informative hover cards for truncated text and evidence | Radix popover positioning, calm 150ms fade |
| Tabs | Radix Tabs | `frontend/src/components/ui/Tabs.tsx` | Module sub-view navigation and filtering | Keyboard arrow navigation, muted active indicator |

### 3. Three.js & Graphic Components
| Component Name | Technology | Target Location | Functional Purpose | Fallback / Guardrail |
|---|---|---|---|---|
| DeviceScene | Three.js + R3F | `frontend/src/three/DeviceScene.tsx` | Procedural 3D phone model on Home view showing scan ring status | Lazy-loaded chunk; paused on window blur/tab hidden; DPR bounded `[1, 2]` |
| StaticFallback | Pure SVG | `frontend/src/three/StaticFallback.tsx` | Static SVG concentric shield graphic for Home view | Automatically rendered when WebGL is unavailable or `prefers-reduced-motion: reduce` is active |
