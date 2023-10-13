install-deps-osx:
	brew install protobuf gmp cmake openssl

install deps-ubuntu:
	sudo apt install -y protobuf-compiler libgmp-dev cmake libssl-dev

gen-proto:
	@echo "Generating protobuf files"
	(cd mods/proto && buf generate)
	@echo "Generating protobuf done."

build-mcl:
	@echo "Building MCL..."
	./mods/trellis/crypto/pairing/mcl/scripts/install-deps.sh
	@echo "Building MCL done."

build-commands:
	( cd mods/trellis/cmd/server && go install && go build )
	( cd mods/trellis/cmd/client && go install && go build )
	( cd mods/trellis/cmd/coordinator && go install && go build )
