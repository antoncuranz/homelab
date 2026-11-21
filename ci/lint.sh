#!/usr/bin/env bash
# Render every Helm chart the way Argo CD does and lint the output.
#
# Usage: ci/lint.sh [chart-dir...]          lint the given charts (default: all charts)
#        ci/lint.sh --changed [base-ref]    lint charts changed since base-ref (default: origin/main)
#        ci/lint.sh --render-only chart...   render charts without running the linters
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
  if grep -qE '^(ci/|\.kube-linter\.yaml$|(\.github|\.forgejo)/workflows/lint\.yaml$)' <<<"$files"; then
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

render_only=false
if [[ ${1:-} == --render-only ]]; then
  render_only=true; shift
  [[ $# -gt 0 ]] || { echo "--render-only requires at least one chart" >&2; exit 1; }
fi

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

# Own repo config/cache, so the user's `helm repo list` is neither used nor refreshed.
# `helm dependency update` would otherwise refresh every configured repo for every chart.
helm_home=${HELM_LINT_HOME:-${XDG_CACHE_HOME:-$HOME/.cache}/homelab-lint}
export HELM_REPOSITORY_CONFIG=$helm_home/repositories.yaml HELM_REPOSITORY_CACHE=$helm_home/repository
mkdir -p "$HELM_REPOSITORY_CACHE"

# true if charts/ already holds every dependency at the version in Chart.yaml
deps_ok() {
  helm dependency list "$1" 2>/dev/null | awk 'NR > 1 && NF && $NF != "ok" { bad = 1 } END { exit bad }'
}

stale=()
for chart in "${charts[@]}"; do
  deps_ok "${chart%/}" || stale+=("${chart%/}")
done

if [[ ${#stale[@]} -gt 0 ]]; then
  echo "Fetching dependencies for: ${stale[*]}"
  # add each HTTP repo once (OCI deps need no repo), then refresh all indexes once
  for url in $(sed -nE 's/^[[:space:]-]*repository:[[:space:]]*"?(https?:\/\/[^" ]+)"?.*/\1/p' \
                 "${stale[@]/%//Chart.yaml}" | sed 's:/*$::' | sort -u); do
    helm repo add --force-update "$(tr -c 'a-zA-Z0-9\n' - <<<"${url#*://}")" "$url" >/dev/null
  done
  [[ -s $HELM_REPOSITORY_CONFIG ]] && helm repo update >/dev/null
fi

failed=()
for chart in "${charts[@]}"; do
  chart=${chart%/}
  name=$(basename "$chart")
  if [[ " ${stale[*]-} " == *" $chart "* ]] && ! helm dependency update --skip-refresh "$chart" >/dev/null; then
    echo "::error file=$chart/Chart.yaml::helm dependency update failed for $chart"
    failed+=("$chart"); continue
  fi
  # same release name/namespace as the ApplicationSet in bootstrap/root/templates/stack.yaml
  if ! helm template "$name" "$chart" --namespace "$name" "${helm_args[@]}" >"$out/$name.yaml"; then
    echo "::error file=$chart/values.yaml::helm template failed for $chart"
    failed+=("$chart"); continue
  fi
done

if [[ $render_only == true ]]; then
  if [[ ${#failed[@]} -gt 0 ]]; then
    echo "FAILED: ${failed[*]}" >&2
    exit 1
  fi
  exit 0
fi

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
