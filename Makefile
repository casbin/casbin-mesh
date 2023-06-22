GOFILES ?= $(shell find . -type f -name '*.go' | grep -v '.pb.go')

.PHONY: license-check
license-check:
	go-license --config=.license.yml --verify $(GOFILES)

.PHONY: license-format
license-format:
	go-license --config=.license.yml $(GOFILES)

.PHONY: format
format: license-format
	goimports -w $(GOFILES)

.PHONY: .install.goimports
.install.goimports:
	go install golang.org/x/tools/cmd/goimports@latest

.PHONY: .install.go-license
.install.go-license:
	go install github.com/palantir/go-license@v1.26.0
