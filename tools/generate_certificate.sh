# see https://github.com/lwsjs/local-web-server/wiki/How-to-get-the-%22green-padlock%22-with-a-new-self-signed-certificate
openssl genrsa -out web/tls/deichwave.key 2048
openssl req -new -nodes -sha256 -key web/tls/deichwave.key -out web/tls/deichwave.crs -config web/tls/openssl.cnf
openssl x509 -req -sha256 -days 3650 -in web/tls/deichwave.crs -signkey web/tls/deichwave.key -out web/tls/deichwave.crt -extfile web/tls/openssl.cnf -extensions v3_req
