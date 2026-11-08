# UI/UX Architecture

**Status:** Active implementation contract for the desktop frontend  
**Platforms:** Linux and Windows  
**Frontend:** Wails + semantic HTML + CSS + JavaScript ES modules  
**Design system:** Material Design 3 semantic tokens and adaptive layout patterns

## 1. Product character

OhMyFreeAI is an operational desktop utility, not a marketing site or a generic AI dashboard. The interface must communicate system state, resource constraints, and available actions without visual noise.

The visual character is calm, precise, technical, and trustworthy:
- Neutral, low-chroma surfaces with one restrained brand accent.
- Clear typography and compact but readable information density.
- Hierarchy through spacing, alignment, and tonal surface roles—not piles of cards, gradients, glow, glassmorphism, or decorative borders.
- Native desktop expectations with Material 3 semantics.
- Never imitate a mobile app enlarged to desktop.

## 2. Material 3 design-system contract

### 2.1 Semantic color roles

Use Material 3 semantic roles through centrally defined CSS custom properties. Feature components consume semantic tokens and must not invent local palettes.

Required role coverage:
- Primary: `primary`, `on-primary`, `primary-container`, `on-primary-container`.
- Secondary: `secondary`, `on-secondary`, `secondary-container`, `on-secondary-container`.
- Tertiary: `tertiary`, `on-tertiary`, `tertiary-container`, `on-tertiary-container`.
- Error: `error`, `on-error`, `error-container`, `on-error-container`.
- Surface: `surface`, `on-surface`, `on-surface-variant`, `surface-dim`, `surface-bright`, `surface-container-lowest`, `surface-container-low`, `surface-container`, `surface-container-high`, `surface-container-highest`.
- Outline and inverse: `outline`, `outline-variant`, `inverse-surface`, `inverse-on-surface`, `inverse-primary`.

Theme rules:
- Use a low-chroma, muted blue-grey seed as a provisional direction, not a hard-coded final brand decision.
- Generate or verify light and dark schemes with Material Color Utilities or a verified Material Theme Builder export.
- Support Light, Dark, and System preference. System is the default unless product settings specify otherwise.
- Keep work surfaces neutral. Use primary for selected navigation, primary actions, focus, and limited emphasis—not large page backgrounds.
- Reserve error roles for actual errors. Warnings and degraded states need distinct semantic treatment and accessible contrast.
- Do not use saturated neon blue/purple, glowing outlines, gradient backgrounds, or pure-black panels as the default visual language.
- Never put raw feature-specific color literals in component styles; add or adjust a central token instead.

### 2.2 Typography

Use the Material 3 type scale and semantic roles:
- Display: rare, high-value numeric summaries only.
- Headline: page and major section headings.
- Title: pane, dialog, and grouped-list headings.
- Body: primary content and supporting descriptions.
- Label: controls, metadata, navigation, and compact status text.

Use Roboto or a metrically compatible system fallback. Keep body text comfortably readable; do not shrink text to fit excess metrics. Use tabular numerals for resource values, durations, versions, and rates.

### 2.3 Shape, spacing, and elevation

Use a consistent 4dp spacing foundation and an 8dp layout rhythm. Prefer 8, 16, 24, and 32px for major spacing; use 4px for fine alignment.

Shape guidance:
- 8px: fields, menus, chips, and compact controls.
- 12px: standard grouped surfaces.
- 16px: prominent panels and navigation containers.
- 28px: dialogs and large overlays.
- Full-pill shapes: compact chips and status indicators only where appropriate.

Prefer tonal surface hierarchy over shadows. Shadows are subtle and reserved for floating menus and dialogs. Do not put every section inside a separate card.

### 2.4 Motion

Motion communicates a state change or spatial relationship:
- 120–180ms for hover, focus, and small state changes.
- 180–240ms for pane and navigation transitions.
- Respect `prefers-reduced-motion`.
- No perpetual animation, pulsing status indicators, or decorative movement.
- Progress indicators must represent real work; never animate fabricated activity.

## 3. Desktop shell and responsive layout

### 3.1 Navigation

Use a persistent desktop navigation rail/sidebar that can expand into a labeled drawer. Primary destinations:
- Overview
- Models and Compatibility
- Runtime
- Engines
- Gateway
- Clients
- Downloads
- Monitoring and Activity
- Persistent Storage and Recovery
- Terminal and Logs
- Settings

Group related destinations with spacing and section labels. Show one selected state using a tonal container. Avoid navigation deeper than two levels in the MVP. Do not expose pages that have no working functionality.

At narrower widths, collapse labels before hiding navigation. At compact widths, use a drawer rather than allowing the content pane to become unusably narrow.

### 3.2 Top app bar

The top bar contains the current page title and optional breadcrumb, runtime connection summary with an explicit state, and contextual actions for the current page. Add global search or command entry only when implemented. Do not duplicate the entire navigation in the top bar.

### 3.3 Content layout and breakpoints

Constrain content to a readable maximum width. Use two-pane/list-detail patterns where they improve browsing and inspection. Use dense tables for model catalogs, process data, and telemetry instead of one card per datum.

Use these layout transition ranges as initial test targets:
- Compact: below 600 CSS px.
- Medium: 600–839px.
- Expanded: 840–1199px.
- Large: 1200–1599px.
- Extra-large: 1600px and above.

These are layout transitions, not device labels. Test intermediate widths and common operating-system scaling settings.

## 4. Information architecture and screen contracts

Only add navigation entries for screens with implemented functionality. The intended product areas are:

### Overview

