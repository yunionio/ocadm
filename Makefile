export GO111MODULES:=on

build: generate
	go build -o ./_output/bin/ocadm cmd/main.go

generate:
	./hack/update-codegen.sh

clean:
	rm -rf ./_output

.PHONY: generate
