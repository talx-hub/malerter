.PHONY : all
all: preproc build-all

.PHONY : preproc
preproc: clean fmt lint test

.PHONY : build-all
build-all: server agent

.PHONY : server
BUILDINFO_SERVER_PATH=github.com/talx-hub/malerter/internal/service/server/buildinfo
BUILD_FLAGS_SERVER=-X '$(BUILDINFO_SERVER_PATH).Version=1.0.0' \
            -X '$(BUILDINFO_SERVER_PATH).Date=$(shell date -u +%Y-%m-%d)' \
            -X '$(BUILDINFO_SERVER_PATH).Commit=$(shell git rev-parse HEAD)'

server:
	go build -ldflags="${BUILD_FLAGS_SERVER}" -o ./bin/server ./cmd/server/main.go

.PHONY : agent
BUILDINFO_AGENT_PATH=github.com/talx-hub/malerter/internal/service/agent/buildinfo
BUILD_FLAGS_AGENT=-X '$(BUILDINFO_AGENT_PATH).Version=1.0.0' \
            -X '$(BUILDINFO_AGENT_PATH).Date=$(shell date -u +%Y-%m-%d)' \
            -X '$(BUILDINFO_AGENT_PATH).Commit=$(shell git rev-parse HEAD)'

agent:
	go build -ldflags="${BUILD_FLAGS_AGENT}" -o ./bin/agent ./cmd/agent/main.go

.PHONY : test
test:
	go test ./... -tags integration_tests -race -coverprofile=cover.out -covermode=atomic
	grep -v "/pkg/pgcontainer/" cover.out > cover.filtered.out

.PHONY : run-agent
run-agent: build-all
	./bin/agent -k='super-secret-key'

.PHONY : run-server
run-server: build-all
	./bin/server -d='postgres://gopher_alerts:gopher_alerts@localhost:5432/gopher_alerts?sslmode=disable' -k='super-secret-key' -i=60

cpu.pprof:
	curl -sk -v http://localhost:8080/debug/pprof/profile?seconds=30 > profiles/cpu.pprof

heap.pprof:
	curl -sk -v http://localhost:8080/debug/pprof/heap > profiles/heap.pprof

allocs.pprof:
	curl -sk -v http://localhost:8080/debug/pprof/allocs > profiles/allocs.pprof

.PHONY : diffpprof
diffpprof:
	go tool pprof -top -diff_base=profiles/base_heap.pprof profiles/result_heap.pprof

.PHONY : clean
clean:
	-rm ./bin/agent 2>/dev/null
	-rm ./bin/server 2>/dev/null
	-rm ./cover.out 2>/dev/null
	-rm ./golangci-lint/report-unformatted.json 2>/dev/null

.PHONY : check-coverage
check-coverage:
	go tool cover -html cover.filtered.out

.PHONY : fmt
fmt:
	go fmt ./...
	goimports -v -w .

.PHONY : lint
lint:
	golangci-lint run --fix -c .golangci.yml > ./golangci-lint/report-unformatted.json

.PHONY : _golangci-lint-format-report
_golangci-lint-format-report:
	cat ./golangci-lint/report-unformatted.json | jq > ./golangci-lint/report.json
	rm ./golangci-lint/report-unformatted.json

SERVER_PORT := 37797
ADDRESS := "localhost:37797"
TEMP_FILE := "temp.bk"
.PHONY : run-autotests
run-autotests: iter01 iter02 iter03 iter04 iter05 iter06 iter07 iter08 iter09 iter10 iter11 iter12 iter13 iter14

.PHONY : iter01
iter01:
	./bin/metricstest -test.run=^TestIteration1$$ -binary-path=./bin/server

.PHONY : iter02
iter02:
	./bin/metricstest -test.run=^TestIteration2A$$ -source-path=. -agent-binary-path=./bin/agent
	./bin/metricstest -test.run=^TestIteration2B$$ -source-path=. -agent-binary-path=./bin/agent

