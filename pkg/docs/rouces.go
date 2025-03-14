package docs

import "embed"

// Mapping elasticsearch mapping
//
//go:embed mapping
var Mapping embed.FS
