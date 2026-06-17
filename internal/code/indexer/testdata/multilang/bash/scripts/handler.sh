#!/usr/bin/env bash
set -euo pipefail

source ./common.sh

health_check() {
  echo "ok"
}

run_handler() {
  health_check
}
