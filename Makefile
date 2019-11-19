BINARY = qbot
REPO = github.com/sleibrock/qbot

SRCS = %.go
VERSION = $(shell git describe --tags --dirty --always 2> /dev/null || echo "dev")
LDFLAGS = -X main.Version=$(VERSION) -extldflags "-static"

all: $(BINARY)

$(BINARY): *.go
	go build $(BUILDFLAGS) qbot.go

clean:
	rm $(BINARY)

release:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 LDFLAGS='$(LDFLAGS)' ./build_release "$(REPO)" README.md LICENSE
	CGO_ENABLED=0 GOOS=linux GOARCH=386 LDFLAGS='$(LDFLAGS)' ./build_release "$(REPO)" README.md LICENSE
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=6 LDFLAGS='$(LDFLAGS)' ./build_release "$(REPO)" README.md LICENSE
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 LDFLAGS='$(LDFLAGS)' ./build_release "$(REPO)" README.md LICENSE
	CGO_ENABLED=0 GOOS=freebsd GOARCH=amd64 LDFLAGS='$(LDFLAGS)' ./build_release "$(REPO)" README.md LICENSE
	CGO_ENABLED=0 GOOS=windows GOARCH=386 LDFLAGS='$(LDFLAGS)' ./build_release "$(REPO)" README.md LICENSE
