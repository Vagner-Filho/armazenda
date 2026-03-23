# CSP Nonce in Templates - Implementation Plan

## Problem

When rendering templates with nested blocks via `{{ template }}` or `{{ range }}`, the context (`.`) gets rebind to the current item, losing access to values like `csp_nonce` that were passed from the handler.

**Example issue:**

```html
<!-- entry-table.html -->
{{ range .Entries }}
{{ template "entry-list-item" . }}  <!-- . is now DisplayEntry, not page data -->
{{ end }}
```

```html
<!-- entry-list-item.html -->
<script nonce="{{ .csp_nonce }}" type="module">  <!-- .csp_nonce is empty! -->
```

## Current Solution

`MergeContextKeys(c, data)` manually merges `c.Keys` into template data at every `c.HTML()` call. This works but:
- Requires explicit wrapping of every `c.HTML()` call
- Doesn't help inside templates where `{{ template }}` rebinds context

## Proposed Solution: Use `{{ index . "csp_nonce" }}`

Gin already makes `c.Keys` available in templates under the dot (`.`). We can access it directly:

```html
<script nonce="{{ index . "csp_nonce" }}" type="module">
```

### Why This Works

1. Gin automatically includes `c.Keys` in template data for `gin.H{}` maps
2. `MergeContextKeys` already does this for struct-based data
3. With `{{ index . "csp_nonce" }}`, we don't need to pass nonce explicitly

### Steps to Implement

1. **Update all template usages** (~38 locations)
   
   Replace:
   ```html
   <script nonce="{{ .csp_nonce }}" type="module">
   ```
   
   With:
   ```html
   <script nonce="{{ index . "csp_nonce" }}" type="module">
   ```

2. **Simplify router handlers**
   
   Many `MergeContextKeys` calls may become unnecessary if Gin already handles it. However, keep `MergeContextKeys` if:
   - Other context values are needed (not just `csp_nonce`)
   - The template data is a struct (not `gin.H{}`)

3. **Test thoroughly**
   - Ensure CSP headers are applied correctly
   - Verify inline scripts have valid nonces
   - Check HTMX requests still work

### Pros & Cons

| Pros | Cons |
|------|------|
| No struct changes needed | Slightly verbose syntax (`index . "csp_nonce"`) |
| Works inside `range` and `{{ template }}` | Still requires Gin context injection |
| Reduces explicit nonce passing | Requires updating ~38 template locations |
| Follows Go template idioms | |

### Alternative Approaches Considered

1. **Template function `{{ nonce }}`** - Requires custom TemplateRenderer to inject context
2. **BaseTemplateData struct** - Embed `CSPNonce` in all entity structs (invasive)
3. **Explicit dict passing** - `{{ template "x" (dict "Data" . "csp_nonce" $.csp_nonce) }}` (verbose)
4. **Move inline scripts to external modules** - Best long-term but requires JS refactoring

## Files to Modify

### Templates (~38 changes)
- `templates/pages/*.html` - Page templates using nonce
- `templates/components/*.html` - Shared components
- `templates/entry/*.html` - Entry-related templates
- `templates/departure/*.html` - Departure-related templates
- `templates/person/*.html` - Person-related templates
- `templates/field/*.html`, `templates/crop/*.html`, etc.

### Routers (evaluate per-case)
- May be able to remove `MergeContextKeys` where only `csp_nonce` is needed
- Keep `MergeContextKeys` where other context values are used
