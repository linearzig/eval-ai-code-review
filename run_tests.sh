#!/bin/bash

# LinearB AI Code Review Testing Suite Automation Script
# Inspired by benchmarking/run.sh
#
# IMPORTANT:
#   - You must set WORKSPACE_DIR in this script to a directory outside this repo (default: $WORKSPACE_DIR)
#   - You must set your GITHUB_TOKEN environment variable for API access (required for opening PRs and evaluation)

set -e  # Exit on any error

# Set the working directory for all git operations (must be outside this repo)

# !!!
# CHANGE THESE TO YOUR OWN VALUES
# !!!
WORKSPACE_DIR="/Users/zig/azigler/ai-code-review/eval-workspace"
REPO_OWNER="linearzig"
REPO_NAME="eval-ai-code-review"



PROJECTS_DIR="projects"
SCENARIOS_DIR="scenarios"
DESCRIPTIONS_DIR="scenario-descriptions"
GITHUB_API="https://api.github.com"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print helpers
print_status() { echo -e "${BLUE}[INFO]${NC} $1"; }
print_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
print_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Usage/help message
show_help() {
  echo "LinearB AI Code Review Testing Suite Automation Script"
  echo ""
  echo "Usage: $0 [OPTIONS]"
  echo ""
  echo "Options:"
  echo "  --deploy-project <project>         Deploy a base project to a new branch (branch name: project-<project-name>)"
  echo "  --deploy-scenario <project> <scenario>  Deploy a scenario to a new branch on top of a project (branch name: <project-name>-<scenario-name>-<moniker>)"
  echo "  --delete-branch <branch>            Delete a branch (cleanup)"
  echo "  --list-projects                     List all available base projects"
  echo "  --list-scenarios                    List all available scenarios"
  echo "  --help                              Show this help message"
  echo ""
  echo "IMPORTANT:"
  echo "  - You must set WORKSPACE_DIR in this script to a directory outside this repo (default: $WORKSPACE_DIR)"
  echo "  - All git operations and deployments will happen in WORKSPACE_DIR"
  echo "  - You must set your GITHUB_TOKEN environment variable for API access (required for opening PRs and evaluation)"
  echo ""
  echo "Examples:"
  echo "  $0 --deploy-project hello-world"
  echo "  $0 --deploy-scenario hello-world task-removal-enhancement"
  echo "  $0 --delete-branch project-hello-world"
  echo "  $0 --delete-branch hello-world-task-removal-enhancement-xyz789"
  echo ""
}

# Generate a random moniker for branch uniqueness
random_moniker() {
  echo $(cat /dev/urandom | tr -dc 'a-z0-9' | fold -w 6 | head -n 1)
}

# List all projects
list_projects() {
  echo "Available projects:"
  for project in "$PROJECTS_DIR"/*/; do
    echo "  - $(basename "$project")"
  done
}

# List all scenarios
list_scenarios() {
  echo "Available scenarios:"
  for scenario in "$SCENARIOS_DIR"/*/; do
    echo "  - $(basename "$scenario")"
  done
}

