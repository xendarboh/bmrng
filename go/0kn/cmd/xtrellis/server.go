package main

import (
	"log"

	"github.com/31333337/bmrng/go/0kn/internal/conf"
	"github.com/31333337/bmrng/go/0kn/pkg/utils"
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

	logger := utils.GetLogger()
	sugar := logger.Sugar()
	defer sugar.Sync()
	sugar.Infow(
		"Launching server",
		"address %s", addr,
	)

	servers, err := config.UnmarshalServersFromFile(serversFile)
	if err != nil {
		sugar.Fatalf("Could not read servers file %s", serversFile)
	}

	// find server id by address in public config
	id, _ := network.FindConfig(addr, servers)
	if id < 0 {
		sugar.Fatalf("Could not find %s in servers file", addr)
	}

	// load private server config
	serversPrivate, err := config.UnmarshalServersFromFile(serverPrivateFile)
	if err != nil {
		sugar.Fatalf("Could not read private servers file %s: %v", serverPrivateFile, err)
	}

	// find server config by address in private config
	_, cfg := network.FindConfig(addr, serversPrivate)
	if cfg == nil {
		sugar.Fatalf("Could not find %s in private servers file", addr)
	}

	// replace public config with private (complete) config
	cfg.Id = id
	servers[id] = cfg

	groups, err := config.UnmarshalGroupsFromFile(groupsFile)
	if err != nil {
		sugar.Fatalf("Could not read group file %s: %v", groupsFile, err)
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

	err := conf.LocalServerConfigSet(addr, args.ServerPrivateFile, args.ServerPublicFile)
	if err != nil {
		log.Printf("Could not create server config for address %s", addr)
		log.Fatalf("%v", err)
	}

	log.Printf("Created server config for address %s", addr)
}
