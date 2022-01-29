package main

import (
	"fmt"
	"log"
	"net"

	"github.com/robertojrojas/grpc-auth/internal/usermanager"
	"github.com/robertojrojas/grpc-auth/internal/vmmanager"
)

func main() {
	userServer, err := usermanager.NewGRPCServer()
	if err != nil {
		log.Fatalf("failed to start UserManager GRPC Server: %v\n", userServer)
	}
	vmServer, err := vmmanager.NewGRPCServer()
	if err != nil {
		log.Fatalf("failed to start VMManager GRPC Server: %v\n", vmServer)
	}

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatalf("unable to listen on port: %#v %v", l, err)
	}

	go func() {
		userServer.Serve(l)
	}()

	go func() {
		vmServer.Serve(l)
	}()

	fmt.Printf("listening on %s\n", l.Addr().String())
	fmt.Scanln()
	userServer.Stop()
	vmServer.Stop()
	fmt.Println(l.Close())
}
