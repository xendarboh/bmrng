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

WORKDIR /opt/trellis


# pre-copy/cache go.mod for pre-downloading dependencies
# and only redownloading them in subsequent builds if they change

COPY pb/go.mod pb/go.sum ./pb/
RUN cd pb && go mod download && go mod verify

COPY go.mod go.sum ./
RUN go mod download && go mod verify


# build and install mcl; use docker build cache
COPY ./crypto/pairing/mcl/scripts ./crypto/pairing/mcl/scripts
RUN ./crypto/pairing/mcl/scripts/install-deps.sh \
  && ldconfig

COPY . .

# generate code from protocol buffer files
RUN cd pb && buf generate

# build trellis; server, client, coordinator
RUN true \
  && cd cmd/server \
    && go install \
    && go build \
  && cd ../client \
    && go install \
    && go build \
  && cd ../coordinator \
    && go install \
    && go build

RUN cd cmd/xtrellis \
  && go install \
  && go build
