package bakery

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
)

// ex "home" -> ["base.go.html", "home.go.html", "footer.go.html"]
type recipeBook map[string][]string

type Config struct {
	IsDev           bool
	WatchExtensions []string //if this is empty we don't have hot reload functionality
	TemplateRootDir string   //relative to your project root
	TemplateFS      fs.FS
}

type Bakery struct {
	isDevEnv        bool
	prodCache       map[string]*template.Template
	recipes         recipeBook
	templateRootDir string
	templateFS      fs.FS
	watchExtensions []string
	httpErrFN       http.HandlerFunc
}

func New(config Config) Bakery {
	recipeBook := make(map[string][]string)

	templateCache := make(map[string]*template.Template)

	return Bakery{
		isDevEnv:        config.IsDev,
		prodCache:       templateCache,
		recipes:         recipeBook,
		templateRootDir: config.TemplateRootDir,
		templateFS:      config.TemplateFS,
		watchExtensions: config.WatchExtensions,
		httpErrFN:       defaultErrorHandler,
	}
}

func (b Bakery) IsDev() bool {
	return b.isDevEnv
}

func (b *Bakery) UseHTTPErrFunc(errFN http.HandlerFunc) {
	b.httpErrFN = errFN
}

func (b *Bakery) AddRecipe(name string, templatePaths ...string) {
	b.recipes[name] = templatePaths
}

func (b *Bakery) AddRecipeUsingTemplateList(name string, existingList []string, additionalPaths ...string) {
	templatePaths := append(existingList, additionalPaths...)
	b.recipes[name] = templatePaths

}

func (b *Bakery) Init() {
	if !b.isDevEnv {
		b.prepProdCache()
	}
	if b.isDevEnv {
		b.prepDevRecipes()
	}
}

// override recipe map with a list of absolute paths for local file template parsing.
func (b *Bakery) prepDevRecipes() {
	filepath.Walk(b.templateRootDir, func(path string, info fs.FileInfo, err error) error {

		if strings.HasSuffix(path, ".go") {
			return nil
		}

		for _, recipeList := range b.recipes {
			for i, templateFile := range recipeList {
				// if path endswith templatefile -> replace.
				if strings.HasSuffix(path, templateFile) {
					recipeList[i] = path
				}
			}
		}
		return nil
	})
}

// utility func if you want to return the template byte buffer instead of a handler.
// useful if you want to handle the error yourself instead of letting the Bake() function return the http handlerfunc for you.
func (b Bakery) BakeBuffer(recipeName string, data any) (*bytes.Buffer, error) {
	return b.process(recipeName, data)
}

func (b Bakery) Bake(recipeName string, data any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		renderBytes, err := b.process(recipeName, data)
		if err != nil {
			b.httpErrFN(w, r)
			return
		}
		renderBytes.WriteTo(w)
	}
}

func (b Bakery) process(recipeName string, data any) (*bytes.Buffer, error) {
	wbuf := bytes.NewBuffer([]byte{})

	//when prod environment, we use the pre-cached templates and execute them.
	if !b.isDevEnv {
		tplate, ok := b.prodCache[recipeName]
		if !ok {
			return nil, fmt.Errorf("recipe name not found: %q", recipeName)
		}

		err := tplate.Execute(wbuf, data)
		if err != nil {
			return nil, fmt.Errorf("issue parsing template for recipe %q - %q: %w", recipeName, tplate.Name(), err)
		}
		return wbuf, nil
	}

	// when dev env we want hot reloading, so we can't use templates that are generated at runtime (though this is preferable in prod) since they won't change even when our local files do.
	// instead we read straight from the local files, and recompile the templates each time we need to execute them.
	// this unlocks hot reloading for us.

	templateFiles, ok := b.recipes[recipeName]
	if !ok {
		return nil, fmt.Errorf("recipe name not found: %q", recipeName)
	}

	tplate, err := template.ParseFiles(templateFiles...)
	if err != nil {
		return nil, fmt.Errorf("issue parsing templates for recipe %q: %w", recipeName, err)
	}

	err = tplate.Execute(wbuf, data)
	if err != nil {
		return nil, fmt.Errorf("issue parsing template for recipe %q - %q: %w", recipeName, tplate.Name(), err)
	}

	return wbuf, nil
}

func (b *Bakery) prepProdCache() error {
	for tname, paths := range b.recipes {
		tplate, err := template.ParseFS(b.templateFS, paths...)
		if err != nil {
			return err
		}
		b.prodCache[tname] = tplate
	}
	return nil
}

func defaultErrorHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}
