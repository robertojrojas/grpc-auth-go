package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	api "github.com/robertojrojas/grpc-auth/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	serverAddr string
	clientCert string
	clientKey  string
	caCert     string
)

func init() {
	flag.StringVar(&serverAddr, "serverAddr", ":50051", "GRPC Server Address")
	flag.StringVar(&clientCert, "clientCert", "../../.ssl/gateway-client.pem", "client Cert")
	flag.StringVar(&clientKey, "clientKey", "../../.ssl/gateway-client-key.pem", "client Key")
	flag.StringVar(&caCert, "caCert", "../../.ssl/ca.pem", "CA Cert")
}

func main() {
	flag.Parse()

	// Load the client certificates from disk
	certificate, err := tls.LoadX509KeyPair(clientCert, clientKey)
	if err != nil {
		log.Fatalf("could not load client key pair: %s - \n\nrun 'make gencert'", err)
	}

	// Create a certificate pool from the certificate authority
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(caCert)
	if err != nil {
		log.Fatalf("could not read ca certificate: %s", err)
	}

	// Append the certificates from the CA
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Fatalf("failed to append ca certs")
	}

	tlsConfig := &tls.Config{}
	tlsConfig.Certificates = make([]tls.Certificate, 1)
	tlsConfig.Certificates[0] = certificate
	tlsConfig.RootCAs = certPool
	opts := []grpc.DialOption{
		// transport credentials.
		grpc.WithTransportCredentials(credentials.NewTLS(
			tlsConfig,
		)),
	}

	conn, err := grpc.Dial(serverAddr, opts...)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	rgc := api.NewUserManagerClient(conn)

	username := &api.Username{
		Value: "josette",
	}
	ctx := context.Background()
	user, err := rgc.GetUser(ctx, username)
	if err != nil {
		log.Fatalf("failed to call GetUser: %v", err)
	}

	fmt.Printf("User is: %#v\n", user)
}
