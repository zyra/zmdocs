APP_VERSION := $(shell git tag | tail -1)

.PHONY: build
build: build_linux build_windows build_darwin ; @echo "Done building!"

build_linux: ; @\
GOOS=linux GOARCH=amd64 go build -mod vendor -ldflags "-X main.AppVersion=${APP_VERSION}" -o zmdocs_linux_amd64 cmd/docs/main.go && \
chmod +x zmdocs_linux_amd64

build_windows: ; @\
GOOS=windows GOARCH=amd64 go build -mod vendor -ldflags "-X main.AppVersion=${APP_VERSION}" -o zmdocs_windows_amd64.exe cmd/docs/main.go

build_darwin: ; @\
GOOS=darwin GOARCH=amd64 go build -mod vendor -ldflags "-X main.AppVersion=${APP_VERSION}" -o zmdocs_darwin_amd64 cmd/docs/main.go && \
chmod +x zmdocs_darwin_amd64