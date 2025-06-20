clean:
	- rm content-pirate
	- rm -rf download
build: clean
	CGO_ENABLED=0 go build
run: build
	./content-pirate scrape

build-arm: clean
	GOOS=linux GOARCH=arm64 go build

restore:
	git restore vendor/github.com/vartanbeno/go-reddit/v2/reddit/things.go 