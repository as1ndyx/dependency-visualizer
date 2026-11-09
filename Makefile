.PHONY: run build test fmt vet demo demo-config clean

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
	./scripts/run_stage2.sh

demo-config:
	./scripts/run_stage1.sh

clean:
	rm -rf bin
