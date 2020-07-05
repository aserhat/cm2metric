.PHONY: lint test build

version=$(shell cat VERSION)
commit=$(shell git rev-parse --short HEAD)
gcflags=all=-trimpath=${PWD}
asmflags=all=-trimpath=${PWD}
ldflags="-s -w"
flags=-gcflags=${gcflags} -asmflags=${asmflags} -ldflags=${ldflags}

build:
	CGO_ENABLED=0 go build ${flags} -o ./build/package/cm2metric ./cmd/cm2metric/cm2metric.go

contain:
	docker build -t docker.pkg.github.com/aserhat/cm2metric/cm2metric:${version} -f build/package/Dockerfile .

push:
	docker push docker.pkg.github.com/aserhat/cm2metric/cm2metric:${version}