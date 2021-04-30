#!/bin/bash

if [[ ${1:0:1} != v ]]; then
  echo "Usage $0 vX.Y.Z [TAG_MESSAGE]"
  exit 1
fi

ver="$1"
msg="${2:-$1}"

echo "Making release for $ver"

set -xe

eval $(pass show sites/github.com | grep GITHUB_TOKEN)
GPG_FINGERPRINT=01230FD4CC29DE17

git tag -s -u $GPG_FINGERPRINT -m "$msg" "$ver"

goreleaser release --rm-dist

git push origin "$ver"

