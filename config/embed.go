package config

import "embed"

// Embedded file host to serve the web interface
//
//go:embed default.toml units device-tree
var Defaults embed.FS

// TODO: add a way to write the embedded config defaults to disk so that users can change them
