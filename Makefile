# Makefile to build wg-easy-vpn
#
#

# Shell
SHELL       := /bin/bash
# Go Compiler
GO          := $(shell command -v go)
GO_VERSION  := $(shell $(GO) version)
# Targeted arch
GOARCH      := $(if $(GOARCH),$(GOARCH),$(shell $(GO) env GOARCH))
# Binary name
BIN		    := wg-easy-vpn
# Installation directory
INSTALL_DIR := $(DESTDIR)/usr/bin
# Fancyness
OK          := "[\e[32mOK\e[0m]"
ERROR       := "[\e[31mERROR\e[0m]"

# Print compiler version
$(info Go compiler located at $(GO) ($(GO_VERSION)))

# Forced operations
.PHONY: test clean install debian

build:
	@echo -n "Building $(BIN)              "
	@GOARCH=$(GOARCH) $(GO) build -o $(BIN) *.go
	@echo -e ${OK}

install: build
	@echo -n "Installing $(BIN) to $(INSTALL_DIR) "
	@mkdir -p $(INSTALL_DIR)
	@install $(BIN) $(INSTALL_DIR)
	@echo -e ${OK}

test:
	@echo -n "Testing $(BIN)               "
	@r=$$(GOARCH=$(GOARCH) $(GO) test -v); \
		if (("$${r:0:2}" == "ok")); then \
			echo -e ${OK}; \
		else \
			echo -e ${ERROR}; \
		fi

cover:
	@echo -n "Code coverage: "
	@GOARCH=$(GOARCH) $(GO) test -cover . | awk -F ' ' '{printf "%.1f%% ",$$5}'
	@echo -e ${OK}

debian:
	@echo "Creating debian package      "
	@dpkg-buildpackage -b -us -uc
	@echo -e ${OK}

doc: 
	@echo -n "Generating documentation     "
	@$(GO) doc . > wg-easy-vpn.md
	@echo -e ${OK}

clean:
	@echo -n "Removing binaries wg-easy-vpn-*             "
	@rm -rf wg-easy-vpn wg-easy-vpn-*
	@echo -e ${OK}