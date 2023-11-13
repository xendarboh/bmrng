.PHONY: build
build: init
	cd go/0kn/cmd/xtrellis && go build

.PHONY: install
install: init
	cd go/0kn/cmd/xtrellis && go install

.PHONY: uninstall
uninstall:
	rm -f $(shell go env GOBIN)/xtrellis

.PHONY: install-deps-osx
install-deps-osx:
	brew install protobuf gmp cmake openssl

.PHONY: install-deps-ubuntu
install-deps-ubuntu:
	sudo apt install -y protobuf-compiler libgmp-dev cmake libssl-dev netcat

.PHONY: install-deps-mcl
install-deps-mcl:
	./go/trellis/crypto/pairing/mcl/scripts/install-deps.sh

.PHONY: init
init:
	./scripts/go-workspace-init.sh

.PHONY: protobuf
protobuf:
	cd api && buf generate

.PHONY: test-go-0kn
test-go-0kn:
	go test ./go/0kn/...

.PHONY: test-go-trellis
test-go-trellis:
	go test -skip 'TestMarshalZero|TestKeyExchange' ./go/trellis/...

.PHONY: test
test: test-go-0kn test-go-trellis

.PHONY: clean
clean:
	git clean -X -f

.PHONY: very-clean
very-clean: clean uninstall

.PHONY: docker-images
docker-images:
	docker compose --project-directory docker/base/ --profile build build
	docker compose --project-directory docker/remote-network-simulation/ --profile build build

.PHONY: docker-clean
docker-clean:
	docker compose --project-directory docker/base --profile test-gateway down
	docker compose --project-directory docker/remote-network-simulation --profile run down

# source env vars
include docker/base/.env
.PHONY: docker-very-clean
docker-very-clean: docker-clean
	docker rmi -f ${IMG_REPO}/${IMG_NAME}:${IMG_TAG}
	docker rmi -f ${IMG_REPO}/${IMG_NAME}-remote:${IMG_TAG}
