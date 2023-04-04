.PHONY: tools
tools:
	go install github.com/RangelReale/trcache/cmd/troptgen

.PHONY: gen
gen: tools
	go generate ./...

.PHONY: test
test:
	go test -count=1 ./...
