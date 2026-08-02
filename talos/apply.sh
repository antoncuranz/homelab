#!/bin/bash

set -euo pipefail

if [ "$#" -lt 2 ]; then
  printf 'usage: %s <node-patch.yaml> <node-ip> [endpoint-ip] [talosctl apply-config args...]\n' "$0" >&2
  exit 1
fi

script_dir=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
patch_file="$1"
node_ip="$2"
endpoint_ip="${3:-$node_ip}"

if [ "$#" -ge 3 ]; then
  shift 3
else
  shift 2
fi

patch_path="$script_dir/$patch_file"
rendered_patch=$(mktemp)
rendered_role_patch=$(mktemp)

trap 'rm -f "$rendered_patch" "$rendered_role_patch"' EXIT

if [ ! -f "$patch_path" ]; then
  printf 'patch file not found: %s\n' "$patch_path" >&2
  exit 1
fi

if grep -q '^  type: worker$' "$patch_path"; then
  base_file="$script_dir/worker.yaml"
  role_patch_path=""
else
  base_file="$script_dir/worker.yaml"
  role_patch_path="$script_dir/controlplane.yaml"
fi

if [ ! -f "$base_file" ]; then
  printf 'base file not found: %s\n' "$base_file" >&2
  exit 1
fi

op inject -i "$patch_path" > "$rendered_patch"

if [ -n "$role_patch_path" ]; then
  op inject -i "$role_patch_path" > "$rendered_role_patch"
fi

if [ -n "$role_patch_path" ]; then
  talosctl --nodes "$node_ip" --endpoints "$endpoint_ip" apply-config \
    -f /dev/stdin \
    --config-patch "@$rendered_role_patch" \
    --config-patch "@$rendered_patch" \
    "$@" < <(op inject -i "$base_file")
else
  talosctl --nodes "$node_ip" --endpoints "$endpoint_ip" apply-config \
    -f /dev/stdin \
    --config-patch "@$rendered_patch" \
    "$@" < <(op inject -i "$base_file")
fi
