package test_static

import "embed"

//go:embed components partials pages base.tmpl
var StaticFS embed.FS
