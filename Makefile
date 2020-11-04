GPATH=httpd/main.go

build:
	go build -ldflags "-s -w" -o bin/main $(GPATH)

run:
	go run $(GPATH)