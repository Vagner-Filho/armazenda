# Implementation Plan: Static Armazenda Landing Page

## Scope and architecture

- Add `landing/index.html`, `landing/styles.css`, and a local `landing/favicon.svg`.
- Keep the site independent: no Go templates, no shared `assets/` references, no Tailwind, no framework, and no build step.
- Link the public page to the confirmed app host at `app.armazenda.com`; hosting/DNS configuration is not part of this first page implementation.
- Do not edit `main.go`, the app's `/` route, or existing app assets.

## Test strategy

- **Static serving:** the focused Playwright config starts Python's standard-library static server and checks the HTML, stylesheet, and favicon. No app or database is required.
- **Content and separation:** inspect the page's language, metadata, verified product descriptions, app/pricing destinations, and absence of Go template syntax, Tailwind, and external runtime dependencies. Run `go list -f '{{range .EmbedFiles}}{{println .}}{{end}}' .` and confirm no `landing/` files are embedded.
- **Visual/accessibility:** add a standalone Playwright config/spec that serves `landing/` with Python's static server and does not run the app/database setup. Run `cd test/e2e && ./node_modules/.bin/playwright test --config landing.playwright.config.js` to check layout at 320px, 390px, 520px, 760px, 768px, 1024px, and 1440px, keyboard navigation, the skip link, heading order, reduced motion, section-anchor offsets, and browser errors. Check primary text contrast ratios against WCAG thresholds.
- **App suites:** no Go or JavaScript behavior changes are planned, so full unit/E2E suites are not applicable. Avoid `make build` because it regenerates `assets/static/css/output.css`, which currently has an unrelated working-tree modification.

## Phases

### 1. Static page and local presentation
- Create semantic Brazilian Portuguese content in `landing/index.html` with a skip link, header/navigation, hero, feature sections, app/pricing calls to action, and footer.
- Create self-contained responsive styles in `landing/styles.css`; keep any illustration inline or local and label any UI preview as illustrative.
- Add an independent local favicon in `landing/favicon.svg`.
- Add `test/e2e/landing.playwright.config.js` and `test/e2e/landing.spec.js` for an isolated static preview; do not attach the normal E2E app/database global setup.
- During this phase, run the static-server checks and focused Playwright test; verify local paths/links.

### 2. Responsive, accessibility, and boundary verification
- Refine mobile and desktop layouts, visible focus states, contrast, reduced-motion support, and metadata.
- Check text contrast and review the focused Playwright results for mobile/desktop sizes and keyboard navigation.
- Run the embedded-file listing check and confirm the app's source/build configuration remains untouched.
- Record exact verification commands/results in this SDD; mark `Done` only if all acceptance criteria are met. Live host/domain checks remain for the separate deployment task.

## Risks and out of scope

- The static hosting provider and DNS/reverse-proxy configuration are not specified; this work only creates the deployable static source directory.
- Marketing claims should remain limited to features verified in the application; pricing remains linked to the app's live pricing page rather than duplicated in static copy.
- Contact forms, analytics, deployment automation, and app-domain/OAuth configuration changes are out of scope.
