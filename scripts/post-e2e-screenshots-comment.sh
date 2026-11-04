#!/usr/bin/env bash
# Publish Playwright PNGs to branch e2e-screenshots and post a PR comment with
# raw.githubusercontent.com URLs (inline on GitHub). GITHUB_TOKEN cannot use
# gh pr comment --attach (ghs_ is rejected); see cli/cli#14180.
set -euo pipefail

SCREENSHOT_ROOT="${SCREENSHOT_ROOT:-frontend/test-results}"
MARKER="<!-- procpane-e2e-screenshots -->"
SCREENSHOTS_BRANCH="e2e-screenshots"

if [[ -z "${PR_NUMBER:-}" ]]; then
  echo "PR_NUMBER unset; skipping PR screenshot comment."
  exit 0
fi

if [[ -z "${GH_TOKEN:-}" ]]; then
  echo "GH_TOKEN unset; cannot publish screenshots."
  exit 1
fi

body_file="$(mktemp)"
work_dir="$(mktemp -d)"
dest_dir="pr-${PR_NUMBER}"
published=()

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
else
  git commit -m "e2e: PR #${PR_NUMBER} screenshots (run ${GITHUB_RUN_ID})"
  git push origin "HEAD:${SCREENSHOTS_BRANCH}"
fi

raw_base="https://raw.githubusercontent.com/${GITHUB_REPOSITORY}/${SCREENSHOTS_BRANCH}/${dest_dir}"

{
  echo "$MARKER"
  echo "## Playwright E2E 截图"
  echo ""
  echo "Workflow run: [\`${GITHUB_RUN_ID}\`](${GITHUB_SERVER_URL}/${GITHUB_REPOSITORY}/actions/runs/${GITHUB_RUN_ID})"
  echo ""
  echo "完整 HTML 报告与 trace 仍可在 Actions **Artifacts** 中下载（\`playwright-report\` / \`playwright-test-results\`）。"
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

if [[ ${#published[@]} -eq 0 ]]; then
  echo "_本次运行未找到 PNG 截图；请查看 Artifacts 或 workflow 日志。_" >>"$body_file"
  echo "" >>"$body_file"
else
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

gh pr comment "$PR_NUMBER" --repo "$GITHUB_REPOSITORY" --body-file "$body_file"
echo "Posted E2E screenshot comment on PR #${PR_NUMBER} (${#published[@]} image(s) on ${SCREENSHOTS_BRANCH})."
