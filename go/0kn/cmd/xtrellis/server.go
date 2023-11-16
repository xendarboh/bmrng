package main

import (
	"log"

	"github.com/31333337/bmrng/go/trellis/config"
	"github.com/31333337/bmrng/go/trellis/errors"
	"github.com/31333337/bmrng/go/trellis/network"
	"github.com/31333337/bmrng/go/trellis/server"
)

func LaunchServer(args ArgsServer) {
	switch {
	case args.Config != nil:
		runServerConfigGenerator(args)
	default:
		runServer(args)
	}
}

// from trellis/cmd/server/
func runServer(args ArgsServer) {
	serversFile := args.ServerFile
	groupsFile := args.GroupFile
	addr := args.Addr
	errors.Addr = addr

	log.Printf("Launching server with address %s", addr)

	servers, err := config.UnmarshalServersFromFile(serversFile)
	if err != nil {
		log.Fatalf("Could not read servers file %s", serversFile)
	}

	groups, err := config.UnmarshalGroupsFromFile(groupsFile)
	if err != nil {
		log.Fatalf("Could not read group file %s", groupsFile)
	}

	// will start in blocked state
	h := server.NewHandler()
	server := server.NewServer(&config.Servers{Servers: servers}, &config.Groups{Groups: groups}, h, addr)

	server.TcpConnections.LaunchAccepts()
	network.RunServer(h, server, servers, addr)
	config.Flush()
}

func runServerConfigGenerator(args ArgsServer) {
	addr := args.Addr
	serverPrivateFile := args.Config.ServerPrivateFile
	serverPublicFile := args.Config.ServerPublicFile

	log.Printf("Creating server config for address %s", addr)

	// create a new server config with a self-signed certificate
	servers := make(map[int64]*config.Server)
	id := int64(0)
	cert, key := config.CreateCertificate(addr)
	servers[id] = config.CreateServerWithCertificate(addr, id, cert, key)

	// write complete (public and private) server config to file
	err := config.MarshalServersToFile(serverPrivateFile, servers)
	if err != nil {
		log.Fatalf("Could not write private server file %s: %v", serverPrivateFile, err)
	}

	// write public server config to file
	servers[id].PrivateKey = nil
	servers[id].PrivateIdentity = nil
	servers[id].SignatureKey = nil
	err = config.MarshalServersToFile(serverPublicFile, servers)
	if err != nil {
		log.Fatalf("Could not write public server file %s: %v", serverPublicFile, err)
	}
}
