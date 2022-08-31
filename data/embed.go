package defaults

import "embed"

// Embedded file host to serve the web interface
//go:embed lights
var Lights embed.FS

// TODO: use embedded data filesystems as defaults if no path is available, add a way to write the defaults to disk
