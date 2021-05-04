#!/bin/bash
set -eo pipefail

function set_globals {
    pipeline=$1
    TARGET="${TARGET:-denver}"
    TEAM="${TEAM:-loggregator}"
    FLY_URL="https://concourse.cf-denver.com/"
    PIPELINES_DIR="$(dirname $0)/pipelines"
}

function validate {
    if [[ "$pipeline" = "-h" ]] || [[ "$pipeline" = "--help" ]] || [[ -z "$pipeline" ]]; then
        print_usage
        exit 1
    fi
}

function set_pipeline {
    pipeline_name=$1
    pipeline_file="${PIPELINES_DIR}/$(ls ${PIPELINES_DIR} | grep ^${pipeline_name})"

    if [[ ${pipeline_file} = *.erb ]]; then
      erb ${pipeline_file} > /dev/null # this way if the erb fails the script bails
    fi

    echo setting pipeline for "$pipeline_name"

    fly -t ${TARGET} set-pipeline -p "$pipeline_name" \
        -c <(erb ${pipeline_file}) \
        -l <(lpass show 'Shared-TAS-Runtime/release-credentials-log-cache.yml' --notes) \
        -l <(lpass show 'Shared-TAS-Runtime/logging-pipeline-secrets' --notes) \
        -l <(lpass show 'Shared-Pivotal Common/pas-releng-fetch-releases' --notes) \
        --var "toolsmiths-api-key=$(lpass show 'Shared-TAS-Runtime/toolsmiths-api-token-loggregator' --notes)" \
        -l ${PIPELINES_DIR}/config/config.yml
}

function sync_fly {
    if ! fly -t ${TARGET} status; then
      fly -t ${TARGET} login -n ${TEAM} -b -c ${FLY_URL}
    fi
    fly -t ${TARGET} sync
}

function set_pipelines {
    if [[ "$pipeline" = all ]]; then
        for pipeline_path in $(find ${PIPELINES_DIR} -maxdepth 1 -type f); do
          pipeline_file=$(basename ${pipeline_path})
          set_pipeline "${pipeline_file%%.*}"
        done
        exit 0
    fi

    set_pipeline "$pipeline"
}

function print_usage {
    echo "usage: $0 <pipeline | all>"
}

function main {
    set_globals $1
    validate
    sync_fly
    set_pipelines
}

lpass ls 1>/dev/null
main $1 $2
