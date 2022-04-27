#!/usr/bin/env bash

openssl req -new -x509 -sha256 -days 365 -nodes -out certs/ca.crt \
  -keyout keys/ca.key -subj "/CN=root-ca"

# Create the server key and CSR and sign with root key
openssl req -new -nodes -out server.csr \
  -keyout keys/server.key -subj "/CN=localhost"

openssl x509 -req -in server.csr -sha256 -days 365 \
    -CA certs/ca.crt -CAkey keys/ca.key -CAcreateserial \
    -out certs/server.crt
