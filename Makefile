VERSION := $(shell cat VERSION.txt)
COMMIT  := $(shell git rev-parse --short HEAD)
DATE    := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

build:
	go build -ldflags "\
	-X 'github.com/Ataraxxia/godin/internal/version.Version=$(VERSION)' \
	-X 'github.com/Ataraxxia/godin/internal/version.Commit=$(COMMIT)' \
	-X 'github.com/Ataraxxia/godin/internal/version.Date=$(DATE)'"
