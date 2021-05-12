#!/bin/bash

set -eo pipefail

cf install-plugin "log-cache" -f

cf tail -n 1000 uaa | grep requests.global.completed


