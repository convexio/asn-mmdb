#!/usr/bin/env bash
# Usage: build.sh TAG BASE_URL
# Writes release assets to dist/ and the download index to site/v1/database.
set -euo pipefail

tag=$1
base_url=$2
source_url=https://iptoasn.com/data/ip2asn-combined.tsv.gz

rm -rf dist site
mkdir -p dist site/v1
curl -fsSLR -o ip2asn-combined.tsv.gz "$source_url"
epoch=$(date -r ip2asn-combined.tsv.gz +%s)

go run . -epoch "$epoch" ip2asn-combined.tsv.gz dist/GeoLite2-ASN.mmdb
tar --sort=name --owner=0 --group=0 --numeric-owner --mtime="@$epoch" -C dist -cf - GeoLite2-ASN.mmdb |
  gzip -9n >dist/GeoLite2-ASN.tgz

md5=$(md5sum dist/GeoLite2-ASN.tgz | cut -d' ' -f1)
index() {
  printf '[{"name":"GeoLite2-ASN.tgz","md5_hash":"%s","url":"%s","updated":%s,"provider":"iptoasn"}]\n' "$md5" "$1" "$epoch"
}
index GeoLite2-ASN.tgz >dist/database
index "$base_url/$tag/GeoLite2-ASN.tgz" >site/v1/database
(cd dist && sha256sum GeoLite2-ASN.mmdb GeoLite2-ASN.tgz database >SHA256SUMS)

{
  echo "Source: $source_url (last modified $(date -u -d "@$epoch" +%Y-%m-%dT%H:%M:%SZ))"
  echo
  echo '```'
  cat dist/SHA256SUMS
  echo '```'
} >dist/NOTES.md
