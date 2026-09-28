# Design System Master File

> **LOGIC:** When building a specific page, first check `design-system/pages/[page-name].md`.
> If that file exists, its rules **override** this Master file.
> If not, strictly follow the rules below.

---

**Project:** Primocrest Pharmacy
**Generated:** 2026-09-28 22:34:10
**Category:** Pharmacy POS / Inventory Management

---

## Global Rules

### Color Palette

The current UI uses a neutral slate palette via shadcn-vue tokens. These are the semantic roles — map brand colors through these tokens rather than using raw hex in components.

| Role | Hex | CSS Variable (shadcn) |
|------|-----|----------------------|
| Primary (brand) | `#15803D` | `--color-primary` |
| On Primary | `#FFFFFF` | `--color-on-primary` |
| Secondary | `#22C55E` | `--color-secondary` |
| On Secondary | `#0F172A` | `--color-on-secondary` |
| Accent | `#0369A1` | `--color-accent` |
| On Accent | `#FFFFFF` | `--color-on-accent` |
| Background | `var(--background)` | shadcn `--background` |
| Foreground | `var(--foreground)` | shadcn `--foreground` |
| Card | `var(--card)` | shadcn `--card` |
| Card Foreground | `var(--card-foreground)` | shadcn `--card-foreground` |
| Muted | `var(--muted)` | shadcn `--muted` |
| Muted Foreground | `var(--muted-foreground)` | shadcn `--muted-foreground` |
| Border | `var(--border)` | shadcn `--border` |
| Destructive | `#DC2626` | `--color-destructive` |
| On Destructive | `#FFFFFF` | `--color-on-destructive` |
| Ring | `var(--ring)` | shadcn `--ring` |

**Color Notes:** Pharmacy green + trust blue. Status colors: green (in-stock/complete), amber (low/mid), red (issues/expired). The shadcn neutral slate (`--background`, `--foreground`, `--card`, etc.) handles the chrome; brand green/blue is reserved for primary actions and highlights.

### Typography

- **Font Family:** System UI font stack (Inter, SF Pro, Segoe UI) — no custom webfonts
- **Body Size:** `text-sm` (14px) minimum for UI chrome, `text-xs` (12px) for data-dense tables
- **Mood:** clean, accessible, readable, professional
- **No Google Fonts import needed** — uses system stack for performance

### Spacing

Use Tailwind spacing utilities consistently:
- `gap-1` (4px) — tight inline gaps
- `gap-2` (8px) — icon-label spacing, small component gaps
- `gap-3` (12px) — standard flex gaps
- `gap-4` (16px) — section spacing
- `gap-6` (24px) — large section margins
- `p-4`/`p-6` — card/container padding
- `py-2`/`py-3` — table cell padding
- Consistent 4px/8px rhythm for all spacing

---

## Style Guidelines

**Style:** Flat Design + Minimalism

**Keywords:** Clean lines, no shadows, 2D, icon-heavy, typography-focused, accessible, data-dense

**Best For:** Web apps, dashboards, inventory management, POS systems, corporate tools

**Key Constraints:**
- No gradients or shadows on UI elements
- `border-radius: 0-4px` (flat/slightly rounded)
- `box-shadow: none` (no drop shadows)
- Simple hover states (color/opacity shift, 150-200ms)
- SVG icons (Lucide), consistent 1.5px stroke width
- High-density data tables with sticky headers
- Fast loading, minimal transitions

### Page Pattern (Internal Dashboard)

- **Structure:** Left sidebar nav → main content with sticky header
- **Pattern:** List/Table + Detail Panel (sales, inventory, stock taking)
- **Metric Cards:** Top row showing 3-4 KPIs
- **Filters:** Inline date range, search bar, dropdown presets
- **Detail Sidebar:** Slides in from right, shows item-level data
- **Modals:** Centered overlay with backdrop blur for forms/confirms
- **Keyboard navigation:** Arrow keys, Enter, Escape for list-detail flows

---

## Anti-Patterns (Do NOT Use)

- ❌ Bright neon colors or AI purple/pink gradients
- ❌ Motion-heavy animations or parallax
- ❌ Emojis as icons — use Lucide SVG icons instead
- ❌ Missing `cursor:pointer` on clickable elements
- ❌ Layout-shifting hovers (avoid scale transforms that shift layout)
- ❌ Low contrast text — maintain 4.5:1 minimum contrast ratio
- ❌ Instant state changes — always use transitions (150-300ms)
- ❌ Invisible focus states — focus rings must be visible for keyboard nav
- ❌ Horizontal scroll on mobile — content must fit viewport width
- ❌ Nested scroll regions that conflict with main page scroll
- ❌ Placeholder-only labels — inputs must have visible labels
- ❌ Emoji for UI icons — use SVG (Lucide) consistently

---

## Pre-Delivery Checklist

Before delivering any UI code, verify:

- [ ] No emojis used as icons (Lucide SVG instead)
- [ ] All icons use consistent stroke width (1.5px)
- [ ] `cursor-pointer` on all clickable elements
- [ ] Hover states with smooth transitions (150-300ms)
- [ ] Text contrast 4.5:1 minimum in light mode
- [ ] Dark mode tested separately (not inferred from light)
- [ ] Focus states visible for keyboard navigation
- [ ] Touch targets ≥44px on inputs, buttons, selects
- [ ] `prefers-reduced-motion` respected (no excessive animation)
- [ ] Responsive: 375px, 768px, 1024px, 1440px
- [ ] Content not hidden behind fixed navbars
- [ ] No horizontal scroll on mobile
- [ ] Keyboard navigation works (Tab, Arrow keys, Enter, Escape)
