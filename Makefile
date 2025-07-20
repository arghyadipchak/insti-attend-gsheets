GOCMD=go
GOBUILD=${GOCMD} build
GOTEST=${GOCMD} test
GOFMT=${GOCMD} fmt
GOVET=${GOCMD} vet
GOLINT=golangci-lint run

BUILD_DIR=./bin/
BIN=${BUILD_DIR}attender

all: build

deps: go.mod go.sum
	${GOCMD} mod download

clean:
	${GOCMD} clean
	rm -rf ${BUILD_DIR}

test:
	${GOTEST} -v ./...

fmt:
	${GOFMT} ./...

vet:
	${GOVET} ./...

lint:
	${GOLINT}

build:
	${GOBUILD} -o ${BIN} .

release:
	${GOBUILD} -o ${BIN} -ldflags="-s -w" -trimpath .

run: build
	${BIN}

docker:
	docker build -t attender:dev .
