package recipeutil

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/futurice/jalapeno/pkg/recipe"
	"github.com/go-yaml/yaml"
)

// All Jalapeno related metadata will be saved to this directory on the project
const (
	ProjectDir = ".jalapeno"
	RecipeFile = "recipe.yml"
)

func SaveRecipe(r *recipe.Recipe, dest string) error {
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		return errors.New("destination path does not exist")
	}

	path := filepath.Join(dest, ProjectDir)
	err := os.MkdirAll(path, 0700)
	if err != nil {
		return err
	}

	out, err := yaml.Marshal(r)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(path, RecipeFile), out, 0700)
	if err != nil {
		return err
	}

	return nil
}

func SaveFiles(files []*recipe.File, dest string) error {
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		return errors.New("destination path does not exist")
	}

	for _, file := range files {
		path := filepath.Join(dest, file.Name)

		// Create file's parent directories (if not already exist)
		err := os.MkdirAll(filepath.Dir(path), 0700)
		if err != nil {
			return err
		}

		// Create the file
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer f.Close()

		// Write the data to the file
		_, err = f.Write(file.Data)
		if err != nil {
			return err
		}

		f.Sync()
	}
	return nil
}
