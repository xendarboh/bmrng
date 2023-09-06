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
    wget \
  && rm -rf /v \ar/lib/apt/lists/*

# install go
ARG VERSION_GO=1.17.3
RUN export F="go${VERSION_GO}.linux-amd64.tar.gz" \
  && wget https://golang.org/dl/${F} \
  && tar -C /usr/local -xzf ${F} \
  && rm ${F}

# update PATH with go/bin
ENV PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/local/go/bin

WORKDIR /opt/trellis

# pre-copy/cache go.mod for pre-downloading dependencies
# and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

# build and install mcl
RUN ./crypto/pairing/mcl/scripts/install-deps.sh \
  && ldconfig

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
