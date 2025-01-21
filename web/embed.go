package web

import "embed"

// Embedded file host to serve the web interface

//go:embed public
var Public embed.FS

//go:embed tls
var TLS embed.FS
