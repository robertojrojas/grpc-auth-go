package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/robertojrojas/grpc-auth/internal/db"
	"github.com/robertojrojas/grpc-auth/internal/usermanager"
	"github.com/robertojrojas/grpc-auth/internal/vmmanager"
)

var (
	serverAddr string
	serverCert string
	serverKey  string
	caCert     string
)

func init() {
	flag.StringVar(&serverAddr, "serverAddr", ":50051", "GRPC Server Address")
	flag.StringVar(&serverCert, "serverCert", ".ssl/server.pem", "server Cert")
	flag.StringVar(&serverKey, "serverKey", ".ssl/server-key.pem", "server Key")
	flag.StringVar(&caCert, "caCert", ".ssl/ca.pem", "CA Cert")
}

func main() {
	flag.Parse()

	err := db.BuildDBIfNeeded()
	if err != nil {
		log.Fatalf("unable to init db: %v\n", err)
	}

	userServer, err := usermanager.NewGRPCServer(serverCert, serverKey, caCert)
	if err != nil {
		log.Fatalf("failed to start UserManager GRPC Server: %v\n", userServer)
	}
	vmServer, err := vmmanager.NewGRPCServer()
	if err != nil {
		log.Fatalf("failed to start VMManager GRPC Server: %v\n", vmServer)
	}

	l, err := net.Listen("tcp", serverAddr)
	if err != nil {
		log.Fatalf("unable to listen on port: %#v %v", l, err)
	}

	go func() {
		userServer.Serve(l)
	}()

	go func() {
		vmServer.Serve(l)
	}()

	fmt.Printf("export GRPC_SERVER='%s'\n", l.Addr().String())
	fmt.Scanln()
	userServer.Stop()
	vmServer.Stop()
	fmt.Println(l.Close())
}
