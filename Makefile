generate:
	go generate ./pkg/...
	go generate ./example/...

e2etest: generate
	go test -count=1 -v ./example

unittest: generate
	ginkgo -r -v -trace -cover ./pkg/ ...

.PHONY: generate e2etest unittest
