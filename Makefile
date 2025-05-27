SERVER_BIN=go-redis
SERVER_SRC=cmd/server/main.go
SERVER_TEST_SRC=cmd/server/main_test.go
SERVER_PORT=6379
REDIS_DIRECTORY=/tmp/redis
REDIS_DB_FILE_NAME=dump.rdb

.PHONY: run clean all test

build:
	go build -o $(SERVER_BIN) $(SERVER_SRC)

run: build
	@echo "Starting server..."
	@./$(SERVER_BIN) --port $(SERVER_PORT) --dir $(REDIS_DIRECTORY) --dbfilename $(REDIS_DB_FILE_NAME) &

test: build
	@echo "Starting server..."
	@./$(SERVER_BIN) --port $(SERVER_PORT) --dir $(REDIS_DIRECTORY) --dbfilename $(REDIS_DB_FILE_NAME) &

	@echo "Running tests..."
	go test -v $(SERVER_TEST_SRC) --port $(SERVER_PORT) --dir $(REDIS_DIRECTORY) --dbfilename $(REDIS_DB_FILE_NAME)

	@make kill clean

clean:
	@echo "Cleaning up..."
	@rm -f $(SERVER_BIN)

kill: 
	@echo "Stopping server..."
	@kill -9 $$(lsof -t -i tcp:$(SERVER_PORT))

all: build run kill clean
