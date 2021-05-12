#!/bin/bash

set -eo pipefail

time=$(date +%s%N)
cf install-plugin "log-cache" -f
sleep 5
cf tail uaa -n 1000 --start-time="${time}" | grep requests.global.completed


