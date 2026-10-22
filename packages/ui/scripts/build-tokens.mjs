#!/usr/bin/env node
// Regenerates the CSS and TS that come from tokens.json (the Homie design system):
//   src/styles/tokens.css   --hm-* variables: colours per theme, scales, and the per-surface layer
//   src/styles/theme.css    Tailwind v4 @theme: hm colours, type roles, spacing, radii, shadows
//   src/lib/type-scale.ts   scale names for tailwind-merge
// Usage: node scripts/build-tokens.mjs   (run after replacing tokens.json with a newer export)
import { readFileSync, writeFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const root = join(dirname(fileURLToPath(import.meta.url)), "..");
const T = JSON.parse(readFileSync(join(root, "tokens.json"), "utf8"));
const themes = T.color.themes.map((t) => t.id);
const val = (t, id) =>
  typeof t.value === "string" ? t.value : (t.value[id] ?? t.value[themes[0]]);
const tail = (name) => name.split("-").slice(1).join("-");
const px = (v) => parseFloat(String(v));

// One rem is one design pixel. The kiosk derives its root font-size from the
// viewport, so every rem-valued token scales with the panel and no token needs
// a multiplier of its own. Desktop and mobile leave the root font-size alone,
// where a rem keeps its usual meaning and the reader's font preference applies.
const REM_BASE = 16;
const toRem = (value) => {
  if (!/^-?[\d.]+px$/.test(String(value).trim())) return value;
  const n = parseFloat(value);
  return n === 0 ? "0" : `${+(n / REM_BASE).toFixed(5)}rem`;
};
/** Rewrites every px length inside a composite value, such as a shadow. */
const remAll = (value) =>
  String(value).replace(/(-?[\d.]+)px/g, (_, n) =>
    Number(n) === 0 ? "0" : `${+(Number(n) / REM_BASE).toFixed(5)}rem`,
  );

const SURFACES = ["kiosk", "desktop", "mobile"];
const BREAKPOINT = px(T.size.tokens.find((t) => t.name === "size-breakpoint")?.value ?? "768px");

// Type roles: the Kiosk / Desktop / Mobile groups hold <surface>-<role> styles.
const roleStyles = {};
for (const g of T.type.groups) {
  for (const s of g.styles) {
    const [surface, ...rest] = s.name.split("-");
    if (!SURFACES.includes(surface)) continue;
    const role = rest.join("-");
    if (!roleStyles[role]) roleStyles[role] = {};
    roleStyles[role][surface] = s;
  }
}
const roles = Object.keys(roleStyles);
const layout = {};
for (const t of T.layout.tokens) {
  const [surface, ...rest] = t.name.split("-");
  const role = rest.join("-");
  if (!layout[role]) layout[role] = {};
  layout[role][surface] = t.value;
}
const screenWidth = T.size.tokens.find((t) => t.name === "size-screen-w")?.value ?? "1920px";
const screenHeight = T.size.tokens.find((t) => t.name === "size-screen-h")?.value ?? "1080px";

/*
 * The kiosk's root font-size: the design is 1920 x 1080, and min() fits it to
 * whichever axis runs out first, so nothing ever overflows and the layout
 * absorbs the slack on the other axis.
 *
 * The ratio stays inside the calc on purpose. A viewport-unit font-size
 * (`min(0.8333vw, 1.4815vh)`) is snapped to 1/60px before anything multiplies
 * it, and that rounding compounds through every rem — a full-bleed element
 * lands ~2px short at every resolution. Multiplying a px base by a unitless
 * ratio resolves at full precision and is exact.
 */
const KIOSK_ROOT_FONT_SIZE = `calc(${REM_BASE}px * min(100vw / ${screenWidth}, 100dvh / ${screenHeight}))`;

// Photo and emoji dimming per theme (the filter family in tokens.json, arranged by theme).
const FILTERS = {
  light: ["none", "none"],
  dark: ["brightness(0.85)", "none"],
  night: ["brightness(0.5) saturate(0.7) sepia(0.25)", "brightness(0.75) saturate(0.8)"],
};
const SELECTOR = {
  light: ':root,\n.light,\n[data-theme="light"]',
  dark: '.dark,\n[data-theme="dark"]',
  night: '.night,\n[data-theme="night"]',
};

const out = [
  "/*\n * Homie tokens. GENERATED from tokens.json by scripts/build-tokens.mjs; edit tokens.json, not this file.\n * Themes: light (Day), dark (Evening), night, set with a class (.dark / .night) or data-theme.\n * Surfaces: desktop by default, mobile under " +
    BREAKPOINT +
    'px, kiosk with data-surface="kiosk". data-surface="mobile" / "desktop" force the others.\n */\n',
];
for (const id of themes) {
  const lines = T.color.tokens.map((t) => `  --hm-${t.name}: ${val(t, id)};`);
  const [photo, emoji] = FILTERS[id] ?? FILTERS.light;
  lines.push(`  --hm-photo-filter: ${photo};`, `  --hm-emoji-filter: ${emoji};`);
  if (id !== "light") lines.push("  color-scheme: dark;");
  out.push(`${SELECTOR[id] ?? `.${id},\n[data-theme="${id}"]`} {\n${lines.join("\n")}\n}\n`);
}
const fixed = [
  "spacing",
  "radius",
  "layout",
  "shadow",
  "size",
  "border",
  "opacity",
  "interaction",
].flatMap((f) => T[f]?.tokens ?? []);
// Breakpoints are compared against real pixels in media queries, and the pill
// radius is a "larger than anything" sentinel rather than a measurement.
const KEEP_PX = new Set(["size-breakpoint", "radius-pill"]);
const fixedValue = (t) =>
  KEEP_PX.has(t.name) ? t.value : t.name.startsWith("shadow-") ? remAll(t.value) : toRem(t.value);

// A rem-valued custom property computes to its token stream ("7rem") unless it
// is registered, so JS that measures one would read 7 instead of 112 — wrong,
// and silent. Registering as <length> makes getComputedStyle resolve it to px.
const JS_READ = [
  "hm-fs-headline",
  "hm-size-panel-portrait",
  "hm-size-panel-square",
  "hm-size-panel-landscape",
];
out.push(
  `/* Read by JS via getComputedStyle; registering them resolves rem to px. */\n` +
    JS_READ.map(
      (name) =>
        `@property --${name} {\n  syntax: "<length>";\n  inherits: true;\n  initial-value: 0px;\n}`,
    ).join("\n") +
    `\n\n/* Scales, in rem: one rem is one design pixel. Spacing and radius are written at kiosk size; components multiply them by --hm-density. */\n:root {\n${fixed.map((t) => `  --hm-${t.name}: ${fixedValue(t)};`).join("\n")}\n}\n\n` +
    `/* The kiosk scales its whole design by setting the root font-size; every rem above follows. */\n[data-surface="kiosk"] {\n  font-size: ${KIOSK_ROOT_FONT_SIZE};\n}\n`,
);

const surfaceBlock = (s) =>
  [
    ...roles.flatMap((r) => [
      `  --hm-fs-${r}: ${toRem(roleStyles[r][s].fontSize)};`,
      `  --hm-lh-${r}: ${roleStyles[r][s].lineHeight};`,
    ]),
    ...Object.keys(layout).map((n) => `  --hm-${n}: var(--hm-${s}-${n});`),
  ].join("\n");
out.push(
  `/* The active surface. */\n:root,\n[data-surface="desktop"] {\n${surfaceBlock("desktop")}\n}\n\n` +
    `@media (max-width: ${BREAKPOINT - 0.02}px) {\n  :root {\n${surfaceBlock("mobile").replace(/^/gm, "  ")}\n  }\n}\n\n` +
    `[data-surface="desktop"] {\n${surfaceBlock("desktop")}\n}\n\n[data-surface="mobile"] {\n${surfaceBlock("mobile")}\n}\n\n[data-surface="kiosk"] {\n${surfaceBlock("kiosk")}\n}\n`,
);
const SEMANTIC = [
  ["background", "bg"],
  ["foreground", "ink"],
  ["card", "surface"],
  ["card-foreground", "ink"],
  ["popover", "surface"],
  ["popover-foreground", "ink"],
  ["primary", "accent"],
  ["primary-foreground", "on-accent"],
  ["secondary", "surface"],
  ["secondary-foreground", "ink"],
  ["muted", "photo"],
  ["muted-foreground", "ink-2"],
  ["accent", "accent-soft"],
  ["accent-foreground", "ink"],
  ["border", "line"],
  ["input", "line"],
  ["ring", "accent"],
];
out.push(
  "/* shadcn semantic variables, declared wherever a theme is set so they resolve against that theme. */\n:root,\n.light,\n.dark,\n.night,\n[data-theme] {\n" +
    SEMANTIC.map(([a, b]) => `  --${a}: var(--hm-${b});`).join("\n") +
    "\n  /* Homie has no error colour yet. This placeholder keeps stock shadcn components rendering; replace it. */\n  --destructive: var(--hm-swatch-clay);\n}\n" +
    "/* shadcn's radius follows the surface's control radius. */\n:root,\n[data-surface] {\n  --radius: var(--hm-control-radius);\n}\n",
);
writeFileSync(join(root, "src/styles/tokens.css"), out.join("\n"));

// Density-scaled spacing and radius keys: p-hm-16 = 16px at kiosk, ~10px on desktop, ~9px on mobile.
const EXTRA = [2, 22, 26, 30, 44, 52, 76, 128];
const spaceSteps = [...new Set([...T.spacing.tokens.map((t) => px(t.value)), ...EXTRA])].sort(
  (a, b) => a - b,
);
const radiusSteps = T.radius.tokens
  .filter((t) => /px$/.test(t.value) && px(t.value) < 999)
  .map((t) => px(t.value));
const SEMANTIC_SPACE = {
  touch: "touch",
  button: "button",
  "badge-lg": "badge-lg",
  badge: "badge",
  "badge-sm": "badge-sm",
  "time-col": "time-col",
  screen: "screen-padding",
  column: "column-gap",
  section: "section-gap",
};
const kioskSizes = T.size.tokens.filter((t) => t.name !== "size-breakpoint");
const styles = roles.map((r) => {
  const k = roleStyles[r].kiosk;
  return [
    `  --text-hm-${r}: var(--hm-fs-${r});`,
    `  --text-hm-${r}--line-height: var(--hm-lh-${r});`,
    ...(k.letterSpacing ? [`  --text-hm-${r}--letter-spacing: ${k.letterSpacing};`] : []),
    `  --text-hm-${r}--font-weight: ${k.fontWeight};`,
  ].join("\n");
});
const theme = `/* GENERATED from tokens.json by scripts/build-tokens.mjs; edit tokens.json, not this file. */

@theme inline {
  /* Colours: bg-hm-surface, text-hm-ink-2 ... Swatches and tint-strength have no utility; they are chosen via lib/people.ts, not Tailwind classes. */
${T.color.tokens
  .filter((t) => !t.name.startsWith("swatch-") && t.name !== "tint-strength")
  .map((t) => `  --color-hm-${t.name}: var(--hm-${t.name});`)
  .join("\n")}

  /* Type roles, sized by the active surface: text-hm-headline, text-hm-body ... */
${styles.join("\n")}

  /* Surface layout: h-hm-touch, h-hm-button, size-hm-badge, p-hm-screen, gap-hm-column, gap-hm-section ... */
${Object.entries(SEMANTIC_SPACE)
  .map(([k, v]) => `  --spacing-hm-${k}: var(--hm-${v});`)
  .join("\n")}

  /* Density-scaled spacing (kiosk rem × density): gap-hm-12, p-hm-24, size-hm-28 ... */
${spaceSteps.map((n) => `  --spacing-hm-${n}: calc(${toRem(`${n}px`)} * var(--hm-density));`).join("\n")}

  /* Fixed kiosk sizes: w-hm-panel-landscape, w-hm-now-card ... */
${kioskSizes.map((t) => `  --spacing-hm-${tail(t.name)}: var(--hm-${t.name});`).join("\n")}

  /* Radii: rounded-hm-panel and rounded-hm-control follow the surface; rounded-hm-18 etc. scale with density. */
  --radius-hm-panel: var(--hm-panel-radius);
  --radius-hm-control: var(--hm-control-radius);
${radiusSteps.map((n) => `  --radius-hm-${n}: calc(${toRem(`${n}px`)} * var(--hm-density));`).join("\n")}

  /* Shadows: shadow-hm-sheet ... */
${T.shadow.tokens.map((t) => `  --shadow-hm-${tail(t.name)}: var(--hm-${t.name});`).join("\n")}
}

@theme {
  --font-sans: "Lexend Variable", "Lexend", system-ui, sans-serif;
  --ease-hm-panel: cubic-bezier(0.4, 0, 0.2, 1);
  --ease-hm-sheet: cubic-bezier(0.2, 0.9, 0.3, 1);
  --animate-hm-blink: hm-blink 2s steps(1, end) infinite;

  @keyframes hm-blink {
    0%, 49.9% { opacity: 1; }
    50%, 100% { opacity: 0.18; }
  }
}

/* Surface variants. mobile: = under ${BREAKPOINT}px (or data-surface="mobile"); kiosk: = data-surface="kiosk". */
@custom-variant kiosk (&:where([data-surface="kiosk"], [data-surface="kiosk"] *));
@custom-variant mobile {
  &:where([data-surface="mobile"], [data-surface="mobile"] *) {
    @slot;
  }
  @media (max-width: ${BREAKPOINT - 0.02}px) {
    &:where(:not([data-surface="desktop"], [data-surface="desktop"] *, [data-surface="kiosk"], [data-surface="kiosk"] *)) {
      @slot;
    }
  }
}
`;
writeFileSync(join(root, "src/styles/theme.css"), theme);

const list = (a) => JSON.stringify(a);
writeFileSync(
  join(root, "src/lib/type-scale.ts"),
  "/** GENERATED from tokens.json by scripts/build-tokens.mjs. Scale names used to teach tailwind-merge about the hm-* utilities. */\n" +
    `export const HM_TEXT_STYLES = ${list(roles)} as const\n` +
    `export const HM_SIZES = ${list([...Object.keys(SEMANTIC_SPACE), ...spaceSteps.map(String), ...kioskSizes.map((t) => tail(t.name))])} as const\n` +
    `export const HM_RADII = ${list(["panel", "control", ...radiusSteps.map(String)])} as const\n` +
    `export const HM_SHADOWS = ${list(T.shadow.tokens.map((t) => tail(t.name)))} as const\n`,
);
console.log(
  `tokens.css, theme.css, type-scale.ts written (${T.color.tokens.length} colours, ${roles.length} type roles × 3 surfaces)`,
);
