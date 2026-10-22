# Revision Checklist

Run these checks in order against the draft. Each check names what to scan for and what to do. A check that finds nothing passes silently.

## Pass 1 — Protected spans

- [ ] Every span matching a category in `protected-content.md` is marked and excluded from all later passes.
- [ ] Every meaning-carrying qualifier listed in `protected-content.md` is marked and excluded from Pass 4.

## Pass 2 — Stock phrases and dead metaphors (Orwell rule 1)

- [ ] Scan for every phrase in the dead-figures list in `orwell-rules.md` rule 1. Replace each with plain phrasing or a specific image.
- [ ] Scan for mixed metaphors. If you cannot picture the image, delete the clause.
- [ ] Scan for these openers and delete them: `In today's fast-paced world`, `Let's dive in`, `Here's the thing`, `It's no secret that`, `Whether you're X or Y`, `Imagine a world where`.
- [ ] Scan for these AI-tell constructions and rewrite: `not only … but also`, `it's not just X, it's Y`, `serves as a testament to`, `plays a crucial role in`, `stands as`, `in the realm of`, `navigate the complexities of`, `unlock the potential of`, `a game-changer for`, `at its core`, `the key takeaway is`.

## Pass 3 — Long words (Orwell rule 2)

- [ ] Replace every word in the substitution table in `orwell-rules.md` rule 2.
- [ ] For any remaining word of four or more syllables, confirm no shorter word carries the same meaning. If one does, use it.
- [ ] Confirm no substitution lost precision. Revert any that did, per rule 6.

## Pass 4 — Needless words (Orwell rule 3)

- [ ] Delete every filler word in the rule-3 list.
- [ ] Collapse every empty frame in the rule-3 table.
- [ ] Delete any sentence that only restates the previous sentence or the heading above it.
- [ ] Delete any closing paragraph that only summarizes what the reader just read, unless the user asked for a summary.
- [ ] Confirm no qualifier marked in Pass 1 was deleted.

## Pass 5 — Voice (Orwell rule 4)

- [ ] Find every passive construction (`be` + past participle). For each, name the actor and rewrite as active, or record it as a justified passive per rule 4.
- [ ] Find every disguised passive: `it was determined`, `a decision was made`, `changes were made`, `it is recommended`, `it is suggested`, `is being tracked`. Name the actor or use the imperative.
- [ ] Find every nominalization and restore the verb: `the implementation of X` → `implementing X`; `provides validation of` → `validates`; `the utilization of` → `using`.

## Pass 6 — Jargon and foreign phrases (Orwell rule 5)

- [ ] Replace every entry in the substitution table in `orwell-rules.md` rule 5.
- [ ] For every remaining technical term, decide: does the stated audience know it? If not, define it in one clause on first use, or link its definition.
- [ ] Expand every acronym on first use, except acronyms the stated audience uses daily.

## Pass 7 — STE baseline

- [ ] Sentence length: no instruction over 20 words, no other sentence over 25 words.
- [ ] Term consistency: build a list of every technical term used. Confirm each concept has exactly one term, one spelling, and one capitalization throughout.
- [ ] Noun groups: no stack of more than three nouns.
- [ ] Procedures: every step is imperative, has one action, and states its expected result.
- [ ] Conditions: every conditional clause comes before its action.
- [ ] Paragraphs: one topic each, six sentences maximum, topic sentence first.
- [ ] References: no `it`, `this`, `that`, or `they` with more than one possible antecedent.
- [ ] Positive form: every instruction states what to do, except explicit prohibitions.

## Pass 8 — Honesty

- [ ] Every claim in the text is supported by the input. Delete or flag any claim you introduced.
- [ ] No number, name, date, version, or citation was invented or changed.
- [ ] No hedge was strengthened into a certainty, and no certainty was weakened into a hedge.
- [ ] No euphemism hides an unpleasant fact: `an issue occurred` → say what broke; `we sunset the feature` → `we removed the feature`; `resources were reallocated` → say what actually happened.

## Pass 9 — Rule 6 restoration

- [ ] Read the full text aloud. Restore any phrasing that is now awkward, ungrammatical, ambiguous, robotic, clipped, or condescending.
- [ ] Confirm the tone still matches the tone the user asked for.
- [ ] Confirm the text is not shorter than the meaning requires.

## Pass 10 — Constraints

- [ ] Format matches what the user asked for (markdown, plain text, bullet count, heading levels).
- [ ] Length is within any stated limit.
- [ ] Every protected span from Pass 1 appears unchanged.
- [ ] The notes list records every rule-6 override, kept exception, and constraint conflict.
