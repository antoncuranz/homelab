#!/usr/bin/env bash
# Render every Helm chart the way Argo CD does and lint the output.
#
# Usage: ci/lint.sh [chart-dir...]          lint the given charts (default: all charts)
#        ci/lint.sh --changed [base-ref]    lint charts changed since base-ref (default: origin/main)
# Requires: helm, kubeconform, kube-linter
set -euo pipefail

cd "$(dirname "$0")/.."

all_charts() {
  for f in bootstrap/*/Chart.yaml system/*/Chart.yaml platform/*/Chart.yaml apps/*/Chart.yaml; do
    dirname "$f"
  done
}

# chart dirs touched since the merge base with $1; all charts if the lint setup itself changed
changed_charts() {
  local files
  files=$(git diff --name-only "$(git merge-base "$1" HEAD)" HEAD)
  if grep -qE '^(ci/|\.kube-linter\.yaml$|\.github/workflows/lint\.yaml$)' <<<"$files"; then
    echo "lint setup changed, linting all charts" >&2
    all_charts; return
  fi
  grep -oE '^(bootstrap|system|platform|apps)/[^/]+/' <<<"$files" | sort -u | while read -r dir; do
    [[ -f $dir/Chart.yaml ]] && echo "${dir%/}" # skips deleted charts
  done
}

# CRD schemas that are not part of upstream Kubernetes (ExternalSecret, HTTPRoute, ServiceMonitor, ...)
CRD_SCHEMAS='https://raw.githubusercontent.com/datreeio/CRDs-catalog/main/{{.Group}}/{{.ResourceKind}}_{{.ResourceAPIVersion}}.json'

# Charts gate optional resources on .Capabilities; Argo CD passes the cluster's APIs, so do the same.
API_VERSIONS=(
  monitoring.coreos.com/v1
  gateway.networking.k8s.io/v1
  cert-manager.io/v1
  external-secrets.io/v1
  snapshot.storage.k8s.io/v1
  postgresql.cnpg.io/v1
)

if [[ ${1:-} == --changed ]]; then
  charts=(); while IFS= read -r c; do charts+=("$c"); done < <(changed_charts "${2:-origin/main}")
elif [[ $# -gt 0 ]]; then
  charts=("$@")
else
  charts=(); while IFS= read -r c; do charts+=("$c"); done < <(all_charts)
fi

if [[ ${#charts[@]} -eq 0 ]]; then
  echo "No charts changed, nothing to lint."
  exit 0
fi
echo "Linting: ${charts[*]}"

# set LINT_OUT=<dir> to keep the rendered manifests for inspection
if [[ -n ${LINT_OUT:-} ]]; then
  out=$LINT_OUT; mkdir -p "$out"
else
  out=$(mktemp -d)
  trap 'rm -rf "$out"' EXIT
fi

# Argo CD ignores helm test hooks, so don't lint them either
helm_args=(--skip-tests)
for v in "${API_VERSIONS[@]}"; do helm_args+=(--api-versions "$v"); done

failed=()
for chart in "${charts[@]}"; do
  chart=${chart%/}
  name=$(basename "$chart")
  # same release name/namespace as the ApplicationSet in bootstrap/root/templates/stack.yaml
  # "update" (not "build") works without `helm repo add` and with the gitignored Chart.lock
  if ! helm dependency update "$chart" >/dev/null; then
    echo "::error file=$chart/Chart.yaml::helm dependency update failed for $chart"
    failed+=("$chart"); continue
  fi
  if ! helm template "$name" "$chart" --namespace "$name" "${helm_args[@]}" >"$out/$name.yaml"; then
    echo "::error file=$chart/values.yaml::helm template failed for $chart"
    failed+=("$chart"); continue
  fi
done

if ! compgen -G "$out/*.yaml" >/dev/null; then
  echo "FAILED: ${failed[*]}" >&2
  exit 1
fi

echo "--- kubeconform"
kubeconform -strict -summary -ignore-missing-schemas \
  -schema-location default -schema-location "$CRD_SCHEMAS" \
  "$out"/*.yaml || failed+=(kubeconform)

echo "--- kube-linter"
kube-linter lint --config .kube-linter.yaml "$out" || failed+=(kube-linter)

if [[ ${#failed[@]} -gt 0 ]]; then
  echo "FAILED: ${failed[*]}" >&2
  exit 1
fi
