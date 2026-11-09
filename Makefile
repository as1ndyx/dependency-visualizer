.PHONY: run build test fmt vet demo clean

run:
	go run ./src

build:
	go build -o bin/depgraph ./src

test:
	go test ./...

fmt:
	gofmt -w src tests

vet:
	go vet ./...

demo:
	./scripts/run_stage1.sh

clean:
	rm -rf bin
