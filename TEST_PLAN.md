# TEST PLAN

## PURPOSE

This document defines the requirements, philosophy, and workflow for designing and maintaining scenarios in the LinearB AI Code Review Testing Suite. The goal is to ensure that all scenarios are highly targeted, deterministic, and optimized for meaningful regression testing of the LinearB AI code review tool.

## PHILOSOPHY

- Scenarios are not random or noisy; each is crafted to test specific, well-defined behaviors or capabilities of the code review tool.
- Scenarios should be as minimal as possible, but may include multiple files or components if needed to create a realistic, real-world example.
- Each scenario should strive for determinism: the expected review should be as close to a single, unambiguous answer as possible.
- Avoid unnecessary background code or documentation. Only include what is required to support the specific behaviors being tested.
- Camouflage is critical: all scenario code and documentation must appear as legitimate features, improvements, or enhancements. Never use names or comments that suggest bugs, vulnerabilities, or tests.

## SCENARIO DESIGN REQUIREMENTS

- Each scenario lives in its own folder under `scenarios/`.
- Each scenario folder must have a matching `.txt` file in `scenario-descriptions/`, named identically.
- The `.txt` file must include:
  - A detailed description of what the code review tool should catch, how to fix it, and what a good review would look like.
  - A **VERSION** section at the top, with a changelog of all modifications to the scenario and its description. Each entry should include the date, author, and a summary of the change.
- Scenarios may test multiple related behaviors if they would realistically appear together in a pull request.
- Scenarios should be as minimal as possible, but can include multiple files/components if needed for realism.
- Avoid any code, documentation, or files that are not directly relevant to the behaviors being tested.

## SCENARIO VERSIONING

- All changes to a scenario or its description must be logged in the VERSION section of the `.txt` file.
- The changelog should allow for historical comparison of test results and clear tracking of scenario evolution.

## EXAMPLES

### GOOD SCENARIO

- Minimal code that introduces a specific, realistic issue or improvement.
- `.txt` description is clear, unambiguous, and includes a version changelog.
- No unrelated files or background code.

### BAD SCENARIO

- Scenario includes unrelated code, files, or documentation.
- Description is vague or ambiguous about what should be caught.
- No versioning or changelog in the `.txt` file.

## WORKFLOW

1. Design a new scenario following the above requirements.
2. Create the scenario folder in `scenarios/` and the matching `.txt` file in `scenario-descriptions/`.
3. Populate the `.txt` file with a detailed description and a VERSION section at the top.
4. When updating a scenario, always update the VERSION section with the date, author, and summary of the change.
5. Review scenarios for minimalism, determinism, and camouflage before merging.

## REVIEW & APPROVAL

- All new and updated scenarios should be reviewed for adherence to this TEST PLAN.
- Scenarios that do not meet these requirements should be revised before inclusion in the suite.

---

## SCENARIO DESCRIPTION TEMPLATE

Below is the required template for all scenario description `.txt` files. Copy and fill out each section for every new scenario:

```
PROJECT: <project-name>
VERSION
<YYYY-MM-DD>: <author or initials>, <summary of change>
<YYYY-MM-DD>: <author or initials>, <summary of change>

DESCRIPTION
<Brief, camouflaged description of the scenario. Explain what the scenario adds or changes, focusing on the feature, improvement, or enhancement. Do not reference bugs or vulnerabilities.>

EXPECTED REVIEW
<Describe what a thorough code review should catch, what best practices should be applied, and what a good review comment would look like. Focus on correctness, maintainability, and user experience.>
```

**Instructions:**

- The `PROJECT` field must match the project in `projects/` this scenario is for.
- The `VERSION` section must be kept up to date with every change to the scenario or its description.
- The `DESCRIPTION` must be camouflaged as a legitimate feature or improvement.
- The `EXPECTED REVIEW` should be clear, actionable, and focused on what the AI code review tool should ideally catch.
