#!/usr/bin/env bash
set -Eeuoxv pipefail
TARBALL="ssl_certificates.tar.gz"
AWS_ACCOUNT_NUMBER="$(aws sts get-caller-identity --query "Account" --output text)"
if [ "${AWS_ACCOUNT_NUMBER}" == "151924297945" ]
then
  tar -czvf ssl_certificates.tar.gz -C private \
  com_imprivata_common_dev_star.crt \
  com_imprivata_common_dev_star.key \
  com_imprivata_common_integration_star.crt \
  com_imprivata_common_integration_star.key

  aws s3 cp  ssl_certificates.tar.gz \
  "s3://com.imprivata.${AWS_ACCOUNT_NUMBER}.us-east-1.deployment-resources/ssl_certificates/ssl_certificates.tar.gz"
  rm -f ssl_certificates.tar.gz
elif [ "${AWS_ACCOUNT_NUMBER}" == "468716396736" ]
then
  tar -czvf ssl_certificates.tar.gz -C private \
  com_imprivata_common_staging_star.crt \
  com_imprivata_common_staging_star.key \
  com_imprivata_cloud_biometric-aware.crt \
  com_imprivata_cloud_biometric-aware.key

  aws s3 cp  ssl_certificates.tar.gz \
  "s3://com.imprivata.${AWS_ACCOUNT_NUMBER}.us-east-1.deployment-resources/ssl_certificates/ssl_certificates.tar.gz"

  rm -f ssl_certificates.tar.gz
fi
