.PHONY: build build-linux build-windows build-mac clean fmt vet run

APP_NAME := edonish-auto
VERSION := 0.2.0
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"

build:
        go build $(LDFLAGS) -o $(APP_NAME) .

build-linux:
        CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(APP_NAME)-linux-amd64 .

build-windows:
        CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(APP_NAME).exe .

build-mac:
        CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(APP_NAME)-mac-amd64 .

build-arm:
        CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(APP_NAME)-linux-arm64 .

clean:
        rm -f $(APP_NAME) $(APP_NAME)-* $(APP_NAME).exe

fmt:
        gofmt -w .
        goimports -w .

vet:
        go vet ./...

run:
        go run .

docker:
        docker build -t $(APP_NAME):$(VERSION) .
