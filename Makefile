BINARY := fixme


# Build a development build
build:
	@go build -o bin/${BINARY}

# run tests
test:
	@go test -v -race ./...

# update the golden files used for the integration tests
update-tests:
	@go test integration/cli_test.go -update
