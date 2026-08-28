package stanalysis

import (
	"sort"

	"github.com/alt1tla/nautilus-f32/lang/ir"
)

// LanguageCatalog is a JSON-friendly snapshot of every callable construct
// recognized by the ST front end. A WebSocket handler can return it directly
// to build palettes, completion lists, and client-side arity checks.
type LanguageCatalog struct {
	Operators      []OperatorInfo      `json:"operators"`
	Functions      []FunctionInfo      `json:"functions"`
	FunctionBlocks []FunctionBlockInfo `json:"functionBlocks"`
}

// OperatorInfo describes an ST operator. MaxArgs is -1 for extensible forms.
type OperatorInfo struct {
	Name    string `json:"name"`
	Token   string `json:"token"`
	MinArgs int    `json:"minArgs"`
	MaxArgs int    `json:"maxArgs"`
}

// FunctionInfo describes a stateless function registered in ir.Builtins.
// A parameter or result named ANY/DYNAMIC is resolved from the call operands
// by semantic analysis.
type FunctionInfo struct {
	Name       string   `json:"name"`
	Parameters []string `json:"parameters"`
	Result     string   `json:"result"`
	MinArgs    int      `json:"minArgs"`
	MaxArgs    int      `json:"maxArgs"`
	Variadic   bool     `json:"variadic"`
}

// FunctionBlockInfo describes a stateful FB type registered in ir.FBs.
type FunctionBlockInfo struct {
	Name    string     `json:"name"`
	Inputs  []SlotInfo `json:"inputs"`
	Outputs []SlotInfo `json:"outputs"`
}

// SlotInfo is one public input or output pin on a function block.
type SlotInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// BuildCatalog returns the current ST language catalog. Results are sorted by
// name so JSON responses and frontend caches remain deterministic. The
// function reads the registries on every call, so functions or FBs registered
// by an embedding application are included automatically.
func BuildCatalog() LanguageCatalog {
	catalog := LanguageCatalog{Operators: operatorCatalog()}

	functionNames := make([]string, 0, len(ir.Builtins))
	for name := range ir.Builtins {
		functionNames = append(functionNames, name)
	}
	sort.Strings(functionNames)
	for _, name := range functionNames {
		sig := ir.Builtins[name]
		parameters := make([]string, len(sig.Params))
		for i, parameter := range sig.Params {
			parameters[i] = typeName(parameter, "ANY")
		}
		minArgs, maxArgs := functionArity(name, sig)
		catalog.Functions = append(catalog.Functions, FunctionInfo{
			Name: name, Parameters: parameters,
			Result:  functionResult(name, sig),
			MinArgs: minArgs, MaxArgs: maxArgs, Variadic: sig.Variadic,
		})
	}

	fbNames := make([]string, 0, len(ir.FBs))
	for name := range ir.FBs {
		fbNames = append(fbNames, name)
	}
	sort.Strings(fbNames)
	for _, name := range fbNames {
		def := ir.FBs[name]
		info := FunctionBlockInfo{Name: name}
		for _, input := range def.Inputs {
			info.Inputs = append(info.Inputs, SlotInfo{Name: input.Name, Type: typeName(input.Type, "ANY")})
		}
		for _, output := range def.Outputs {
			info.Outputs = append(info.Outputs, SlotInfo{Name: output.Name, Type: typeName(output.Type, "ANY")})
		}
		catalog.FunctionBlocks = append(catalog.FunctionBlocks, info)
	}
	return catalog
}

func functionResult(name string, sig ir.BuiltinSig) string {
	if name == "BTF" {
		return "REAL"
	}
	return typeName(sig.Result, "DYNAMIC")
}

func functionArity(name string, sig ir.BuiltinSig) (int, int) {
	minArgs, maxArgs := len(sig.Params), len(sig.Params)
	if sig.Variadic {
		maxArgs = -1
	}
	// These functions have constraints stricter than BuiltinSig's generic
	// "last parameter repeats" representation can express.
	switch name {
	case "MUX":
		minArgs = 3
	case "BTF":
		minArgs, maxArgs = 1, 32
	}
	return minArgs, maxArgs
}

func typeName(t *ir.Type, fallback string) string {
	if t == nil {
		return fallback
	}
	return t.String()
}

func operatorCatalog() []OperatorInfo {
	return []OperatorInfo{
		{Name: "ASSIGN", Token: ":=", MinArgs: 2, MaxArgs: 2},
		{Name: "OR", Token: "OR", MinArgs: 2, MaxArgs: 2},
		{Name: "XOR", Token: "XOR", MinArgs: 2, MaxArgs: 2},
		{Name: "AND", Token: "AND", MinArgs: 2, MaxArgs: 2},
		{Name: "EQ", Token: "=", MinArgs: 2, MaxArgs: 2},
		{Name: "NE", Token: "<>", MinArgs: 2, MaxArgs: 2},
		{Name: "GT", Token: ">", MinArgs: 2, MaxArgs: 2},
		{Name: "GE", Token: ">=", MinArgs: 2, MaxArgs: 2},
		{Name: "LT", Token: "<", MinArgs: 2, MaxArgs: 2},
		{Name: "LE", Token: "<=", MinArgs: 2, MaxArgs: 2},
		{Name: "ADD", Token: "+", MinArgs: 2, MaxArgs: 2},
		{Name: "SUB", Token: "-", MinArgs: 2, MaxArgs: 2},
		{Name: "MUL", Token: "*", MinArgs: 2, MaxArgs: 2},
		{Name: "DIV", Token: "/", MinArgs: 2, MaxArgs: 2},
		{Name: "MOD", Token: "MOD", MinArgs: 2, MaxArgs: 2},
		{Name: "NEG", Token: "-", MinArgs: 1, MaxArgs: 1},
		{Name: "NOT", Token: "NOT", MinArgs: 1, MaxArgs: 1},
	}
}
