.PHONEY: all docs test install get_vendor_deps ensure_tools

GOTOOLS = \
	github.com/Masterminds/glide
REPO:=github.com/tendermint/go-wire

docs:
	@go get github.com/davecheney/godoc2md
	godoc2md $(REPO) > README.md

all: install test

install: 
	go install github.com/tendermint/go-wire/cmd/...

test:
	go test `glide novendor`

get_vendor_deps: ensure_tools
	@rm -rf vendor/
	@echo "--> Running glide install"
	@glide install

ensure_tools:
	go get $(GOTOOLS)

pigeon:
	pigeon -o expr/expr.go expr/expr.peg

tools:
	@go get github.com/clipperhouse/gen
