GPATH=httpd/main.go

build:
	go build -o bin/main $(GPATH)

run:
	go run $(GPATH)