.PHONY : iter03
iter03:
	./bin/metricstest -test.run=^TestIteration3A$$ -source-path=. -agent-binary-path=./bin/agent -binary-path=./bin/server
	./bin/metricstest -test.run=^TestIteration3B$$ -source-path=. -agent-binary-path=./bin/agent -binary-path=./bin/server

.PHONY : iter04
iter04:
	./bin/metricstest -test.run=^TestIteration4$$ -source-path=. -agent-binary-path=./bin/agent -binary-path=./bin/server -server-port=$(SERVER_PORT)

.PHONY : iter05
iter05:
	./bin/metricstest -test.run=^TestIteration5$$ -agent-binary-path=./bin/agent -binary-path=./bin/server -server-port=$(SERVER_PORT) -source-path=.

.PHONY : iter06
iter06:
	./bin/metricstest -test.run=^TestIteration6$$ -agent-binary-path=./bin/agent -binary-path=./bin/server -server-port=$(SERVER_PORT) -source-path=.

.PHONY : iter07
iter07:
	./bin/metricstest -test.run=^TestIteration7$$ -agent-binary-path=./bin/agent -binary-path=./bin/server -server-port=$(SERVER_PORT) -source-path=.

.PHONY : iter08
iter08:
	./bin/metricstest -test.run=^TestIteration8$$ -agent-binary-path=./bin/agent -binary-path=./bin/server -server-port=$(SERVER_PORT) -source-path=.

.PHONY : iter09
iter09:
	./bin/metricstest -test.run=^TestIteration9$$ -agent-binary-path=./bin/agent -binary-path=./bin/server -server-port=$(SERVER_PORT) -source-path=. -file-storage-path=$(TEMP_FILE)

.PHONY : iter10
iter10:
	./bin/metricstest -test.run=^TestIteration10A$$ \
	  -agent-binary-path=./bin/agent \
	  -binary-path=./bin/server \
	  -server-port=$(SERVER_PORT) \
	  -source-path=. \
      -database-dsn='postgres://gopher_alerts:gopher_alerts@localhost:5432/gopher_alerts?sslmode=disable'
	./bin/metricstest -test.run=^TestIteration10B$$ \
	  -agent-binary-path=./bin/agent \
	  -binary-path=./bin/server \
	  -server-port=$(SERVER_PORT) \
	  -source-path=. \
      -database-dsn='postgres://gopher_alerts:gopher_alerts@localhost:5432/gopher_alerts?sslmode=disable'

.PHONY : iter11
iter11:
	./bin/metricstest -test.run=^TestIteration11$$ \
	  -agent-binary-path=./bin/agent \
	  -binary-path=./bin/server \
	  -server-port=$(SERVER_PORT) \
	  -source-path=. \
      -database-dsn='postgres://gopher_alerts:gopher_alerts@localhost:5432/gopher_alerts?sslmode=disable'

.PHONY : iter12
iter12:
	./bin/metricstest -test.run=^TestIteration12$$ \
	  -agent-binary-path=./bin/agent \
	  -binary-path=./bin/server \
	  -server-port=$(SERVER_PORT) \
	  -source-path=. \
      -database-dsn='postgres://gopher_alerts:gopher_alerts@localhost:5432/gopher_alerts?sslmode=disable'

.PHONY : iter13
iter13:
	./bin/metricstest -test.run=^TestIteration13$$ \
	  -agent-binary-path=./bin/agent \
	  -binary-path=./bin/server \
	  -server-port=$(SERVER_PORT) \
	  -source-path=. \
      -database-dsn='postgres://gopher_alerts:gopher_alerts@localhost:5432/gopher_alerts?sslmode=disable'

.PHONY : iter14
iter14:
	./bin/metricstest -test.run=^TestIteration14$$ \
	  -agent-binary-path=./bin/agent \
	  -binary-path=./bin/server \
	  -server-port=$(SERVER_PORT) \
	  -source-path=. \
      -database-dsn='postgres://gopher_alerts:gopher_alerts@localhost:5432/gopher_alerts?sslmode=disable' \
      -key="super-secret-key"
