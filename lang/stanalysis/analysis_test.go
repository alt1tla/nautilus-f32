package stanalysis

import (
	"strings"
	"testing"

	"github.com/alt1tla/nautilus-f32/lang/ir"
)

func TestAnalyzeReturnsASTIRAndSymbols(t *testing.T) {
	source := `PROGRAM Main
VAR
    Input : REAL;
    Output : REAL;
END_VAR
Output := Input + 1.0;
END_PROGRAM`

	result := Analyze(source, Options{})
	if !result.Valid() {
		t.Fatalf("Analyze() diagnostics = %#v", result.Diagnostics)
	}
	if result.AST == nil || result.IR == nil {
		t.Fatal("Analyze() did not return both AST and IR")
	}
	if len(result.Symbols) != 2 {
		t.Fatalf("symbols = %#v, want two variables", result.Symbols)
	}
	if result.IR.SlotIndex["Output"] < 0 {
		t.Fatal("IR does not contain Output")
	}
}

func TestAnalyzeReportsSyntaxDiagnostic(t *testing.T) {
	result := Analyze("PROGRAM Main\nIF THEN\nEND_PROGRAM", Options{})
	if result.Valid() || len(result.Diagnostics) != 1 {
		t.Fatalf("result = %#v, want one diagnostic", result)
	}
	if result.Diagnostics[0].Stage != StageSyntax || result.Diagnostics[0].Pos.Line != 2 {
		t.Fatalf("diagnostic = %#v", result.Diagnostics[0])
	}
	if result.AST != nil || result.IR != nil {
		t.Fatal("syntax failure unexpectedly returned AST or IR")
	}
}

func TestAnalyzeReportsSemanticDiagnosticAndKeepsAST(t *testing.T) {
	source := `PROGRAM Main
VAR Output : REAL; END_VAR
Output := Missing;
END_PROGRAM`
	result := Analyze(source, Options{})
	if result.Valid() || len(result.Diagnostics) != 1 {
		t.Fatalf("result = %#v, want one diagnostic", result)
	}
	diagnostic := result.Diagnostics[0]
	if diagnostic.Stage != StageSemantic || diagnostic.Pos.Line != 3 {
		t.Fatalf("diagnostic = %#v", diagnostic)
	}
	if !strings.Contains(diagnostic.Message, "undeclared identifier") {
		t.Fatalf("message = %q", diagnostic.Message)
	}
	if result.AST == nil || result.IR != nil {
		t.Fatal("semantic failure should retain AST but not return IR")
	}
}

func TestAnalyzeUsesImplicitGlobals(t *testing.T) {
	source := `PROGRAM Main
VAR Output : REAL; END_VAR
Output := Input;
END_PROGRAM`
	result := Analyze(source, Options{ImplicitGlobals: map[string]*ir.Type{"Input": ir.RealT}})
	if !result.Valid() {
		t.Fatalf("Analyze() diagnostics = %#v", result.Diagnostics)
	}
	if got := result.IR.Globals["Input"]; got != ir.RealT {
		t.Fatalf("implicit Input type = %v, want REAL", got)
	}
}

func TestAnalyzeWithPreludeKeepsDocumentCoordinates(t *testing.T) {
	prelude := "TYPE\n  Header : STRUCT\n    Valid : BOOL;\n  END_STRUCT;\nEND_TYPE\n"
	source := "PROGRAM Main\nVAR H : Header; END_VAR\nH.Valid := Missing;\nEND_PROGRAM\n"
	result := Analyze(source, Options{Prelude: prelude})
	if result.ContextAST == nil || result.ContextAST == result.AST {
		t.Fatal("Analyze() did not retain the combined prelude context AST")
	}
	if len(result.Diagnostics) != 1 {
		t.Fatalf("diagnostics = %#v, want one", result.Diagnostics)
	}
	if result.Diagnostics[0].Pos.Line != 3 {
		t.Fatalf("diagnostic line = %d, want document line 3", result.Diagnostics[0].Pos.Line)
	}
	for _, symbol := range result.Symbols {
		if symbol.Name == "Header" {
			t.Fatal("document symbols unexpectedly contain a prelude declaration")
		}
	}
}

