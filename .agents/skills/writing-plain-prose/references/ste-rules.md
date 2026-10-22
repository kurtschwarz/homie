# STE Baseline

An STE-style baseline derived from ASD-STE100 Simplified Technical English. Apply all fourteen rules to technical, instructional, business, product, and marketing prose.

**Scope limit:** ASD-STE100 has a licensed controlled dictionary that is not bundled here. This file reproduces the writing rules and an approximate word discipline only. Never state that output "conforms to ASD-STE100" or "is valid STE". Say "STE baseline" or "STE-style".

## 1. Keep sentences short

- Instructions: 20 words maximum.
- Descriptive and explanatory text: 25 words maximum.
- One main action or one statement per sentence.

When a sentence exceeds the limit, split it at the conjunction or convert the list into a bulleted list.

- **Bad:** If the deploy fails and the health check does not pass within five minutes, roll back to the previous release and notify the on-call engineer in the incident channel, then open a ticket. (39 words)
- **Good:** If the deploy fails, wait five minutes for the health check. If the check still fails, do these steps:
  1. Roll back to the previous release.
  2. Notify the on-call engineer in the incident channel.
  3. Open a ticket.

## 2. Use a clear subject and an active verb

Put the actor first. Use one verb per clause. Do not stack auxiliary verbs.

- **Bad:** It will have been retried by the consumer.
- **Good:** The consumer retries the message.

## 3. Use the same term for the same thing

Pick one term per concept and repeat it in every sentence. Never vary a term for style.

- **Bad:** The job is queued. The task then runs. When the unit of work finishes, the record updates.
- **Good:** The job is queued. The job then runs. When the job finishes, the job updates the record.

Same rule for spelling, capitalization, and hyphenation of a term: `webhook` everywhere, not `web hook` and `Webhook`.

## 4. Use one verb form for one meaning

Do not use a noun as a verb, or a verb as a noun, when a plain form exists.

- **Bad:** Action the request. Do a rebuild of the index.
- **Good:** Process the request. Rebuild the index.

## 5. Use familiar words with one precise meaning

Avoid idioms, slang, figurative language, humor that depends on culture, and vague verbs (`handle`, `deal with`, `support`, `manage`, `leverage`, `drive`) when a specific verb exists.

- **Bad:** The service handles auth.
- **Good:** The service validates the JWT and rejects expired tokens.

## 6. Use a specific technical term when accuracy requires it

Use the exact term. On first use, define it in one clause or link the definition, unless the stated audience already knows it. Do not invent a plain-English substitute for a defined term.

## 7. Keep noun groups to three words or fewer

Break longer noun stacks with prepositions.

- **Bad:** production database connection pool exhaustion alert threshold
- **Good:** the alert threshold for connection pool exhaustion in the production database

## 8. Write procedures as direct instructions

Every step states the condition, the action, and the expected result. Use the imperative. One action per step.

- **Bad:** The migration should then be run and you will see the output.
- **Good:** Run `npm run migrate`. The command prints `migrations applied: 3`.

## 9. Prefer positive instructions

State what the reader must do. Use a negative only for a prohibition, and make the prohibition explicit.

- **Bad:** Do not proceed unless the backup is not stale.
- **Good:** Confirm the backup is less than 24 hours old, then continue.
- **Good (prohibition):** Never run this command against production.

## 10. Put conditions before actions

The reader must know whether the step applies before reading the action.

- **Bad:** Rotate the key if the audit log shows a failed access.
- **Good:** If the audit log shows a failed access, rotate the key.

## 11. Use articles and full sentences

Do not write telegraphic text in prose. Keep `the` and `a`. Exception: table cells, list labels, and UI strings may omit articles.

- **Bad:** Set flag, restart service.
- **Good:** Set the flag, then restart the service.

## 12. Write one topic per paragraph

Six sentences maximum per paragraph. Start the paragraph with its topic sentence.

## 13. Use consistent American English spelling

Use American spelling and `MM/DD` avoidance: write dates as `2026-07-30` or `July 30, 2026`. If the project style guide requires another English variety, follow the style guide and record the override in the notes.

## 14. Use consistent, unambiguous references

Replace `it`, `this`, `that`, and `they` with the noun when more than one candidate antecedent exists in the previous sentence.

- **Bad:** The worker reads from the queue and writes to the table. If it fails, retry.
- **Good:** The worker reads from the queue and writes to the table. If the write fails, retry the write.

## When strict STE is not possible

If a required term, a legal phrase, or a domain convention breaks a rule above, keep it. Record the exception in the notes list with its reason. Do not silently degrade accuracy to satisfy a rule.
