
bin/networkestop: go.mod *.go cmd/module/*.go
	go build -o bin/networkestop cmd/module/cmd.go


lint:
	gofmt -s -w .

updaterdk:
	go get go.viam.com/rdk@latest
	go mod tidy

test:
	go test

module: bin/networkestop
	tar czf module.tar.gz bin/networkestop

all: test module 


