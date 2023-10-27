#!/bin/bash

set -o errexit
set -o pipefail
set -x
REPO_ROOT="$(git rev-parse --show-toplevel)"
declare -r REPO_ROOT
OUTPUT_DIR="$REPO_ROOT/_output"
PKG_DIR="$OUTPUT_DIR/pkg"
YUNION_BIN="$PKG_DIR/opt/yunion/bin"
TAG="${TAG:-$(git describe --tags --abbrev=0)}"
export CGO_ENABLED=0
cd "${REPO_ROOT}"

# GIT_VERSION=$TAG GIT_BRANCH=tags/$TAG make

mkdir -p "$YUNION_BIN"

docker run --platform linux/arm64 --rm -v "$(pwd):/root/go/src/yunion.io/x/ocadm" \
	--workdir "/root/go/src/yunion.io/x/ocadm" \
	registry.cn-beijing.aliyuncs.com/yunionio/debian10-base:arm-with-apt-utils-go120-git ./hack/build_deb.sh
