# UI Design Reference - Armazenda

> AI Agent Guide for Creating Consistent GUI Components

## Quick Reference

### Design System
- **Style**: Glass morphism with backdrop blur
- **Background**: Sky-to-green gradient (`#0ea5e9` → `#64b5dc` → `#30b60b`) or `inside_bg_3.jpg`
- **Primary Color**: Emerald green (`#22c55e`, `#16a34a`)
- **Active State**: Blue (`bg-blue-500/70`)
- **Typography**: Noto Sans (body), Noto Serif (headings)

### Essential Imports (Head Section)
```html
<link rel="preconnect" href="https://fonts.googleapis.com">
<link href="https://fonts.googleapis.com/css2?family=Noto+Sans:ital,wght@0,100..900;1,100..900&family=Noto+Serif:ital,wght@0,100..900;1,100..900&display=swap" rel="stylesheet">
<link href="/public/assets/static/css/output.css" rel="stylesheet">
<script src="/public/assets/static/htmx.min.js.gz" defer></script>
<script src="https://cdn.jsdelivr.net/npm/iconify-icon@2.1.0/dist/iconify-icon.min.js" defer></script>
```

### Page Skeleton
```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <!-- Imports above -->
  <title>Page Name - Armazenda</title>
</head>
<body class="main-bg">
  {{ template "navbar" }}
  {{ template "menu" "page-id" }}
  <section class="mt-14 lg:mt-20 mb-4 lg:pl-12 px-2 max-w-screen-xl mx-auto">
    <!-- Content -->
  </section>
</body>
</html>
```

---

## Layout Components

### Fixed Navbar (Top-Right)
```html
{{ template "navbar" }}
```
- Fixed position, glass panel style
- Shows username from cookie
- Always include as first element in body

### Side Menu (Responsive)
```html
{{ template "menu" "page-id" }}
```
- Parameter matches route name (e.g., "romaneio", "pessoa")
- Auto-highlights active page with emerald border
- Hamburger menu on mobile (< lg breakpoint)
- Fixed left, full height on desktop

---

## Glass Panel System

### Standard Container
```html
<div class="glass-panel rounded-xl shadow-center p-4">
  <!-- Content -->
</div>
```

### Section Spacing
- Section wrapper: `mt-14 lg:mt-20 mb-4 lg:pl-12 px-2 max-w-screen-xl mx-auto`
- Header: `text-center mb-8`
- Between panels: `mb-8` or `mt-2`

---

## Typography Patterns

### Page Title
```html
<h1 class="text-xl lg:text-3xl font-bold text-white font-[Noto_Sans] drop-shadow-lg">
  Page Title
</h1>
<p class="text-sm text-white/90 mt-2">
  Description text
</p>
```

### Section Headers
```html
<h2 class="text-xl lg:text-2xl font-bold text-white/95 text-center">
  Section Title
</h2>
```

### Text Colors
- Primary: `text-white` or `text-white/95`
- Secondary: `text-white/90`, `text-white/80`, `text-white/70`
- Muted: `text-white/60`

---

## Button Patterns

### Primary Action Button
```html
<button class="primary-glass-btn w-36 flex items-center gap-1 whitespace-nowrap">
  <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24">
    <path fill="currentColor" d="..."/>
  </svg>
  Button Text
</button>
```

### Secondary/Glass Button
```html
<button class="glass-btn flex items-center gap-1">
  <!-- Icon -->
  Text
</button>
```

### Toggle Button Group
```html
<div class="glass-panel rounded-2xl p-1 flex" id="type-toggler">
  <button class="flex-1 py-2 px-4 rounded-2xl flex items-center justify-center gap-2 bg-blue-500/70 text-white shadow-sm">
    <!-- Active state icon -->
    <span class="font-semibold">Option 1</span>
  </button>
  <button class="flex-1 py-2 px-4 rounded-2xl flex items-center justify-center gap-2 text-white/70 hover:bg-sky-50/10">
    <!-- Inactive state icon -->
    <span class="font-semibold">Option 2</span>
  </button>
</div>
```

### Button with HTMX
```html
<button hx-get="/endpoint" hx-target="#content-container" hx-swap="innerHTML">
  Load Content
</button>
```

---

## HTMX Patterns

### Load Content into Container
```html
<div id="content-container">
  {{ template "initial-content" . }}
</div>

<button hx-get="/api/list" hx-target="#content-container" hx-swap="innerHTML">
  Load List
</button>
```

