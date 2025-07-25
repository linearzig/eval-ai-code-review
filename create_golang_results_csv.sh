#!/bin/bash

NO_LINEARB=false
for arg in "$@"; do
  if [ "$arg" == "--no-linearb" ]; then
    NO_LINEARB=true
    echo "[INFO] Will append 'no-linearb' to PR titles to skip LinearB reviews."
  fi
done

# Create CSV header
echo "Bug #,Scenario Name,Tool,PR Link,Expected Review,Actual Review" > golang_results.csv

# Function to extract expected review from scenario description
get_expected_review() {
    local scenario=$1
    local desc_file="scenario-descriptions/${scenario}.txt"
    if [ -f "$desc_file" ]; then
        awk '/^EXPECTED REVIEW$/{flag=1;next} /^[A-Z]+$/{flag=0} flag' "$desc_file" | tr '\n' ' ' | sed 's/^[[:space:]]*//;s/[[:space:]]*$//'
    else
        echo "Description file not found"
    fi
}

# Function to extract actual review from JSON result
get_actual_review() {
    local result_file=$1
    if [ -f "$result_file" ]; then
        jq -r '.ai_reviews[0].body // "No review found"' "$result_file" | tr '\n' ' ' | sed 's/^[[:space:]]*//;s/[[:space:]]*$//'
    else
        echo "Result file not found"
    fi
}

# Function to get PR URL from JSON result
get_pr_url() {
    local result_file=$1
    if [ -f "$result_file" ]; then
        jq -r '.pr_url // "No PR URL found"' "$result_file"
    else
        echo "Result file not found"
    fi
}

# Function to get Copilot PR URL by searching for [no-linearb] in PR title
get_copilot_pr_url() {
    local scenario=$1
    # Find the most recent PR with [no-linearb] in the title for this scenario
    gh pr list --state all --json title,url --search "[no-linearb] in:title $scenario" | jq -r '.[0].url // "No Copilot PR found"'
}

# Function to get LinearB PR URL by searching for PRs without [no-linearb] in the title
get_linearb_pr_url() {
    local scenario=$1
    # Find the most recent PR without [no-linearb] in the title for this scenario
    gh pr list --state all --json title,url --search "$scenario" | jq -r '.[] | select(.title | contains("[no-linearb]") | not) | .url' | head -1
}

# Function to properly escape CSV fields
escape_csv_field() {
    local field="$1"
    field=$(echo "$field" | sed 's/"/""/g')
    echo "\"$field\""
}

# Define scenarios in order
scenarios=(
    "zero-value-initialization"
    "interface-nil-check"
    "method-receiver-optimization"
    "slice-capacity-enhancement"
    "channel-buffering-enhancement"
    "goroutine-leak-prevention"
    "context-cancellation-enhancement"
    "map-iteration-enhancement"
    "defer-execution-enhancement"
    "type-embedding-enhancement"
    "channel-select-enhancement"
    "interface-type-assertion-enhancement"
    "memory-pool-enhancement"
    "reflection-safety-enhancement"
    "goroutine-coordination-enhancement"
)

# Process each scenario
for i in "${!scenarios[@]}"; do
    bug_num=$((i + 1))
    scenario=${scenarios[$i]}
    expected_review=$(get_expected_review "$scenario")
    expected_review_escaped=$(escape_csv_field "$expected_review")

    # LinearB row
    result_file=$(ls results/ | grep "golang-weave-${scenario}" | grep -v '\[no-linearb\]' | tail -1)
    if [ -n "$result_file" ]; then
        pr_url=$(get_pr_url "results/$result_file")
        actual_review=$(get_actual_review "results/$result_file")
        pr_url_escaped=$(escape_csv_field "$pr_url")
        actual_review_escaped=$(escape_csv_field "$actual_review")
        echo "$bug_num,$scenario,LinearB,$pr_url_escaped,$expected_review_escaped,$actual_review_escaped" >> golang_results.csv
    else
        # Try to find the PR via gh if no result file
        pr_url=$(get_linearb_pr_url "$scenario")
        pr_url_escaped=$(escape_csv_field "$pr_url")
        echo "$bug_num,$scenario,LinearB,$pr_url_escaped,$expected_review_escaped," >> golang_results.csv
    fi

    # Copilot row
    copilot_pr_url=$(get_copilot_pr_url "$scenario")
    copilot_pr_url_escaped=$(escape_csv_field "$copilot_pr_url")
    echo "$bug_num,$scenario,Copilot,$copilot_pr_url_escaped,$expected_review_escaped," >> golang_results.csv

done

echo "CSV file created: golang_results.csv" 