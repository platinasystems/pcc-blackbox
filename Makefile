build: build-linux

build-linux:
			CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go test -c
