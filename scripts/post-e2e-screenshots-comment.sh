#!/usr/bin/env bash
# Publish Playwright PNGs to branch e2e-screenshots and post a PR comment with
# raw.githubusercontent.com URLs (inline on GitHub). GITHUB_TOKEN cannot use
# gh pr comment --attach (ghs_ is rejected); see cli/cli#14180.
# Push failures must not fail CI; fall back to an artifact-pointing PR comment.
set -euo pipefail

SCREENSHOT_ROOT="${SCREENSHOT_ROOT:-frontend/test-results}"
MARKER="<!-- procpane-e2e-screenshots -->"
SCREENSHOTS_BRANCH="e2e-screenshots"
PUSH_RETRIES="${PUSH_RETRIES:-3}"
PUSH_RETRY_SLEEP_SEC="${PUSH_RETRY_SLEEP_SEC:-3}"

if [[ -z "${PR_NUMBER:-}" ]]; then
  echo "PR_NUMBER unset; skipping PR screenshot comment."
  exit 0
fi

if [[ -z "${GH_TOKEN:-}" ]]; then
  echo "GH_TOKEN unset; skipping PR screenshot comment."
  exit 0
fi

body_file="$(mktemp)"
work_dir="$(mktemp -d)"
dest_dir="pr-${PR_NUMBER}"
published=()
push_ok=false
committed=false

cleanup() {
  rm -f "$body_file"
  rm -rf "$work_dir"
}
trap cleanup EXIT

mkdir -p "${work_dir}/repo"
git clone --quiet --depth 1 \
  "https://x-access-token:${GH_TOKEN}@github.com/${GITHUB_REPOSITORY}.git" \
  "${work_dir}/repo"

cd "${work_dir}/repo"
git config user.name "github-actions[bot]"
git config user.email "41898282+github-actions[bot]@users.noreply.github.com"

if git ls-remote --heads origin "${SCREENSHOTS_BRANCH}" | grep -q "${SCREENSHOTS_BRANCH}"; then
  git fetch origin "${SCREENSHOTS_BRANCH}:${SCREENSHOTS_BRANCH}"
  git checkout "${SCREENSHOTS_BRANCH}"
else
  git checkout --orphan "${SCREENSHOTS_BRANCH}"
  git rm -rf . >/dev/null 2>&1 || true
fi

mkdir -p "${dest_dir}"
shopt -s nullglob globstar
if [[ -d "${GITHUB_WORKSPACE:-}/$SCREENSHOT_ROOT" ]]; then
  abs_root="${GITHUB_WORKSPACE}/$SCREENSHOT_ROOT"
elif [[ -d "$SCREENSHOT_ROOT" ]]; then
  abs_root="$SCREENSHOT_ROOT"
else
  abs_root=""
fi

if [[ -n "$abs_root" && -d "$abs_root" ]]; then
  while IFS= read -r -d '' f; do
    base="$(basename "$f")"
    case "$base" in
      procpane-smoke.png|procpane-expanded.png)
        dest_name="$base"
        ;;
      *)
        dest_name="failure-${base}"
        ;;
    esac
    # Avoid collisions when flattening nested Playwright output paths.
    if [[ -f "${dest_dir}/${dest_name}" ]]; then
      dest_name="failure-${RANDOM}-${base}"
    fi
    cp "$f" "${dest_dir}/${dest_name}"
    published+=("${dest_name}")
  done < <(find "$abs_root" -type f -name '*.png' -print0)
fi

git add "${dest_dir}" || true
if git diff --staged --quiet; then
  echo "No PNG screenshots to publish under ${SCREENSHOT_ROOT}."
  # Branch may already contain images from a prior run; inline URLs remain valid.
  push_ok=true
else
  git commit -m "e2e: PR #${PR_NUMBER} screenshots (run ${GITHUB_RUN_ID})"
  committed=true
  attempt=1
  while [[ "$attempt" -le "$PUSH_RETRIES" ]]; do
    if git push origin "HEAD:${SCREENSHOTS_BRANCH}"; then
      push_ok=true
      break
    fi
    echo "git push to ${SCREENSHOTS_BRANCH} failed (attempt ${attempt}/${PUSH_RETRIES})."
    if [[ "$attempt" -lt "$PUSH_RETRIES" ]]; then
      sleep "$((PUSH_RETRY_SLEEP_SEC * attempt))"
    fi
    attempt=$((attempt + 1))
  done
  if [[ "$push_ok" != true ]]; then
    echo "Warning: could not push to ${SCREENSHOTS_BRANCH}; will post artifact-only PR comment."
  fi
