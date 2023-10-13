package example

import (
	"fmt"

	"github.com/hudsn/bakery"
	"github.com/hudsn/bakery/example/test_static"
)

func main() {
	tmplFiles := bakery.NewSourceFiles(
		// views.RecipeFsys(os.DirFS("test")),
		bakery.RecipeFsys(test_static.StaticFS),
		bakery.RecipeBaseTemplate("base.tmpl"),
		bakery.RecipePartialsGlobs("partials/*.tmpl", "partials/**/*.tmpl"),
		bakery.RecipePagesGlobs("pages/*.tmpl", "pages/**/*/tmpl"),
		bakery.RecipeStandaloneGlobs("components/*.tmpl", "components/**/*.tmpl"),
	)

	v, err := bakery.New(false, tmplFiles)
	if err != nil {
		panic(err)
	}

	buf, err := v.Process("home.tmpl", nil)
	if err != nil {
		panic(err)
	}
	fmt.Println(buf.String())
	buf, err = v.Process("component1.tmpl", nil)
	if err != nil {
		panic(err)
	}
	fmt.Println(buf.String())
	buf, err = v.Process("component2.tmpl", nil)
	if err != nil {
		panic(err)
	}

	fmt.Println(buf.String())
}
