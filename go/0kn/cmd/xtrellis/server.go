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
	serverPrivateFile := args.ServerPrivateFile
	errors.Addr = addr

	log.Printf("Launching server with address %s", addr)

	// load public server config for mix-net round
	servers, err := config.UnmarshalServersFromFile(serversFile)
	if err != nil {
		log.Fatalf("Could not read servers file %s: %v", serversFile, err)
	}

	// load private server config
	serversPrivate, err := config.UnmarshalServersFromFile(serverPrivateFile)
	if err != nil {
		log.Fatalf("Could not read private servers file %s: %v", serverPrivateFile, err)
	}

	// find server id by address in public config
	id, _ := network.FindConfig(addr, servers)
	if id < 0 {
		log.Fatalf("Could not find %s in servers file", addr)
	}

	// find server config by address in private config
	_, cfg := network.FindConfig(addr, serversPrivate)
	if id < 0 {
		log.Fatalf("Could not find %s in private servers file", addr)
	}

	// replace public config with private (complete) config
	servers[id] = cfg

	groups, err := config.UnmarshalGroupsFromFile(groupsFile)
	if err != nil {
		log.Fatalf("Could not read group file %s: %v", groupsFile, err)
	}

	// will start in blocked state
	h := server.NewHandler()
	server := server.NewServer(
		&config.Servers{Servers: servers},
		&config.Groups{Groups: groups},
		h,
		addr,
	)

	server.TcpConnections.LaunchAccepts()
	network.RunServer(h, server, servers, addr)
	config.Flush()
}

func runServerConfigGenerator(args ArgsServer) {
	addr := args.Addr
	serverPrivateFile := args.ServerPrivateFile
	serverPublicFile := args.ServerPublicFile

	var servers map[int64]*config.Server

	log.Printf("Creating server config for address %s", addr)

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
		log.Fatalf("Could not write private server file %s: %v", serverPrivateFile, err)
	}

	// save public server config to file without private info
	for id := range servers {
		servers[id].PrivateKey = nil
		servers[id].PrivateIdentity = nil
		servers[id].SignatureKey = nil
	}
	if err := config.MarshalServersToFile(serverPublicFile, servers); err != nil {
		log.Fatalf("Could not write public server file %s: %v", serverPublicFile, err)
	}
}
