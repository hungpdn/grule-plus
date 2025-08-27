APP_NAME := grule-plus

.PHONY: test test-coverage test-coverage-html lint benchmark clean

test:
	go test -v ./...

test-coverage:
	go test -v ./... -coverprofile=coverage.out

test-coverage-html: test-coverage
	go tool cover -html=coverage.out

lint:
	golangci-lint run

benchmark:
    go test -bench=. -benchmem ./benchmark/...

clean:
	rm -rf coverage.out