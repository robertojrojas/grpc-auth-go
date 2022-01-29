# START: begin
CONFIG_PATH=.ssl/
SSL_CONF=ssl_conf

.PHONY: init
init:
	mkdir -p ${CONFIG_PATH}

.PHONY: gencert
gencert: init
	cfssl gencert \
		-initca ${SSL_CONF}/ca-csr.json | cfssljson -bare ca

	cfssl gencert \
		-ca=ca.pem \
		-ca-key=ca-key.pem \
		-config=${SSL_CONF}/ca-config.json \
		-profile=server \
		${SSL_CONF}/server-csr.json | cfssljson -bare server
# END: begin

# START: client
	cfssl gencert \
		-ca=ca.pem \
		-ca-key=ca-key.pem \
		-config=${SSL_CONF}/ca-config.json \
		-profile=client \
		${SSL_CONF}/client-csr.json | cfssljson -bare client
# END: client

# START: multi_client
	cfssl gencert \
		-ca=ca.pem \
		-ca-key=ca-key.pem \
		-config=${SSL_CONF}/ca-config.json \
		-profile=client \
		-cn="root" \
		${SSL_CONF}/client-csr.json | cfssljson -bare root-client

	cfssl gencert \
		-ca=ca.pem \
		-ca-key=ca-key.pem \
		-config=${SSL_CONF}/ca-config.json \
		-profile=client \
		-cn="nobody" \
		${SSL_CONF}/client-csr.json | cfssljson -bare nobody-client
# END: multi_client

# START: begin
	mv *.pem *.csr ${CONFIG_PATH}

# END: begin

# START: begin
.PHONY: test
# END: auth
test:
# END: begin
# START: auth
test: $(CONFIG_PATH)/policy.csv $(CONFIG_PATH)/model.conf
#: START: begin
	go test -race ./...
# END: auth

.PHONY: compile
compile:
	protoc api/v1/*.proto \
		--go_out=. \
		--go-grpc_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_opt=paths=source_relative \
		--proto_path=.

# END: begin
