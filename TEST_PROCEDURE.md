# TESTING PROCEDURE

This document provides a detailed, step-by-step guide for product engineers on how to use the LinearB AI Code Review Testing Suite. Follow these procedures to deploy projects, test scenarios, review results, and extend the suite as needed.

---

## 1. PREPARING YOUR ENVIRONMENT

1.1. Ensure you have `git` and `bash` installed on your machine.

1.2. Clone the repository:

```bash
git clone https://github.com/linearzig/eval-ai-code-review.git
cd eval-ai-code-review
```

1.3. Make sure you are on the `main` branch before running any commands:

```bash
git checkout main
```

---

## 2. DEPLOYING A BASE PROJECT

2.1. List available base projects:

```bash
./run_tests.sh --list-projects
```

2.2. Deploy a base project to a new branch (replace `<project-name>`):

```bash
./run_tests.sh --deploy-project <project-name>
```

- This creates and pushes a new branch with the base project files.
- The branch name will be `project-<project-name>`.

---

## 3. DEPLOYING A SCENARIO ON TOP OF A PROJECT

3.1. List available scenarios:

```bash
./run_tests.sh --list-scenarios
```

3.2. Deploy a scenario to a new branch on top of a project (replace `<project-name>` and `<scenario-name>`):

```bash
./run_tests.sh --deploy-scenario <project-name> <scenario-name>
```

- This creates and pushes a new branch with both the project and scenario files applied.
- The branch name will include both names and a random moniker (e.g., `hello-world-task-removal-enhancement-xyz789`).

---

## 4. REVIEWING RESULTS (STUB)

4.1. After deploying a scenario, the branch is ready for code review.

4.2. **[STUB]** The process for triggering and collecting the AI code review results will depend on your LinearB deployment:

- If the AI code review tool runs automatically on new branches/PRs, check the review output in your code review platform (e.g., GitHub, GitLab).
- If you need to trigger the review manually or via a CI/CD pipeline, follow your internal process.
- **Future integration:** The script can be extended to fetch and evaluate AI review output automatically. (See Section 8)

---

## 5. ADDING A NEW SCENARIO

5.1. Create a new folder in `scenarios/` with a camouflaged, feature-focused name.

5.2. Add the scenario code and documentation, following camouflage requirements (see `TEST_PLAN.md`).

5.3. Create a matching `.txt` file in `scenario-descriptions/` using the template in `TEST_PLAN.md`.

- Fill out the `PROJECT`, `VERSION`, `DESCRIPTION`, and `EXPECTED REVIEW` sections.

5.4. Commit and push your changes to a new branch for review.

---

## 6. SWAPPING OR ADDING A NEW BASE PROJECT

6.1. Add a new project folder to `projects/` with a realistic, feature-focused name.

6.2. Add all necessary files for the base project.

6.3. Use the deployment steps in Section 2 to deploy the new project to a branch.

---

## 7. CLEANING UP BRANCHES

7.1. To delete a branch (both locally and remotely):

```bash
./run_tests.sh --delete-branch <branch-name>
```

---

## 8. EXTENDING THE SUITE & AI REVIEW INTEGRATION (STUB)

8.1. The current script stubs the AI review and scoring process. To fully automate evaluation:

- Integrate with your AI code review tool's API or output.
- Add logic to the script to fetch, parse, and compare AI review results to the `EXPECTED REVIEW` in the scenario description.
- See `TEST_PLAN.md` for guidance on what a good review should catch.

---

## 9. FURTHER REFERENCE

- See `README.md` for a quick overview and script usage.
- See `TEST_PLAN.md` for scenario design philosophy, formatting, and the scenario description template.
- For questions or improvements, contact the engineering team or open a pull request.
