package network

// GRPC network setup things

import (
	"crypto/tls"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/31333337/bmrng/go/0kn/pkg/utils"
	"github.com/31333337/bmrng/go/trellis/config"
	coord "github.com/31333337/bmrng/go/trellis/coordinator/messages"
	"github.com/31333337/bmrng/go/trellis/network/messages"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func RunServer(handler messages.MessageHandlersServer, coordHandler coord.CoordinatorHandlerServer, servercfgs map[int64]*config.Server, addr string) {
	logger := utils.GetLogger()
	sugar := logger.Sugar()
	defer sugar.Sync()
	server := StartServer(handler, coordHandler, servercfgs, addr)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	server.Stop()
	sugar.Infof("Server stopped",
		"address", addr,
	)
}

func StartServer(handler messages.MessageHandlersServer, coordHandler coord.CoordinatorHandlerServer, servercfgs map[int64]*config.Server, addr string) *grpc.Server {
	logger := utils.GetLogger()
	sugar := logger.Sugar()
	defer sugar.Sync()
	id, myCfg := FindConfig(addr, servercfgs)
	if id < 0 {
		sugar.Fatalf("Could not find %s", addr)
		sugar.Sync()
		panic("Could not find " + addr)
	}
	cert, err := tls.X509KeyPair(myCfg.Identity, myCfg.PrivateIdentity)
	if err != nil {
		sugar.Fatalw(
			"error generating X509 Key Pair",
			"error", err,
		)
		sugar.Sync()
		panic(err)
	}
	cred := credentials.NewServerTLSFromCert(&cert)
	grpcServer := grpc.NewServer(grpc.Creds(cred),
		grpc.MaxRecvMsgSize(2*config.StreamSize), grpc.MaxSendMsgSize(2*config.StreamSize))
	if handler != nil {
		messages.RegisterMessageHandlersServer(grpcServer, handler)
	}
	if coordHandler != nil {
		coord.RegisterCoordinatorHandlerServer(grpcServer, coordHandler)
	}
	lis, err := net.Listen("tcp", config.Port(addr))
	if err != nil {
		sugar.Fatalw("Server could not listen over tcp",
			"addr", addr,
			"err", err,
		)
		sugar.Sync()
		panic(err)
	}

	go func() {
		err := grpcServer.Serve(lis)
		if err != nil && err != grpc.ErrServerStopped {
			sugar.Fatalf("grpcServer err: %v", err)
			sugar.Sync()
			panic(err)
		}
	}()
	sugar.Infow("Server started",
		"address", addr,
	)
	return grpcServer
}

func FindConfig(addr string, servercfgs map[int64]*config.Server) (int64, *config.Server) {
	for id, cfg := range servercfgs {
		if cfg.Address == addr {
			return id, cfg
		}
	}
	return -1, nil
}
