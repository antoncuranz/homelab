#!/usr/bin/env bash
# Compare changed charts against their merge-base render. LINT_OUT must contain
# the current renders produced by ci/lint.sh --changed for the same base.
set -euo pipefail
cd "$(dirname "$0")/.."
base=${1:-origin/main}
: "${LINT_OUT:?Set LINT_OUT to the directory containing the lint renders}"
base_commit=$(git merge-base "$base" HEAD)
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
mkdir -p "$tmp/base" "$tmp/rendered"
printf '# Rendered manifest diff\n\n' > "$tmp/summary"
git archive "$base_commit" | tar -x -C "$tmp/base"
# Use the current rendering logic on both revisions, even if the script changed.
mkdir -p "$tmp/base/ci"
cp ci/lint.sh "$tmp/base/ci/lint.sh"

charts=()
while IFS= read -r chart; do
  [[ -n $chart ]] && charts+=("${chart%/}")
done < <(git diff --name-only "$base_commit" HEAD -- bootstrap system platform apps | \
  grep -oE '^(bootstrap|system|platform|apps)/[^/]+/' | sort -u || true)

if [[ ${#charts[@]} -eq 0 ]]; then
  echo "No chart manifests to diff."
  echo "No chart manifests to diff." >> "$tmp/summary"
  [[ -z ${DIFF_SUMMARY:-} ]] || cat "$tmp/summary" >> "$DIFF_SUMMARY"
  exit 0
fi
base_charts=()
changed=false
for chart in "${charts[@]}"; do
  [[ -f $tmp/base/$chart/Chart.yaml ]] && base_charts+=("$chart")
done
if [[ ${#base_charts[@]} -gt 0 ]]; then
  LINT_OUT="$tmp/rendered" bash "$tmp/base/ci/lint.sh" --render-only "${base_charts[@]}"
fi

for chart in "${charts[@]}"; do
  name=${chart##*/}
  old=$tmp/rendered/$name.yaml
  new=$LINT_OUT/$name.yaml
  [[ -f $tmp/base/$chart/Chart.yaml ]] || old=/dev/null
  if [[ ! -f $chart/Chart.yaml ]]; then
    new=/dev/null
  elif [[ ! -f $new ]]; then
    echo "Missing current render for $chart: $new" >&2
    exit 1
  fi
  echo "--- Manifest diff: $chart"
  status=0
  diff -u --label "base/$chart" --label "head/$chart" "$old" "$new" > "$tmp/diff" || status=$?
  if [[ $status -eq 0 ]]; then
    echo "No rendered changes."
  elif [[ $status -eq 1 ]]; then
    changed=true
    cat "$tmp/diff"
    {
      printf '## `%s`\n\n' "$chart"
      echo '````diff'
      cat "$tmp/diff"
      printf '\n````\n\n'
    } >> "$tmp/summary"
  elif [[ $status -ne 1 ]]; then
    exit "$status"
  fi
done
if [[ $changed == false ]]; then
  echo "No rendered changes." >> "$tmp/summary"
fi
[[ -z ${DIFF_SUMMARY:-} ]] || cat "$tmp/summary" >> "$DIFF_SUMMARY"
