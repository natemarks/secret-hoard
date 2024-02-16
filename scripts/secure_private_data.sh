#!/usr/bin/env bash
set -Eeuoxv pipefail

DATE="$(date '+%Y%m%d-%H%M%S')"
TARBALL="${DATE}.tar.gz"

mkdir -p encrypt secure private
tar -czvf "encrypt/${TARBALL}" -C private/  .

HASH="$(shasum -a 256 "encrypt/${TARBALL}" | cut -f1 -d' ')"
echo "HASH: ${HASH}"

gpg --armor --symmetric --cipher-algo AES256 --output "secure/${TARBALL}.gpg" "encrypt/${TARBALL}"
rm -f encrypt/*
