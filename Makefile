export GO111MODULE:=on

build:
	go build -mod vendor -o ./_output/bin/ocadm cmd/main.go

generate:
	./hack/codegen.sh

clean:
	rm -rf ./_output

.PHONY: generate
