---
name: creating-git-commits
description: Creates a cryptographically signed git commit with a Conventional Commits message that prioritizes reasoning over description, including a Co-Authored-By trailer naming the current Claude model.
---

# creating-git-commits

Creates a signed git commit whose message follows Conventional Commits, carries a body only when the diff is not self-explanatory, and ends with a Co-Authored-By trailer for the current Claude model. Use whenever committing staged or unstaged changes.

## Prerequisites / Inputs

- Changed files in the working tree (staged or unstaged). If there are none, tell the user there is nothing to commit and stop.
- A configured commit signing key. If signing is not configured, `git commit -S` fails — surface the error rather than committing unsigned.

## Process Steps

1. Run `git status` and `git diff` (and `git diff --staged`) to review what changed and understand the scope.
2. Decide which files to stage. If all changes relate to one logical change, stage them. If there are unrelated or ambiguous changes, ask the user what to include.
3. Stage the chosen files with `git add`.
4. Load `references/conventions.md`. Choose the Conventional Commits type, optional scope, and write the subject following the "Subject Format" section.
5. Decide whether a body is needed using the "When to Include a Body" section of `references/conventions.md`. Default to no body when the diff explains itself.
6. Draft the subject, and the body if step 5 called for one, with the `writing-plain-prose` skill. Keep it terse.
7. Load `references/examples.md`. Compare your draft subject and body against the good/bad pairs and revise to match.
8. Determine your current model's human-friendly name for the trailer (for example `Opus 4.8`, `Sonnet 5`).
9. Create the signed commit with `git commit -S`, passing the subject, the body (if any), and the Co-Authored-By trailer as separate `-m` arguments per the "Co-Authored-By Trailer" section of `references/conventions.md`.
10. If the commit fails to sign, stop and report the error to the user. Do not retry without `-S`.
11. Run `git log -1` to show the created commit and confirm the message and signature are correct.

## Guardrails

- Never create an unsigned commit as a fallback when `-S` fails. Report the signing error and stop.
- Never fabricate a body to fill space. A subject-only commit is the common, correct outcome.
- Never `git add` unrelated changes without asking the user first.

## Output

- One cryptographically signed commit: a Conventional Commits subject, a body only when the diff is not self-explanatory, and a Co-Authored-By trailer naming the current Claude model.
- The `git log -1` output confirming the commit.
