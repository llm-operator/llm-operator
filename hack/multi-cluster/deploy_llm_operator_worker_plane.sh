#!/usr/bin/env bash

set -euo pipefail

basedir=$(dirname "$0")

helm upgrade \
  --install \
  --wait \
  -n llm-operator-wp \
  llm-operator \
  oci://public.ecr.aws/cloudnatix/llm-operator-charts/llm-operator \
  --set tags.control-plane=false \
  -f "${basedir}"/../llm-operator-values.yaml \
  -f "${basedir}"/../llm-operator-values-cpu-only.yaml \
  -f "${basedir}"/llm-operator-values-worker-plane.yaml
