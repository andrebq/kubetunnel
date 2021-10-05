.PHONY: test generate dist

include Runner.mk

dist: generate
	go build -o ./dist/kubetunnel ./cmd/kubetunnel

test: generate
	go test ./...

generate: internal/protocol/protocol_grpc.pb.go
	go build ./...

internal/protocol/protocol_grpc.pb.go: internal/protocol/protocol.proto
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		internal/protocol/protocol.proto
