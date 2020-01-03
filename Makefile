BIN_DIR=bin
MODULE=github.com/maybetheresloop/charter-go

.PHONY: all
all: charterd charter-pw

charterd:
	go build -o bin/$@ -v ${MODULE}/cmd/daemon

charter-pw:
	go build -o bin/$@ -v ${MODULE}/cmd/pw

.PHONY: clean test cov

clean:
	rm -rf bin/

test:
	go test ${MODULE}/...

cov:
	go test ${MODULE}/... --coverprofile coverage.txt --covermode atomic