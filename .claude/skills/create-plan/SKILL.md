---
name: create-plan
description: Use when you have a spec or requirements for a multi-step task, before touching code.
---
# Writing Plans

## Overview

Write comprehensive implementation plans assuming the engineer has zero context for our codebase. Document everything they need to know: which files to touch for each task, code, testing, docs they might need to check, how to test it. Give them the whole plan as bite-sized tasks. DRY. YAGNI. TDD. Frequent commits.

Assume they are a skilled developer, but know almost nothing about our toolset or problem domain. Assume they don't know good test design very well.

**Announce at start:** "I'm using the create-plan skill to create the implementation plan."

**Context:** This should be run in a dedicated worktree (created by brainstorming skill).

**Save plans to:** `.tmp/plans/YYYY-MM-DD-<feature-name>.md`

## Bite-Sized Task Granularity

**Each step is one action (2-5 minutes):**
- "Write the failing test" - step
- "Run it to make sure it fails" - step
- "Implement the minimal code to make the test pass" - step
- "Run the tests and make sure they pass" - step
- "Commit" - step

## Plan Document Header

**Every plan MUST start with this header:**

```markdown
# [Feature Name] Implementation Plan


**Goal:** [One sentence describing what this builds]

**Architecture:** [2-3 sentences about approach]

**Tech Stack:** [Key technologies/libraries]

---
```

## Task Structure

````markdown
### Task N: [Component Name]

**Files:**
- Create: `exact/path/to/file.go`
- Modify: `exact/path/to/existing.go:123-145`
- Test: `tests/exact/path/to/test.go`

**Step 1: Write the failing test**

```go
func TestSpecificBehavior(t *testing.T) {
    result := Function(input)
    if result != expected {
        t.Errorf("Expected %v, got %v", expected, result)
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test tests/path/test.go -v`
Expected: FAIL with "function not defined"

**Step 3: Write minimal implementation**

```go
func Function(input interface{}) interface{} {
    return expected
}
```

**Step 4: Run test to verify it passes**

Run: `go test tests/path/test.go -v`
Expected: PASS

**Step 5: Commit**

```bash
git add tests/path/test.go src/path/file.go
git commit -m "feat: add specific feature"
```
```
## Remember
- Exact file paths always
- Complete code in plan (not "add validation")
- Exact commands with expected output
- Reference relevant skills with @ syntax
- DRY, YAGNI, TDD, frequent commits


**"Plan complete and saved to `.tmp/plans/<filename>.md`**




