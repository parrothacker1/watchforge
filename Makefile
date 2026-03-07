BINARY_DIR = .bin
APP_NAME = watchforge

TARGET ?= linux
ARCH ?= amd64

SRC := $(shell find . -type f -name "*.go" -not -path "./.bin/*")
MAIN_SRC := .

UPX ?= upx

BUILD ?= debug
VERSION ?= dev

DEBUG_FLAGS =
RELEASE_FLAGS = -ldflags="-s -w -X main.version=$(VERSION)" -trimpath -buildvcs=false

DEBUG_BIN = $(BINARY_DIR)/$(APP_NAME)-debug-$(TARGET)-$(ARCH)
RELEASE_BIN = $(BINARY_DIR)/$(APP_NAME)-release-$(TARGET)-$(ARCH)

EXT :=
ifeq ($(TARGET),windows)
EXT := .exe
endif

DEBUG_BIN := $(DEBUG_BIN)$(EXT)
RELEASE_BIN := $(RELEASE_BIN)$(EXT)

BIN := $(DEBUG_BIN)
FLAGS := $(DEBUG_FLAGS)

ifeq ($(BUILD),release)
BIN := $(RELEASE_BIN)
FLAGS := $(RELEASE_FLAGS)
endif

.PHONY: setup build clean run

setup:
	@echo "[X] Setting up dir"
	@mkdir -p $(BINARY_DIR)

$(BIN): $(SRC) | setup
	@echo "[X] Building the $(BUILD) binary"
	@CGO_ENABLED=0 GOOS=$(TARGET) GOARCH=$(ARCH) \
		go build $(FLAGS) -o $(BIN) $(MAIN_SRC)
	@echo "[X] Done building the $(BUILD) binary"
ifeq ($(BUILD),release)
	@echo "[X] Compressing binary with UPX"
	@$(UPX) -9 $(BIN)
endif

build: $(BIN)

clean:
	@echo "[X] Removing all binaries from $(BINARY_DIR)"
	@rm -rf $(BINARY_DIR)

run: $(BIN)
	@echo "[X] Running the $(BUILD) binary"
	@./$(BIN)