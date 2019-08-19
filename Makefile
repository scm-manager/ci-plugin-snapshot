VERSION:=1.1.0

build: dependencies
	GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o scm-plugin-snapshot

dependencies:
	GO111MODULE=on go get

image:
	docker build -t cloudogu/scm-plugin-snapshot:${VERSION} .

deploy: image
	docker push cloudogu/scm-plugin-snapshot:${VERSION}
