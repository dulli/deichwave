package web

import "embed"

// Embedded file host to serve the web interface

//go:embed public
var Public embed.FS

//go:embed tls/deichwave-ca.crt tls/deichwave-server.crt tls/deichwave-server.key
var TLS embed.FS
