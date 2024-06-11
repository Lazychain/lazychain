

###############################################################################
###                                  Build                                  ###
###############################################################################

install:
	@echo "Installing slothchaind"
	@ignite chain build

local-docker:
	@echo "Building slothchain:local"
	@docker build -t slothchain:local .

###############################################################################
###                              Run locally                                ###
###############################################################################

localinterslothchain:
	@echo "Spinning up localinterslothchain environment"
	@cd interslothtest && go run ./localinterslothchain

###############################################################################
###                                Linting                                  ###
###############################################################################

interslothtest:
	@echo "Running interslothtest"
	@cd interslothtest && go test -race -v -run TestICS20TestSuite/TestIBCTokenTransfers ./...
	@cd interslothtest && go test -race -v -run TestICS20TestSuite/TestTIAGasToken ./...
	@cd interslothtest && go test -race -v -run TestICS721TestSuite/TestICS721 ./...

###############################################################################
###                                Linting                                  ###
###############################################################################

lint:
	@echo "--> Running linter"
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint run --timeout=10m

format:
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint run --fix

.PHONY: *