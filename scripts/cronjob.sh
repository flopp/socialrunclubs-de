#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &> /dev/null && pwd)
TARGETS=(/var/www/virtual/floppnet/socialrunclubs.run/)

rm -rf "${SCRIPT_DIR}/out"

(cd "${SCRIPT_DIR}/repo" && ../generate-linux \
    -config "${SCRIPT_DIR}/production.json")

for TARGET in ${TARGETS[@]}; do
    cp -a "${SCRIPT_DIR}/out/." "${TARGET}"
    chmod -R a+rx "${TARGET}"
done
