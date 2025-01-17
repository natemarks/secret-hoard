#!/usr/bin/env bash
TARBALL="ssl_certificates.tar.gz"
AWS_ACCOUNT_NUMBER="$(aws sts get-caller-identity --query "Account" --output text)"

cd private
aws s3 cp "s3://com.imprivata.${AWS_ACCOUNT_NUMBER}.us-east-1.deployment-resources/ssl_certificates/${TARBALL}" .
tar -xzvf "${TARBALL}"
rm -f "${TARBALL}"
