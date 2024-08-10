
GO = go
BUILD_DIR = bin
HTTP_SERVER_DIR = http-server-go
GIT_CLI_DIR = git-go

http-server-build:
	$(GO) build -o $(BUILD_DIR)/$(HTTP_SERVER_DIR)/app $(HTTP_SERVER_DIR)/app/*.go

http-server-test: 
	cd $(HTTP_SERVER_DIR)/app && ${GO} test  *.go

git-go-build:
	$(GO) build -o $(BUILD_DIR)/$(GIT_CLI_DIR)/app $(GIT_CLI_DIR)/app/*.go

git-go-test: git-go-build
	cd $(GIT_CLI_DIR)/test && ${GO} test  *.go