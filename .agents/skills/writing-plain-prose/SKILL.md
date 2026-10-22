---
name: writing-plain-prose
description: Drafts, rewrites, edits, reviews, or copyedits prose using George Orwell's six rules and ASD-STE100 Simplified Technical English. Use for documentation, technical writing, product copy, emails, posts, essays, scripts, and any other style-sensitive text.
---

# writing-plain-prose

Drafts and revises prose so it is clear, direct, and honest, using Orwell's six rules and ASD-STE100 Simplified Technical English as filters. Use this skill whenever the task produces or edits prose, including documentation, commit bodies, PR descriptions, ADRs, runbooks, product copy, emails, and creative writing.

## Prerequisites / Inputs

- **Text or brief** (required): the draft to revise, or a description of what to write. If neither is present, ask what to write and about what.
- **Audience** (required): who reads the text. If not provided, ask.
- **Purpose** (required): what the reader must know or do after reading. If not provided, ask.
- **Genre** (optional): technical, instructional, business, marketing, or creative. If not stated, infer from the request and state the inference in the output notes.
- **Tone, length, and format constraints** (optional): honor exactly as given.
- **House style guide** (optional): if the project has one, it overrides this skill on conflicts.

## Process Steps

1. Record the audience, purpose, genre, tone, length, and format from the inputs. Ask the user for any required input that is missing. Do not guess the audience or purpose.

2. Classify the genre.
   - If technical, instructional, business, product, or marketing: apply the full STE baseline loaded in step 4.
   - If creative (fiction, poetry, memoir, script, lyrical prose): load `references/genre-exceptions.md` and follow the "Creative Writing" section instead of the full STE baseline.

3. Load `references/orwell-rules.md`. Note each rule and its stated test.

4. Load `references/ste-rules.md`. Note the fourteen baseline rules.

5. Load `references/protected-content.md`. List every span in the input that must not be rewritten. Copy those spans verbatim into the output.

6. Load `references/examples.md`. Use the before/after pairs to calibrate how far to rewrite. Note Example 5 — a draft that already passes needs no edit.

7. Draft or revise the text.
   - When drafting: write in concrete, direct English, one main action or statement per sentence.
   - When revising: keep the author's meaning and every explicit constraint. Cut only words that do no work.

8. Load `references/revision-checklist.md`. Run every check in it against the draft, in the listed order. Rewrite any span that fails a check.

9. Re-read `references/orwell-rules.md` rule 6. Restore any phrasing that the checklist made awkward, ungrammatical, or misleading.

10. Verify the protected spans from step 5 appear unchanged in the final text. If any changed, restore them.

11. Verify the output still satisfies the audience, purpose, tone, length, and format recorded in step 1. If a constraint conflicts with a rule, keep the constraint and record the conflict for step 12.

12. Build the notes list. Include: every term or passage kept against a rule and why, every conflict found in step 11, and every genre inference made in step 2. If the list is empty, say so in one line.

## Guardrails

- Never alter text inside the protected categories in `references/protected-content.md`. That includes code, commands, identifiers, error strings, legal text, quotations, and API field names.
- Never remove a qualifier that carries legal, medical, safety, or statistical meaning in order to shorten a sentence. Flag it instead.
- Never claim the output conforms to ASD-STE100. This skill applies an STE-derived baseline, not the licensed dictionary. Say "STE-style" or "STE baseline".
- Never change a technical term to a synonym for variety. Repetition of the exact term is required.

## Output

- The drafted or revised text, in the format the user requested. No commentary inside the text.
- A short notes list after the text, as bullets. Each bullet names one kept exception, constraint conflict, or inference, and its one-line reason.
- When revising existing text and the user asked what changed, add a bullet list of the substantive edits by category (cut filler, active voice, term consistency, split sentence, replaced jargon). Do not include a diff.
