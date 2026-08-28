package ir

import "fmt"

// Target/compiler intrinsics are valid ST calls that an embedding compiler
// must lower for its own environment. The inherited VM deliberately does not
// emulate them.
func init() {
	RegisterBuiltin(BuiltinSig{
		Name:   "VAL",
		Params: []*Type{StringT},
		Result: RealT,
		Fn: func(_ []Value) (Value, error) {
			return Value{}, fmt.Errorf("VAL is a target intrinsic and must be handled by the embedding compiler")
		},
	})
}
