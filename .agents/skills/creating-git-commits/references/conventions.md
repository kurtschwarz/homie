# Conventions

## Subject Format

Every commit subject follows the Conventional Commits spec:

```
<type>(<optional scope>): <description>
```

- **type** — one of: `feat`, `fix`, `refactor`, `perf`, `docs`, `test`, `build`, `ci`, `chore`, `style`, `revert`.
- **scope** (optional) — the area of the codebase in parentheses, e.g. `feat(auth):`. Omit it when the change is not tied to one area.
- **description** — the intent of the change, lowercase, no trailing period, imperative or descriptive mood.
- **breaking changes** — append `!` after type/scope (`feat(api)!: ...`) and add a `BREAKING CHANGE:` footer in the body.

### Do

- `fix(orders): dispatch events after commit`
- `refactor(copilot): extract tool scoping`

### Don't

- `fix: update files` — no intent, says nothing
- `feat(copilot): add a new patient picker component and update the shared package` — lists what was touched, over the 50-character budget
- `Fixed the bug.` — not Conventional Commits, capitalized, trailing period

## The 50/72 Rule

- **Subject line — 50 characters max.** Keep the summary short so it displays cleanly in git logs without truncation. This includes the `type(scope):` prefix.
- **Blank line.** Insert one empty line between the subject and the body. This tells Git where the title ends and the details begin.
- **Body lines — 72 characters max.** Wrap every body line at 72 characters so text does not wrap awkwardly in terminals or email patches.
- **No total length limit.** The body can have as many lines as needed, provided each individual line stays at or under 72 characters.

## When to Include a Body

The body exists to carry **reasoning the diff cannot show**. Prioritize *why* over *what*.

Include a body only when at least one is true:

- The change is not self-explanatory from the diff alone.
- A reviewer would ask "why was this done this way?"
- The change encodes a non-obvious constraint, trade-off, or a fix for a specific incident/bug.

Omit the body entirely when the subject already tells the whole story (renames, small mechanical edits, obvious fixes).

### Body Rules

- Separate the subject from the body with one blank line.
- Explain the motivation and the reasoning, not the mechanics. The diff already shows the mechanics.
- Wrap body lines at 72 characters (see "The 50/72 Rule").
- Do not narrate the file list. Do not restate the subject.

## Writing With writing-plain-prose

Both the subject and the body are prose. Draft them with the `writing-plain-prose` skill so they stay clear and direct:

- **Audience**: a future engineer reading `git log` / `git blame` to understand this change.
- **Purpose**: convey the intent and reasoning behind the change in as few words as carry the meaning.

Do not be verbose. A subject with no body is the common, correct outcome.

## Signing

Every commit must be cryptographically signed. Pass `-S` to `git commit`.

If signing fails (no signing key configured, or `gpg`/SSH signing errors), stop and report the failure to the user. Do not create an unsigned commit as a fallback.

## Co-Authored-By Trailer

Every commit ends with a trailer naming the model that is creating the commit. Use the format for your provider:

| Provider  | Trailer                                                |
| --------- | ------------------------------------------------------ |
| Anthropic | `Co-Authored-By: Claude (<Model>) <noreply@anthropic.com>` |
| Moonshot  | `Co-Authored-By: Kimi (<Model>) <noreply@moonshot.ai>`     |
| OpenAI    | `Co-Authored-By: OpenAI (<Model>) <noreply@openai.com>`    |

Replace `<Model>` with the human-friendly name of the model creating the commit (for example `Opus 4.8`, `Sonnet 5`, `Haiku 4.5`, `K3`, `GPT-5`). The trailer is the last thing in the message. When there is no body, the trailer is the only content below the subject, separated from it by one blank line.
