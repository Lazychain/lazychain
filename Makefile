
install:
	@echo "Installing slothchaind"
	@ignite chain build

local-docker:
	@echo "Building slothchain:local"
	@docker build -t slothchain:local .

localinterslothchain:
	@echo "Spinning up localinterslothchain environment"
	@cd interslothtest && go run ./localinterslothchain

interslothtest:
	@echo "Running interslothtest"
	@cd interslothtest && go test -race -v -run TestICS20TestSuite/TestIBCTokenTransfers ./...
	@cd interslothtest && go test -race -v -run TestICS20TestSuite/TestTIAGasToken ./...
	@cd interslothtest && go test -race -v -run TestICS721TestSuite/TestICS721 ./...

.PHONY: *