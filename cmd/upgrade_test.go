package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	re "github.com/futurice/jalapeno/pkg/recipe"
)

func iUpgradeSauce(ctx context.Context, recipe string) (context.Context, error) {
	recipesDir := ctx.Value(recipesDirectoryPathCtxKey{}).(string)
	projectDir := ctx.Value(projectDirectoryPathCtxKey{}).(string)

	cmd, cmdStdOut, cmdStdErr := wrapCmdOutputs(newUpgradeCmd)

	cmd.SetArgs([]string{projectDir, filepath.Join(recipesDir, recipe)})

	if err := cmd.Execute(); err != nil {
		return ctx, err
	}

	ctx = context.WithValue(ctx, cmdStdOutCtxKey{}, cmdStdOut.String())
	ctx = context.WithValue(ctx, cmdStdErrCtxKey{}, cmdStdErr.String())

	return ctx, nil
}

func conflictsAreReported(ctx context.Context) (context.Context, error) {
	cmdStdOut := ctx.Value(cmdStdOutCtxKey{}).(string)
	cmdStdErr := ctx.Value(cmdStdErrCtxKey{}).(string)
	if matched, _ := regexp.MatchString("modified", cmdStdOut); matched {
		return ctx, nil
	}
	return ctx, fmt.Errorf("Expecting conflicts in recipe but none reported\nstdout:\n%s\n\nstderr:\n%s\n", cmdStdOut, cmdStdErr)
}

func noConflictsWereReported(ctx context.Context) (context.Context, error) {
	cmdStdOut := ctx.Value(cmdStdOutCtxKey{}).(string)
	cmdStdErr := ctx.Value(cmdStdErrCtxKey{}).(string)
	if matched, _ := regexp.MatchString("modified", cmdStdOut); matched {
		return ctx, fmt.Errorf("Conflict in recipe\nstdout:\n%s\n\nstderr:\n%s\n", cmdStdOut, cmdStdErr)
	}
	return ctx, nil
}

func iChangeProjectFileToContain(ctx context.Context, filename, content string) (context.Context, error) {
	projectDir := ctx.Value(projectDirectoryPathCtxKey{}).(string)
	if err := os.WriteFile(filepath.Join(projectDir, filename), []byte(content), 0644); err != nil {
		return ctx, err
	}
	return ctx, nil
}

func iChangeRecipeTemplateToRender(ctx context.Context, recipeName, filename, content string) (context.Context, error) {
	recipesDir := ctx.Value(recipesDirectoryPathCtxKey{}).(string)
	templateFilePath := filepath.Join(recipesDir, recipeName, "templates", filename)
	if err := os.WriteFile(templateFilePath, []byte(content), 0644); err != nil {
		return ctx, err
	}
	return ctx, nil
}

func iChangeRecipeToVersion(ctx context.Context, recipeName, version string) (context.Context, error) {
	recipesDir := ctx.Value(recipesDirectoryPathCtxKey{}).(string)
	recipeFile := filepath.Join(recipesDir, recipeName, re.RecipeFileName+re.YAMLExtension)
	recipeData, err := os.ReadFile(recipeFile)
	if err != nil {
		return ctx, err
	}

	newData := strings.Replace(string(recipeData), "v0.0.1", version, 1)

	if err := os.WriteFile(filepath.Join(recipesDir, recipeName, re.RecipeFileName+re.YAMLExtension), []byte(newData), 0644); err != nil {
		return ctx, err
	}

	return ctx, nil
}