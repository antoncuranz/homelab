#!/bin/bash
set -euo pipefail

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
release_file="$root_dir/infrastructure/networking/cilium/app/release.yaml"
repo_file="$root_dir/infrastructure/networking/cilium/app/repo.yaml"

for bin in helm kubectl yq; do
  if ! command -v "$bin" >/dev/null 2>&1; then
    echo "missing required binary: $bin" >&2
    exit 1
  fi
done

release_name="$(yq -r '.metadata.name' "$release_file")"
namespace="$(yq -r '.metadata.namespace' "$release_file")"
chart_name="$(yq -r '.spec.chart.spec.chart' "$release_file")"
chart_version="$(yq -r '.spec.chart.spec.version' "$release_file")"
repo_name="$(yq -r '.spec.chart.spec.sourceRef.name' "$release_file")"
repo_url="$(yq -r '.spec.url' "$repo_file")"
values_file="$(mktemp)"
trap 'rm -f "$values_file"' EXIT

yq -y '.spec.values' "$release_file" > "$values_file"

kubectl create namespace "$namespace" --dry-run=client -o yaml | kubectl apply -f - >/dev/null
helm repo add "$repo_name" "$repo_url" >/dev/null 2>&1 || helm repo add "$repo_name" "$repo_url" --force-update >/dev/null
helm repo update "$repo_name" >/dev/null
helm upgrade --install "$release_name" "$repo_name/$chart_name" \
  --namespace "$namespace" \
  --version "$chart_version" \
  --values "$values_file" \
  --wait
