package ir

import "testing"

func TestVALSignatureAndRuntimeGuard(t *testing.T) {
	sig, ok := Builtins["VAL"]
	if !ok {
		t.Fatal("VAL is not registered")
	}
	if len(sig.Params) != 1 || sig.Params[0] != StringT || sig.Result != RealT {
		t.Fatalf("VAL signature = %#v", sig)
	}
	if _, err := sig.Fn([]Value{StringVal("Controller1.Temperature")}); err == nil {
		t.Fatal("VAL unexpectedly executed in the inherited VM")
	}
}
