.PHONY: build clean

build:
	go build -o kocproxy main.go

windows:
	GOOS=windows GOARCH=amd64 go build -o kocproxy.exe main.go

linux:
	GOOS=linux GOARCH=amd64 go build -o kocproxy main.go

clean:
	rm ./kocproxy
