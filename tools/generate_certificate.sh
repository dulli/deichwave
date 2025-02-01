CANAME=deichwave-ca
CERTNAME=deichwave-server

# Generate the server certificate request
openssl req -new -nodes \
    -newkey rsa:2048 -keyout web/tls/$CERTNAME.key \
    -out web/tls/$CERTNAME.csr \
    -subj "/CN=deichwave.internal" -config web/tls/openssl.cnf

# Sign the certificate request to create a certificate
openssl x509 -req \
    -in web/tls/$CERTNAME.csr \
    -CA web/tls/$CANAME.crt -CAkey web/tls/$CANAME.key -CAcreateserial \
    -sha256 -days 3650 -out web/tls/$CERTNAME.crt \
    -extfile web/tls/openssl.cnf -extensions v3_ext

# Verify the certificate
openssl x509 -in web/tls/$CERTNAME.crt -text -noout
