.PHONY: run build test fmt vet demo demo-graph demo-config demo-maven clean

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
	./scripts/run_stage4.sh

demo-graph:
	./scripts/run_stage3.sh

demo-maven:
	./scripts/run_maven_graph.sh

demo-config:
	./scripts/run_stage1.sh

clean:
	rm -rf bin
