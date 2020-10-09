package server

import (
	"fmt"
	"net"

	"github.com/onflow/flow-go/access"
	legacyaccess "github.com/onflow/flow-go/access/legacy"
	"github.com/onflow/flow-go/model/flow"
	grpcprometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	accessproto "github.com/onflow/flow/protobuf/go/flow/access"
	legacyaccessproto "github.com/onflow/flow/protobuf/go/flow/legacy/access"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/dapperlabs/flow-emulator/server/backend"
)

type GRPCServer struct {
	logger     *logrus.Logger
	port       int
	grpcServer *grpc.Server
}

func NewGRPCServer(logger *logrus.Logger, b *backend.Backend, port int, debug bool) *GRPCServer {
	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(grpcprometheus.StreamServerInterceptor),
		grpc.UnaryInterceptor(grpcprometheus.UnaryServerInterceptor),
	)

	chain := flow.Emulator.Chain()
	adaptedBackend := backend.NewAdapter(b)

	legacyaccessproto.RegisterAccessAPIServer(grpcServer, legacyaccess.NewHandler(adaptedBackend, chain))
	accessproto.RegisterAccessAPIServer(grpcServer, access.NewHandler(adaptedBackend, chain))

	grpcprometheus.Register(grpcServer)

	if debug {
		reflection.Register(grpcServer)
	}

	return &GRPCServer{
		logger:     logger,
		port:       port,
		grpcServer: grpcServer,
	}
}

func (g *GRPCServer) Server() *grpc.Server {
	return g.grpcServer
}

func (g *GRPCServer) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", g.port))
	if err != nil {
		return err
	}

	return g.grpcServer.Serve(lis)
}

func (g *GRPCServer) Stop() {
	g.grpcServer.GracefulStop()
}
