CANAME=deichwave-ca

# Generate CA certificate.
openssl req -x509 -new -nodes \
    -newkey rsa:2048 -keyout web/tls/$CANAME.key \
    -sha256 -days 3650 -out web/tls/$CANAME.crt \
    -config web/tls/openssl.cnf -extensions v3_ca
