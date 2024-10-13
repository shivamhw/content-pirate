clean:
	- rm reddit-pirate
build: clean
	CGO_ENABLED=0 go build
run: build
	./reddit-pirate scrape

build-arm: clean
	GOOS=linux GOARCH=arm64 go build
