#!/bin/bash

tag=true

while true; do
  case "$1" in
    --no-tag)
      tag=false
      shift
      ;;
    *)
      break
      ;;
  esac
done

if tag && [[ ${1:0:1} != v ]]; then
  echo "Usage $0 [--no-tag] vX.Y.Z [TAG_MESSAGE]"
  exit 1
fi

ver="$1"
msg="${2:-$1}"

echo "Making release for $ver"

set -xe

eval $(pass show sites/github.com | grep GITHUB_TOKEN)
GPG_FINGERPRINT=01230FD4CC29DE17

if $tag; then
  git tag -s -u $GPG_FINGERPRINT -m "$msg" "$ver"
  git push origin "$ver"
fi

goreleaser release --rm-dist


