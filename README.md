# WikiCrawl

Tool for crawling a wiki and accessing its health.

The only current use case is to find broken links but this may be expanded.

## Dependencies

    go get golang.org/x/net/html
    go get github.com/Sirupsen/logrus

## Testing

    go test jalandis.com/wikicrawl

### Test Coverage

    go test -cover jalandis.com/wikicrawl

    go test -coverprofile=coverage.out jalandis.com/wikicrawl
    go tool cover -html=coverage.out

## Linting

    gofmt -w jalandis.com/wikicrawl/crawler.go
    gofmt -w jalandis.com/wikicrawl/crawler_test.go
    gofmt -s -w jalandis.com/wikicrawl/crawler.go
    gofmt -s -w jalandis.com/wikicrawl/crawler_test.go

## Documentation

    godoc -http=:6060

## Execute

    go run jalandis.com/wikicrawl/cli/cli.go --wiki http://wiki-url

## Git Hooks

[Pre commit check](https://golang.org/misc/git/pre-commit)
