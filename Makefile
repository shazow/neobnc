BINARY = neobnc
BIND = "localhost:6668"

all: $(BINARY)

$(BINARY): *.go
	go build -ldflags "-X main.version=`git describe --long --tags --dirty --always`" .

.PHONY: deps build clean run debug test

deps:
	go get ./...

build: $(BINARY)

clean:
	rm $(BINARY)

run: $(BINARY)
	./$(BINARY) --bind $(BIND) -vv

debug: $(BINARY)
	./$(BINARY) --pprof 6060 -vv

test:
	go test ./...
	golint ./...
