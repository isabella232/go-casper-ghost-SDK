test:
	go test -cover -race ./...

#spec_test_minimal:
#	./scripts/download-spec-tests.sh v0.12.3 minimal

spec_test_mainnet:
	./scripts/download-spec-tests.sh v0.12.3 mainnet
	go test ./src/state_transition/spec_tests/...
	rm -r ./src/state_transition/spec_tests/.temp

generate_proto:
	find . -type f -name '*.pb.go' -delete
	${info "make sure you have protoc-go-gen v1.3.5 ONLY!"}
	protoc -I=${GOPATH}/src -I=./ --gofast_out=./src/core ./src/core/*.proto
	sszgen --path ./src/core/types.pb.go --output ./src/core/types_generated.pb.go --include ./src/core/block.pb.go,./src/core/attestation.pb.go
	sszgen --path ./src/core/block.pb.go --output ./src/core/block_generated.pb.go --include ./src/core/attestation.pb.go
	sszgen --path ./src/core/attestation.pb.go --output ./src/core/attestation_generated.pb.go

build:
	go build ./...