open_pr() {
  local branch_name="$1"
  local title="$2"
  local body="$3"
  print_status "Opening PR for branch $branch_name ..."
  local pr_json=$(curl -s -X POST -H "Authorization: token $GITHUB_TOKEN" \
    -d "{\"title\": \"$title\", \"head\": \"$branch_name\", \"base\": \"main\", \"body\": \"$body\"}" \
    "$GITHUB_API/repos/$REPO_OWNER/$REPO_NAME/pulls")
  local pr_url=$(echo "$pr_json" | grep '"html_url":' | head -1 | sed -E 's/.*"html_url": "([^"]+)".*/\1/')
  if [[ "$pr_url" == https://* ]]; then
    print_success "Opened PR: $pr_url"
    echo "$pr_url"
  else
    print_error "Failed to open PR. Response: $pr_json"
    return 1
  fi
}

# Deploy a base project to a new branch
deploy_project() {
  local project_name="$1"
  local branch_name="project-${project_name}"
  local project_path="$PROJECTS_DIR/$project_name"

  if [ ! -d "$project_path" ]; then
    print_error "Project folder not found: $project_path"
    exit 1
  fi

  print_status "Preparing workspace at $WORKSPACE_DIR"
  mkdir -p "$WORKSPACE_DIR"
  if [ ! -d "$WORKSPACE_DIR/.git" ]; then
    print_status "Workspace is not a git repo. Cloning fresh copy..."
    git clone https://github.com/linearzig/eval-ai-code-review "$WORKSPACE_DIR"
  fi
  cd "$WORKSPACE_DIR"

  git checkout main
  git pull origin main
  git checkout -b "$branch_name"
  rsync -av --exclude='.DS_Store' --exclude='.git' "$OLDPWD/$project_path/" ./
  git add .
  git commit -m "feat: deploy base project $project_name"
  git push -u origin "$branch_name"
  print_success "Deployed $project_name to branch $branch_name"

  # Open PR
  local pr_url=$(open_pr "$branch_name" "Deploy base project $project_name" "Automated PR for $branch_name")
  if [ -n "$pr_url" ]; then
    # Evaluate PR
    print_status "Running eval_test.sh for $pr_url ..."
    "$OLDPWD/eval_test.sh" "$pr_url"
  fi
  cd - > /dev/null
}

# Deploy a scenario to a new branch on top of a project
deploy_scenario() {
  local project_name="$1"
  local scenario_name="$2"
  local moniker=$(random_moniker)
  local branch_name="${project_name}-${scenario_name}-${moniker}"
  local project_path="$PROJECTS_DIR/$project_name"
  local scenario_path="$SCENARIOS_DIR/$scenario_name"
  local description_path="$DESCRIPTIONS_DIR/$scenario_name.txt"

  if [ ! -d "$project_path" ]; then
    print_error "Project folder not found: $project_path"
    exit 1
  fi
  if [ ! -d "$scenario_path" ]; then
    print_error "Scenario folder not found: $scenario_path"
    exit 1
  fi
  if [ ! -f "$description_path" ]; then
    print_error "Description file not found: $description_path"
    exit 1
  fi

  print_status "Preparing workspace at $WORKSPACE_DIR"
  mkdir -p "$WORKSPACE_DIR"
  if [ ! -d "$WORKSPACE_DIR/.git" ]; then
    print_status "Workspace is not a git repo. Cloning fresh copy..."
    git clone https://github.com/linearzig/eval-ai-code-review "$WORKSPACE_DIR"
  fi
  cd "$WORKSPACE_DIR"

  git checkout main
  git pull origin main
  git checkout -b "$branch_name"
  rsync -av --exclude='.DS_Store' --exclude='.git' "$OLDPWD/$project_path/" ./
  rsync -av --exclude='.DS_Store' --exclude='.git' "$OLDPWD/$scenario_path/" ./
  git add .
  git commit -m "feat: apply scenario $scenario_name to $project_name"
  git push -u origin "$branch_name"
  print_success "Deployed scenario $scenario_name on $project_name to branch $branch_name"

  # Open PR
  local pr_url=$(open_pr "$branch_name" "Apply scenario $scenario_name to $project_name" "Automated PR for $branch_name")
  if [ -n "$pr_url" ]; then
    # Evaluate PR
    print_status "Running eval_test.sh for $pr_url ..."
    "$OLDPWD/eval_test.sh" "$pr_url"
  fi
  cd - > /dev/null
}

# Delete a branch (cleanup)
delete_branch() {
  cd "$WORKSPACE_DIR"
  local branch_name="$1"
  git branch -D "$branch_name" || print_warning "Branch $branch_name not found locally."
  git push origin --delete "$branch_name" || print_warning "Branch $branch_name not found on remote."
  print_success "Deleted branch $branch_name (local and remote if existed)"
  cd - > /dev/null
}

# Main script logic
main() {
  if [[ $# -eq 0 ]]; then
    show_help
    exit 0
  fi

  while [[ $# -gt 0 ]]; do
    case $1 in
      --deploy-project)
        deploy_project "$2"
        exit 0
        ;;
      --deploy-scenario)
        deploy_scenario "$2" "$3"
        exit 0
        ;;
      --delete-branch)
        delete_branch "$2"
        exit 0
        ;;
      --list-projects)
        list_projects
        exit 0
        ;;
      --list-scenarios)
        list_scenarios
        exit 0
        ;;
      --help)
        show_help
        exit 0
        ;;
      *)
        print_error "Unknown option: $1"
        show_help
        exit 1
        ;;
    esac
  done
}

main "$@" 