Generate Server and Client certificates

Create CA key and cert:

openssl genrsa -out ca.key 4096

openssl req -x509 -new -nodes -key ca.key -sha256 -days 365 -out ca.crt -subj "/CN=MyCA"


Create server cert (for order_service):

openssl genrsa -out order_service/certs/server.key 4096

openssl req -new -key order_service/certs/server.key -out order_service/certs/server.csr -config san.cnf

openssl x509 -req -in order_service/certs/server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out order_service/certs/server.crt -days 365 -sha256 -extensions v3_req -extfile san.cnf


Create client cert (for user_service):

openssl genrsa -out user_service/certs/client.key 4096

openssl req -new -key user_service/certs/client.key -out user_service/certs/client.csr -config san.cnf

openssl x509 -req -in user_service/certs/client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out user_service/certs/client.crt -days 365 -sha256 -extensions v3_req -extfile san.cnf


Verify the Certificate - Confirm the server certificate includes the SAN for localhost:

openssl x509 -in order_service/certs/server.crt -text -noout


This ensures server.crt includes:

X509v3 Subject Alternative Name:
    DNS:localhost
Subject: C = LK, ST = Southern, L = Matara, O = MyOrg, OU = Dev, CN = localhost
