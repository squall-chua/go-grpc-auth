package web

import "embed"

//go:embed all:.output/public
var Assets embed.FS
