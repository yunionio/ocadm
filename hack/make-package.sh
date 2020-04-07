#!/bin/bash

set -o errexit
set -o pipefail

REPO_ROOT="$(git rev-parse --show-toplevel)"
declare -r REPO_ROOT
OUTPUT_DIR="$REPO_ROOT/_output"
PKG_DIR="$OUTPUT_DIR/pkg"
YUNION_BIN="$PKG_DIR/opt/yunion/bin"
TAG="${TAG:-$(git describe --tags --abbrev=0)}"
cd "${REPO_ROOT}"

GIT_VERSION=$TAG GIT_BRANCH=tags/$TAG make

mkdir -p $YUNION_BIN

cp -a "$OUTPUT_DIR/bin/ocadm" "$YUNION_BIN"

docker run --rm -v "$(pwd):/src/" cdrx/fpm-centos:7 fpm  -n yunion-ocadm -v "${TAG#v}" -s dir -t rpm  -C "_output/pkg"
