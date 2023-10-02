## Network Data Flow

The Gateway:

- receives incoming data streams and packetizes them into messages for the mix-net
- serves messages for mix-net clients to retrieve and send through the mix-net
- receives messages leaving the mix-net
- serves outgoing data streams reassembled from mix-net messages

```mermaid
flowchart LR
    subgraph mixnet ["mix-net"]
    direction LR

    subgraph L1 ["L 1"]
        direction LR
        X[" "]
        Y[" "]
        Z[" "]
    end

    subgraph L2 ["..."]
        direction LR
        X1[" "]
        Y1[" "]
        Z1[" "]
    end

    subgraph L3 ["L N"]
        direction LR
        X2[" "]
        Y2[" "]
        Z2[" "]
    end
end

subgraph gateway1 ["gateway"]
    subgraph proxyIn ["In"]
    end
end

subgraph gateway2 ["gateway"]
    subgraph proxyOut ["Out"]
    end
end

L1 --> L2 --> L3
proxyIn --> mixnet --> proxyOut
```
