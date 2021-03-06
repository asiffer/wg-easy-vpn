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
GOARM       := 
ARCH        := GOARCH=$(GOARCH)
DPKG_ARCH   := $(GOARCH)

# ARM 32 bits cases
ifeq ($(GOARCH), arm)
	GOARM     := $(if $(GOARM), $(GOARM), 7)
	ARCH      := $(ARCH) GOARM=$(GOARM)
	DPKG_ARCH := armhf
	ifneq ($(GOARM), 7)
		DPKG_ARCH := armel
	endif
endif


# Binary name
BIN		    := wg-easy-vpn
# Binary directory (where binaries are copied)
BIN_DIR     := ./bin
# Where the debian packages are stored
DIST_DIR    := ./dist
# Installation directory
INSTALL_DIR := $(DESTDIR)/usr/bin
# Fancyness
OK          := "[\e[32mOK\e[0m]"
ERROR       := "[\e[31mERROR\e[0m]"

# Print compiler version
$(info Go compiler located at $(GO) ($(GO_VERSION)))

# Forced operations
.PHONY: test clean install debian

default: build

deps:
	@echo -n "Retrieving dependencies      "
	@$(ARCH) $(GO) get -u ./...
	@echo -e ${OK}

build:
	@echo -n "Building $(BIN)              "
	@$(ARCH) $(GO) build -o $(BIN) *.go
	@mkdir -p $(BIN_DIR)/$(DPKG_ARCH)
	@cp $(BIN) $(BIN_DIR)/$(BIN)-$(DPKG_ARCH)
	@echo -e ${OK}

install:
	@echo -n "Installing $(BIN) to $(INSTALL_DIR) "
	@mkdir -p $(INSTALL_DIR)
	@install $(BIN) $(INSTALL_DIR)
	@echo -e ${OK}

test:
	@echo -n "Testing $(BIN)               "
	@r="$(shell GOARCH=$(GOARCH) $(GO) test -v | tail -n1)"; \
		if (("$${r:0:2}" == "ok")); then \
			echo -e ${OK}; \
		else \
			echo -e ${ERROR}; \
		fi

cover:
	@echo -n "Code coverage: "
	@GOARCH=$(GOARCH) $(GO) test -cover -coverprofile coverage.txt . | awk -F ' ' '{printf "%.1f%% ",$$5}'
	@echo -e ${OK}

debian:
	@echo "Creating debian package      "
	@dpkg-buildpackage -a $(DPKG_ARCH) -b
	@mkdir -p dist/
	@mv ../wg-easy-vpn_*.deb dist/
	@echo -e ${OK}

debian-no-sign:
	@echo "Creating debian package      "
	@dpkg-buildpackage -a $(DPKG_ARCH) -b --no-sign
	@mkdir -p dist/
	@mv ../wg-easy-vpn_*.deb dist/
	@echo -e ${OK}

doc: 
	@echo -n "Generating documentation          "
	@$(GO) test -run TestGenDoc 
	@mv /tmp/wg-easy-vpn.ex wg-easy-vpn.1
	@echo -e ${OK}

clean:
	@echo -n "Removing binaries wg-easy-vpn-*   "
	@rm -rf $(BIN) $(BIN)-*
	@rm -rf debian/$(BIN)
	@echo -e ${OK}