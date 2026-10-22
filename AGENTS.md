# Local development

- Run `just format` after editing TypeScript, CSS, or JSON. Pass file paths to format or lint only those files, e.g. `just format apps/display/src/app.tsx` or `just lint apps/display/src/app.tsx`. `just lint` runs Biome's recommended rules.
- Run `just dev` before verifying changes to `apps/display` or `apps/manager`. It starts the API and Vite servers, then opens Firefox Developer Edition with the local pages in separate tabs.
- Use the Firefox DevTools MCP (`firefox-devtools`) to inspect the running app, console, and network requests. It connects to the Firefox instance launched by `just dev`; it cannot launch its own browser in this mode.
- The display is designed for a 1920x1080 CSS viewport (`apps/display/src/index.css`), which is larger than most laptop screens. `just dev` seeds the window size into the profile's `xulstore.json` and scales the browser with `layout.css.devPixelsPerPx`, so the whole viewport fits on a smaller screen. The values live near the top of `scripts/dev/justfile` and suit a 1800x1169 point display; lower `firefox_dppx` (or set `FIREFOX_DPPX`) for a smaller screen. The window ends up one CSS pixel taller than 1080, which the display hides.
- `xulstore.json` is seeded only when absent, because Firefox owns it afterwards and rewrites it as the window moves. Delete the profile to pick up a changed `firefox_css_width` or `firefox_css_height`. `user.js` is rewritten every run, so `firefox_dppx` takes effect on the next start.
- The display URL is `http://localhost:5173/`. The manager URL is `http://localhost:5174/manage/` (the `/manage/` path is required by its Vite base configuration).
- A fresh manager profile may show the account setup screen. The setup code is logged by the API; inspect the API container logs with `docker compose logs api`.
- The Firefox profile is disposable local state, ignored by Git under `.scratch/*`. Keep it separate from any personal Firefox profile. `just dev` closes the dev Firefox when it exits.
- Firefox writes its own output to `.scratch/firefox-dev-profile/firefox.log` instead of the terminal, which `just dev` leaves to the compose logs. Read that file when the browser fails to start.
