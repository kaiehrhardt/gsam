generate:
	templ generate

build:
	go build -o app

run:
	./app

full: generate build run
