#!/bin/bash
set -euo pipefail
cd "$(dirname "$0")"

devcontainer up
devcontainer exec -- "$@"
