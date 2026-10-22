# Orwell's Six Rules

From "Politics and the English Language" (1946). These are the primary filters. Each rule below states the rule, then the test to apply, then a worked pair.

## Rule 1 — Never use a metaphor, simile, or other figure of speech which you are used to seeing in print

**Test:** Could you have written this phrase without thinking? Have you seen it in a press release? If yes, it is dead. Replace it with plain phrasing or a specific, concrete image.

Dead figures to cut on sight: `at the end of the day`, `moving forward`, `low-hanging fruit`, `move the needle`, `boil the ocean`, `drink from the firehose`, `single source of truth` (unless it is the project's defined term), `north star`, `game changer`, `paradigm shift`, `bleeding edge`, `heavy lifting`, `deep dive`, `circle back`, `unlock value`, `journey` (for anything other than travel), `battle-tested`, `rock solid`, `blazing fast`, `under the hood`, `first-class citizen`, `source of truth`, `swim lane`, `tip of the iceberg`, `perfect storm`.

- **Bad:** At the end of the day, this migration is low-hanging fruit that moves the needle on latency.
- **Good:** The migration is cheap to do and cuts p99 latency by about 40 ms.

A mixed metaphor is a hard failure. If you cannot picture the image, delete it.

- **Bad:** This unlocks a runway for the team to hit the ground running.
- **Good:** The team can start next sprint without waiting on the schema change.

## Rule 2 — Never use a long word where a short one will do

**Test:** Is there a shorter, more common word with the same meaning? "The same meaning" is strict — do not swap a precise word for a vague short one.

| Replace | With |
|---|---|
| utilize, leverage (verb) | use |
| facilitate | help, cause, run |
| commence, initiate | start |
| terminate | stop, end, kill |
| endeavor | try |
| ascertain | find out, confirm |
| approximately | about |
| in order to | to |
| prior to | before |
| subsequent to | after |
| with regard to, in respect of | about, for |
| in the event that | if |
| due to the fact that | because |
| at this point in time | now |
| a number of | some, several, or the count |
| the majority of | most |
| necessitate | need, require |
| methodology | method |
| functionality | features, behavior, or what it does |
| additional | more, extra |
| individuals | people |
| demonstrate | show |
| sufficient | enough |
| remainder | rest |
| implement (noun sense) | tool |

Do not shorten a word that carries exact meaning. `idempotent`, `authenticate`, `deprecate`, `nullable`, `eventual consistency` stay.

## Rule 3 — If it is possible to cut a word out, always cut it out

**Test:** Delete the word. Re-read. If the meaning is intact, the deletion stands.

Delete these on sight when they add nothing: `very`, `really`, `quite`, `rather`, `basically`, `essentially`, `actually`, `simply`, `just`, `literally`, `truly`, `certainly`, `clearly`, `obviously`, `of course`, `needless to say`, `it should be noted that`, `it is important to note that`, `in terms of`, `the fact that`, `kind of`, `sort of`, `overall`, `generally speaking`.

Collapse these empty frames:

| Cut | Keep |
|---|---|
| It is important to note that X | X |
| There are several reasons why X happens | X happens because … |
| What this means is that X | X |
| In this section we will discuss X | (delete; the heading already says it) |
| X is something that Y | Y X |
| make a decision | decide |
| perform an analysis of | analyze |
| provide assistance to | help |
| is able to | can |
| has the ability to | can |
| gives you the option to | lets you |
| in the process of doing X | doing X |

- **Bad:** It is important to note that the worker is basically able to process a number of jobs in parallel.
- **Good:** The worker processes jobs in parallel.

Do not cut a word that carries a condition, limit, scope, or hedge that is factually required. `usually`, `only`, `at most`, `after the first retry` earn their place.

## Rule 4 — Never use the passive where you can use the active

**Test:** Who did it? If the actor is known and matters, name it and put it first.

- **Bad:** The record is updated when the webhook is received.
- **Good:** The consumer updates the record when it receives the webhook.

Keep the passive in these cases, and say why in the notes:
- The actor is unknown or irrelevant: "The row was deleted at 03:12 UTC."
- The object is the topic of the paragraph and moving it breaks focus.
- Naming the actor assigns blame that the text should not assign.
- The convention of the document requires it (some release notes, incident timelines, legal text).

Watch for the disguised passive: `it was determined that`, `a decision was made to`, `changes were introduced`, `it is recommended that`. Each hides an actor. Name the actor or use an imperative.

- **Bad:** It is recommended that the cache be warmed before the cutover.
- **Good:** Warm the cache before the cutover.

## Rule 5 — Never use a foreign phrase, a scientific word, or a jargon word if you can think of an everyday English equivalent

**Test:** Would a competent new engineer on the team, or the stated audience, know this word without looking it up? If not, and a plain word exists, use the plain word.

| Replace | With |
|---|---|
| inter alia | among other things |
| per se | in itself, or delete |
| vis-à-vis | compared with, about |
| ipso facto | therefore |
| ceteris paribus | all else equal |
| ergo | so |
| a priori | in advance, assumed |
| de facto | in practice |
| ad hoc | one-off, unplanned |
| orthogonal (non-math use) | unrelated, independent |
| non-trivial | hard, large, or the actual size |
| performant | fast, or the measured number |
| synergize | work together, or delete |
| operationalize | put into practice, run |
| ideate | think up, plan |
| socialize (an idea) | share, discuss |
| going forward | (delete) |
| learnings | lessons, findings |
| verbiage | wording, text |

Keep the exact technical term when accuracy needs it. When the term is necessary and the audience may not know it, define it on first use in one clause, or link its definition. Do not define a term the stated audience already knows.

- **Bad:** The handler is idempotent.  (audience: support agents)
- **Good:** The handler is idempotent — running it twice has the same effect as running it once.

## Rule 6 — Break any of these rules sooner than say anything outright barbarous

**Test:** Read the result aloud. If it is awkward, ungrammatical, ambiguous, robotic, insulting, or now says something untrue, revert the change.

Rule 6 outranks rules 1 through 5. Use it when:
- Cutting a word made a sentence ambiguous.
- Forcing the active voice invented an actor the text cannot verify.
- Replacing a term lost precision.
- Shortening produced a clipped, hostile, or condescending tone.
- The plain word collides with a defined product term.

Record every rule-6 override in the notes list, one line each, with the reason.
