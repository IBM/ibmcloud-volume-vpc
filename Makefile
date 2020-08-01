
GOPACKAGES=$(shell go list ./... | grep -v /vendor/ | grep -v /samples)
GOFILES=$(shell find . -type f -name '*.go' -not -path "./vendor/*")
ARCH = $(shell uname -m)
LINT_VERSION="1.27.0"

.PHONY: all
all: deps dofmt vet test

.PHONY: deps
deps:
	go get github.com/pierrre/gotestcover
	@if ! which golangci-lint >/dev/null || [[ "$$(golangci-lint --version)" != *${LINT_VERSION}* ]]; then \
		curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v${LINT_VERSION}; \
	fi

.PHONY: fmt
fmt:
	golangci-lint run --disable-all --enable=gofmt

.PHONY: dofmt
dofmt:
	golangci-lint run --disable-all --enable=gofmt --fix

.PHONY: lint
lint:
	golangci-lint run

.PHONY: makefmt
makefmt:
	gofmt -l -w ${GOFILES}

.PHONY: test
test:
ifeq ($(ARCH), ppc64le)
	# POWER
	$(GOPATH)/bin/gotestcover -v -coverprofile=cover.out ${GOPACKAGES} -timeout 90m
else
	# x86_64
	$(GOPATH)/bin/gotestcover -v -race -coverprofile=cover.out ${GOPACKAGES} -timeout 90m
endif

.PHONY: coverage
coverage:
	go tool cover -html=cover.out -o=cover.html

.PHONY: vet
vet:
	go vet ${GOPACKAGES}