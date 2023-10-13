package bakery

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"path/filepath"
)

type Views struct {
	Cache           map[string]*template.Template
	Files           Files
	IsDev           bool
	TemplateRecipes Recipes
}

func (v Views) Process(templateName string, data any) (*bytes.Buffer, error) {
	wbuf := bytes.NewBuffer([]byte{})

	if !v.IsDev {
		t, found := v.Cache[templateName]
		if !found {
			return nil, fmt.Errorf("unable to render template for %v. template does not exist.", templateName)
		}
		err := t.Execute(wbuf, data)
		if err != nil {
			return nil, err
		}
		return wbuf, nil
	}
	reqPaths, found := v.TemplateRecipes[templateName]
	if !found {
		return nil, fmt.Errorf("unable to parse template for %v. template does not exist.", templateName)
	}
	t, err := template.ParseFS(v.Files.Fsys, reqPaths...)
	if err != nil {
		return nil, fmt.Errorf("unable to parse template for %v. parsing required templates failed: %w", templateName, err)
	}

	err = t.Execute(wbuf, data)
	if err != nil {
		return nil, err
	}

	return wbuf, nil
}

func New(isDevEnv bool, files Files) (Views, error) {
	retView := Views{
		Files: files,
		IsDev: isDevEnv,
	}

	recipes, err := files.MakeRecipes()
	if err != nil {
		return Views{}, err
	}

	retView.TemplateRecipes = recipes

	cacheMap := make(map[string]*template.Template)
	if !isDevEnv {

		for file, requiredPaths := range recipes {
			parsedTemplate, err := template.ParseFS(retView.Files.Fsys, requiredPaths...)
			if err != nil {
				return Views{}, err
			}

			cacheMap[file] = parsedTemplate
		}

		retView.Cache = cacheMap
	}

	return retView, nil

}

type Recipes map[string][]string

type Files struct {
	Fsys            fs.FS
	BaseTemplate    string
	PartialsGlobs   []string
	PagesGlobs      []string
	StandaloneGlobs []string
}

type FileConfigFunc func(*Files)

func NewSourceFiles(configs ...FileConfigFunc) Files {
	var f Files

	for _, configFunc := range configs {
		configFunc(&f)
	}

	return f
}

func RecipeStandaloneGlobs(globs ...string) FileConfigFunc {
	return func(files *Files) {
		files.StandaloneGlobs = globs
	}
}

func RecipeFsys(fsys fs.FS) FileConfigFunc {
	return func(files *Files) {
		files.Fsys = fsys
	}
}

func RecipeBaseTemplate(basePath string) FileConfigFunc {
	return func(files *Files) {
		files.BaseTemplate = basePath
	}
}

func RecipePartialsGlobs(globs ...string) FileConfigFunc {
	return func(files *Files) {
		files.PartialsGlobs = globs
	}
}

func RecipePagesGlobs(globs ...string) FileConfigFunc {
	return func(files *Files) {
		files.PagesGlobs = globs
	}
}

func (f Files) MakeRecipes() (Recipes, error) {
	retMap := make(map[string][]string)

	baseTemplatePath, err := fs.Glob(f.Fsys, f.BaseTemplate)
	if err != nil {
		return nil, err
	}

	i := len(baseTemplatePath)
	switch {
	case i <= 0:
		return nil, fmt.Errorf("unable to find files matching base glob: %q - expected 1, got 0", f.BaseTemplate)
	case i >= 2:
		return nil, fmt.Errorf("too many files matched base glob: %q - expected 1, got %d", f.BaseTemplate, i)
	default:
		break
	}

	var partialsPaths []string
	for _, glob := range f.PartialsGlobs {
		globPaths, err := fs.Glob(f.Fsys, glob)
		if err != nil {
			return nil, err
		}
		partialsPaths = append(partialsPaths, globPaths...)
	}

	var pagesPaths []string
	for _, glob := range f.PagesGlobs {
		globPaths, err := fs.Glob(f.Fsys, glob)
		if err != nil {
			return nil, err
		}
		pagesPaths = append(pagesPaths, globPaths...)
	}

	var standalonePaths []string
	for _, glob := range f.StandaloneGlobs {
		globPaths, err := fs.Glob(f.Fsys, glob)
		if err != nil {
			return nil, err
		}
		standalonePaths = append(standalonePaths, globPaths...)
	}

	for _, component := range standalonePaths {
		var componentTemplateList []string
		var shortenedName string
		shortenedName = filepath.Base(component)
		componentTemplateList = append(componentTemplateList, component)
		componentTemplateList = append(componentTemplateList, partialsPaths...)
		retMap[shortenedName] = componentTemplateList
	}

	for _, page := range pagesPaths {
		var pageTemplateList []string
		var shortenedName string
		shortenedName = filepath.Base(page)
		pageTemplateList = append(pageTemplateList, baseTemplatePath[0])
		pageTemplateList = append(pageTemplateList, partialsPaths...)
		pageTemplateList = append(pageTemplateList, page)

		retMap[shortenedName] = pageTemplateList
	}

	return retMap, nil
}