### Modal/Dialog Pattern
```html
<button hx-get="/form" hx-target="body" hx-swap="beforeend">
  Open Form
</button>
```
Returns: `<dialog>` element that auto-opens with `showModal()`

### Dynamic List Loading
```html
<main id="list-container" hx-get="/api/items" hx-trigger="load" hx-swap="innerHTML">
  <div class="flex justify-center items-center p-4">
    <img src="/public/assets/static/spinner.svg" class="animate-spin h-8 w-8" alt="Loading...">
  </div>
</main>
```

---

## Template Syntax

### Include Component
```html
{{ template "component-name" }}
{{ template "component-with-data" .Data }}
```

### Block Definition (Reusable)
```html
{{ block "block-name" . }}
  <!-- Default content -->
{{ end }}
```

### Conditional Rendering
```html
{{ if .Condition }}
  <!-- Show if true -->
{{ else }}
  <!-- Show if false -->
{{ end }}
```

### Loop Through Data
```html
{{ range .Items }}
  {{ template "item-template" . }}
{{ end }}
```

---

## CSS Utility Classes

### Custom Utilities (from input.css)

| Class | Purpose |
|-------|---------|
| `glass-panel` | Main glass morphism container |
| `glass-btn` | Transparent button with border |
| `primary-glass-btn` | Emerald primary action button |
| `add-btn` | Green bordered add button |
| `icon-btn` | Icon-only button with hover scale |
| `cancel-btn` | Orange cancel button |
| `label-peer` | Floating label for peer inputs |
| `required-star` | Adds red asterisk for required fields |
| `filter-card` | Filter panel card style |
| `fieldset-legend` | Styled fieldset legend |

### Common Tailwind Combinations

**Glass Container:**
```
glass-panel rounded-xl shadow-center p-4
```

**Flex Row:**
```
flex justify-between items-center gap-2
```

**Responsive Grid:**
```
grid grid-cols-1 md:grid-cols-2 gap-4
```

**Button Base:**
```
flex items-center gap-1 whitespace-nowrap
```

---

## Table Patterns

### Table Header
```html
<thead class="border-b border-white/10 text-xs text-gray-400 uppercase tracking-wider *:font-medium">
  <tr>
    <th class="py-3 px-4 text-left">Column Header</th>
  </tr>
</thead>
```

**Classes breakdown:**
- `border-b border-white/10` - subtle bottom border
- `text-xs` - small text size
- `text-gray-400` - muted text color (works well on glass-panel)
- `uppercase` - uppercase text
- `tracking-wider` - wider letter spacing
- `*:font-medium` - medium font weight for all child elements

### Table Body
```html
<tbody class="text-white/80">
  <!-- Table rows -->
</tbody>
```

### Table Row
```html
<tr class="border-b border-white/10 hover:bg-white/5">
  <td class="py-3 px-4">Data</td>
</tr>
```

### Complete Table Example
```html
<table class="responsive-table">
  <thead class="border-b border-white/10 text-xs text-gray-400 uppercase tracking-wider *:font-medium">
    <tr>
      <th class="py-3 px-4 text-left">Column 1</th>
      <th class="py-3 px-4 text-left">Column 2</th>
    </tr>
  </thead>
  <tbody class="text-white/80">
    <tr class="border-b border-white/10 hover:bg-white/5">
      <td class="py-3 px-4" data-label="Column 1">Data 1</td>
      <td class="py-3 px-4" data-label="Column 2">Data 2</td>
    </tr>
  </tbody>
</table>
```

### Responsive Table Class

Use `.responsive-table` class for automatic mobile card transformation:

```html
<table class="responsive-table">
  <!-- Table content -->
</table>
```

**Requirements:**
- Add `data-label` attribute to each `<td>` matching the column header:
```html
<td data-label="NOME">John Doe</td>
<!-- Mobile renders: "NOME: John Doe" as a card row -->
```

**How it works on mobile (< 600px):**
1. Hides `<thead>` visually but keeps it accessible
2. Converts `<tr>` to `display: block` (stacks as cards)
3. Converts `<td>` to `display: block` with right-aligned values
4. Uses `::before` pseudo-element to show column label from `data-label`

### Table Cell Padding
All table cells use consistent padding:
- `py-3 px-4` - vertical 0.75rem, horizontal 1rem

