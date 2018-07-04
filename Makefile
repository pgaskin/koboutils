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
	go get github.com/aktau/github-release
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

	archiver make build/koboutils_windows.zip build/*_windows.exe
	archiver make build/koboutils_linux-x64.tar.gz build/*_linux-x64
	archiver make build/koboutils_darwin-x64.tar.gz build/*_darwin-x64

ifdef GITHUB_TOKEN
	github-release delete --user geek1011 --repo koboutils --tag $(ver) >/dev/null 2>/dev/null || true
	github-release release --user geek1011 --repo koboutils --tag $(ver) --name "$(shell date +%Y-%m-%d)" --description "$(ver)"

	github-release upload --user geek1011 --repo koboutils --tag $(ver) --name "koboutils_windows.zip" --file "koboutils_windows.zip" --replace
	github-release upload --user geek1011 --repo koboutils --tag $(ver) --name "koboutils_linux-x64.tar.gz" --file "koboutils_linux-x64.tar.gz" --replace
	github-release upload --user geek1011 --repo koboutils --tag $(ver) --name "koboutils_darwin-x64.tar.gz" --file "koboutils_darwin-x64.tar.gz" --replace
endif