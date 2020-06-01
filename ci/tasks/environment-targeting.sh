#!/bin/bash
API_URL=$(cat toolsmiths-env/metadata | jq -r '.cf.api_url')
eval "$(bbl print-env --metadata-file toolsmiths-env/metadata)"
password=$(credhub get --name=$(credhub find | grep cf_admin | awk '{print $3}') --output-json | jq -r .value)
cf api $API_URL --skip-ssl-validation
cf auth admin $password
cf target -o system
