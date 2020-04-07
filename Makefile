export GO111MODULE:=on
export GOPROXY:=direct

VERSION_PKG := yunion.io/x/pkg/util/version

GIT_COMMIT := $(shell git rev-parse --short HEAD)

ifndef GIT_BRANCH
	override GIT_BRANCH := $(shell git name-rev --name-only HEAD)
endif
ifndef GIT_VERSION
	override GIT_VERSION := $(shell git describe --tags --abbrev=14 $(GIT_COMMIT)^{commit})
endif

GIT_TREE_STATE := $(shell s=`git status --porcelain 2>/dev/null`; if [ -z "$$s" ]; then echo "clean"; else echo "dirty"; fi)
BUILD_DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

LDFLAGS := "-w \
        -X $(VERSION_PKG).gitVersion=$(GIT_VERSION) \
        -X $(VERSION_PKG).gitCommit=$(GIT_COMMIT) \
        -X $(VERSION_PKG).gitBranch=$(GIT_BRANCH) \
        -X $(VERSION_PKG).buildDate=$(BUILD_DATE) \
        -X $(VERSION_PKG).gitTreeState=$(GIT_TREE_STATE) \
        -X $(VERSION_PKG).gitMajor=0 \
        -X $(VERSION_PKG).gitMinor=0"

build:
	go build -mod vendor -ldflags $(LDFLAGS) -o ./_output/bin/ocadm cmd/main.go

generate:
	./hack/codegen.sh

clean:
	rm -rf ./_output

RELEASE_BRANCH:=release/3.0
mod:
	go get yunion.io/x/onecloud@$(RELEASE_BRANCH)
	go get yunion.io/x/onecloud-operator@$(RELEASE_BRANCH)
	go get $(patsubst %,%@master,$(shell GO111MODULE=on go mod edit -print | sed -n -e 's|.*\(yunion.io/x/[a-z].*\) v.*|\1|p' | grep -v '/onecloud$$' | grep -v '/onecloud-operator$$'))
	go mod tidy
	go mod vendor -v

.PHONY: generate
