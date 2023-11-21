package conf

import (
	"errors"
	"fmt"

	"github.com/31333337/bmrng/go/trellis/config"
	"github.com/31333337/bmrng/go/trellis/network"
)

// Generates a new server config for the given address and save it to the given files.
// If the server already exists in the config, it is overwritten.
func LocalServerConfigSet(addr string, serverPrivateFile string, serverPublicFile string) error {
	// load private server config, if it exists
	servers, err := config.UnmarshalServersFromFile(serverPrivateFile)
	if err != nil {
		servers = make(map[int64]*config.Server)
	}

	// find server id by address in config, or get next available id
	id, _ := network.FindConfig(addr, servers)
	if id < 0 {
		id = int64(len(servers))
	}

	// create a new server config with a self-signed certificate
	cert, key := config.CreateCertificate(addr)
	servers[id] = config.CreateServerWithCertificate(addr, id, cert, key)

	// write complete (public and private) server config to file
	if err := config.MarshalServersToFile(serverPrivateFile, servers); err != nil {
		return errors.New(
			fmt.Sprintf("Could not write private server file %s: %v", serverPrivateFile, err),
		)
	}

	// save public server config to file without private info
	for id := range servers {
		// servers[id].PrivateKey = nil // !!! TODO/WIP: client server fails if not public
		servers[id].PrivateIdentity = nil
		servers[id].SignatureKey = nil
	}
	if err := config.MarshalServersToFile(serverPublicFile, servers); err != nil {
		return errors.New(
			fmt.Sprintf("Could not write public server file %s: %v", serverPublicFile, err),
		)
	}

	return nil
}
