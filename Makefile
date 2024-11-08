all: preproc build-all
.PHONY : all

.PHONY : preproc
preproc: fmt lint test

.PHONY : build-all
build-all: clean server agent

server:
	go build -o ./bin/server ./cmd/server/main.go

agent:
	go build -o ./bin/agent ./cmd/agent/main.go

test:
	go test ./... -race -coverprofile=cover.out -covermode=atomic

.PHONY : clean
clean:
	-rm ./bin/agent 2>/dev/null
	-rm ./bin/server 2>/dev/null
	-rm ./cover.out 2>/dev/null

.PHONY : check-coverage
check-coverage:
	go tool cover -html cover.out

.PHONY : fmt
fmt:
	go fmt ./...
	goimports -v -w .

.PHONY : lint
lint:
	golangci-lint run ./...

SERVER_PORT := 37797
ADDRESS := "localhost:37797"
TEMP_FILE := "./temp"
.PHONY : run-autotests
run-autotests: iter1 iter2 iter3 iter4 iter5 iter6

.PHONY : iter1
iter1:
	./bin/metricstest -test.run=^TestIteration1$$ -binary-path=./bin/server

.PHONY : iter2
iter2:
	./bin/metricstest -test.run=^TestIteration2A$$ -source-path=. -agent-binary-path=./bin/agent

.PHONY : iter3
iter3:
	./bin/metricstest -test.run=^TestIteration3A$$ -source-path=. -agent-binary-path=./bin/agent -binary-path=./bin/server
	./bin/metricstest -test.run=^TestIteration3B$$ -source-path=. -agent-binary-path=./bin/agent -binary-path=./bin/server

.PHONY : iter4
iter4:
	./bin/metricstest -test.run=^TestIteration4$$ -source-path=. -agent-binary-path=./bin/agent -binary-path=./bin/server -server-port=$(SERVER_PORT)

.PHONY : iter5
iter5:
	./bin/metricstest -test.run=^TestIteration5$$ -agent-binary-path=./bin/agent -binary-path=./bin/server -server-port=$(SERVER_PORT) -source-path=.

.PHONY : iter6
iter6:
	./bin/metricstest -test.run=^TestIteration6$$ -agent-binary-path=./bin/agent -binary-path=./bin/server -server-port=$(SERVER_PORT) -source-path=.
