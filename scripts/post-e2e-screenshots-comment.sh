#!/usr/bin/env bash
# Post Playwright screenshots on a pull request (inline via gh --attach).
# Requires: gh >= 2.99, GH_TOKEN, PR_NUMBER, GITHUB_REPOSITORY, GITHUB_RUN_ID
set -euo pipefail

SCREENSHOT_ROOT="${SCREENSHOT_ROOT:-frontend/test-results}"
MARKER="<!-- procpane-e2e-screenshots -->"

if [[ -z "${PR_NUMBER:-}" ]]; then
  echo "PR_NUMBER unset; skipping PR screenshot comment."
  exit 0
fi

if ! gh pr comment --help 2>&1 | grep -qE '\-\-attach'; then
  echo "gh with --attach support is required (>= 2.99); found: $(gh --version | head -1)"
  exit 1
fi

body_file="$(mktemp)"
attach_args=()
attached_paths=()

append_shot() {
  local file="$1"
  local title="$2"
  local rel="${file#./}"
  if [[ ! -f "$file" ]]; then
    return 0
  fi
  for a in "${attached_paths[@]:-}"; do
    if [[ "$a" == "$file" ]]; then
      return 0
    fi
  done
  {
    echo "### ${title}"
    echo ""
    echo "![${title}](${rel})"
    echo ""
  } >>"$body_file"
  attach_args+=(--attach "$file")
  attached_paths+=("$file")
}

{
  echo "$MARKER"
  echo "## Playwright E2E 截图"
  echo ""
  echo "Workflow run: [\`${GITHUB_RUN_ID}\`](${GITHUB_SERVER_URL}/${GITHUB_REPOSITORY}/actions/runs/${GITHUB_RUN_ID})"
  echo ""
  echo "完整 HTML 报告与 trace 仍可在 Actions **Artifacts** 中下载（\`playwright-report\` / \`playwright-test-results\`）。"
  echo ""
} >"$body_file"

append_shot "$SCREENSHOT_ROOT/procpane-smoke.png" "列表概览 (smoke)"
append_shot "$SCREENSHOT_ROOT/procpane-expanded.png" "展开卡片 (detail)"

# Playwright failure screenshots (only-on-failure) and any other PNGs under test-results.
if [[ -d "$SCREENSHOT_ROOT" ]]; then
  mapfile -t extra_pngs < <(
    find "$SCREENSHOT_ROOT" -type f -name '*.png' | sort
  )
  failure_any=false
  for f in "${extra_pngs[@]:-}"; do
    case "$f" in
      "$SCREENSHOT_ROOT/procpane-smoke.png"|"$SCREENSHOT_ROOT/procpane-expanded.png") continue ;;
    esac
    failure_any=true
    base="$(basename "$f")"
    append_shot "$f" "失败截图: ${base}"
  done
  if [[ "$failure_any" == true ]]; then
    echo "_（含 Playwright 自动捕获的失败截图）_" >>"$body_file"
    echo "" >>"$body_file"
  fi
fi

if [[ ${#attach_args[@]} -eq 0 ]]; then
  {
    echo "_本次运行未找到可附带的 PNG 截图；请查看 Artifacts 或 workflow 日志。_"
    echo ""
  } >>"$body_file"
fi

# Replace prior bot comment with the same marker (one sticky comment per PR).
existing_id="$(
  gh api "repos/${GITHUB_REPOSITORY}/issues/${PR_NUMBER}/comments" \
    --paginate \
    --jq ".[] | select(.body | contains(\"${MARKER}\")) | .id" \
    | head -1 \
    || true
)"

if [[ -n "$existing_id" ]]; then
  gh api -X DELETE "repos/${GITHUB_REPOSITORY}/issues/comments/${existing_id}" >/dev/null
fi

gh pr comment "$PR_NUMBER" --repo "$GITHUB_REPOSITORY" --body-file "$body_file" "${attach_args[@]}"
echo "Posted E2E screenshot comment on PR #${PR_NUMBER} (${#attach_args[@]} attachment(s))."
