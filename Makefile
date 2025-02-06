SHELL=/bin/bash
BINARY_NAME=metric-momentum
MAKEFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
PROJECT_ROOT := $(dir $(MAKEFILE_PATH))

build:
	cd $(PROJECT_ROOT) && go build -o $(BINARY_NAME) main.go

run:
	cd $(PROJECT_ROOT) && source colors && MEMO_WIDTH=10 MEMO_FILEPATH=/home/francis/.local/share/memo ./$(BINARY_NAME) print -color

clean:
	cd $(PROJECT_ROOT) && go clean && rm -f $(BINARY_NAME)

.PHONY: build run clean
