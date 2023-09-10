package main

import (
	"log"
	"os"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/simonlangowski/lightning1/config"
	"github.com/simonlangowski/lightning1/coordinator"
)

type Args struct {
	Mode string `arg:"positional,required" help:"coordinator, server, or client"`

	F           float64 `default:"0"`
	RunType     int     `default:"1"`
	NumUsers    int     `default:"100"`
	NumServers  int     `default:"10"`
	MessageSize int     `default:"1024"`

	NumGroups int `default:"3"`
	GroupSize int `default:"3"`
	BinSize   int `default:"0"`
	LimitSize int `default:"0"`
	NumLayers int `default:"10"`
	Overflow  int `default:"32"`

	NumClientServers int  `default:"0"`
	SkipPathGen      bool `default:"False"`
	NoCheck          bool `default:"False"`
	LoadMessages     bool `default:"False"`
	StartIdx         int  `default:"0"`
	Interval         int  `default:"0"`

	// files
	ServerFile  string `default:"servers.json"`
	GroupFile   string `default:"groups.json"`
	ClientFile  string `default:"clients.json"`
	KeyFile     string `default:"keys.json"`
	MessageFile string `default:"messages.json"`
	OutFile     string `default:"res.json"`
}

func LaunchCoordinator(args Args) {
	p := arg.MustParse(&args)

	////////////////////////////////////////////////////////////////////////
	// process args
	////////////////////////////////////////////////////////////////////////
	if args.GroupSize == 0 {
		if args.F != 0 {
			if args.NumGroups != 0 {
				args.GroupSize = config.CalcGroupSize(args.NumServers, args.NumGroups, args.F)
			} else {
				args.GroupSize, args.NumGroups = config.CalcFewGroups2(args.F, args.NumServers)
			}
		} else {
			log.Printf("Set groupsize or f")
			p.WriteHelp(os.Stdout)
			return
		}
	}
	if args.NumGroups == 0 {
		// Integer division takes floor
		args.NumGroups = args.NumServers / args.GroupSize
	}
	if args.NumLayers == 0 {
		if args.F != 0 {
			args.NumLayers = config.NumLayers(args.NumUsers, args.F)
		} else {
			log.Printf("Set numlayers or f")
			p.WriteHelp(os.Stdout)
			return
		}
	}
	if args.BinSize == 0 {
		args.BinSize = config.BinSize2(args.NumLayers, args.NumServers, args.NumUsers, -args.Overflow)
	}
	if args.LimitSize == 0 {
		// the limit for how many servers needs to check the boomerang message
		// the size of an anytrust group ensures one honest server
		// we need replacement because servers can be selected multiple times
		args.LimitSize = config.GroupSizeWithReplacement(args.NumGroups, args.F)
	}
	if args.LoadMessages {
		args.NumClientServers = 0
	}

	log.Printf("%+v", args)

	////////////////////////////////////////////////////////////////////////
	// setup network
	////////////////////////////////////////////////////////////////////////
	var net *coordinator.CoordinatorNetwork

	switch args.RunType {
	case 0: // run in the same process
		net = coordinator.NewInProcessNetwork(args.NumServers, args.NumGroups, args.GroupSize)

	case 1: // run in separate process on the same machine
		serverConfigs, groupConfigs, clientConfigs := coordinator.NewLocalConfig(args.NumServers, args.NumGroups, args.GroupSize, args.NumClientServers, false)
		if args.LoadMessages {
			oldServers, err := config.UnmarshalServersFromFile(args.ServerFile)
			if err != nil {
				log.Fatalf("Could not read servers file %s", args.ServerFile)
			}
			// copy old keys
			for id, s := range serverConfigs {
				old := oldServers[id]
				s.PrivateKey = old.PrivateKey
				s.PublicKey = old.PublicKey
			}
		}
		net = coordinator.NewLocalNetwork(serverConfigs, groupConfigs, clientConfigs)
		defer net.KillAll()
	}

	////////////////////////////////////////////////////////////////////////
	// run experiment
	////////////////////////////////////////////////////////////////////////
	numLayers := args.NumLayers
	numServers := args.NumServers
	numMessages := args.NumUsers
	numLightning := 5
	c := coordinator.NewCoordinator(net)
	if args.LoadMessages {
		c.LoadKeys(args.KeyFile)
		c.LoadMessages(args.MessageFile)
	}
	l := 0
	if !args.SkipPathGen {
		for i := 0; i < numLayers; i++ {
			log.Printf("Round %v", i)
			exp := c.NewExperiment(i, numLayers, numServers, numMessages, args)
			if i == 0 {
				exp.KeyGen = !args.LoadMessages
				exp.LoadKeys = args.LoadMessages
			}
			exp.Info.PathEstablishment = true
			exp.Info.LastLayer = (i == numLayers-1)
			exp.Info.Check = !args.NoCheck
			exp.Info.Interval = int64(args.Interval)
			if args.BinSize > 0 {
				exp.Info.BinSize = int64(args.BinSize)
			} else if i == 0 {
				log.Printf("Using bin size %d", exp.Info.BinSize)
			}
			if args.LimitSize > 0 {
				exp.Info.BoomerangLimit = int64(args.LimitSize)
			} else {
				exp.Info.BoomerangLimit = int64(numLayers)
			}
			exp.Info.ReceiptLayer = 0
			if i-int(exp.Info.BoomerangLimit) > 0 {
				exp.Info.ReceiptLayer = int64(i) - exp.Info.BoomerangLimit
			}
			exp.Info.NextLayer = int64(i)
			if args.RunType == 5 {
				exp.Info.StartId = int64(args.StartIdx)
				exp.Info.EndId = exp.Info.StartId + int64(numMessages)
				c.WriteKeys(args.KeyFile)
				c.WriteMessages(args.MessageFile, exp)
				return
			}
			err := c.DoAction(exp)
			if err != nil {
				log.Print(err)
				return
			}
			log.Printf("Path round %v took %v", i, time.Since(exp.ExperimentStartTime))
			exp.RecordToFile(args.OutFile)
		}
		l = numLayers
	}
	for i := l; i < l+numLightning; i++ {
		log.Printf("Round %v", i)
		exp := c.NewExperiment(i, numLayers, numServers, numMessages, args)
		exp.Info.PathEstablishment = false
		exp.Info.MessageSize = int64(args.MessageSize)
		exp.Info.Check = !args.NoCheck
		if args.BinSize > 0 {
			exp.Info.BinSize = int64(args.BinSize)
		}
		if args.SkipPathGen && (i == 0) {
			exp.Info.SkipPathGen = true
			exp.KeyGen = true
		}
		exp.Info.Interval = int64(args.Interval)
		err := c.DoAction(exp)
		if err != nil {
			log.Print(err)
			return
		}
		log.Printf("Lightning round %v took %v", i, time.Since(exp.ExperimentStartTime))
		exp.RecordToFile(args.OutFile)
	}
	l += numLightning
}

