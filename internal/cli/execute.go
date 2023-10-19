package cli

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/futurice/jalapeno/internal/cli/option"
	"github.com/futurice/jalapeno/pkg/oci"
	"github.com/futurice/jalapeno/pkg/recipe"
	"github.com/futurice/jalapeno/pkg/recipeutil"
	"github.com/futurice/jalapeno/pkg/survey"
	"github.com/gofrs/uuid"
	"github.com/spf13/cobra"
)

type executeOptions struct {
	RecipeURL string
	option.Values
	option.Styles
	option.OCIRepository
	option.WorkingDirectory
	option.Common
}

func NewExecuteCmd() *cobra.Command {
	var opts executeOptions
	var cmd = &cobra.Command{
		Use:     "execute RECIPE_PATH",
		Aliases: []string{"exec", "e"},
		Short:   "Execute a recipe",
		Long:    "TODO",
		Args:    cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			opts.RecipeURL = args[0]
			return option.Parse(&opts)
		},
		Run: func(cmd *cobra.Command, args []string) {
			runExecute(cmd, opts)
		},
	}

	if err := option.ApplyFlags(&opts, cmd.Flags()); err != nil {
		return nil
	}

	return cmd
}

func runExecute(cmd *cobra.Command, opts executeOptions) {
	if _, err := os.Stat(opts.Dir); os.IsNotExist(err) {
		cmd.PrintErrln("Error: output path does not exist")
		return
	}

	var (
		re              *recipe.Recipe
		err             error
		wasRemoteRecipe bool
	)

	if strings.HasPrefix(opts.RecipeURL, "oci://") {
		wasRemoteRecipe = true
		ctx := context.Background()
		re, err = oci.PullRecipe(ctx,
			oci.Repository{
				Reference: strings.TrimPrefix(opts.RecipeURL, "oci://"),
				PlainHTTP: opts.PlainHTTP,
				Credentials: oci.Credentials{
					Username:      opts.Username,
					Password:      opts.Password,
					DockerConfigs: opts.Configs,
				},
				TLS: oci.TLSConfig{
					CACertFilePath: opts.CACertFilePath,
					Insecure:       opts.Insecure,
				},
			})

	} else {
		re, err = recipe.LoadRecipe(opts.RecipeURL)
	}

	if err != nil {
		cmd.PrintErrf("Error: can not load the recipe: %s\n", err)
		return
	}

	style := lipgloss.NewStyle().Foreground(opts.Colors.Primary)
	cmd.Printf("%s: %s\n", style.Render("Recipe name"), re.Metadata.Name)

	if re.Metadata.Description != "" {
		cmd.Printf("%s: %s\n", style.Render("Description"), re.Metadata.Description)
	}

	// Load all existing sauces
	existingSauces, err := recipe.LoadSauces(opts.Dir)
	if err != nil {
		cmd.PrintErrf("Error: %s", err)
		return
	}

	reusedValues := make(recipe.VariableValues)
	if opts.ReuseSauceValues && len(existingSauces) > 0 {
		for _, sauce := range existingSauces {
			overlappingSauceValues := make(recipe.VariableValues)
			for _, v := range re.Variables {
				if val, found := sauce.Values[v.Name]; found {
					overlappingSauceValues[v.Name] = val
				}
			}

			if len(overlappingSauceValues) > 0 {
				reusedValues = recipeutil.MergeValues(reusedValues, overlappingSauceValues)
			}
		}
	}

	providedValues, err := recipeutil.ParseProvidedValues(re.Variables, opts.Values.Flags, opts.Values.CSVDelimiter)
	if err != nil {
		cmd.PrintErrf("Error when parsing provided values: %s\n", err)
		return
	}

	values := recipeutil.MergeValues(reusedValues, providedValues)

	// Filter out variables which don't have value yet
	varsWithoutValues := recipeutil.FilterVariablesWithoutValues(re.Variables, values)
	if len(varsWithoutValues) > 0 {
		promptedValues, err := survey.PromptUserForValues(cmd.InOrStdin(), cmd.OutOrStdout(), varsWithoutValues, values)
		if err != nil {
			if errors.Is(err, survey.ErrUserAborted) {
				return
			} else {
				cmd.PrintErrf("Error when prompting for values: %s\n", err)
				return
			}
		}
		values = recipeutil.MergeValues(values, promptedValues)
	}

	sauce, err := re.Execute(values, uuid.Must(uuid.NewV4()))
	if err != nil {
		cmd.PrintErrf("Error: %s", err)
		return
	}

	// Check for conflicts
	for _, s := range existingSauces {
		if conflicts := s.Conflicts(sauce); conflicts != nil {
			cmd.PrintErrf("Error: conflict in recipe '%s': file '%s' was already created by recipe '%s'\n", re.Name, conflicts[0].Path, s.Recipe.Name)
			return
		}
	}

	// Automatically add recipe origin if the recipe was remote
	if wasRemoteRecipe {
		sauce.CheckFrom = opts.RecipeURL
	}

	err = sauce.Save(opts.Dir)
	if err != nil {
		cmd.PrintErrf("Error: %s", err)
		return
	}

	cmd.Println("\nRecipe executed successfully!")

	tree := recipeutil.CreateFileTree(opts.Dir, sauce.Files)
	cmd.Printf("The following files were created:\n\n%s", tree)

	if re.InitHelp != "" {
		cmd.Printf("\nNext up: %s\n", re.InitHelp)
	}
}
