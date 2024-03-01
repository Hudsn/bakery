package static

import "embed"

//go:embed style/*.css
//go:embed script/*.js
var StaticFS embed.FS
