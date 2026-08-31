package stanalysis

import (
	"os"
	"strings"
	"testing"
)

func TestAnalyzeAllSupportedFeatures(t *testing.T) {
	sourceBytes, err := os.ReadFile("testdata/all_supported_features.st")
	if err != nil {
		t.Fatal(err)
	}
	source := string(sourceBytes)

	result := Analyze(source, Options{})
	if !result.Valid() {
		for _, diagnostic := range result.Diagnostics {
			t.Errorf("%s diagnostic at %d:%d: %s", diagnostic.Stage,
				diagnostic.Pos.Line, diagnostic.Pos.Col, diagnostic.Message)
		}
		t.FailNow()
	}
	if result.AST == nil || result.IR == nil {
		t.Fatal("Analyze() did not produce both AST and IR")
	}

	// Keep this fixture exhaustive as the public catalog grows.
	catalog := BuildCatalog()
	upperSource := strings.ToUpper(source)
	for _, function := range catalog.Functions {
		if !strings.Contains(upperSource, function.Name+"(") {
			t.Errorf("fixture does not call catalog function %s", function.Name)
		}
	}
	for _, block := range catalog.FunctionBlocks {
		if !strings.Contains(upperSource, ": "+block.Name+";") {
			t.Errorf("fixture does not declare function block %s", block.Name)
		}
	}
}
