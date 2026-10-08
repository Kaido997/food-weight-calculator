---
version: alpha
name: Food Weight
description: A kitchen companion with a photographic introduction and a precise, quiet calculator.
colors:
  primary: "#176149"
  background: "#f6f8f5"
  surface: "#ffffff"
  text: "#193c32"
  muted: "#596d62"
  border: "#d4ded7"
  result: "#e9f0e5"
  citrus: "#deed9c"
  danger: "#a02e33"
  dark-background: "#14231f"
  dark-surface: "#1c3028"
  dark-text: "#edf3e9"
  dark-primary: "#c8dfa2"
typography:
  display:
    fontFamily: '"Libre Baskerville", Georgia, serif'
  body:
    fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif'
rounded:
  control: "12px"
  segment: "9px"
spacing:
  field-gap: "14px"
  panel-padding: "20px"
components:
  button:
    backgroundColor: "{colors.primary}"
    textColor: "{colors.surface}"
  result:
    backgroundColor: "{colors.result}"
    textColor: "{colors.text}"
---

# Food Weight design system

## Overview

A useful object on a kitchen counter: warm food photography next to a legible scale-like readout. The brief asks for a half-width landing panel and half-width calculator, then a compact image header on phones. Audience: people portioning everyday meals. English and Italian are supported; no specific national market is assumed.

This is a hybrid screen. The introduction carries the brand; the calculator stays familiar. The signature is the large “From raw to ready” headline over the original meal-prep image. Avoid dashboard tiles, invented statistics, decorative charts, and excessive animation.

Business evidence: README.md shared food API; internal/food/service.go units and calculation invariants; current user brief. There are no account, billing, destructive, or regulated workflows in this screen. Admin analytics remains outside this redesign.

## Colors

Runtime CSS custom properties in `web/assets/app.css` are canonical. This file mirrors the accepted palette and intent; no token generator or build pipeline is needed. `colors.primary` → `--accent`, background → `--bg`, surface → `--surface`, text → `--text`, muted → `--muted`, border → `--border`, result → `--result`, danger → `--danger`, citrus → `--citrus`. Dark-prefixed values map to the corresponding variables under `[data-theme="dark"]`. Component recipes consume those variables.

Control outlines use `--control-border` (#85958b light, #6e8978 dark) for a minimum 3:1 contrast against their surface.

Evergreen actions recall the existing site's green palette; mist surfaces keep the tool legible. Citrus is reserved for photo-panel accents and dark-theme actions. Errors use text and ARIA, never just red. Light, dark, and system preferences persist locally; system mode follows OS changes. Forced colors preserves system control outlines and scrollbars.

## Typography

Headings use Libre Baskerville Regular (400), self-hosted as a preloaded Latin WOFF2 subset with Georgia fallback. Serif tracking is relaxed to -0.025em with 1.12/1.2 heading line heights. Body text and numeric results retain the system UI stack, tabular figures, and locale-aware formatting. The font is included in the PWA cache; no third-party runtime font requests. A local search field filters native option menus by case/accent-insensitive words anywhere in the name. Exact matches, whole phrases, and shorter matches rank first; all query words must match. Full food names remain in native option menus; results wrap long names.

## Layout

Desktop above 800px: equal columns, sticky photographic introduction, document-scrolling calculator. Icon theme radios and compact spacing prioritize the result above the fold. The form is capped at 460px (500px on wide screens). Below 800px: introduction first, typically less than a third of the viewport; mobile viewports below 760px tall omit secondary intro and calculator copy. Below 650px, the hero wordmark and visible field hints are omitted to give the answer space. Zoom, long errors, and open help remain naturally scrollable rather than clipped. Text zoom may grow content rather than clipping it. Safe-area bottom padding supports installed apps.

Both desktop panels share `--panel-gutter` (32–72px). Heading-to-form spacing is 20px and fields use 14px gaps; mobile retains 12px/10px gaps and smaller result padding to protect the answer area.

Reserve result and status space. Do not constrain the calculator to a fixed viewport height. At narrow widths, theme and language controls remain reachable. The page has one normal document scroller.

## Elevation & Depth

Use borders and tinted surfaces rather than floating cards. Only selected segmented controls have a small shadow. The image has an evergreen gradient to maintain text contrast, independent of the selected theme.

## Shapes

Controls and result panels share `--radius` (12px); segmented radio groups use 9px outer and 6px inner corners. Icons use two-pixel strokes and simple geometry. No decorative pills or emoji flags.

## Components

| Capability | Canonical owner | Source of truth | Allowed variants | Verification |
|---|---|---|---|---|
| Select/Listbox | `.control` and native select | This file | Local search plus native OS popup for food; native language select; platform geometry is accepted | Go rendering tests; browser check when connected |
| Form | `calculator.js`, shared field/control CSS | Service contract and this file | Calculator-specific validation with inline errors | Node interaction tests and Go handler tests |
| Scrollbar | Global `app.css` rules | Runtime theme tokens | Standards plus WebKit fallback; forced-colors auto | Static audit; browser check when connected |
| Feedback | Inline form error and result live region | This file | Error, pending, empty, success | Node interaction tests |
| Theme and install | `theme.js`, `shell.js`, `shell.html` | This file | Light/dark/system, native install or instructions | Node tests and manifest/asset tests |

Shared head, brand, preferences, typography, buttons, fields, segmented radios, and result recipes are reusable for future calculators. Calculation behavior stays in its own script; the shared shell owns no food arithmetic. No framework or duplicated calculation logic. `food-search.js` filters and ranks the existing option snapshot without server calls. Search is transient, not URL state. It preserves a selected food while it still matches, otherwise selects the highest-ranked match; an empty result leaves no valid food ID. Clearing restores all options and focuses search. The search uses native input/select semantics rather than a custom combobox.

Validation happens on submit, focuses the first invalid field, and uses localized associated errors. Editing clears obsolete results and cancels pending work. Requests time out after 12 seconds; resubmission is explicit. Buttons retain their size while busy. No native alert/confirm or validation bubbles. Result formatting follows page locale. Inputs are kept after request errors and are not stored remotely or in URLs.

The PWA caches only public shell pages and assets. Network-first fetch keeps food options current online; calculations and admin/API responses are never cached. Offline openings clearly explain that calculations need a connection. Install instructions cover browsers without an install event. New worker versions activate after existing tabs close; bump the shell cache version when assets change.

Motion: only short color transitions and a busy spinner, disabled under reduced motion. Focus rings and native keyboard interaction take priority over custom popup styling. Copy is conversational and specific; results are estimates, not promises of exact cooking yield.

## Do's and Don'ts

- Do reuse shell templates, semantic tokens, and native controls in later calculators.
- Do keep result context tied to the exact submitted food, weight, and unit.
- Don't cache arithmetic results or silently calculate with stale food factors offline.
- Don't add decorative metrics, signup steps, or external frontend dependencies.
