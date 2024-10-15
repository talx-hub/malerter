all: preproc build-all
.PHONY : all

.PHONY : preproc
preproc: clean fmt lint test check-coverage

.PHONY : build-all
build-all: clean server agent

server:
	go build -o ./bin/server ./cmd/server/main.go

agent:
	go build -o ./bin/agent ./cmd/agent/main.go

test:
	go test ./... -coverprofile cover.out

.PHONY : clean
clean:
	-rm ./bin/agent 2>/dev/null
	-rm ./bin/server 2>/dev/null
	-rm ./cover.out 2>/dev/null

check-coverage:
	go tool cover -html cover.out

.PHONY : fmt
fmt:
	go fmt ./...

.PHONY : lint
lint:
	golangci-lint run ./...

SERVER_PORT := 37797
ADDRESS := "localhost:37797"
TEMP_FILE := "./temp"
run-autotests: iter1 iter2 iter3 iter4 iter5
.PHONY : run-autotests

iter1:
	./bin/metricstest $(v) -test.run=^TestIteration1$$ -binary-path=./bin/server

iter2:
	./bin/metricstest $(v) -test.run=^TestIteration2A$$ -source-path=. -agent-binary-path=./bin/agent

iter3:
	./bin/metricstest $(v) -test.run=^TestIteration3A$$ -source-path=. -agent-binary-path=./bin/agent -binary-path=./bin/server
	./bin/metricstest $(v) -test.run=^TestIteration3B$$ -source-path=. -agent-binary-path=./bin/agent -binary-path=./bin/server

iter4:
	./bin/metricstest $(v) -test.run=^TestIteration4$$ -source-path=. -agent-binary-path=./bin/agent -binary-path=./bin/server -server-port=12345

iter5:
	./bin/metricstest $(v) -test.run=^TestIteration5$$ -agent-binary-path=./bin/agent -binary-path=./bin/server -server-port=$(SERVER_PORT) -source-path=.

