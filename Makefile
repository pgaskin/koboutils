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
	
	GOOS=windows GOARCH=386 go build -ldflags "$(ldflags)" -o "build/kobo-find_windows.exe" ./kobo-find
	GOOS=linux GOARCH=amd64 go build -ldflags "$(ldflags)" -o "build/kobo-find_linux-x64" ./kobo-find
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(ldflags)" -o "build/kobo-find_darwin-x64" ./kobo-find

	GOOS=windows GOARCH=386 go build -ldflags "$(ldflags)" -o "build/kobo-info_windows.exe" ./kobo-info
	GOOS=linux GOARCH=amd64 go build -ldflags "$(ldflags)" -o "build/kobo-info_linux-x64" ./kobo-info
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(ldflags)" -o "build/kobo-info_darwin-x64" ./kobo-info

#ifdef GITHUB_TOKEN
#	github-release delete --user geek1011 --repo koboutils --tag latest >/dev/null 2>/dev/null || true
#	github-release release --user geek1011 --repo koboutils --tag latest --name "$(shell date +%Y%m%d)" --description "$(ver)"
#
#	github-release upload --user geek1011 --repo koboutils --tag latest --name "kobo-find_windows.exe" --file "kobo-find_windows.exe" --replace
#	github-release upload --user geek1011 --repo koboutils --tag latest --name "kobo-find_linux-x64" --file "kobo-find_linux-x64" --replace
#	github-release upload --user geek1011 --repo koboutils --tag latest --name "kobo-find_darwin-x64" --file "kobo-find_darwin-x64" --replace
#
#	github-release upload --user geek1011 --repo koboutils --tag latest --name "kobo-info_windows.exe" --file "kobo-info_windows.exe" --replace
#	github-release upload --user geek1011 --repo koboutils --tag latest --name "kobo-info_linux-x64" --file "kobo-info_linux-x64" --replace
#	github-release upload --user geek1011 --repo koboutils --tag latest --name "kobo-info_darwin-x64" --file "kobo-info_darwin-x64" --replace
#endif