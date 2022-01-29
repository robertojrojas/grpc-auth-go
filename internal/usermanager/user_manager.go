package usermanager

import (
	api "github.com/robertojrojas/grpc-auth/api/v1"
	"google.golang.org/grpc"
)

type grpcServer struct {
	api.UnimplementedUserManagerServer
}

func newgrpcServer() (*grpcServer, error) {
	srv := &grpcServer{}
	return srv, nil
}

func NewGRPCServer() (
	*grpc.Server,
	error,
) {
	// opts = append(opts, grpc.StreamInterceptor(
	// 	grpc_middleware.ChainStreamServer(
	// 		grpc_auth.StreamServerInterceptor(authenticate),
	// 	)), grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
	// 	grpc_auth.UnaryServerInterceptor(authenticate),
	// )))
	opts := []grpc.ServerOption{}
	gsrv := grpc.NewServer(opts...)
	srv, err := newgrpcServer()
	if err != nil {
		return nil, err
	}
	api.RegisterUserManagerServer(gsrv, srv)
	return gsrv, nil
}
