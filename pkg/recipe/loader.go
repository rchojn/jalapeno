package recipe

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/go-yaml/yaml"
)

const (
	RecipeFileName         = "recipe.yml"
	RecipeTemplatesDirName = "templates"
	RenderedRecipeDirName  = ".jalapeno"
)

func Load(path string) (*Recipe, error) {
	// Later on here we can add additional load mechanisms (example from URL)
	return LoadFromDir(path)
}

func LoadFromDir(path string) (*Recipe, error) {
	rootdir, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	// TODO: Check if root path was not a dir

	recipeFile := filepath.Join(rootdir, RecipeFileName)
	dat, err := os.ReadFile(recipeFile)
	if err != nil {
		return nil, err
	}

	recipe := &Recipe{}
	err = yaml.Unmarshal(dat, recipe)
	if err != nil {
		return nil, err
	}

	if err := recipe.Validate(); err != nil {
		return nil, err
	}

	templates := make(map[string][]byte)
	templatesDir := filepath.Join(rootdir, RecipeTemplatesDirName)

	walk := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Continue walking if the path is directory
		if info.IsDir() {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		prefix := filepath.Join(rootdir, RecipeTemplatesDirName)
		prefix += string(filepath.Separator)
		name := filepath.ToSlash(strings.TrimPrefix(path, prefix))

		templates[name] = data
		return nil
	}

	err = filepath.Walk(templatesDir, walk)
	if err != nil {
		return recipe, err
	}

	recipe.Templates = templates

	return recipe, nil
}

// Load recipe which already has been rendered
func LoadRenderedFromDir(path string) (*Recipe, error) {
	rootDir, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	// TODO: Check if root path was not a dir

	recipeFile := filepath.Join(rootDir, RenderedRecipeDirName, RecipeFileName)
	dat, err := os.ReadFile(recipeFile)
	if err != nil {
		return nil, err
	}

	recipe := &Recipe{}
	err = yaml.Unmarshal(dat, recipe)
	if err != nil {
		return nil, err
	}

	if err := recipe.Validate(); err != nil {
		return nil, err
	}

	files := make(map[string][]byte)

	walk := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Continue walking if the path is directory
		if info.IsDir() {
			return nil
		}

		trimmedPath := strings.TrimPrefix(path, rootDir+string(filepath.Separator))

		// Skip recipe directory
		if filepath.Dir(trimmedPath) == RenderedRecipeDirName {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		name := filepath.ToSlash(trimmedPath)

		files[name] = data
		return nil
	}

	err = filepath.Walk(rootDir, walk)
	if err != nil {
		return nil, err
	}

	recipe.RenderedTemplates = files

	return recipe, nil
}
