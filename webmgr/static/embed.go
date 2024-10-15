package static

import "embed"

//go:embed files
var StaticFiles embed.FS
