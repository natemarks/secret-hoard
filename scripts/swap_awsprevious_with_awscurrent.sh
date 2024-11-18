#!/usr/bin/env bash
#set -Eeuoxv pipefail
set -Eeuo pipefail

# secret_name="ssl_certificate/dev/star.dev.common.imprivata.com"
secret_name="${1}"
echo "swapping AWSPREVIOUS <-> AWSCURRENT for: ${secret_name}"

secret_versions=$(aws secretsmanager list-secret-version-ids --secret-id "$secret_name" --query 'Versions[]' --output json)
AWS_PREVIOUS_ID=$(echo "${secret_versions}" | jq -r '.[] | select(.VersionStages[] == "AWSPREVIOUS") | .VersionId')
echo "PREVIOUS ID: ${AWS_PREVIOUS_ID}"
AWS_CURRENT_ID=$(echo "${secret_versions}" | jq -r '.[] | select(.VersionStages[] == "AWSCURRENT") | .VersionId')
echo "CURRENT ID: ${AWS_CURRENT_ID}"
exit
aws secretsmanager update-secret-version-stage \
--secret-id "${secret_name}" \
--version-stage AWSCURRENT \
--move-to-version-id ${AWS_PREVIOUS_ID} \
--remove-from-version-id "${AWS_CURRENT_ID}"

