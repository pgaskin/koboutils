.PHONY: default
default: clean test build

ver := $(shell git describe --tags --always --dirty)
ldflags := -X main.version=$(ver)

.PHONY: clean test build
clean:
	rm -rfv build

test:
	go test ./...

build:
	go build -ldflags "$(ldflags)" -o "build/kobo-find" ./kobo-find
	go build -ldflags "$(ldflags)" -o "build/kobo-info" ./kobo-info

release: clean
	go get github.com/tcnksm/ghr
	go get github.com/mholt/archiver/cmd/archiver
	
	GOOS=windows GOARCH=386 go build -ldflags "$(ldflags)" -o "build/kobo-find_windows.exe" ./kobo-find
	GOOS=linux GOARCH=amd64 go build -ldflags "$(ldflags)" -o "build/kobo-find_linux-x64" ./kobo-find
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(ldflags)" -o "build/kobo-find_darwin-x64" ./kobo-find

	GOOS=windows GOARCH=386 go build -ldflags "$(ldflags)" -o "build/kobo-info_windows.exe" ./kobo-info
	GOOS=linux GOARCH=amd64 go build -ldflags "$(ldflags)" -o "build/kobo-info_linux-x64" ./kobo-info
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(ldflags)" -o "build/kobo-info_darwin-x64" ./kobo-info

	GOOS=windows GOARCH=386 go build -ldflags "$(ldflags)" -o "build/kobo-versionextract_windows.exe" ./kobo-versionextract
	GOOS=linux GOARCH=amd64 go build -ldflags "$(ldflags)" -o "build/kobo-versionextract_linux-x64" ./kobo-versionextract
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(ldflags)" -o "build/kobo-versionextract_darwin-x64" ./kobo-versionextract

	mkdir build/dist
	archiver make build/dist/koboutils_windows.zip build/*_windows.exe
	archiver make build/dist/koboutils_linux-x64.tar.gz build/*_linux-x64
	archiver make build/dist/koboutils_darwin-x64.tar.gz build/*_darwin-x64

ifdef GITHUB_TOKEN
	ghr -delete $(ver) build/dist
endif
