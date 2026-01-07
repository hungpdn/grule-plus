APP_NAME := grule-plus

.PHONY: test test-coverage test-coverage-html lint benchmark protobuf clean

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

protobuf:
	protoc --go_out=. --go_opt=paths=source_relative examples/protobuf/discount.proto

clean:
	rm -rf coverage.out