#!/usr/bin/env bash
cd "$(dirname "$0")"
while true; do ./sample.sh || sleep 2; done
