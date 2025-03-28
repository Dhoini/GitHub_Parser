PROTO_PATH=internal/infrastructure/api/proto
PROTO_FILES=$(PROTO_PATH)/github_parser.proto

.PHONY: proto
proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		$(PROTO_FILES)

.PHONY: build
build:
	go build -o bin/server ./cmd/server

.PHONY: run
run:
	go run ./cmd/server

.PHONY: docker-build
docker-build:
	docker-compose build

.PHONY: docker-up
docker-up:
	docker-compose up -d

.PHONY: docker-down
docker-down:
	docker-compose down

.PHONY: test
test:
	go test -v ./...

.PHONY: lint
lint:
	golangci-lint run