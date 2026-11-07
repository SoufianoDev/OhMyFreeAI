# UI/UX Architecture

## Design intent

Professional desktop utility, calm and information-dense without visual noise. Avoid generic AI-dashboard styling, oversized hero sections, excessive cards, gradients, glow, glassmorphism and decorative animation.

## Material 3 use

Use the vendored `.agents/skills/material-3/` skill and references. Apply semantic color roles, tonal surfaces, accessible typography, consistent 4dp spacing, restrained shape/elevation and predictable interaction states. Use CSS custom properties and semantic HTML. Material Web is in maintenance mode; do not make it a hard dependency or claim full Compose Expressive parity.

## Layout

Desktop-first navigation with a persistent navigation rail/sidebar and a focused content area. Constrain content width, use responsive two-pane/list-detail patterns where useful, and keep diagnostics/logs readable. Use dense tables for model catalogs and process/telemetry data rather than one card per datum.

## Accessibility and localization

Keyboard navigation, visible focus, semantic labels, screen-reader announcements, sufficient contrast, reduced-motion support, scalable text, Arabic/English localization and true RTL/LTR layout. Do not mirror icons or technical charts blindly under RTL.

## Main pages

Dashboard, Models, Compatibility Catalog, Runtime, Engines, Monitoring, Terminal, Clients, Downloads, Persistent Storage, Recovery, Settings, Logs/Diagnostics, Legal/Privacy. Add only pages with working functionality; do not ship empty navigation.

## Status design

Use explicit textual labels and timestamps for Live, Stale, Unavailable, Error, Starting, Recovering and Paused. Color is supplemental, never the only status signal.
