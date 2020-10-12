APP = $(shell basename $$PWD)
BINDIR ?= bin
DESTDIR ?= /usr/bin
GC = go build
GCFLAGS =
LDFLAGS = -ldflags="-s -w"
GO111MODULES = on
UPX := $(shell command -v upx 2> /dev/null)
#LDFLAGS=-ldflags "-X=main.Version=$(VERSION) -X=main.Build=$(BUILD)"
#VERSION := $(shell git describe --tags)
#BUILD := $(shell git rev-parse --short HEAD)
#ROJECTNAME := $(shell basename "$(PWD)")
#COMMIT_SHA = $(shell git rev-parse --short HEAD)
.DEFAULT_GOAL = help

# Ignore dependencies since we are using command aliases instead of targets
.PHONY: hello run build uninstall compile clean help

## hello: Print the app name we are building
hello:
	@echo Makefile for $(APP) output in $(BINDIR)

## run: Build and run the program
run:
	@go run cmd/$(APP)/main.go

## build: Build the package binary (in the bin/ directory)
build:
	$(GC) $(GCFLAGS) -o $(BINDIR)/$(APP) cmd/$(APP)/main.go

## buildsmall: Build a stripped package binary
buildsmall:
	$(GC) $(GCFLAGS) $(LDFLAGS) -o $(BINDIR)/$(APP) cmd/$(APP)/main.go

## install: Install the program on the local system
install: buildsmall
	@echo Installing $(BINDIR)/$(APP) in $(DESTDIR)/$(APP)
	@cp $(BINDIR)/$(APP) $(DESTDIR)/$(APP) \
		&& chown root.root $(DESTDIR)/$(APP) \
		&& chmod 755 $(DESTDIR)/$(APP)

## uninstall: Remove an installed program
uninstall:
	@echo Removing $(APP) from $(DESTDIR)
	@sudo rm $(DESTDIR)/$(APP)

## compile: Compile binaries for Linux and Windows
compile:
	@echo Compiling stripped binaries for Linux, Mac, and Windows
	@GOOS=linux GOARCH=amd64 $(GC) $(GCFLAGS) $(LDFLAGS) -o $(BINDIR)/$(APP)-linux-amd64 cmd/$(APP)/main.go
	@GOOS=darwin GOARCH=amd64 $(GC) $(GCFLAGS) $(LDFLAGS) -o $(BINDIR)/$(APP)-darwin-amd64 cmd/$(APP)/main.go
	@GOOS=windows GOARCH=amd64 $(GC) $(GCFLAGS) $(LDFLAGS) -o $(BINDIR)/$(APP)-windows-amd64.exe cmd/$(APP)/main.go

## package: Package stripped binaries for distribition
package: compile
	@echo Compressing binaries for distribution
ifdef UPX
	@$(UPX) --brute $(BINDIR)/$(APP)-linux-amd64
	@$(UPX) --brute $(BINDIR)/$(APP)-darwin-amd64
	@$(UPX) --brute $(BINDIR)/$(APP)-windows-amd64.exe
endif
	@GZIP=-9 tar -C $(BINDIR) -czvf $(APP)-linux-amd64.tar.gz $(APP)-linux-amd64
	@GZIP=-9 tar -C $(BINDIR) -czvf $(APP)-darwin-amd64.tar.gz $(APP)-darwin-amd64
	@GZIP=-9 tar -C $(BINDIR) -czvf $(APP)-windows-amd64.tar.gz $(APP)-windows-amd64.exe

## clean: Run "go clean"
clean:
	go clean
	@rm -f $(BINDIR)/$(APP) $(BINDIR)/$(APP)-linux-amd64 $(BINDIR)/$(APP)-windows-amd64.exe

## help: Show available make targets and a brief description of each
help: Makefile
	@echo
	@echo Available make targets:
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo