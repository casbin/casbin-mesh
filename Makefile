.PHONY: license-check
license-check:
	find . -type f -name '*.go' | grep -v '.pb.go' | xargs go-license --config=.license.yml --verify

.PHONY: license-format
license-format:
	find . -type f -name '*.go' | grep -v '.pb.go' | xargs go-license --config=.license.yml
