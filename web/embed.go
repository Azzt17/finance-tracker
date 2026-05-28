package web

import "embed"

//go:embed static/* index.html manifest.json sw.js
var StaticFS embed.FS
