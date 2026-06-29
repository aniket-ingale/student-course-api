---
name: create-userstories
description: Break a PRD down into a list of individual, structured user stories. Use when the user asks to split/decompose a PRD into user stories, generate a backlog from requirements, or turn a spec into actionable stories. Output is saved to the .tmp folder.
---

# Breakdown PRD into User Stories

Decompose a Product Requirements Document into a flat list of small,
independently-implementable user stories, each in a consistent structured format.

## Process

1. **Locate the source PRD.** Use the path the user gives. If none is provided,
   look for a PRD in `.tmp/`, `docs/prd/`, or the current file in the IDE, and
   confirm which one before proceeding. Read it fully before decomposing.

2. **Extract the units of work.** Walk the PRD's goals, user stories, and
   functional/non-functional requirements (FR-/NFR- items). Group related
   requirements into stories that are:
   - **Independent** — deliverable on its own where possible.
   - **Small** — one coherent capability; split anything that mixes concerns.
   - **Testable** — has concrete, verifiable acceptance criteria.

   Map each story back to the PRD requirements it covers so nothing is dropped
   and nothing is invented. Cross-cutting NFRs (validation, error contract,
   testing) become acceptance criteria on the relevant stories rather than
   stories of their own, unless they are substantial enough to stand alone.

3. **Write each story** using the structure below. Keep titles imperative and
   specific. Acceptance criteria should be checkable Given/When/Then or bullet
   assertions tied to PRD requirement IDs.

4. **Save the result** as Markdown to
   `.tmp/<kebab-case-prd-name>-stories.md`. Create `.tmp/` if needed.

5. **Summarize**: report the file path, the number of stories, and call out any
   PRD open questions or assumptions that affect the breakdown.

## Story structure

Produce a document with a short header, then one block per story:

```markdown
# User Stories: <PRD Name>

Source: <path to PRD>
Generated: <YYYY-MM-DD>

## US-1: <imperative title>

- **Story:** As a <user>, I want <action> so that <benefit>.
- **Priority:** High | Medium | Low
- **Estimate:** S | M | L
- **Covers:** <PRD requirement IDs, e.g. FR-1, FR-2, NFR-2>
- **Dependencies:** <other US-n, or None>
- **Acceptance Criteria:**
  - [ ] <verifiable condition>
  - [ ] <verifiable condition>
- **Notes:** <assumptions / open questions, or omit>

## US-2: ...
```

## Guidelines

- One capability per story — if a story needs "and" to describe it, consider
  splitting it.
- Every FR/NFR in the PRD must be traceable to at least one story's **Covers**
  field. Don't introduce requirements the PRD doesn't state; surface gaps as
  Notes instead.
- Order stories roughly by dependency / delivery sequence and number them
  `US-1, US-2, ...`.
- Keep acceptance criteria concrete and testable — mirror the PRD's status codes,
  validation rules, and error contracts exactly.
- Preserve the PRD's domain language; don't rename its concepts.
- Carry PRD open questions into the relevant story's **Notes** so they aren't lost.
