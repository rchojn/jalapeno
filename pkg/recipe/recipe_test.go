package recipe

import (
	"html/template"
	"strings"
	"testing"

	"github.com/Masterminds/sprig"
)

type TestRenderEngine struct{}

var _ RenderEngine = TestRenderEngine{}

func (e TestRenderEngine) Render(templates map[string][]byte, values map[string]interface{}) (map[string][]byte, error) {
	rendered := make(map[string][]byte)

	for name, data := range templates {
		t := template.New("gotpl")
		t.Funcs(sprig.TxtFuncMap())

		if _, err := t.New(name).Parse(string(data)); err != nil {
			return nil, err
		}
		var buf strings.Builder
		if err := t.ExecuteTemplate(&buf, name, values); err != nil {
			return nil, err
		}
		rendered[name] = []byte(buf.String())
	}

	return rendered, nil
}

func TestRecipeRenderChecksums(t *testing.T) {
	recipe := NewRecipe()
	recipe.Metadata.Name = "test"
	recipe.Metadata.Version = "v1.0.0"
	recipe.Variables = []Variable{{Name: "foo"}}
	recipe.Templates = map[string][]byte{
		"README.md": []byte("{{ .Variables.foo }}"),
	}

	recipe.SetEngine(TestRenderEngine{})

	sauce, err := recipe.Execute(VariableValues{"foo": "bar"})
	if err != nil {
		t.Fatalf("Failed to render recipe: %s", err)
	}

	readme := sauce.Files["README.md"]
	sumWithAlgo := "sha256:fcde2b2edba56bf408601fb721fe9b5c338d10ee429ea04fae5511b68fbf8fb9"
	if readme.Checksum != sumWithAlgo {
		t.Fatalf("Expected checksum %s for content %s to match %s", readme.Content, readme.Checksum, sumWithAlgo)
	}
}

func TestRecipeRenderAnchor(t *testing.T) {
	recipe := NewRecipe()
	recipe.Metadata.Name = "test"
	recipe.Metadata.Version = "v1.0.0"

	recipe.SetEngine(TestRenderEngine{})

	sauce, err := recipe.Execute(nil)
	if err != nil {
		t.Fatalf("Failed to render recipe: %s", err)
	}

	if sauce.Anchor.IsNil() {
		t.Fatal("Sauce anchor was empty")
	}

}

func TestRecipeRenderEmptyFiles(t *testing.T) {
	recipe := NewRecipe()

	recipe.Metadata.Name = "test"
	recipe.Metadata.Version = "v1.0.0"
	recipe.Variables = []Variable{{Name: "foo"}}
	recipe.Templates = map[string][]byte{
		"empty-file":                           []byte(""),
		"empty-file-with-spaces":               []byte(" "),
		"empty-file-with-tabulator":            []byte("\t"),
		"empty-file-with-spaces-and-newline-1": []byte(" \n"),
		"empty-file-with-spaces-and-newline-2": []byte(" \n "),
		"file-with-empty-variable":             []byte(" {{ .Variables.foo }} "),
	}

	recipe.SetEngine(TestRenderEngine{})

	sauce, err := recipe.Execute(VariableValues{"foo": ""})
	if err != nil {
		t.Fatalf("Failed to render recipe: %s", err)
	}

	if len(sauce.Files) > 0 {
		failingFiles := make([]string, len(sauce.Files))

		i := 0
		for k := range sauce.Files {
			failingFiles[i] = k
			i++
		}
		t.Fatalf("Rendered templates contains empty files, which exist on the output. Failing files: %s", failingFiles)
	}
}