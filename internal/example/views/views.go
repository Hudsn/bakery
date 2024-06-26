package views

import (
	"embed"

	"github.com/hudsn/bakery"
)

//go:embed layouts/*go.html
//go:embed components/*.go.html
//go:embed components/*/*.go.html
//go:embed pages/*.go.html
var viewFS embed.FS

func New(isDev bool) bakery.Bakery {

	config := bakery.Config{
		IsDev: isDev,
		WatchExtensions: []string{
			".html",
			".go.html",
			".css",
			".js",
		},
		TemplateFS:      viewFS,
		TemplateRootDir: "internal/example/views",
		// StaticFS:        static.StaticFS,
		// StaticRootDir:   "internal/example/static",
	}

	myBakery := bakery.New(config)

	myBakery.AddRecipe("home", "layouts/app.go.html", "pages/home.go.html", "components/button.go.html", "components/sub/sub.go.html")

	myBakery.Init()

	return myBakery
}
