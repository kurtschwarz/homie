# Examples

## Subject Only — the diff is self-explanatory

Most commits look like this. No body.

```
refactor(copilot): extract tool scoping

Co-Authored-By: Claude (Opus 4.8) <noreply@anthropic.com>
```

Why: a reader who opens the diff sees the extraction. There is nothing to
explain, so there is no body.

## Subject Plus Body — the reasoning is not in the diff

```
fix(orders): dispatch events after commit

Analytics and Slack are best-effort and must never break an order write.
Dispatching inside the tx callback meant a slow or failing sink could
roll back the status change. Move the dispatch to after commit so order
visibility never trades off against write durability.

Co-Authored-By: Claude (Opus 4.8) <noreply@anthropic.com>
```

Why: the diff shows a call moved a few lines down. It cannot show *why* that
move matters. The body carries the reasoning a reviewer would otherwise ask
for.

## Good vs. Bad

### Good

```
fix(copilot): settle patient picker in sync
```

Names the intent and the bug it closes, in 43 characters.

### Bad

```
fix: patient picker changes
```

Says nothing. No scope, no intent.

### Bad

```
feat(copilot): add tools

Added get_soap_note tool. Added get_internal_note tool. Edited the
controller. Edited the shared package. Ran the tests.
```

The body narrates the file list and restates the subject. It carries no
reasoning. Either drop the body or replace it with *why* the tools were added.

## Full Workflow Example

Starting state: you have just finished moving an event dispatch out of a
database transaction.

1. Review the diff to confirm all changes are related.
2. Stage the files with `git add`.
3. Decide the type and scope: `fix(orders)`.
4. Judge whether the diff explains itself — it does not; the reason for the
   move is invisible in the diff, so write a short body.
5. Draft subject and body with the `writing-plain-prose` skill.
6. Commit:
   ```
   git commit -S \
     -m "fix(orders): dispatch events after commit" \
     -m "Analytics and Slack are best-effort and must never break an order write. Dispatching inside the tx callback meant a slow or failing sink could roll back the status change. Move the dispatch to after commit so order visibility never trades off against write durability." \
     -m "Co-Authored-By: Claude (Opus 4.8) <noreply@anthropic.com>"
   ```
