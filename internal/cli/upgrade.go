package cli

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/futurice/jalapeno/internal/cli/internal/option"
	"github.com/futurice/jalapeno/pkg/recipe"
	"github.com/futurice/jalapeno/pkg/recipeutil"
	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
)

type upgradeOptions struct {
	ProjectPath string
	SourcePath  string
	option.Values
	option.Common
}

func newUpgradeCmd() *cobra.Command {
	var opts upgradeOptions
	var cmd = &cobra.Command{
		Use:   "upgrade PROJECT_PATH RECIPE_PATH",
		Short: "Upgrade sauce in a project",
		Long:  "TODO",
		Args:  cobra.ExactArgs(2),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			opts.ProjectPath = args[0]
			opts.SourcePath = args[1]
			return option.Parse(&opts)
		},
		Run: func(cmd *cobra.Command, args []string) {
			runUpgrade(cmd, opts)
		},
	}

	if err := option.ApplyFlags(&opts, cmd.Flags()); err != nil {
		return nil
	}

	return cmd
}

func runUpgrade(cmd *cobra.Command, opts upgradeOptions) {
	re, err := recipe.LoadRecipe(opts.SourcePath)
	if err != nil {
		cmd.PrintErrf("Error: %s", err)
		return
	}

	sauces, err := recipe.LoadSauces(opts.ProjectPath)
	if err != nil {
		cmd.PrintErrf("Error: %s", err)
		return
	}
	var oldSauce *recipe.Sauce
	for _, s := range sauces {
		if s.Recipe.Name == re.Name {
			oldSauce = s
			break
		}
	}
	if oldSauce == nil {
		cmd.PrintErrf("Error: project %s does not contain sauce %s. Recipe name used in the project should match the recipe which is used for upgrading", opts.ProjectPath, re.Name)
		return
	}

	if semver.Compare(re.Metadata.Version, oldSauce.Recipe.Metadata.Version) <= 0 {
		cmd.PrintErrln("Error: new recipe version is lower or same than the existing one")
		return
	}

	cmd.Printf("Upgrading recipe %s from version %s to %s\n", oldSauce.Recipe.Name, oldSauce.Recipe.Metadata.Version, re.Metadata.Version)

	// Check if the new version of the recipe has removed some variables
	// which existed on previous version
	for valueName := range oldSauce.Values {
		found := false
		for _, variable := range re.Variables {
			if variable.Name == valueName {
				found = true
			}
		}
		if !found {
			delete(oldSauce.Values, valueName)
		}
	}

	predefinedValues, err := recipeutil.ParsePredefinedValues(re.Variables, opts.Values.Flags)
	if err != nil {
		cmd.PrintErrf("Error when parsing provided values: %v\n", err)
		return
	}

	// Don't prompt variables which already has a value in existing sauce or is predefined
	varsWithoutValues := make([]recipe.Variable, 0, len(re.Variables))
	for _, v := range re.Variables {
		_, oldValueExists := oldSauce.Values[v.Name]
		_, predefinedValueExists := predefinedValues[v.Name]
		if !oldValueExists && !predefinedValueExists {
			varsWithoutValues = append(varsWithoutValues, v)
		}
	}

	values, err := recipeutil.PromptUserForValues(varsWithoutValues)
	if err != nil {
		cmd.PrintErrf("Error: %s", err)
		return
	}

	newSauce, err := re.Execute(recipeutil.MergeValues(values, oldSauce.Values, predefinedValues), oldSauce.Anchor)
	if err != nil {
		cmd.PrintErrf("Error: %s", err)
		return
	}

	// read common ignore file if it exists
	ignorePatterns := make([]string, 0)
	if data, err := os.ReadFile(filepath.Join(opts.ProjectPath, recipe.IgnoreFileName)); err == nil {
		ignorePatterns = append(ignorePatterns, strings.Split(string(data), "\n")...)
	} else if !errors.Is(err, fs.ErrNotExist) {
		// something else happened than trying to read an ignore file that does not exist
		cmd.PrintErrf("Error: failed to read ignore file: %s\n", err)
		return
	}
	ignorePatterns = append(ignorePatterns, re.IgnorePatterns...)

	// Collect files which should be written to the destination directory
	output := make(map[string]recipe.File, len(newSauce.Files))
	overrideNoticed := false

	for path := range newSauce.Files {
		skip := false
		for _, pattern := range ignorePatterns {
			if matched, err := filepath.Match(pattern, path); err != nil {
				cmd.PrintErrf("Error: bad ignore pattern '%s': %s\n", pattern, err)
				return
			} else if matched {
				// file was marked as ignored for upgrades
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		if prevFile, exists := oldSauce.Files[path]; exists {
			// Check if file was modified after rendering
			filePath := filepath.Join(opts.ProjectPath, path)
			if modified, err := recipeutil.IsFileModified(filePath, prevFile); err != nil {
				cmd.PrintErrf("Error: %s", err)
				return
			} else if modified {
				// The file contents has been modified
				if !overrideNoticed {
					cmd.Println("Some of the files has been manually modified. Do you want to override the following files:")
					overrideNoticed = true
				}

				// TODO: We could do better in terms of merge conflict management. Like show the diff or something
				var override bool
				prompt := &survey.Confirm{
					Message: path,
					Default: true,
				}
				err = survey.AskOne(prompt, &override)
				if err != nil {
					cmd.PrintErrf("Error: %s", err)
					return
				}

				if !override {
					// User decided not to override the file with manual changes, remove from
					// list of changes to write
					continue
				}
			}
		}

		// Add new file or replace existing one
		output[path] = newSauce.Files[path]
	}

	newSauce.Files = output

	err = newSauce.Save(opts.ProjectPath)
	if err != nil {
		cmd.PrintErrf("Error: %s", err)
		return
	}
}