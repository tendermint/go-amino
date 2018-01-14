GOTOOLS = \
	github.com/tendermint/glide \
	gopkg.in/alecthomas/gometalinter.v2
GOTOOLS_CHECK = glide gometalinter.v2

all: check_tools get_vendor_deps test metalinter

########################################
###  Build

build:
	# Nothing to build!

install:
	# Nothing to install!


########################################
### Tools & dependencies

check_tools:
	@# https://stackoverflow.com/a/25668869
	@echo "Found tools: $(foreach tool,$(GOTOOLS_CHECK),\
        $(if $(shell which $(tool)),$(tool),$(error "No $(tool) in PATH")))"

get_tools:
	@echo "--> Installing tools"
	go get -u -v $(GOTOOLS)
	@gometalinter.v2 --install

update_tools:
	@echo "--> Updating tools"
	@go get -u $(GOTOOLS)

get_vendor_deps:
	@rm -rf vendor/
	@echo "--> Running glide install"
	@glide install


########################################
### Testing

test:
	go test `glide novendor`


########################################
### Formatting, linting, and vetting

fmt:
	@go fmt ./...

metalinter:
	@echo "==> Running linter"
	gometalinter.v2 --vendor --deadline=600s --disable-all  \
		--enable=deadcode \
		--enable=goconst \
		--enable=goimports \
		--enable=gosimple \
		--enable=ineffassign \
		--enable=megacheck \
		--enable=misspell \
		--enable=staticcheck \
		--enable=safesql \
		--enable=structcheck \
		--enable=unconvert \
		--enable=unused \
		--enable=varcheck \
		--enable=vetshadow \
		./...

		#--enable=maligned \
		#--enable=gas \
		#--enable=aligncheck \
		#--enable=dupl \
		#--enable=errcheck \
		#--enable=gocyclo \
		#--enable=golint \ <== comments on anything exported
		#--enable=gotype \
		#--enable=interfacer \
		#--enable=unparam \
		#--enable=vet \

metalinter_all:
	protoc $(INCLUDE) --lint_out=. types/*.proto
	gometalinter.v2 --vendor --deadline=600s --enable-all --disable=lll ./...


test_golang1.10rc:
	docker run -it -v "$(CURDIR):/go/src/github.com/tendermint/go-wire" -w "/go/src/github.com/tendermint/go-wire" golang:1.10-rc /bin/bash -ci "make get_tools all"

# To avoid unintended conflicts with file names, always add to .PHONY
# unless there is a reason not to.
# https://www.gnu.org/software/make/manual/html_node/Phony-Targets.html
.PHONY: build install check_tools get_tools update_tools get_vendor_deps test fmt metalinter metalinter_all
