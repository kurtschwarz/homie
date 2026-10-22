# Examples

Each example shows a before, an after, and the key patterns that produced the change. Use these as calibration for how aggressive to be.

## Example 1 — Technical documentation

**Before**

> It is important to note that in order to facilitate the utilization of the new caching layer, a configuration change will need to be made by the operator prior to deployment. At the end of the day, this is low-hanging fruit that will move the needle on our performance metrics, and it should be noted that approximately a 40ms improvement in p99 latency has been observed in staging environments.

**After**

> Before you deploy, set `CACHE_BACKEND=redis` in the service config. In staging, this cut p99 latency by about 40 ms.

**Key patterns**
- Cut `It is important to note that`, `in order to`, `it should be noted that` (rule 3 empty frames).
- `facilitate the utilization of` → `use`, then dropped entirely once the instruction became concrete (rule 2, rule 3).
- `a configuration change will need to be made by the operator` → imperative with the actual key name (rule 4, STE rule 8).
- Cut `at the end of the day`, `low-hanging fruit`, `move the needle` (rule 1).
- Kept `about` in front of `40 ms` — it modifies a measurement (protected qualifier).
- `approximately` → `about` (rule 2), meaning preserved.
- 63 words → 24 words, in two sentences, each under the STE limit.

## Example 2 — Procedure

**Before**

> The database migration should be run by the engineer, and if there are any errors that are encountered during the process, then the rollback script can be utilized in order to restore the previous state of the schema, though it is recommended that a backup be taken beforehand.

**After**

> 1. Confirm a backup exists and is less than 24 hours old.
> 2. Run `npm run migrate`. The command prints `migrations applied: N`.
> 3. If the command exits non-zero, run `npm run migrate:rollback`. The command restores the previous schema.

**Key patterns**
- One 47-word sentence → three imperative steps, each with an action and an expected result (STE rules 1, 8).
- `should be run by the engineer` → `Run` (rule 4).
- `it is recommended that a backup be taken` → step 1, moved to the front because it must happen first (rule 4 disguised passive, STE rule 10).
- `can be utilized in order to` → `run` (rules 2, 3).
- Command names and output strings are protected and appear verbatim.
- Made the failure condition testable (`exits non-zero`) instead of vague (`any errors encountered`).

## Example 3 — Passive voice that should stay

**Before**

> At 03:12 UTC, 14,000 rows were deleted from the `patients` table.

**After (unchanged)**

> At 03:12 UTC, 14,000 rows were deleted from the `patients` table.

**Notes bullet produced**

> - Kept passive in the timeline entry: the actor is not yet identified, and naming a suspected actor would assign blame the report cannot support.

**Key patterns**
- Rule 4 has an explicit exception for unknown actors. Do not invent one.
- Table name is protected.

## Example 4 — Business email

**Before**

> Hi team, I wanted to circle back on the action items from our sync. Going forward, we'll need to socialize the new process with stakeholders and operationalize the feedback loop. Let's touch base offline to align on next steps and make sure we're all rowing in the same direction.

**After**

> Hi team,
>
> Two things from yesterday's sync:
>
> 1. I'll walk Support and Billing through the new refund process this week.
> 2. Priya will set up the weekly feedback review, starting Monday.
>
> Reply if either owner or date is wrong.

**Key patterns**
- Cut `circle back`, `going forward`, `touch base`, `align on next steps`, `rowing in the same direction` (rule 1).
- `socialize` → `walk … through`; `operationalize` → `set up` (rule 5).
- `stakeholders` → the actual teams. Vague nouns get replaced with the specific thing (STE rule 5).
- Every item got an owner and a date, because an action item without them is not an action item.
- The email is shorter and now says something the reader can act on. If the input had not named owners, the correct move is to ask, not to invent them.

## Example 5 — Over-correction to avoid

**Before**

> Because the token refresh is idempotent, the client can safely retry the request after a 401 without risking a duplicate session.

**Bad after**

> The client can retry after a 401.

**Correct after (unchanged)**

> Because the token refresh is idempotent, the client can safely retry the request after a 401 without risking a duplicate session.

**Key patterns**
- The bad rewrite deleted the reason (`idempotent`) and the guarantee (`without risking a duplicate session`). Both carry meaning.
- `idempotent` is the exact technical term. Rule 5 keeps it; the audience is client developers who know it.
- 21 words, active voice, condition before action, no filler. It already passes. Leave it alone.
- A draft that passes the checklist needs no edit. Do not edit to show effort.

## Example 6 — Creative prose

**Before**

> She waited. The rain came down in sheets, and she thought that it seemed as though the whole city was holding its breath, waiting alongside her, for something that might never actually arrive at all.

**After**

> She waited. The rain came down in sheets. The whole city held its breath with her, waiting for something that might never come.

**Key patterns**
- Kept the short opening sentence and the deliberate rhythm.
- Cut `she thought that it seemed as though` — filter words that push the reader out of the scene.
- `the whole city was holding its breath, waiting alongside her` → `held its breath with her`. Kept the image, cut the words.
- Cut `actually` and `at all` (rule 3 filler).
- `in sheets` is a worn phrase, but it is the character's register and it is doing work. Kept under rule 6.
- Did not apply STE sentence limits or term-consistency rules. This is creative prose.
