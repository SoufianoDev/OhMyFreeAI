# OhMyFreeAI UI Foundation

**Status:** Initial design baseline  
**Version:** 0.1  
**Scope:** Desktop GUI for Linux and Windows  
**Design system:** Material Design 3 semantic tokens and adaptive layout patterns

This document is the implementation contract for the first GUI iteration. It deliberately defines a restrained, information-dense desktop product rather than a marketing dashboard.

## 1. Product character

OhMyFreeAI is an operational desktop tool. The interface must communicate system state, resource constraints, and available actions without visual noise.

The visual character is:
- Calm, precise, technical, and trustworthy.
- Neutral surfaces with one restrained brand accent.
- Clear typography and compact but readable information density.
- Strong hierarchy through spacing, alignment, and surface roles—not piles of cards, gradients, glow, or decorative borders.
- Consistent with native desktop expectations while retaining Material 3 semantics.

Do not imitate a mobile app enlarged to desktop. Do not use a generic AI-dashboard template.

## 2. Design-system rules

### 2.1 Semantic color, not arbitrary colors

Use Material 3 semantic color roles through CSS custom properties. Components must consume roles, never hard-code their own palette.

Required roles:
- `primary`, `on-primary`, `primary-container`, `on-primary-container`
- `secondary`, `on-secondary`, `secondary-container`, `on-secondary-container`
- `tertiary`, `on-tertiary`, `tertiary-container`, `on-tertiary-container`
- `error`, `on-error`, `error-container`, `on-error-container`
- `surface`, `on-surface`, `on-surface-variant`
- `surface-container-lowest`, `surface-container-low`, `surface-container`, `surface-container-high`, `surface-container-highest`
- `surface-dim`, `surface-bright`, `outline`, `outline-variant`
- `inverse-surface`, `inverse-on-surface`, `inverse-primary`

Theme direction:
- Start with a low-chroma, muted blue-grey brand seed (provisional seed: `#526B7A`).
- Generate light and dark schemes using Material Color Utilities or a verified Material Theme Builder export.
- Default to the operating-system theme preference; allow an explicit Light / Dark / System setting.
- Keep large work surfaces neutral. Use primary for selected navigation, primary actions, focus, and small emphasis—not as a page background.
- Reserve error colors for actual errors. Warnings and degraded runtime states need a distinct semantic warning treatment with accessible contrast.
- Never use raw color literals inside feature components. Add or change a token centrally.

Do not use saturated neon blue/purple, glowing outlines, gradient backgrounds, or pure-black panels as the default visual language.

### 2.2 Typography

Use the Material 3 type scale and its semantic roles:
- Display: only for rare, high-value numeric summaries.
- Headline: page and major section headings.
- Title: pane, dialog, and card/list group headings.
- Body: primary content and supporting descriptions.
- Label: controls, metadata, navigation labels, and compact status text.

Use Roboto or a metrically compatible system fallback. Keep body text comfortably readable; do not use tiny grey text to squeeze in more metrics. Use tabular numerals for resource values, durations, versions, and rates.

### 2.3 Shape, spacing, and elevation

Use a consistent 4/8-based spacing scale. Prefer 8, 16, 24, and 32px for layout rhythm; use 4px only for fine alignment.

Shape roles:
- 8px: fields, menus, chips, compact controls.
- 12px: standard grouped surfaces.
- 16px: prominent panels and navigation containers.
- 28px: dialogs and large overlays.
- Full pill: chips and compact status indicators only where appropriate.

Use surface tonal hierarchy before shadows. Shadows should be subtle and reserved for floating menus/dialogs. Avoid placing every section inside a separate card.

### 2.4 Motion

Motion is functional:
- 120–180ms for hover, focus, and small state changes.
- 180–240ms for pane and navigation transitions.
- Respect reduced-motion preferences.
- No perpetual animations, pulsing status indicators, or decorative movement.
- Progress must represent real work and never animate as fake activity.

## 3. Desktop application shell

### 3.1 Primary navigation

Use a persistent left navigation rail that can expand into a labeled drawer:
- Overview
- Models
- Runtime
- Engines
- Gateway
- Clients
- Downloads
- Recovery
- Activity / Logs
- Settings

Group related destinations with spacing and section labels. Show one clear selected state using a tonal container and an icon plus text when expanded. Avoid nested navigation deeper than two levels in the MVP.

At narrow window widths, collapse labels before hiding navigation. At compact widths, use a drawer rather than allowing the content pane to become unusably narrow.

### 3.2 Top app bar

The top bar contains:
- Current page title and optional breadcrumb.
- Global search or command entry when implemented.
- Runtime connection summary with explicit state.
- Contextual actions for the current page.

Do not duplicate the entire left navigation in the top bar.

### 3.3 Main content

Use a responsive content grid with a readable maximum width. Overview uses a dashboard/feed structure; model and engine browsing use list-detail; settings use a section navigation plus detail pane where width permits.

