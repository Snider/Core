### Installation
- `npm install` (install dependencies)
- `npm outdated` (verify dependency status)

### Development
- `npm run start`
- Visit http://localhost:4200

## Lint
- `npm run lint`

## Tests (headless-ready, no Chrome required)
- Unit/integration: `npm run test` (opens browser), or:
  - Headless (uses Puppeteer Chromium): `npm run test:headless`
  - Coverage report (HTML + text-summary): `npm run coverage`
- Coverage thresholds are enforced in Karma (≈80% statements/lines/functions, 70% branches for global). Adjust in `karma.conf.js` if needed.

### TDD workflow and test naming (Good/Bad/Ugly)
- Follow strict TDD:
  1) Write failing tests from user stories + acceptance criteria
  2) Implement minimal code to pass
  3) Refactor
- Test case naming convention: each logical test should have three variants to clarify intent and data quality.
  - Example helpers in `src/testing/gbu.ts`:
    ```ts
    import { itGood, itBad, itUgly, trio } from 'src/testing/gbu';

    itGood('saves profile', () => {/* valid data */});
    itBad('saves profile', () => {/* incorrect data (edge) */});
    itUgly('saves profile', () => {/* invalid data/conditions */});

    // Or use trio
    trio('process order', {
      good: () => {/* ... */},
      bad:  () => {/* ... */},
      ugly: () => {/* ... */},
    });
    ```
- Do not modify router-outlet containers in tests/components.

### Standalone Angular 20+ patterns (migration notes)
- This app is moving to Angular standalone APIs. Prefer:
  - Standalone components (`standalone: true`, add `imports: []` per component)
  - `provideRouter(...)`, `provideHttpClient(...)`, `provideServiceWorker(...)` in `app.config.ts`
  - Translation is configured via `app.config.ts` using `TranslateModule.forRoot(...)` and an HTTP loader.
- Legacy NgModules should be converted progressively. If an `NgModule` remains but is unrouted/unreferenced, keep it harmlessly until deletion is approved. Do not alter the main router-outlet page context panel.

### Web Awesome + Font Awesome (Pro)
- Both Font Awesome and Web Awesome are integrated. Do not remove. Web Awesome assets are copied via `angular.json` assets, and its base path is set at runtime in `app.ts`:
  ```ts
  import('@awesome.me/webawesome').then(m => m.setBasePath('/assets/web-awesome'));
  ```
- CSS includes are defined in `angular.json` and `src/styles.css`.

### SSR and production
- Build (browser + server): `npm run build`
- Serve SSR bundle: `npm run serve` → http://localhost:4000

### Notes for other LLMs / contributors
- Respect the constraints:
  - Do NOT edit the router-outlet main panel; pages/services are the focus
  - Preserve existing functionality; do not remove Web Awesome/Font Awesome
  - Use strict TDD and Good/Bad/Ugly naming for tests
  - Keep or improve code coverage ≥ configured thresholds for changed files
- Use Angular 20+ standalone patterns; update `app.config.ts` for global providers.
- For tests, prefer headless runs via Puppeteer (no local Chrome needed).

### Author
- Author: danny
