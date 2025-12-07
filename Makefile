.PHONY: build run test clean tidy

build:
	go build -o bin/img2ppt cmd/server/main.go

run:
	export https_proxy=http://127.0.0.1:7890 http_proxy=http://127.0.0.1:7890 all_proxy=socks5://127.0.0.1:7890 && go run cmd/server/main.go

test:
	go test -v ./...

clean:
	rm -rf bin/ output/

tidy:
	go mod tidy
