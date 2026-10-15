# sdd:static_landing_page

## Status
Done

## Goal
Create a public, static marketing page for Armazenda at `armazenda.com`, separate from the application hosted at `app.armazenda.com`. The page should explain the product and direct visitors to the existing app without sharing its runtime or frontend build stack.

## Requirements

### Functional
- Serve a complete Portuguese (Brazil) landing page from static files in a top-level `landing/` directory.
- Describe verified Armazenda capabilities, including grain entry/exit records (romaneios), humidity tracking, reports, and NF-e workflows.
- Provide clear navigation and calls to action linking to `https://app.armazenda.com/` and `https://app.armazenda.com/pricing`.
- Work when served directly by a static file server, without the Go app or a backend.

### Non-Functional
- Use plain HTML and CSS, with no Tailwind, JavaScript framework, package manager, or required build step.
- Keep all landing source and assets outside the Go `go:embed` paths (`templates/` and `assets/`); do not change the app's `/` login route.
- Be responsive, keyboard navigable, semantically structured, and usable with reduced motion enabled.
- Include page title, description, language, and canonical URL metadata; use local assets and avoid runtime dependence on third-party CDNs.

## Acceptance Criteria
- [x] `landing/index.html` and its local assets are served successfully from a standalone static server, with no Go server required. Verify with `cd test/e2e && ./node_modules/.bin/playwright test --config landing.playwright.config.js`; the isolated config starts Python's static server and checks the HTML, stylesheet, and favicon.
- [x] Main login and pricing links point to `https://app.armazenda.com/` and `https://app.armazenda.com/pricing`. Verify their exact destinations in the focused Playwright test; live destination checks are deferred until a preview deployment exists.
- [x] The landing directory is not included in the Go package's embedded files. Verify with `go list -f '{{range .EmbedFiles}}{{println .}}{{end}}' .` and confirm no result begins with `landing/`.
- [x] The page source has no Tailwind utilities, template syntax, or required external script/font dependencies. Verify with the static source checks and `page.locator('script').toHaveCount(0)` in the focused Playwright test.
- [x] At mobile (320px, 390px, and 520px) and desktop/tablet (760px, 768px, 1024px, and 1440px) viewports, content has no horizontal overflow and both primary calls to action remain visible. Verify with `cd test/e2e && ./node_modules/.bin/playwright test --config landing.playwright.config.js`.
- [x] Keyboard navigation reaches every link, the skip link is the first focus stop, the heading outline has no skipped levels, the dashboard preview has an accessible description, and reduced-motion preferences are honored. Verify with the focused Playwright test; calculated text contrast ratios meet WCAG AA thresholds for body text and large headings.

## Verification Results
- `cd test/e2e && ./node_modules/.bin/playwright test --config landing.playwright.config.js` — PASS; 2 Firefox tests, including static asset responses, app URLs, 7 viewport widths, keyboard link traversal, skip link, heading outline, reduced motion, and anchor offset.
- `go list -f '{{range .EmbedFiles}}{{println .}}{{end}}' .` — PASS; 185 embedded app files listed and none begin with `landing/`.
- Static source and HTTP smoke checks — PASS; `/`, `/styles.css`, and `/favicon.svg` return HTTP 200; source has no template syntax, Tailwind reference, script, or external runtime asset; local/section links resolve.
- WCAG contrast calculation — PASS for checked content pairs: body copy 4.62:1, section eyebrow 4.84:1, workflow copy 4.61:1, footer description 4.64:1, footer fine print 4.96:1, hero copy 11.47:1, CTA label 11.77:1; large italic heading 3.94:1 (3:1 threshold).
- `make test-go`, `make test-js`, and full `make test-e2e` — NOT RUN; no app behavior changed, and the focused standalone Playwright suite covers this page.
- `make build` — NOT RUN; it regenerates `assets/static/css/output.css`, which has a pre-existing working-tree modification. The Go embed listing confirms the landing directory is excluded without rebuilding.
- Live preview deployment and live app-link navigation — NOT RUN; hosting/DNS configuration is out of scope for this first static-page implementation.
