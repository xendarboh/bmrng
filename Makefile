install-deps-osx:
	brew intall protobuf gmp cmake openssl

install deps-ubuntu:
	sudo apt install -y protobuf-compiler libgmp-dev cmake libssl-dev

gen-proto:
	@echo "Generating protobuf files"
	(cd pb && buf generate)
	@echo "Generating protobuf done."

build-mcl:
	@echo "Building MCL..."
	@cd mcl && make
	@echo "Building MCL done."

build-protobuf:
	@echo "Building protobuf..."
	@cd protobuf && make
	@echo "Building protobuf done."

build-commands:
	( cd cmd/server && go install && go build )
	( cd cmd/client && go install && go build )
	( cd cmd/coordinator && go install && go build )