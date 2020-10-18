test:
	go test -cover -race ./...

generate_proto:
	${info "make sure you have protoc-go-gen v1.3.5 ONLY!"}
	protoc -I=${GOPATH}/src -I=./ --gofast_out=./src/core ./src/core/*.proto

build:
	go build ./...
