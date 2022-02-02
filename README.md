# grpc-auth-go
Sample gRPC app demonstrating mTLS and Authorization using [Casbin](github.com/casbin/casbin).

# Run

```
make gencert

```

# Terminal 1
```
make server && ./grpc-auth-server
```

# Terminal 2
```
make client && ./grpc-auth-client
```