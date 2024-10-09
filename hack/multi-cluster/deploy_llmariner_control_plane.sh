#!/usr/bin/env bash

set -euo pipefail

basedir=$(dirname "$0")

helm upgrade \
  --install \
  --wait \
  -n llmariner \
  llmariner \
  oci://public.ecr.aws/cloudnatix/llmariner-charts/llmariner \
  -f "${basedir}"/../../provision/common/llmariner-values.yaml \
  -f "${basedir}"/llmariner-values-control-plane.yaml