Answer at a glance:
1. Is the runtime connected?
2. Which model is active, if any?
3. Is the gateway healthy?
4. What resources are actually available?
5. Is anything waiting for attention?

Use a concise runtime summary, compact resource summary, and recent activity list. Avoid a wall of equal-weight metric cards. A stopped runtime is not automatically an error.

### Models and Compatibility

Use list-detail:
- List: model name, family, quantization, artifact size, compatibility state.
- Detail: architecture support, checkpoint/GGUF status, engine/backend compatibility, runtime resource assessment, minimum and recommended requirements, suggested configuration, and verification history.
- Filters must not silently hide unsupported or unknown models.
- Distinguish Unsupported, Resource Insufficient, Backend Incompatible, Runtime Restricted, Unknown, and Verified. Do not collapse these into one generic red status.

### Runtime

Show provider, session state, runtime age when known, detected GPU, VRAM, RAM, storage, CUDA/backend information, and last successful probe. Separate provider policy eligibility from technical connectivity. Actions must reflect actual backend capability.

### Engines and Gateway

Show installed engine build/version, backend capabilities, model-loaded state, gateway bind address/port, health, and connected client status. Distinguish process-running from ready-to-serve.

### Clients

Show configured client integrations and their actual connection/configuration status. Do not claim an integration is active merely because its configuration file exists.

### Downloads and Persistent Storage

Show artifact identity, version, size, source, integrity status, destination, and lifecycle state where available. Destructive or replacement actions require clear confirmation and truthful progress.

### Monitoring, Activity, Terminal, and Logs

Use a readable event timeline and a dedicated log viewer with filtering, copy, and follow controls. Keep diagnostics readable; do not scatter raw terminal blocks throughout the interface. Terminal capabilities must be backed by the native application service.

### Recovery and Settings

Recovery must expose inspect/preview and execution states separately; never imply a restore occurred when only inspection passed. Settings should group appearance, language, storage, and application preferences without exposing unsupported controls.

## 5. Component interaction contract

Every interactive component defines applicable default, hover, focus-visible, pressed, disabled, loading, success, warning, and error states.

Every control must have:
- Keyboard operation and visible focus.
- An accessible name and correct semantic HTML role.
- A clear empty state and failure state.
- Loading behavior that avoids unnecessary layout shifts.
- A truthful connection between displayed state and backend state.

Prefer native HTML semantics and accessible project-owned components. Material Web is in maintenance mode; do not make `@material/web` a core dependency or claim full Compose Expressive parity. The implementation stack remains Wails with HTML/CSS/JavaScript. Kotlin/Jetpack Compose APIs from KotlinComposeKit are not to be copied into this frontend; only transferable design and engineering practices may be reused.

## 6. Accessibility and internationalization

- Meet WCAG 2.2 AA contrast for text and interactive controls.
- Minimum contrast targets: 4.5:1 for normal text and 3:1 for large text and relevant UI boundaries.
- Never communicate status by color alone; include text and/or a meaningful icon label.
- Support full keyboard navigation, logical focus order, visible focus, and Escape behavior for dialogs and menus.
- Respect system font scaling and reduced motion.
- Use semantic landmarks and ordered heading levels.
- Support Arabic and English with true RTL/LTR layout. Use CSS logical properties rather than hard-coded left/right assumptions.
- Do not mirror icons or technical charts blindly under RTL.
- Keep numbers, paths, code, and technical identifiers readable in RTL contexts.

## 7. Product truth and status language

UI state must come from backend evidence, not optimistic presentation.

Use explicit states and timestamps where applicable:
- **Stopped:** no managed runtime process is active.
- **Starting:** startup began, but readiness checks have not passed.
- **Connected:** the runtime session is reachable.
- **Ready:** the model is loaded and a real inference health check passed.
- **Degraded:** the service is reachable, but one or more declared checks or capabilities are failing.
- **Live / Stale / Unavailable / Error / Recovering / Paused:** use only when the corresponding data source and lifecycle state support the label.
- **Unknown:** the application lacks sufficient evidence.

A successful HTTP response alone is not enough to label a model Ready. Never fabricate GPU metrics, throughput, context length, progress, activity, or timestamps.

## 8. Visual QA acceptance checklist

A screen is not accepted until:
- It uses semantic tokens; no scattered hard-coded feature colors.
- Light, dark, and system themes remain legible and balanced.
- Layout works at 600, 840, 1200, and 1600px widths and common OS scaling settings.
- Navigation, dialogs, forms, tables, and menus are keyboard-operable.
- Text and controls meet contrast requirements.
- Empty, loading, error, stopped, degraded, ready, and unknown states are designed where relevant.
- Resource values use tabular numerals and meaningful units.
- No decorative metric-card wall, gratuitous gradient, glowing border, or unexplained status color appears.
- The primary task and next action are obvious within a few seconds.
- RTL and LTR are both tested, including mixed-direction technical identifiers.

## 9. Source references

- Project Material 3 skill: https://github.com/SoufianoDev/material-3-skill/tree/master/skills/material-3
- Material 3 layout: https://m3.material.io/foundations/layout/layout-overview
- Material 3 color roles: https://m3.material.io/styles/color/roles
- Material 3 typography: https://m3.material.io/styles/typography/applying-type
- Material 3 accessibility structure: https://m3.material.io/foundations/designing/structure
- KotlinComposeKit: https://github.com/SoufianoDev/KotlinComposeKit

Implementation note: KotlinComposeKit targets Kotlin/Jetpack Compose. OhMyFreeAI uses Wails with HTML/CSS/JavaScript. Reuse transferable design and engineering practices only; do not copy Compose-specific APIs into this frontend.
