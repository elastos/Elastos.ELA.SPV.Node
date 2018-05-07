BUILD=go build
VERSION := $(shell git describe --abbrev=4 --dirty --always --tags)

BUILD_SPV_NODE =$(BUILD) -ldflags "-X main.Version=$(VERSION)" -o spv-node main.go

all:
	$(BUILD_SPV_NODE)