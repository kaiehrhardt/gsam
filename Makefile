generate:
	templ generate

build:
	go build

run:
	./glabsa2

full: generate build run
