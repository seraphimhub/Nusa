BINARY  := nusa
GO      ?= go
PKG     := ./cmd/nusa
EXAMPLE := examples/fitur.nusa

.PHONY: all build test cover fmt vet run repl clean

all: fmt vet test build

## build: bangun binary nusa
build:
	$(GO) build -o $(BINARY) $(PKG)

## test: jalankan seluruh test
test:
	$(GO) test ./...

## cover: jalankan test beserta laporan cakupan
cover:
	$(GO) test -cover ./...

## fmt: rapikan format seluruh berkas Go
fmt:
	gofmt -w .

## vet: analisis statis
vet:
	$(GO) vet ./...

## run: jalankan contoh utama
run: build
	./$(BINARY) run $(EXAMPLE)

## repl: buka REPL
repl: build
	./$(BINARY) repl

## clean: hapus binary hasil build
clean:
	rm -f $(BINARY)

## help: tampilkan daftar target
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //'
