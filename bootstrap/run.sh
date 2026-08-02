#!/bin/bash
set -euo pipefail

environment="${1:-prod}"

case "$environment" in
  prod|production)
    environment=prod
    flux_path=./clusters/prod
    flux_branch=talos
    ;;
  local)
    flux_path=./clusters/local
    flux_branch=talos
    ;;
  *)
    echo "usage: $0 [prod|local]" >&2
    exit 1
    ;;
esac

cat initial-secrets.yaml | op inject | kubectl apply --server-side --field-manager flux-client-side-apply -f -

"$(dirname "$0")/install-cilium.sh"

echo "Bootstrapping ${environment} from ${flux_path}. Use prod and local against separate target clusters; both reuse the standard flux-system names."

flux bootstrap github \
  --owner=antoncuranz \
  --repository=homelab \
  --branch="${flux_branch}" \
  --path="${flux_path}" \
  --personal
