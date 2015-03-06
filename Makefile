BINARY = neobnc

all: $(BINARY)

$(BINARY): *.go
	go build -ldflags "-X main.version `git describe --long --tags --dirty --always`" .

deps:
	go get .

build: $(BINARY)

clean:
	rm $(BINARY)

run: $(BINARY)
	./$(BINARY) -vv

debug: $(BINARY)
	./$(BINARY) --pprof 6060 -vv

test:
	go test ./...
	golint ./...
