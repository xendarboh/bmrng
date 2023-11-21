FROM ubuntu:22.04

# use the "noninteractive" debconf frontend
ENV DEBIAN_FRONTEND noninteractive

# use bash for RUN commands
SHELL ["/bin/bash", "--login", "-c"]

# install things
RUN apt update \
  && apt install --no-install-recommends -y -q \
    build-essential \
    ca-certificates \
    cmake \
    git \
    libgmp-dev \
    libssl-dev \
    netcat \
    sudo \
    unzip \
    wget \
  && rm -rf /v \ar/lib/apt/lists/*

# install go
ARG VERSION_GO=1.21.1
RUN export F="go${VERSION_GO}.linux-amd64.tar.gz" \
  && wget https://golang.org/dl/${F} \
  && tar -C /usr/local -xzf ${F} \
  && rm ${F}

# update PATH with go/bin
ENV PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/local/go/bin:/root/go/bin

# install protocol buffers
ARG VERSION_PB=latest
RUN cd /tmp \
  && if [ "${VERSION_PB}" = "latest" ]; then \
    # get the latest release
    export V=$( \
      wget -O - -q https://api.github.com/repos/protocolbuffers/protobuf/releases/latest \
      | sed -n -e 's/"tag_name": "\(.*\)",/\1/p' \
      | sed -e 's/^.*v//' \
    ) \
  ; else \
    export V=${VERSION_PB} \
  ; fi \
  && export F="protoc-${V}-linux-x86_64.zip" \
  && wget "https://github.com/protocolbuffers/protobuf/releases/download/v${V}/${F}" \
  && unzip ${F} \
    -x readme.txt \
    -d /usr/local \
  && rm -f ${F}

# install buf protocol buffer tool
ARG VERSION_BUF=latest
RUN go install github.com/bufbuild/buf/cmd/buf@${VERSION_BUF} \
  && go clean --cache

WORKDIR /src


# pre-copy and cache downloaded go modules and other dependencies
# and only redownloading them in subsequent builds if they change

# build and install mcl
COPY go/trellis/crypto/pairing/mcl/scripts ./go/trellis/crypto/pairing/mcl/scripts
RUN ./go/trellis/crypto/pairing/mcl/scripts/install-deps.sh \
  && ldconfig

COPY api/go.* ./api/
COPY go/0kn/go.* ./go/0kn/
COPY go/trellis/go.* ./go/trellis/

RUN cd api && go mod download && go mod verify
RUN cd go/0kn && go mod download && go mod verify
RUN cd go/trellis && go mod download && go mod verify


COPY . .

RUN \
  # go install to /usr/local/bin
  go env -w GOBIN=/usr/local/bin \
  # setup go workspace
  && make init \
  # generate code from protocol buffer files
  && make protobuf \
  # build and install binaries
  && make install
