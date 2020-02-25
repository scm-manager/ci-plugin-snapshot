APP:=ci-plugin-snapshot
VERSION:=1.1.5

build: dependencies
	GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o ${APP}

dependencies:
	GO111MODULE=on go get

image:
	docker build -t scm-manager/${APP}:${VERSION} .

deploy: image
	docker push scm-manager/${APP}:${VERSION}
