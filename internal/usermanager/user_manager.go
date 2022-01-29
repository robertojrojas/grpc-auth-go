package usermanager

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/google/uuid"
	api "github.com/robertojrojas/grpc-auth/api/v1"
	"github.com/robertojrojas/grpc-auth/internal/db"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type grpcServer struct {
	api.UnimplementedUserManagerServer
}

func newgrpcServer() (*grpcServer, error) {
	srv := &grpcServer{}
	return srv, nil
}

func NewGRPCServer(serverCert, serverKey, caCert string) (*grpc.Server, error) {

	// opts = append(opts, grpc.StreamInterceptor(
	// 	grpc_middleware.ChainStreamServer(
	// 		grpc_auth.StreamServerInterceptor(authenticate),
	// 	)), grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
	// 	grpc_auth.UnaryServerInterceptor(authenticate),
	// )))

	cert, err := tls.LoadX509KeyPair(serverCert, serverKey)
	if err != nil {
		log.Fatalf("failed to load key pair: %s", err)
	}

	// Create a certificate pool from the certificate authority
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(caCert)
	if err != nil {
		log.Fatalf("could not read ca certificate: %s", err)
	}

	// Append the client certificates from the CA
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Fatalf("failed to append client certs")
	}

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(unaryInterceptor),
		// Enable TLS for all incoming connections.
		grpc.Creds( // Create the TLS credentials
			credentials.NewTLS(&tls.Config{
				ClientAuth:   tls.RequireAndVerifyClientCert,
				Certificates: []tls.Certificate{cert},
				ClientCAs:    certPool,
			},
			)),
	}

	gsrv := grpc.NewServer(opts...)
	srv, err := newgrpcServer()
	if err != nil {
		return nil, err
	}
	api.RegisterUserManagerServer(gsrv, srv)

	// Register reflection service on gRPC server.
	reflection.Register(gsrv)

	return gsrv, nil
}

func (s *grpcServer) Create(ctx context.Context, req *api.User) (*api.User, error) {
	req.VmUuid = uuid.New().String()

	// Insert User to DB
	// create the postgres db connection
	dbConn, err := db.CreateConnection()
	if err != nil {
		return nil, err
	}
	// close the db connection
	defer dbConn.Close()

	tx := dbConn.MustBegin()
	_, err = dbConn.NamedExec(`INSERT INTO users (name,uuid) VALUES (:name,:uuid)`,
		map[string]interface{}{
			"name": req.UserName,
			"uuid": req.VmUuid,
		})
	if err != nil {
		return nil, err
	}
	tx.Commit()

	return req, nil
}

func (s *grpcServer) GetUser(ctx context.Context, req *api.Username) (*api.User, error) {
	dbConn, err := db.CreateConnection()
	if err != nil {
		return nil, err
	}
	defer dbConn.Close()

	rows, err := dbConn.NamedQuery(`SELECT name, uuid FROM users WHERE name=:fn`, map[string]interface{}{"fn": req.GetValue()})
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	user := &api.User{}
	if !rows.Next() {
		return user, nil
	}
	err = rows.Scan(
		&user.UserName,
		&user.VmUuid,
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// // authentication (token verification)
	// md, ok := metadata.FromIncomingContext(ctx)
	// if !ok {
	// 	return nil, errMissingMetadata
	// }
	// if !valid(md["authorization"]) {
	// 	return nil, errInvalidToken
	// }
	// m, err := handler(ctx, req)
	// if err != nil {
	// 	logger("RPC failed with error %v", err)
	// }
	// return m, err

	peer, ok := peer.FromContext(ctx)
	if !ok {
		return ctx, status.New(
			codes.Unknown,
			"couldn't find peer info",
		).Err()
	}

	fmt.Printf("request: %#v\n", req)
	fmt.Printf("info: %#v\n", info)
	fmt.Printf("Peer: %#v\n", peer)

	if peer.AuthInfo == nil {
		fmt.Println("AuthInfo is nil....")
		return context.WithValue(ctx, subjectContextKey{}, ""), nil
	}

	tlsInfo := peer.AuthInfo.(credentials.TLSInfo)

	subject := tlsInfo.State.PeerCertificates[0].Subject.CommonName
	ctx = context.WithValue(ctx, subjectContextKey{}, subject)
	fmt.Printf("PeerCertificates - subject: %s\n", subject)

	for id, peerCert := range tlsInfo.State.PeerCertificates {
		fmt.Printf("peerCert[%d]: %#v\n\n", id, peerCert)
	}

	subject = tlsInfo.State.VerifiedChains[0][0].Subject.CommonName
	ctx = context.WithValue(ctx, subjectContextKey{}, subject)
	fmt.Printf("VerifiedChains - subject: %s\n", subject)

	return handler(ctx, req)
}

func authenticate(ctx context.Context) (context.Context, error) {
	peer, ok := peer.FromContext(ctx)
	if !ok {
		return ctx, status.New(
			codes.Unknown,
			"couldn't find peer info",
		).Err()
	}

	if peer.AuthInfo == nil {
		return context.WithValue(ctx, subjectContextKey{}, ""), nil
	}

	tlsInfo := peer.AuthInfo.(credentials.TLSInfo)
	subject := tlsInfo.State.VerifiedChains[0][0].Subject.CommonName
	ctx = context.WithValue(ctx, subjectContextKey{}, subject)

	return ctx, nil
}

func subject(ctx context.Context) string {
	return ctx.Value(subjectContextKey{}).(string)
}

type subjectContextKey struct{}
