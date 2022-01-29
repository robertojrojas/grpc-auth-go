package main

import (
	"fmt"
	"log"

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
	fmt.Printf("userServer: %#v\n", userServer)
	fmt.Printf("vmServer: %#v\n", vmServer)
}