func TestFloat32ScalarAcceptsNumericBoolAndTime(t *testing.T) {
	source := `PROGRAM Main
VAR
    Enable : REAL;
    Delay : REAL;
    Timer : TON;
    Running : REAL;
END_VAR
Timer(IN := Enable, PT := Delay);
Running := Timer.Q;
IF Delay THEN
    Running := TRUE;
END_IF
END_PROGRAM`

	strict := Analyze(source, Options{ScalarMode: IECStrict})
	if strict.Valid() {
		t.Fatal("strict IEC analysis unexpectedly accepted scalar type mixing")
	}

	float32Result := Analyze(source, Options{ScalarMode: Float32Scalar})
	if !float32Result.Valid() {
		t.Fatalf("float32 analysis diagnostics = %#v", float32Result.Diagnostics)
	}
}

func TestFloat32ScalarStillRejectsStringsAsControlValues(t *testing.T) {
	source := `PROGRAM Main
VAR Text : STRING; END_VAR
IF Text THEN
END_IF
END_PROGRAM`
	result := Analyze(source, Options{ScalarMode: Float32Scalar})
	if result.Valid() || len(result.Diagnostics) != 1 {
		t.Fatalf("result = %#v, want a semantic diagnostic", result)
	}
}

func TestFloat32ScalarOperatorFunctions(t *testing.T) {
	source := `PROGRAM Main
VAR
    A : REAL;
    B : REAL;
    Logic : REAL;
    Remainder : REAL;
END_VAR
Logic := AND(A, OR(B, XOR(A, B)));
Remainder := MOD(A, B);
END_PROGRAM`
	result := Analyze(source, Options{ScalarMode: Float32Scalar})
	if !result.Valid() {
		t.Fatalf("float32 operator functions diagnostics = %#v", result.Diagnostics)
	}
}

func TestDefaultModeTreatsAllNumericLiteralsAsFloat32Real(t *testing.T) {
	source := `PROGRAM Main
VAR
    a : REAL;
    b : REAL;
END_VAR
a := 10;
b := MOD(a, 5);
END_PROGRAM`
	result := Analyze(source, Options{})
	if !result.Valid() {
		t.Fatalf("default float32 analysis diagnostics = %#v", result.Diagnostics)
	}
	for _, statement := range result.IR.Body {
		assignment, ok := statement.(*ir.Assign)
		if !ok {
			continue
		}
		if literal, ok := assignment.Value.(*ir.Lit); ok && literal.ExprType() != ir.RealT {
			t.Fatalf("numeric literal type = %s, want REAL", literal.ExprType())
		}
		if call, ok := assignment.Value.(*ir.Call); ok {
			for _, argument := range call.Args {
				if literal, ok := argument.(*ir.Lit); ok && literal.ExprType() != ir.RealT {
					t.Fatalf("function literal argument type = %s, want REAL", literal.ExprType())
				}
			}
		}
	}
}

func TestFloat32ScalarBitConversionFunctions(t *testing.T) {
	source := `PROGRAM Main
VAR
    Value : REAL;
    Bit0 : REAL;
    Bit1 : REAL;
    Rebuilt : REAL;
END_VAR
Bit0 := FTB(Value, 0.0);
Bit1 := FTB(Value, 1.0);
Rebuilt := BTF(Bit0, Bit1);
END_PROGRAM`
	result := Analyze(source, Options{ScalarMode: Float32Scalar})
	if !result.Valid() {
		t.Fatalf("float32 bit conversion diagnostics = %#v", result.Diagnostics)
	}
}

func TestAnalyzeTargetVALIntrinsic(t *testing.T) {
	source := `PROGRAM Main
VAR Temperature : REAL; END_VAR
Temperature := VAL('Controller1.Temperature');
END_PROGRAM`
	result := Analyze(source, Options{ScalarMode: Float32Scalar})
	if !result.Valid() {
		t.Fatalf("VAL diagnostics = %#v", result.Diagnostics)
	}
	if result.IR == nil || len(result.IR.Body) != 1 {
		t.Fatalf("VAL did not produce IR: %#v", result.IR)
	}
	assignment, ok := result.IR.Body[0].(*ir.Assign)
	if !ok {
		t.Fatalf("VAL statement IR = %T, want *ir.Assign", result.IR.Body[0])
	}
	call, ok := assignment.Value.(*ir.Call)
	if !ok || call.Name != "VAL" || call.ExprType() != ir.RealT {
		t.Fatalf("VAL expression IR = %#v", assignment.Value)
	}
}
