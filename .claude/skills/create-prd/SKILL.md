---
name: create-prd
description: Create a Product Requirements Document (PRD) for a feature or product. Use when the user asks to write a PRD, draft requirements, spec out a feature, or plan a new product/capability before implementation.
---

# Create PRD

Produce a clear, actionable Product Requirements Document (PRD) for the feature or product the user describes.

## Process

1. **Gather context.** Identify the feature, its users, and the problem it solves. If any of the following are unclear and would materially change the PRD, ask the user before writing:
   - The core problem and who has it
   - The target users / personas
   - The primary goal and how success is measured
   - Scope boundaries (what is explicitly out of scope)

   Don't over-ask — fill obvious gaps with reasonable, clearly-labeled assumptions.

2. **Draft the PRD** using the structure below. Keep it concise and skimmable; prefer bullet points and tables over prose. Every requirement should be testable.

3. **Save the document** as a Markdown file. Default to `docs/prd/<kebab-case-feature-name>.md` unless the user specifies another location. Create the directory if needed.

4. **Summarize** what you wrote and call out any open questions or assumptions that need confirmation.

## PRD structure

```markdown
# PRD: <Feature Name>

| Field    | Value                          |
| -------- | ------------------------------ |
| Author   | <author>                       |
| Status   | Draft                          |
| Date     | <YYYY-MM-DD>                   |

## 1. Overview
One paragraph: what this is and why it matters.

## 2. Problem Statement
The problem being solved and who experiences it.

## 3. Goals & Non-Goals
- **Goals:** measurable objectives this delivers.
- **Non-Goals:** explicitly out of scope.

## 4. Users & Personas
Who uses this and what they need.

## 5. User Stories
- As a <user>, I want <action> so that <benefit>.

## 6. Requirements
### Functional
- FR-1: <requirement> (testable, unambiguous)

### Non-Functional
- NFR-1: performance, security, accessibility, etc.

## 7. Success Metrics
How success is measured (KPIs, targets).

## 8. Assumptions & Dependencies
Assumptions made and external dependencies.

## 9. Risks & Open Questions
Known risks and unresolved questions.

## 10. Milestones (optional)
Rough phases or timeline.
```

## Guidelines

- Write requirements that are specific and verifiable — avoid vague terms like "fast" or "user-friendly" without a measurable definition.
- Label every assumption clearly so reviewers can challenge it.
- Keep the document implementation-agnostic: describe *what* and *why*, not *how* (leave technical design to a separate doc unless asked).
- Match the user's domain language and any existing PRD conventions in the repo.
