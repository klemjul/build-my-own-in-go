
GO = go
BUILD_DIR = bin
HTTP_SERVER_BINARY = http-server-go/app

http-server-build:
	$(GO) build -o $(BUILD_DIR)/$(HTTP_SERVER_BINARY) $(HTTP_SERVER_BINARY)/*.go

http-server-run: http-server-build
	./$(BUILD_DIR)/$(HTTP_SERVER_BINARY)

http-server-test: 
	${GO} test  $(HTTP_SERVER_BINARY)/*.go