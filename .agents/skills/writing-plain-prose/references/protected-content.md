# Protected Content

Spans in these categories must appear in the output byte-for-byte identical to the input. Never simplify, shorten, reword, reformat, or "fix" them. If one of them is genuinely wrong, leave it unchanged and flag it in the notes list.

## Always Protected

1. **Code** — any fenced code block, indented code block, or inline `code span`.
2. **Commands and shell snippets** — including flags, paths, and environment variable names.
3. **Identifiers** — variable, function, class, module, package, repo, branch, table, column, and file names.
4. **API surface** — endpoint paths, HTTP methods, header names, JSON field names, query parameters, enum values.
5. **Error and log strings** — anything a reader will grep for. `ECONNRESET`, `429 Too Many Requests`, `panic: nil map`.
6. **Configuration keys and values** — YAML/JSON/TOML/env keys, Helm values paths, feature flag names.
7. **Product, brand, and company names** — including their exact capitalization: `Google`, `CircleCI`, `Datadog`, `PostgreSQL`, `Node.js`.
8. **Version numbers, SHAs, IDs, and URLs** — including ticket IDs such as `#90210`.
9. **Legal, contractual, licensing, privacy, and compliance text** — terms of service, disclaimers, license headers, HIPAA/SOC 2 language, regulatory notices.
10. **Direct quotations and citations** — anything inside quotation marks attributed to a person or document, plus the attribution.
11. **Standardized or regulated terminology** — clinical, pharmaceutical, financial, and safety terms whose meaning is fixed by a standard or regulator.
12. **UI strings being documented** — the literal label a user sees: "Click **Save and continue**."
13. **User-supplied text the user asked you to keep** — any span the user marked as fixed, verbatim, or do-not-touch.
14. **Frontmatter values and metadata** — YAML keys, slugs, canonical paths, tags.

## Qualifiers That Must Not Be Cut

Rule 3 (cut every needless word) does not apply to words that carry meaning of these kinds. Cutting them changes a claim into a stronger, false claim.

- **Scope and limit:** `only`, `at most`, `up to`, `per request`, `in the `us-east-1` region`
- **Frequency:** `usually`, `rarely`, `in most cases`, `intermittently`
- **Statistical:** `p99`, `median`, `on average`, `approximately` when it modifies a measurement
- **Conditional:** `if the flag is enabled`, `after the first retry`, `unless overridden`
- **Safety and medical:** `do not`, `never`, `consult a clinician`, `contraindicated`
- **Legal hedges:** `may`, `is not intended to`, `without warranty`, `where permitted`
- **Attribution of certainty:** `we believe`, `reported`, `alleged`, `estimated`

Do / Don't:

- **Don't:** "Approximately 40 ms" → "40 ms". This turns an estimate into a measurement.
- **Do:** "approximately 40 ms" → "about 40 ms". Same meaning, shorter word.
- **Don't:** "The drug may cause drowsiness" → "The drug causes drowsiness".
- **Don't:** "Most requests complete in under 200 ms" → "Requests complete in under 200 ms".

## Handling a Protected Span That Looks Wrong

Leave it exactly as written. Add one notes bullet in this form:

`Left "<span>" unchanged (protected: <category>). Possible issue: <one line>.`
