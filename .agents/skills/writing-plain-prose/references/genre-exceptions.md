# Genre Exceptions

The STE baseline is written for technical and instructional prose. Other genres need a different mix. Use the section that matches the genre classified in step 2 of the process.

## Genre Classification

| Genre | Signals | Rules to apply |
|---|---|---|
| Technical / instructional | docs, runbook, README, ADR, API reference, how-to, error message | Orwell 1–6 + full STE baseline |
| Engineering communication | commit body, PR description, incident report, design doc, code comment | Orwell 1–6 + STE baseline, plus the section below |
| Business / internal | email, memo, status update, proposal, meeting notes | Orwell 1–6 + STE rules 1, 2, 3, 5, 7, 12, 14 |
| Product / marketing copy | landing page, feature announcement, in-app copy, changelog | Orwell 1–6 + STE rules 1, 3, 5, 7; allow one fresh, specific image per section |
| Creative | fiction, poetry, memoir, script, lyrics, personal essay | Orwell 1–6 as aids only; see "Creative Writing" below |

If the request mixes genres, classify per section, not per document.

## Engineering Communication

- **Commit messages and PR descriptions:** subject line is imperative, present tense, and states the change, not the activity. Body states why the change is needed and what a reviewer should look at. STE sentence limits apply to the body.
- **Incident reports:** the passive voice is allowed in the timeline where naming the actor would assign blame the report should not assign. Everywhere else, name the actor. Never use a euphemism for an outage.
- **Code comments:** say why, not what. Delete any comment that restates the line below it.
- **Error messages:** state what failed, why, and the next action, in that order. No apology, no exclamation, no blame on the user.

## Product and Marketing Copy

Marketing copy may use one fresh, concrete image per section. It must be specific and earned.

- **Bad (dead):** Our platform is a game changer that unlocks unprecedented value.
- **Bad (vague):** Our platform is very powerful and easy to use.
- **Good (specific):** Set up a new environment in four commands. Most teams ship their first deploy the same afternoon.

Never make a claim the input does not support. Do not invent metrics, customer counts, or testimonials.

## Creative Writing

For fiction, poetry, memoir, scripts, and lyrical prose, treat Orwell's rules as clarity aids and ignore the STE baseline unless the user explicitly asks for STE.

### Preserve

- Intentional ambiguity, ellipsis, and withheld information.
- Cadence, meter, rhyme, and sentence-length variation, including long sentences and one-word fragments.
- Dialogue that reflects how a character actually speaks, including slang, dialect, jargon, verbosity, and cliché when the cliché characterizes the speaker.
- Repetition used as a device (anaphora, refrain, motif).
- Fresh, specific figurative language and extended metaphor.
- Second- and third-person interiority, present tense, and other deliberate formal choices.
- The narrator's voice, even when it is pompous, evasive, or unreliable by design.

### Still remove

- Figurative language the writer clearly inherited rather than chose.
- Inflated diction used to sound serious rather than to mean something.
- Adverbs and adjectives that add no image: "she said angrily" when the line is already angry.
- Filter words that put distance between reader and scene: "she felt that", "he noticed that", "it seemed as though" — unless the distance is the point.
- Abstraction that stands where a concrete detail belongs.
- Passive voice that is accidental rather than chosen.

### Conflict handling

If the user asks for strict STE on creative text, apply it, then add one notes bullet naming the specific effects the request destroyed. Do not refuse the request and do not apply it silently.

- **Example note:** Applied strict STE as requested. This flattened the three-clause opening sentence into three short sentences and removed the deliberate breathlessness.
