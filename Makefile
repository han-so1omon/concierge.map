gen:
	- go generate ./...

dev:
	- air

build:
	- go build -a -gcflags='' 

start:
	- go run main.go

lint:
	- go vet
	- golint

gen_keys:
	go run tools/gen_keys.go

.PHONY: gen dev build start lint gen_keys