fi

raw_base="https://raw.githubusercontent.com/${GITHUB_REPOSITORY}/${SCREENSHOTS_BRANCH}/${dest_dir}"
run_url="${GITHUB_SERVER_URL}/${GITHUB_REPOSITORY}/actions/runs/${GITHUB_RUN_ID}"
artifacts_url="${run_url}#artifacts"

{
  echo "$MARKER"
  echo "## Playwright E2E 截图"
  echo ""
  echo "Workflow run: [\`${GITHUB_RUN_ID}\`](${run_url})"
  echo ""
} >"$body_file"

title_for() {
  case "$1" in
    procpane-smoke.png) echo "列表概览 (smoke)" ;;
    procpane-expanded.png) echo "展开卡片 (detail)" ;;
    failure-*) echo "失败截图: ${1#failure-}" ;;
    *) echo "$1" ;;
  esac
}

use_inline_images=false
if [[ "$push_ok" == true && ${#published[@]} -gt 0 ]]; then
  use_inline_images=true
fi

if [[ "$use_inline_images" != true ]]; then
  if [[ "$committed" == true && "$push_ok" != true ]]; then
    echo "_本次运行已生成 PNG，但推送到 \`${SCREENSHOTS_BRANCH}\` 分支失败（GitHub 可能暂时拒绝 ref 更新）。请从本 run 的 [Artifacts](${artifacts_url}) 下载 \`playwright-report\` / \`playwright-test-results\` 查看截图与 trace。_" >>"$body_file"
  elif [[ ${#published[@]} -eq 0 ]]; then
    echo "_本次运行未找到 PNG 截图；请从 [Artifacts](${artifacts_url}) 或 workflow 日志查看。_" >>"$body_file"
  else
    echo "完整 HTML 报告与 trace 请在本 run 的 [Artifacts](${artifacts_url}) 中下载（\`playwright-report\` / \`playwright-test-results\`）。" >>"$body_file"
  fi
  echo "" >>"$body_file"
else
  {
    echo "完整 HTML 报告与 trace 仍可在 Actions **Artifacts** 中下载（\`playwright-report\` / \`playwright-test-results\`）。"
    echo ""
  } >>"$body_file"
fi

if [[ "$use_inline_images" == true ]]; then
  # Stable order: smoke, expanded, then failures.
  for name in procpane-smoke.png procpane-expanded.png; do
    if [[ -f "${dest_dir}/${name}" ]]; then
      title="$(title_for "$name")"
      {
        echo "### ${title}"
        echo ""
        echo "![${title}](${raw_base}/${name})"
        echo ""
      } >>"$body_file"
    fi
  done
  for name in $(printf '%s\n' "${published[@]}" | sort); do
    case "$name" in
      procpane-smoke.png|procpane-expanded.png) continue ;;
    esac
    title="$(title_for "$name")"
    {
      echo "### ${title}"
      echo ""
      echo "![${title}](${raw_base}/${name})"
      echo ""
    } >>"$body_file"
  done
fi

if ! existing_id="$(
  gh api "repos/${GITHUB_REPOSITORY}/issues/${PR_NUMBER}/comments" \
    --paginate \
    --jq ".[] | select(.body | contains(\"${MARKER}\")) | .id" \
    | head -1 \
    || true
)"; then
  existing_id=""
fi

if [[ -n "$existing_id" ]]; then
  gh api -X DELETE "repos/${GITHUB_REPOSITORY}/issues/comments/${existing_id}" >/dev/null || \
    echo "Warning: could not delete prior screenshot comment ${existing_id}."
fi

if gh pr comment "$PR_NUMBER" --repo "$GITHUB_REPOSITORY" --body-file "$body_file"; then
  if [[ "$use_inline_images" == true ]]; then
    echo "Posted E2E screenshot comment on PR #${PR_NUMBER} (${#published[@]} image(s) on ${SCREENSHOTS_BRANCH})."
  else
    echo "Posted E2E artifact-pointer comment on PR #${PR_NUMBER} (push_ok=${push_ok}, images=${#published[@]})."
  fi
else
  echo "Warning: failed to post PR comment on #${PR_NUMBER}; CI step remains non-fatal."
fi

exit 0
