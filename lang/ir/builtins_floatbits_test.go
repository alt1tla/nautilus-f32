package ir

import "testing"

func TestFloatBitFunctionsRoundTripFloat32(t *testing.T) {
	ftb := Builtins["FTB"]
	btf := Builtins["BTF"]
	want := float32(-13.625)
	bits := make([]Value, 32)
	for index := range bits {
		bit, err := ftb.Fn([]Value{RealVal(float64(want)), IntVal(int64(index))})
		if err != nil {
			t.Fatalf("FTB bit %d: %v", index, err)
		}
		bits[index] = bit
	}
	got, err := btf.Fn(bits)
	if err != nil {
		t.Fatal(err)
	}
	if float32(got.F) != want {
		t.Fatalf("BTF(FTB(%v)) = %v", want, got.F)
	}
}

func TestFTBRejectsOutOfRangeIndex(t *testing.T) {
	_, err := Builtins["FTB"].Fn([]Value{RealVal(1), IntVal(32)})
	if err == nil {
		t.Fatal("FTB accepted bit index 32")
	}
}
