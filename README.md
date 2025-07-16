# LinearB AI Code Review Testing Suite

## OVERVIEW

This repository provides a robust, extensible framework for evaluating the LinearB AI Code Review Tool against a suite of realistic, camouflaged scenarios. It is designed for long-term regression testing and feature evaluation, and is now packaged for handoff to the product team.

- **Main branch** contains only the framework, scripts, and documentation.
- **Projects** and **scenarios** are deployed to their own branches as needed for testing.
- The suite is designed for extensibility: new projects and scenarios can be added, and the workflow supports future integration with AI review and scoring.

## GETTING STARTED

### 1. Set Up Your Workspace

- **Designate a working directory outside this repo** for all git operations and deployments (e.g., `/Users/zig/azigler/ai-code-review/eval-workspace`).
- Edit `run_tests.sh` and set the `WORKSPACE_DIR` variable to this directory.
- Edit `eval_test.sh` and set the `default_owner` and `default_repo` variables to your own values.
- **Set your GitHub personal access token** in your shell, for `eval_test.sh`:

  ```bash
  export GITHUB_TOKEN=your_token_here
  ```

### 2. List All Projects

```bash
./run_tests.sh --list-projects
```

### 3. List All Scenarios

```bash
./run_tests.sh --list-scenarios
```

### 4. Deploy a Base Project to a New Branch

This will create a new branch from `main` and copy the selected project into it. The branch will be named `project-<project-name>`.

```bash
./run_tests.sh --deploy-project <project-name>
```

Example:

```bash
./run_tests.sh --deploy-project hello-world
```

- This creates and pushes a new branch named `project-hello-world` in your designated workspace directory.

### 5. Deploy a Scenario to a New Branch on Top of a Project

This will create a new branch from `main`, copy the selected project and scenario into it, and push the branch. The branch will be named `<project-name>-<scenario-name>-<moniker>` (e.g., `hello-world-task-removal-enhancement-xyz789`).

```bash
./run_tests.sh --deploy-scenario <project-name> <scenario-name>
```

Example:

```bash
./run_tests.sh --deploy-scenario hello-world task-removal-enhancement
```

### 6. Delete a Branch (Cleanup)

This will delete the specified branch both locally and remotely.

```bash
./run_tests.sh --delete-branch <branch-name>
```

### 7. Get Help

Prints a detailed help message with all available options.

```bash
./run_tests.sh --help
```

## IMPORTANT NOTES

- Only the orchestrator directory contains the framework and scripts. Projects and scenarios are deployed to their own branches for testing in the designated workspace directory.
- All scenario and project names must be camouflaged as legitimate features or improvements. Never use names that hint at bugs or vulnerabilities.
- See `TEST_PLAN.md` for scenario design philosophy, requirements, and workflow.
- The script stubs AI review and scoring logic; future integration is supported.
- **You must set your `GITHUB_TOKEN` environment variable for API access.**

## HANDOFF

This framework is now ready for the product team to:

- Select or add base projects in `projects/`
- Write and maintain scenarios in `scenarios/` and `scenario-descriptions/`
- Use the provided script to deploy, test, and clean up branches as needed, using a designated workspace directory outside this repo
- Extend the workflow as new requirements arise

For any questions or to contribute improvements, please refer to the `TEST_PLAN.md` or contact the engineering team.