func LaunchServer() {
	log.Printf("TODO: server")

	// // read configuration files
	// serversFile := os.Args[1]
	// groupsFile := os.Args[2]
	// addr := os.Args[len(os.Args)-1]

	// errors.Addr = addr
	// servers, err := config.UnmarshalServersFromFile(serversFile)
	// if err != nil {
	// 	log.Fatalf("Could not read servers file %s", serversFile)
	// }

	// groups, err := config.UnmarshalGroupsFromFile(groupsFile)
	// if err != nil {
	// 	log.Fatalf("Could not read group file %s", groupsFile)
	// }

	// fmt.Println("addr", addr)

	// // will start in blocked state
	// h := server.NewHandler()
	// server := server.NewServer(&config.Servers{Servers: servers}, &config.Groups{Groups: groups}, h, addr)

	// server.TcpConnections.LaunchAccepts()
	// network.RunServer(h, server, servers, addr)
	// config.Flush()
}

func LaunchClient() {
	log.Printf("TODO: client")
}

func main() {
	var args Args
	p := arg.MustParse(&args)

	switch args.Mode {
	case "coordinator":
		LaunchCoordinator(args)

	case "client":
		LaunchClient()

	case "server":
		LaunchServer()

	default:
		p.WriteHelp(os.Stdout)
		os.Exit(1)
	}

}
