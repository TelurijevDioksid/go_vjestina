build:
	@go build -o bin/gogasprices

run: build
	@./bin/gogasprices

test:
	@go test -v ./...

