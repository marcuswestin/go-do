test:
	bash test.sh

GO_PACKAGES := .
test-go-all: lint-go vet-go test-go-serial test-go-race
test-go-race:
	go test --race -v ${GO_PACKAGES}
test-go-serial:
	go test --parallel 1 -v ${GO_PACKAGES}
vet-go:
	go vet ${GO_PACKAGES}
lint-go:
	golint -set_exit_status ${GO_PACKAGES}
