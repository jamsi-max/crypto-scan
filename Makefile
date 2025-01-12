build:
	go build -o bin/arbitrage

run: build
	./bin/arbitrage

test:
	go test -v ./...
	