### Pagination Controls
```html
<div class="flex justify-center items-center p-4">
  <button class="px-2 lg:px-4 py-1 lg:py-2 bg-sky-50/20 hover:bg-sky-50/30 rounded-lg transition-colors">
    Anterior
  </button>
  <span class="text-sm lg:text-base mx-2 lg:mx-4 text-white/90">
    Página 1 de 10
  </span>
  <button class="px-2 lg:px-4 py-1 lg:py-2 bg-sky-50/20 hover:bg-sky-50/30 rounded-lg transition-colors">
    Próxima
  </button>
</div>
```

**Disabled state:**
```html
<button class="px-2 lg:px-4 py-1 lg:py-2 bg-sky-50/10 text-white/40 rounded-lg cursor-not-allowed" disabled>
  Anterior
</button>
```

**Pagination Classes:**
- `bg-sky-50/20` - semi-transparent sky background for enabled buttons
- `hover:bg-sky-50/30` - slightly lighter on hover
- `bg-sky-50/10` - more transparent for disabled state
- `text-white/40` - muted text for disabled state
- `rounded-lg` - rounded corners
- Responsive padding: `px-2 lg:px-4 py-1 lg:py-2`

---

## Icon Patterns

### SVG Inline Icons
```html
<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24">
  <path fill="currentColor" d="..."/>
</svg>
```

### Iconify Icons
```html
<iconify-icon icon="ic:baseline-add-circle"></iconify-icon>
```

### Common Icons Used
- Add/Plus: `ic:baseline-add-circle`
- Edit: `mdi:pencil`
- Delete: `mdi:delete`
- Search: `mdi:magnify`
- Close: `mdi:close`

---

## File Organization

### Template Directory Structure
```
templates/
├── pages/
│   └── {page-name}.html       # Full page templates
├── layout/
│   ├── navbar.html            # Fixed navbar
│   └── menu.html              # Side navigation
├── components/
│   └── {component-name}.html  # Reusable components
├── {feature}/
│   ├── {feature}-form.html    # Form dialogs
│   ├── {feature}-list.html    # List/table views
│   └── {feature}-item.html    # Individual items
└── icons/
    └── {icon-name}.html       # SVG icons
```

### Naming Conventions
- Pages: `pages/{route-name}.html` (e.g., `romaneio.html`)
- Components: `components/{action}-{type}.html` (e.g., `add-btn-icon.html`)
- Feature blocks: `{feature}/{feature}-{type}.html` (e.g., `entry/entry-form.html`)
- Use kebab-case for all file names

---

## Responsive Breakpoints

- **Mobile first**: Base styles for mobile
- **sm**: 640px
- **md**: 768px  
- **lg**: 1024px (main breakpoint for layout changes)
- **xl**: 1280px

### Common Responsive Patterns
```html
<!-- Typography -->
text-xl lg:text-3xl

<!-- Spacing -->
mt-14 lg:mt-20
px-2 lg:px-4

<!-- Layout -->
grid-cols-1 md:grid-cols-2 lg:grid-cols-3

<!-- Show/Hide -->
hidden lg:flex      <!-- Hide mobile, show desktop -->
lg:hidden          <!-- Show mobile, hide desktop -->
```

---

## Complete Example: New Page

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link href="https://fonts.googleapis.com/css2?family=Noto+Sans:ital,wght@0,100..900;1,100..900&family=Noto+Serif:ital,wght@0,100..900;1,100..900&display=swap" rel="stylesheet">
  <link href="/public/assets/static/css/output.css" rel="stylesheet">
  <title>New Page - Armazenda</title>
  <script src="/public/assets/static/htmx.min.js.gz" defer></script>
  <script src="https://cdn.jsdelivr.net/npm/iconify-icon@2.1.0/dist/iconify-icon.min.js" defer></script>
</head>
<body class="main-bg">
  {{ template "navbar" }}
  {{ template "menu" "new-page" }}
  
  <section class="mt-14 lg:mt-20 mb-4 lg:pl-12 px-2 max-w-screen-xl mx-auto">
    <header class="text-center mb-8">
      <h1 class="text-xl lg:text-3xl font-bold text-white font-[Noto_Sans] drop-shadow-lg">
        Page Title
      </h1>
    </header>
    
    <div class="glass-panel rounded-xl shadow-center p-4">
      <!-- Content here -->
    </div>
  </section>
</body>
</html>
```
