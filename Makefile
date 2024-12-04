.PHONY : all
all: preproc build-all

.PHONY : preproc
preproc: clean fmt lint test

.PHONY : build-all
build-all: server agent

server:
	go build -o ./bin/server ./cmd/server/main.go

agent:
	go build -o ./bin/agent ./cmd/agent/main.go

.PHONY : test
test:
	go test ./... -race -coverprofile=cover.out -covermode=atomic

.PHONY : clean
clean:
	-rm ./bin/agent 2>/dev/null
	-rm ./bin/server 2>/dev/null
	-rm ./cover.out 2>/dev/null
	-rm ./golangci-lint/report-unformatted.json 2>/dev/null

.PHONY : check-coverage
check-coverage:
	go tool cover -html cover.out

.PHONY : fmt
fmt:
	go fmt ./...
	goimports -v -w .

.PHONY : lint
lint:
	golangci-lint run -c .golangci.yml > ./golangci-lint/report-unformatted.json

.PHONY : _golangci-lint-format-report
_golangci-lint-format-report:
	cat ./golangci-lint/report-unformatted.json | jq > ./golangci-lint/report.json
	rm ./golangci-lint/report-unformatted.json

SERVER_PORT := 37797
ADDRESS := "localhost:37797"
TEMP_FILE := "temp.bk"
.PHONY : run-autotests
run-autotests: iter1 iter2 iter3 iter4 iter5 iter6 iter7 iter8 iter9 iter10 iter11

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

.PHONY : iter7
iter7:
	./bin/metricstest -test.run=^TestIteration7$$ -agent-binary-path=./bin/agent -binary-path=./bin/server -server-port=$(SERVER_PORT) -source-path=.

.PHONY : iter8
iter8:
	./bin/metricstest -test.run=^TestIteration8$$ -agent-binary-path=./bin/agent -binary-path=./bin/server -server-port=$(SERVER_PORT) -source-path=.

.PHONY : iter9
iter9:
	./bin/metricstest -test.run=^TestIteration9$$ -agent-binary-path=./bin/agent -binary-path=./bin/server -server-port=$(SERVER_PORT) -source-path=. -file-storage-path=$(TEMP_FILE)

.PHONY : iter10
iter10:
	./bin/metricstest -test.run=^TestIteration10A$$ \
	  -agent-binary-path=./bin/agent \
	  -binary-path=./bin/server \
	  -server-port=$(SERVER_PORT) \
	  -source-path=. \
      -database-dsn='postgres://godevops:godevops@localhost:5432/godevops_alerts?sslmode=disable'
	./bin/metricstest -test.run=^TestIteration10B$$ \
	  -agent-binary-path=./bin/agent \
	  -binary-path=./bin/server \
	  -server-port=$(SERVER_PORT) \
	  -source-path=. \
      -database-dsn='postgres://godevops:godevops@localhost:5432/godevops_alerts?sslmode=disable'

.PHONY : iter11
iter11:
	./bin/metricstest -test.run=^TestIteration11$$ \
	  -agent-binary-path=./bin/agent \
	  -binary-path=./bin/server \
	  -server-port=$(SERVER_PORT) \
	  -source-path=. \
      -database-dsn='postgres://godevops:godevops@localhost:5432/godevops_alerts?sslmode=disable'
