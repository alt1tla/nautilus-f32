package stanalysis

import "testing"

func TestBuildCatalogContainsRequiredSTSurface(t *testing.T) {
	catalog := BuildCatalog()
	functions := map[string]FunctionInfo{}
	for _, function := range catalog.Functions {
		functions[function.Name] = function
	}
	for _, name := range []string{"AND", "OR", "XOR", "MOD", "MIN", "MAX", "MUX", "SEL", "LIMIT", "BTF", "FTB"} {
		if _, ok := functions[name]; !ok {
			t.Errorf("function catalog is missing %s", name)
		}
	}
	if got := functions["VAL"]; got.MinArgs != 1 || got.MaxArgs != 1 || got.Result != "REAL" || len(got.Parameters) != 1 || got.Parameters[0] != "STRING" {
		t.Errorf("VAL metadata = %#v", got)
	}
	if got := functions["FTB"]; got.MinArgs != 2 || got.MaxArgs != 2 || got.Result != "BOOL" {
		t.Errorf("FTB metadata = %#v", got)
	}
	if got := functions["BTF"]; got.MinArgs != 1 || got.MaxArgs != 32 || got.Result != "REAL" {
		t.Errorf("BTF metadata = %#v", got)
	}
	if got := functions["MUX"]; got.MinArgs != 3 || got.MaxArgs != -1 {
		t.Errorf("MUX metadata = %#v", got)
	}

	blocks := map[string]FunctionBlockInfo{}
	for _, block := range catalog.FunctionBlocks {
		blocks[block.Name] = block
	}
	for _, name := range []string{"TON", "TOF", "TP", "SR", "RS", "R_TRIG", "F_TRIG"} {
		if _, ok := blocks[name]; !ok {
			t.Errorf("function-block catalog is missing %s", name)
		}
	}
	if got := blocks["TON"]; len(got.Inputs) != 2 || got.Inputs[0].Name != "IN" || got.Inputs[1].Name != "PT" {
		t.Errorf("TON metadata = %#v", got)
	}
}

func TestBuildCatalogIsDeterministicallySorted(t *testing.T) {
	catalog := BuildCatalog()
	for i := 1; i < len(catalog.Functions); i++ {
		if catalog.Functions[i-1].Name > catalog.Functions[i].Name {
			t.Fatalf("functions are not sorted at %q, %q", catalog.Functions[i-1].Name, catalog.Functions[i].Name)
		}
	}
	for i := 1; i < len(catalog.FunctionBlocks); i++ {
		if catalog.FunctionBlocks[i-1].Name > catalog.FunctionBlocks[i].Name {
			t.Fatalf("function blocks are not sorted at %q, %q", catalog.FunctionBlocks[i-1].Name, catalog.FunctionBlocks[i].Name)
		}
	}
}
