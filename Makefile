# Makefile to build wg-easy-vpn

# Go Compiler
GO         := $(shell command -v go)
ARCHES     := 386 amd64 arm64 armv6 armv7
OUTPUT_DIR := dist

# $(call goarch,<arch>)
goarch = \
  $(if $(filter armv6 armv7,$1),arm,$1)

# $(call goarm,<arch>)
goarm = $(if $(filter armv6,$1),6,$(if $(filter armv7,$1),7,))

.PHONY = test clean

# basic make command
wg-easy-vpn: main.go */*.go
	$(GO) build -o $@ $<

# arch specific builds
$(OUTPUT_DIR)/wg-easy-vpn-linux-%: main.go */*.go
	@mkdir -p $(OUTPUT_DIR)
	GOOS=linux GOARCH=$(call goarch,$*) GOARM=$(call goarm,$*) $(GO) build -ldflags "-w -s" -o $@ $<

all: $(foreach arch,$(ARCHES),$(OUTPUT_DIR)/wg-easy-vpn-linux-$(arch))

coverage.out: main.go */*.go
	$(GO) test -coverprofile=coverage.out -covermode=atomic ./...

test: coverage.out

clean:
	rm -f coverage.out
	rm -f wg-easy-vpn
	rm -rf $(OUTPUT_DIR)