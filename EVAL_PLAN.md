# EVAL PLAN

This document outlines the technical plan for fetching and evaluating LinearB AI code review results using the GitHub API, as part of the local testing suite workflow.

---

## 1. WORKFLOW OVERVIEW

1. A project and scenario are deployed to a new branch using `run_tests.sh`.
2. A pull request (PR) is manually created from the new branch to `main` (or the appropriate base branch).
3. LinearB AI code review runs on the PR and posts its review as a comment.
4. The evaluation script (to be implemented) will:
   - Use the GitHub API to find the PR for the branch.
   - Poll for new comments on the PR, waiting for the LinearB AI review comment to appear.
   - Extract the AI review comment content.
   - Compare the AI review output to the scenario's `.txt` file (EXPECTED REVIEW section).
   - Score the result and write the evaluation to a results file for record keeping.

---

## 2. GITHUB API USAGE

### 2.1. Authentication

- Use a GitHub personal access token (PAT) with `repo` and `read:org` permissions.
- Store the token in an environment variable (e.g., `GITHUB_TOKEN`).

### 2.2. Finding the PR for a Branch

- Use the [List pull requests API](https://docs.github.com/en/rest/pulls/pulls?apiVersion=2022-11-28#list-pull-requests) to find open PRs with `head=<branch-name>`.
- Endpoint: `GET /repos/{owner}/{repo}/pulls?head={owner}:{branch}`
- Parse the response to get the PR number.

### 2.3. Polling for AI Review Comments

- Use the [List issue comments API](https://docs.github.com/en/rest/issues/comments?apiVersion=2022-11-28#list-issue-comments) to fetch comments on the PR (`GET /repos/{owner}/{repo}/issues/{issue_number}/comments`).
- Poll at a regular interval (e.g., every 30 seconds) until a new comment from the LinearB bot appears, or until a timeout (e.g., 10-15 minutes).
- Identify the LinearB AI review comment by author (e.g., `linear-b[bot]`) or by a unique marker in the comment body.

### 2.4. Extracting the AI Review Output

- Once the LinearB comment is found, extract the comment body as the AI review output.
- Save the raw output for record keeping.

---

## 3. EVALUATION & SCORING

### 3.1. Fetching the Scenario's Expected Review

- Read the corresponding scenario description `.txt` file.
- Extract the `EXPECTED REVIEW` section.

### 3.2. Comparing Outputs

- Compare the AI review output to the expected review using string similarity, keyword matching, or a more advanced NLP approach.
- Log both outputs for transparency.

### 3.3. Scoring

- Implement a scoring function (initially stubbed) to rate how well the AI review matches the expected review.
- Output: PASS/FAIL or a numeric score (e.g., 0-100).

### 3.4. Record Keeping

- Write the evaluation results (branch, PR, AI review, expected review, score, timestamp) to a results file (e.g., `results/<branch-name>.json`).

---

## 4. POLLING & TIMEOUTS

- Poll for new comments every 30 seconds (configurable).
- Timeout after 10-15 minutes if no AI review comment is found.
- Print status updates to the console during polling.

---

## 5. IMPLEMENTATION NOTES

- The evaluation script should be implemented in bash.
- The script should be callable as a standalone command (e.g., `./run_tests.sh --eval <branch-name>` or a separate script `eval_test.sh`).
- When multiple scenarios are tested (e.g., after running `run_tests.sh`), the main script can launch multiple `eval_test.sh` processes in the background, one for each branch/PR.
- Each `eval_test.sh` process will monitor its assigned PR, poll for the AI review comment, and write results to a file upon completion or error.
- The main script can enter a **dashboard mode** after launching all evals:
  - Periodically aggregate and display the status of all running `eval_test.sh` processes.
  - Show which branches are still polling, how long they've been running, and which have completed.
  - Display live updates as results come in (e.g., "[DONE] hello-world-task-removal-enhancement-xyz789: PASS (score: 95)").
  - Optionally, allow the user to interrupt/cancel running evals.
- Support for batch evaluation: allow evaluating multiple branches in sequence or in parallel.
- Ensure all API calls handle rate limits and errors gracefully.
- All results should be written to files for record keeping and later analysis.

---

## 6. FUTURE EXTENSIONS

- Integrate with CI/CD to trigger evaluation automatically after PR creation.
- Support for other code review tools or bots.
- Advanced scoring using LLMs or semantic similarity.

---

## 7. REFERENCES

- [GitHub REST API v3 - Pull Requests](https://docs.github.com/en/rest/pulls/pulls?apiVersion=2022-11-28)
- [GitHub REST API v3 - Issue Comments](https://docs.github.com/en/rest/issues/comments?apiVersion=2022-11-28)