Desktop window targets:
- Compact: below 600 CSS px.
- Medium: 600–839px.
- Expanded: 840–1199px.
- Large: 1200–1599px.
- Extra-large: 1600px and above.

These are layout transitions, not device labels. Test at intermediate widths and with OS scaling enabled.

## 4. Information architecture and first screens

### Overview

The first screen answers, at a glance:
1. Is the runtime connected?
2. Which model is active?
3. Is the gateway healthy?
4. What resources are available?
5. Is anything waiting for attention?

Use a single runtime summary region, a compact resource summary, and a recent activity list. Avoid a wall of equal-weight metric cards. A disconnected runtime must be clearly described as disconnected, not shown as a generic red error if it is simply stopped.

### Models

Use a list-detail layout:
- List: model name, family, quantization, size, compatibility state.
- Detail: architecture support, checkpoint/GGUF status, engine compatibility, current Colab resource assessment, minimum/recommended requirements, recommended configuration, verification history.
- Filters must not hide unsupported or unknown models by default.
- Distinguish Unsupported, Resource Insufficient, Backend Incompatible, Runtime Restricted, Unknown, and Verified. Never collapse them into one red badge.

### Runtime

Show provider, session state, runtime age when available, detected GPU, VRAM, RAM, storage, CUDA, and last successful probe. Actions must be explicit and reflect actual backend capability.

### Engines and Gateway

Show installed engine build/version, backend capabilities, model loaded state, gateway bind address/port, health, and connected client status. Distinguish process-running from ready-to-serve.

### Activity / Logs

Use a readable event timeline and a dedicated log viewer with filtering, copy, and follow controls. Do not render raw logs as decorative terminal blocks throughout the application.

## 5. Component behavior

Every interactive component must define:
- Default, hover, focus-visible, pressed, disabled, loading, success, warning, and error states as applicable.
- Keyboard operation and a visible focus indicator.
- Accessible name and semantic HTML role.
- Empty state and failure state.
- Loading behavior that does not shift the entire layout unnecessarily.

Prefer native HTML semantics and accessible custom components. Material Web is maintenance-only; do not make `@material/web` a core dependency. Implement the required subset with project-owned HTML/CSS/JavaScript components and Material 3 tokens.

## 6. Accessibility and internationalization

- Meet WCAG 2.2 AA contrast for text and interactive controls.
- Small text target: 4.5:1 contrast; large text: 3:1.
- Never communicate status by color alone; include text and/or icon labels.
- Support full keyboard navigation, logical focus order, visible focus, and Escape behavior for dialogs/menus.
- Respect system font scaling and reduced motion.
- Use semantic landmarks and ordered heading levels.
- Design layouts for RTL from the start. Do not hard-code left/right assumptions; use logical CSS properties such as `margin-inline-start`.
- Ensure numbers, paths, code, and technical identifiers remain readable in RTL contexts.

## 7. Product truth and state language

UI state must come from backend state, not optimistic presentation.

Use explicit terms:
- **Stopped**: no managed runtime process is active.
- **Starting**: startup has begun but readiness checks have not passed.
- **Connected**: the remote session is reachable.
- **Ready**: the model has loaded and a real inference health check has passed.
- **Degraded**: service is reachable but one or more declared capabilities/checks are failing.
- **Unknown**: the application lacks sufficient evidence.

A successful HTTP response alone is not enough to label the model Ready. Do not display fabricated GPU metrics, throughput, context length, progress, or activity.

## 8. Visual QA acceptance checklist

A screen is not accepted until:
- It uses semantic tokens only; no scattered hard-coded feature colors.
- Light and dark themes both remain legible and balanced.
- Layout works at 600, 840, 1200, and 1600px widths and at common OS scaling settings.
- Navigation, dialogs, forms, tables, and menus are keyboard-operable.
- Text and controls meet contrast requirements.
- Empty, loading, error, stopped, degraded, and ready states are designed.
- Resource numbers use tabular numerals and meaningful units.
- The screen contains no decorative metric cards, gratuitous gradients, glowing borders, or unexplained status colors.
- The visual hierarchy makes the primary task obvious within a few seconds.

## 9. Source references

- Project Material 3 skill: https://github.com/SoufianoDev/material-3-skill/tree/master/skills/material-3
- Material 3 layout: https://m3.material.io/foundations/layout/layout-overview
- Material 3 color roles: https://m3.material.io/styles/color/roles
- Material 3 typography: https://m3.material.io/styles/typography/applying-type
- Material 3 accessibility structure: https://m3.material.io/foundations/designing/structure
- KotlinComposeKit repository: https://github.com/SoufianoDev/KotlinComposeKit

Implementation note: KotlinComposeKit targets Kotlin/Jetpack Compose. OhMyFreeAI's approved UI stack is Wails with HTML/CSS/JavaScript, so Compose-specific APIs must not be copied into this frontend. Reuse only transferable engineering practices where relevant.
