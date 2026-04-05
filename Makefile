build:
	CGO_ENABLED=0 go build -o reservation ./cmd/reservation/

run: build
	./reservation

test:
	go test ./...

clean:
	rm -f reservation

.PHONY: build run test clean
