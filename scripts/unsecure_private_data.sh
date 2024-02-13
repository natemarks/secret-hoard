#!/usr/bin/env bash
set -Eeuoxv pipefail

GPG_FILE="${1}"
TARBALL="${GPG_FILE%.gpg}"

mkdir -p encrypt secure private
gpg --output "encrypt/${TARBALL}" --decrypt "secure/${GPG_FILE}"

tar -xzvf "encrypt/${TARBALL}" -C private

rm -f encrypt/*

