#!/usr/bin/env bash
# Render every Helm chart the way Argo CD does and lint the output.
#
# Usage: ci/lint.sh [chart-dir...]   (default: all charts in bootstrap/ system/ platform/ apps/)
# Requires: helm, kubeconform, kube-linter
set -euo pipefail

cd "$(dirname "$0")/.."

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

if [[ $# -gt 0 ]]; then
  charts=("$@")
else
  charts=()
  for f in bootstrap/*/Chart.yaml system/*/Chart.yaml platform/*/Chart.yaml apps/*/Chart.yaml; do
    charts+=("$(dirname "$f")")
  done
fi

out=$(mktemp -d)
trap 'rm -rf "$out"' EXIT

helm_args=()
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
