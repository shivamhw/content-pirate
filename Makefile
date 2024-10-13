clean:
	rm reddit-pirate
build: clean
	go build
run: build
	./reddit-pirate scrape
