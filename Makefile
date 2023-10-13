install-deps-osx:
	brew install protobuf gmp cmake openssl

install deps-ubuntu:
	sudo apt install -y protobuf-compiler libgmp-dev cmake libssl-dev

gen-proto:
	@echo "Generating protobuf files"
	(cd api && buf generate)
	@echo "Generating protobuf done."

build-mcl:
	@echo "Building MCL..."
	./go/trellis/crypto/pairing/mcl/scripts/install-deps.sh
	@echo "Building MCL done."

build-commands:
	( cd go/trellis/cmd/server && go install && go build )
	( cd go/trellis/cmd/client && go install && go build )
	( cd go/trellis/cmd/coordinator && go install && go build )
	( cd go/0kn/cmd/xtrellis && go install && go build )

clean:
	git clean -X -f
