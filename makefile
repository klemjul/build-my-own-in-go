
GO = go
BUILD_DIR = bin
HTTP_SERVER_BINARY = http-server-go/app
GIT_CLI_BINARY = git-go/app

http-server-build:
	$(GO) build -o $(BUILD_DIR)/$(HTTP_SERVER_BINARY) $(HTTP_SERVER_BINARY)/*.go

http-server-run: http-server-build
	./$(BUILD_DIR)/$(HTTP_SERVER_BINARY)

http-server-test: 
	cd $(HTTP_SERVER_BINARY)/ && ${GO} test  *.go

git-go-build:
	$(GO) build -o $(BUILD_DIR)/$(GIT_CLI_BINARY) $(GIT_CLI_BINARY)/*.go

git-go-run: git-go-build
	./$(BUILD_DIR)/$(GIT_CLI_BINARY)

git-go-test: git-go-build
	cd $(GIT_CLI_BINARY)/ && ${GO} test  *.go