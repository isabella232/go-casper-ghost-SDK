test:
	go test -cover -race ./...

generate_proto:
	find . -type f -name '*.pb.go' -delete
	${info "make sure you have protoc-go-gen v1.3.5 ONLY!"}
	protoc -I=${GOPATH}/src -I=./ --gofast_out=./src/core ./src/core/*.proto
	sszgen --path ./src/core/ --objs=HistoricalBatch,ForkData,State --output ./src/core/generated.pb.go

build:
	go build ./...
