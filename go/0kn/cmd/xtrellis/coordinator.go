package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	arg "github.com/alexflint/go-arg"

	"github.com/31333337/bmrng/go/0kn/internal/utils"
	"github.com/31333337/bmrng/go/0kn/pkg/gateway"
	"github.com/31333337/bmrng/go/trellis/config"
	"github.com/31333337/bmrng/go/trellis/coordinator"
)

const (
	// run network in the same process
	NETWORK_TYPE_IN_PROCESS = iota

	// run local network in separate processes on the same machine
	NETWORK_TYPE_LOCAL
)

func LaunchCoordinator(args ArgsCoordinator, argParser *arg.Parser) {
	processArgs(&args, argParser)

	switch {
	case args.Config != nil:
		runConfigGenerator(args)

	case args.Experiment != nil:
		runExperiment(args)

	case args.Mixnet != nil:
		runMixnet(args)
	}
}

func processArgs(args *ArgsCoordinator, argParser *arg.Parser) {
	if args.GroupSize == 0 {
		if args.F != 0 {
			if args.NumGroups != 0 {
				args.GroupSize = config.CalcGroupSize(args.NumServers, args.NumGroups, args.F)
			} else {
				args.GroupSize, args.NumGroups = config.CalcFewGroups2(args.F, args.NumServers)
			}
		} else {
			log.Printf("Set groupsize or f")
			argParser.WriteHelp(os.Stdout)
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
			argParser.WriteHelp(os.Stdout)
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
}

func setupNetwork(args ArgsCoordinator) *coordinator.CoordinatorNetwork {
	var net *coordinator.CoordinatorNetwork

	switch args.NetworkType {
	case NETWORK_TYPE_IN_PROCESS: // run in the same process
		net = coordinator.NewInProcessNetwork(args.NumServers, args.NumGroups, args.GroupSize)

	case NETWORK_TYPE_LOCAL: // run in separate process on the same machine
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
	}

	return net
}

func runExperiment(args ArgsCoordinator) {
	net := setupNetwork(args)
	if args.NetworkType == NETWORK_TYPE_LOCAL {
		defer net.KillAll()
	}

	numLayers := args.NumLayers
	numServers := args.NumServers
	numMessages := args.NumUsers

	c := coordinator.NewCoordinator(net)
	if args.LoadMessages {
		c.LoadKeys(args.KeyFile)
		c.LoadMessages(args.MessageFile)
	}

	l := 0
	if !args.SkipPathGen {
		for i := 0; i < numLayers; i++ {
			log.Printf("\n")
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

	numLightning := 5
	for i := l; i < l+numLightning; i++ {
		log.Printf("\n")
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
}

func runMixnet(args ArgsCoordinator) {
	log.Printf("Running mix-net...")

	// setup gateway and start proxy if enabled
	if args.MessageSize <= int(gateway.GetMaxProtocolSize()) {
		log.Fatal("Error: MessageSize too small for Gateway packet protocol")
	}
	gateway.Init(int64(args.MessageSize), args.GatewayEnable, args.GatewayAddrIn, args.GatewayAddrOut)

	// setup network
	net := setupNetwork(args)
	if args.NetworkType == NETWORK_TYPE_LOCAL {
		defer net.KillAll()
	}

	numLayers := args.NumLayers
	numServers := args.NumServers
	numMessages := args.NumUsers

	c := coordinator.NewCoordinator(net)
	if args.LoadMessages {
		c.LoadKeys(args.KeyFile)
		c.LoadMessages(args.MessageFile)
	}

	//////////////////////////////////////////////////////
	// run rounds for each layer to establish paths
	//////////////////////////////////////////////////////
	for i := 0; i < numLayers; i++ {
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

		err := c.DoAction(exp)
		if err != nil {
			log.Print(err)
			return
		}
		log.Printf("Path round %v took %v", i, time.Since(exp.ExperimentStartTime))
		exp.RecordToFile(args.OutFile)
	}

	////////////////////////////////////////////////////////////////////////
	// wait for CTRL-C to exit, leave servers running
	// https://stackoverflow.com/a/18158859
	////////////////////////////////////////////////////////////////////////
	ctrl := make(chan os.Signal)
	signal.Notify(ctrl, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ctrl
		log.Println("Exiting")
		os.Exit(1)
	}()
	log.Println("Coordinator running... CTRL-C to exit.")

	//////////////////////////////////////////////////////
	// continually run lightning rounds to transmit messages
	//////////////////////////////////////////////////////
	var stats = utils.NewTimeStats()
	for i := numLayers; ; i++ {
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

		stats.RecordTime(float64(time.Since(exp.ExperimentStartTime)))

		// print time stats
		c := 10 // print cycle, so terminal not too busy
		if i%c == 0 {
			log.Printf("Lightning round %v : %s", i, stats.GetStatsString())
		}

		exp.RecordToFile(args.OutFile)

		// sleep between rounds
		time.Sleep(time.Duration(args.RoundInterval) * time.Millisecond)
	}
}

}
