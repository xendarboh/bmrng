# 0KN

## Development

### Documentation

- [Architecture](/docs/README.md#architecture)
- [Protocol](/docs/README.md#protocol)
- [Trellis](/go/trellis/README.md)
- [API](/api/README.md)

### Project Structure

The basic structure of this project:

```
.
├── api/                            # Protocol Buffers
├── docker/
│   ├── base
│   └── remote-network-simulation
├── docs/                           # Documentation
├── go/                             # Go Modules
│   ├── 0kn/                        # 0KN; integrated launcher, libs
│   └── trellis/                    # Trellis
└── scripts/
    ├── go-workspace-init.sh        # init go workspace for local dev
    ├── simulate-remote-network.sh  # run remote network simulation
    ├── test-gateway-ci.sh          # full gateway test
    ├── test-gateway-io.sh          # test gateway I/O
    └── test-gateway-pipe.sh        # test gateway pipe
```

### Prerequisites

- [go](https://go.dev/doc/install) `>= 1.21`
- [Trellis dependencies](/go/trellis/README.md#dependencies)

Optional; for generating code from Protocol Buffers:

- [Protocol Buffer Compiler](https://grpc.io/docs/protoc-installation/)
- [Buf](https://buf.build/docs/installation)

Utilities used by test scripts:

- netcat
- pkill
- wget

### Build

Prepare go workspace:

```sh
make init
```

(Optional) Generate code from Protocol Buffers:

```sh
make protobuf
```

Build:

```sh
make build
```

### Test

```sh
make test
```

### Run

Typical invocation involves sub processes and various config files within a
working directory so installation is necessary. The following is an example:

```sh
# build and install executable(s) to directory `go env GOBIN`
make install

# optional: set working directory; see Environment Variables

# `xtrellis` should now be in your PATH
xtrellis --help

# run complete gateway test run by ci
./scripts/test-gateway-ci.sh

# remove installed executable(s)
make uninstall
```

### Configure

#### Environment Variables

- `_0KN_WORKDIR` runtime working directory; default = `~/.0KN`

### E2E Tests

#### Full Automated Test

```sh
./scripts/test-gateway-ci.sh
```

#### With Local Mix-Net

1. Run a coordinated local mix-net with gateway enabled, for example:

   ```sh
   xtrellis coordinator mixnet --gatewayenable --debug
   ```

   `CTRL-C` to exit.

2. Then, in a separate terminal (from project root):

   Send `100KB` random data through the mix-net and compare data in/out:

   ```sh
   ./scripts/test-gateway-io.sh 102400
   ```

   Pipe generic data through the mix-net:

   ```sh
   cat in.png | ./scripts/test-gateway-pipe.sh > out.png
   ```

#### With Docker Compose

```sh
cd docker/base

# build and run container for gateway test
docker compose --profile test-gateway up --build --abort-on-container-exit

# remove container
docker compose --profile test-gateway down
```

### Local Remote Network Simulator

```sh
# build docker images
make docker-images

# run remote network simulator
./scripts/simulate-remote-network.sh

# optional: ssh into coordinator's container
cd docker/remote-network-simulation
./coordinator-ssh.sh

# remove containers
make docker-clean
```
