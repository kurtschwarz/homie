# @homie/ui

The Homie design system as a shared UI package: tokens, a Tailwind v4 theme on shadcn conventions, and React components built with Base UI. It covers three surfaces: phones, desktop browsers, and the 1920 × 1080 kitchen kiosk read from across the room.

## What's in here

```
homie-ui/
├── tokens.json              source of truth, exported from the Homie design system
├── scripts/build-tokens.mjs regenerates the files marked (generated) from tokens.json
├── components.json          shadcn CLI config for this package
├── examples/screens.tsx     Home, Week and Weather screens with sample data
└── src/
    ├── styles/
    │   ├── globals.css      Tailwind entry: imports everything below, variants, base layer
    │   ├── tokens.css       --hm-* variables per theme and per surface + shadcn mapping (generated)
    │   └── theme.css        Tailwind @theme, type roles, mobile:/kiosk: variants (generated)
    ├── lib/                 cn(), people, time formatting, scale names (generated)
    ├── hooks/               useNow, useHomieTheme, useIdle
    └── components/          the Homie components
```

## Add it to your monorepo

1. Copy `homie-ui/` to `packages/homie-ui`. If your workspace uses another scope (shadcn's template uses `@workspace/ui`), rename `name` in `package.json` and the aliases in `components.json` to match.
2. Install: `pnpm install`, then add it to your app: `pnpm add @homie/ui --workspace --filter <your-app>`.
3. In your app's main CSS, replace `@import "tailwindcss";` with:
   ```css
   @import "@homie/ui/globals.css";
   ```
   That brings in Tailwind, the Lexend font (bundled through Fontsource, so the Pi doesn't need the internet for type), the tokens, and the theme. It also tells Tailwind to scan this package's components.
4. Next.js only: add `transpilePackages: ["@homie/ui"]` to `next.config`. Vite needs nothing.
5. Wrap the app:
   ```tsx
   import { HomieThemeProvider } from "@homie/ui"

   // phones and browsers
   <HomieThemeProvider theme="auto" surface="auto">{children}</HomieThemeProvider>
   // the kitchen display
   <HomieThemeProvider theme="auto" surface="kiosk">{children}</HomieThemeProvider>
   ```

## Surfaces

| Surface | When | Type | Targets |
| --- | --- | --- | --- |
| `mobile` | viewport under 768px | body 15px, headline 34px | 44px |
| `desktop` | 768px and up (the default) | body 16px, headline 56px | 40px |
| `kiosk` | `surface="kiosk"` / `data-surface="kiosk"` | body 28px, headline 112px | 64px |

Kiosk is opt-in because it's a place, not a width: a 1920px monitor should still get desktop sizes. `data-surface="mobile"` or `"desktop"` on any element forces that surface for its subtree (handy for previews and Storybook).

You mostly don't think about it: pick a type role or token and the surface picks the size.

- **Type roles** (`text-hm-headline`, `text-hm-lead`, `text-hm-body` ...) carry size, line height, tracking and weight for the active surface.
- **Surface layout:** `h-hm-touch`, `h-hm-button`, `size-hm-badge`, `p-hm-screen`, `gap-hm-column`, `gap-hm-section`, `rounded-hm-panel`, `rounded-hm-control`.
- **Density-scaled spacing:** `p-hm-16`, `gap-hm-12`, `size-hm-28`, `rounded-hm-18` are kiosk pixels × the surface's density (kiosk 1, desktop 0.6, mobile 0.55). Plain Tailwind spacing (`p-4`) stays fixed.
- **Variants:** `mobile:` and `kiosk:` for layout changes. `dark:` covers Evening and Night; `night:` is Night only.
- **Layouts:** `HomeLayout`/`HomeEvents` put the photo beside the events on kiosk and desktop and stack them (events first) on mobile. `WeekView` is seven columns, one on mobile. `HourlyForecast` scrolls sideways on mobile. `ScreenHeader` hides its clock on mobile. Kiosk alone is a fixed 1920 × 1080 canvas.

## Themes

Three themes: `light` (Day), `dark` (Evening) and `night`. With `theme="auto"` they follow the clock: Day until 7 pm, Evening from 7 pm, Night from 9 pm to 6:30 am. The provider sets both `data-theme` and a class (`.light` / `.dark` / `.night`) on `<html>`, so stock shadcn components follow along. Use `scope="local"` to theme just a subtree.

Variants: `dark:` applies in Evening **and** Night (both are dark); `night:` applies only at Night. Night is warm and blue-free, and photos and emoji dim automatically through `--hm-photo-filter` and `--hm-emoji-filter`.

## Using the tokens

Everything Homie is prefixed `hm-` so it never collides with shadcn's own names (Homie's `accent` is the now/next blue, shadcn's `accent` is a hover ground).

| What | Utilities | Examples |
| --- | --- | --- |
| Colours | `bg-hm-*`, `text-hm-*`, `border-hm-*` | `bg-hm-surface`, `text-hm-ink-2`, `bg-hm-accent-soft`, `bg-hm-person-a-tint`, `bg-hm-scrim` |
| Type roles | `text-hm-<role>` sets size, line height, tracking and weight for the surface | `text-hm-headline`, `text-hm-clock`, `text-hm-lead`, `text-hm-item`, `text-hm-body`, `text-hm-caption`, `text-hm-eyebrow` (add `uppercase`) |
| Surface layout | `h-`/`w-`/`size-`/`p-`/`gap-` | `h-hm-touch`, `h-hm-button`, `size-hm-badge-lg`, `p-hm-screen`, `gap-hm-column`, `w-(--hm-time-col)` |
| Density spacing | any spacing utility with `hm-<kiosk px>` | `p-hm-16`, `gap-hm-12`, `size-hm-28`, `mb-hm-8` |
| Kiosk sizes | fixed | `w-hm-panel-landscape`, `w-hm-now-card`, `h-hm-screen-h` |
| Radii | `rounded-hm-*` | `rounded-hm-panel`, `rounded-hm-control`, `rounded-hm-18` (density), `rounded-full` |
| Shadows | `shadow-hm-*` | `shadow-hm-sheet`, `shadow-hm-float` |
| Motion | `ease-hm-*`, `animate-hm-blink` | `ease-hm-sheet`, `ease-hm-panel` |
| Raw values | `var(--hm-*)` | `var(--hm-fs-headline)`, `var(--hm-density)`, `var(--hm-kiosk-touch)` |

The 22 type roles: `display-xl`, `headline`, `clock`, `display`, `title-lg`, `clock-sm`, `title`, `stat`, `title-sm`, `lead`, `item-lg`, `item`, `strong`, `body`, `body-sm`, `label-lg`, `label`, `caption`, `eyebrow-lg`, `eyebrow`, `small`, `micro`. Change a role's weight with the usual utilities (`text-hm-body font-semibold` for day names).

`cn()` from `@homie/ui` knows the Homie scales, so `cn("text-hm-label text-hm-ink", "text-hm-ink-2")` keeps the size and swaps the colour, and `cn("h-hm-button", "h-hm-touch")` resolves to the second.

### shadcn mapping

Stock shadcn components pick up Homie through these:

| shadcn | Homie |
| --- | --- |
| `background` / `foreground` | `bg` / `ink` |
| `card`, `popover`, `secondary` | `surface` |
| `primary` / `primary-foreground` | `accent` / `on-accent` |
| `muted` / `muted-foreground` | `photo` / `ink-2` |
| `accent` / `accent-foreground` | `accent-soft` / `ink` |
| `border`, `input` | `line` |
| `ring` | `accent` |
| `radius` | the surface's `control-radius` (22 / 12 / 12px) |
| `destructive` | **placeholder** (`person-a`). Homie doesn't define an error colour yet. |

To add stock shadcn components into this package, run the shadcn CLI from `packages/homie-ui`. `components.json` says `"style": "base-nova"`; change it to whichever Base UI style your app already uses.

## Components

All are plain React (19) with Tailwind classes. Interactive ones are `"use client"`.

| Component | Base UI | What it is |
| --- | --- | --- |
| `Button` (`variant="outline" \| "text"`, `block`) | `Button` | Outlined nav button (`button` tall) and text button (`touch` tall). Supports `render` (set `nativeButton={false}` for links). |
| `Clock` (`size="default" \| "sm"`) | | Time with a blinking colon; follows the clock unless you pass `date`. |
| `WeatherGlance` (`stale`) | `Button` | Header icon + temperature that opens the Weather screen; dims a cached reading served after an upstream failure. |
| `WeatherIcon`, `CONDITION_LABEL` | | Lucide weather icons for `sun`, `cloud-sun`, `cloud`, `rain`, `moon`. |
| `PersonBadge`, `PersonBadges`, `PersonLegend` | | Emoji badge in a person's colour; legend pill. |
| `ReminderChip` | | One-line nudge under the clock. |
| `NextUp` | | The headline. Measures itself and steps down from the surface's `headline` size (to 0.54×) to fit two lines; empty state when there's no `title`. |
| `EventRow`, `EventList` | | When / who / what rows; `compact` for the sheet. |
| `ComingUp`, `ComingUpSwipeArea`, `ComingUpHandle`, `ComingUpSheet` | `Drawer` (non-modal) | The Coming up sheet: swipe up on the list to open, swipe down to close, closes itself after 45 s. Rises inside the events column (pass `container`) on kiosk and desktop; a normal bottom sheet on mobile. |
| `ScreenHeader` | | Home button, title, middle slot, small clock (hidden on mobile). |
| `Screen`, `HomeLayout`, `HomeEvents`, `WeatherLayout` | | Screen layouts that switch arrangement per surface. |
| `WeekView`, `WeekDay`, `EventCard` | | The Week screen: seven columns, a stacked list on mobile; person-tinted cards. |
| `WeatherNow`, `Callout` | | The Now card and the what-to-wear tip. |
| `HourlyForecast`, `DailyForecast` | | 12 hour tiles (a sideways scroller on mobile) and the weekly range bars. |
| `PhotoPanel` | | Crossfading slideshow with an EXIF place/date caption. Kiosk: resizes to each photo's shape. Desktop: fills beside the events. Mobile: a 4:3 card. |
| `HomieThemeProvider`, `useHomieTheme`, `useNow`, `useIdle` | | Theme schedule and surface, a ticking clock, and the kiosk's 60 s drift back home. |

People are `{ key: "a" | "b" | "c" | "family" | "grownups", name, emoji? }`. Colour always comes with the emoji and the name.

Time helpers in `@homie/ui` follow the content rules: `formatTime` → `5:45 pm`, `formatRange` → `3:30–4:15 pm`, `until` → `in 2 hr 17 min`, `dayLabel` → `Today` / `Tomorrow` / weekday, `formatTakenAt` → `Fri, Aug 14 · 8:12 pm`.

See `examples/screens.tsx` for the Home, Week and Weather screens, written once and working on every surface.

## Updating the tokens

When the design system changes, export its `tokens.json` over this one and run:

```sh
pnpm --filter @homie/ui tokens
```

That rewrites `tokens.css`, `theme.css` and `type-scale.ts`. Don't hand-edit those three.

## Notes

- Contrast: `ink-3` on `accent-soft` is 4.31:1 in Day and 4.00:1 at Night, a little under 4.5:1. It's only used for the date under Today in the week view; at desktop and mobile caption size that's a real (if small) miss, so consider `ink-2` there. The 2px `line` borders are decorative (about 1.3:1), so buttons always carry a label or icon.
- The focus outline (3px `accent`, offset 3px) matters on desktop; on the kiosk you'll never see it.
