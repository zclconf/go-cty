package stdlib

import (
	"fmt"
	"math"
	"math/big"
	"testing"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function/functest"
)

func TestAbsolute(t *testing.T) {
	tests := []struct {
		Input cty.Value
		Want  cty.Value
	}{
		{
			cty.NumberIntVal(15),
			cty.NumberIntVal(15),
		},
		{
			cty.NumberIntVal(-15),
			cty.NumberIntVal(15),
		},
		{
			cty.NumberIntVal(0),
			cty.NumberIntVal(0),
		},
		{
			cty.PositiveInfinity,
			cty.PositiveInfinity,
		},
		{
			cty.NegativeInfinity,
			cty.PositiveInfinity,
		},
		{
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Number),
		},
		{
			cty.DynamicVal,
			cty.UnknownVal(cty.Number),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Absolute(%#v)", test.Input), func(t *testing.T) {
			got, err := Absolute(test.Input)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}

	// Property-based tests against randomly-selected inputs
	t.Run(
		"succeeds for all numbers, including unknown numbers",
		functest.TestSuccessfulType(
			functest.GenFixedArgs(functest.GenNumbers().MaybeInfinity().MaybeAnnotated().MaybeDynamicVal()),
			cty.Number,
			AbsoluteFunc.Call,
		).Run,
	)
	t.Run(
		"result is always positive or zero",
		functest.Test(
			functest.GenFixedArgs(
				functest.GenNumbers().MaybeInfinity(),
			),
			func(args []cty.Value) bool {
				v, err := AbsoluteFunc.Call(args)
				return err == nil && v.AsBigFloat().Sign() >= 0
			},
		).Run,
	)
}

func TestAdd(t *testing.T) {
	tests := []struct {
		A    cty.Value
		B    cty.Value
		Want cty.Value
	}{
		{
			cty.NumberIntVal(1),
			cty.NumberIntVal(2),
			cty.NumberIntVal(3),
		},
		{
			cty.NumberIntVal(1),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Number),
		},
		{
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Number),
		},
		{
			cty.NumberIntVal(1),
			cty.DynamicVal,
			cty.UnknownVal(cty.Number),
		},
		{
			cty.DynamicVal,
			cty.DynamicVal,
			cty.UnknownVal(cty.Number),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Add(%#v,%#v)", test.A, test.B), func(t *testing.T) {
			got, err := Add(test.A, test.B)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}

	// Property-based tests against randomly-selected inputs
	t.Run(
		"succeeds for all numbers, including unknown numbers",
		functest.TestSuccessfulType(
			functest.GenFixedArgs(
				// We only allow one operand to be an infinity, because
				// it's not allowed to add a negative infinity to a positive
				// infinity.
				functest.GenNumbers().MaybeInfinity().MaybeAnnotated().MaybeDynamicVal(),
				functest.GenNumbers().MaybeAnnotated().MaybeDynamicVal(),
			),
			cty.Number,
			AddFunc.Call,
		).Run,
	)
	t.Run(
		"zero is the additive identity",
		functest.Test(
			functest.GenFixedArgs(
				functest.GenNumbers().MaybeInfinity(),
				functest.GenConstant(cty.Zero),
			),
			func(args []cty.Value) bool {
				v, err := AddFunc.Call(args)
				return err == nil && v.RawEquals(args[0])
			},
		).Run,
	)
	t.Run(
		"addition is commutative aside from infinities",
		functest.TestCommutative(functest.GenNumbers(), Add).Run,
	)
	t.Run(
		"addition is associative within reasonable precision bounds, aside from infinities",
		functest.TestAssociative(functest.GenNumbers(), Add).Run,
	)
	t.Run(
		"subtraction is the inverse of addition, aside from infinities",
		functest.Test(
			functest.GenFixedArgs(
				functest.GenNumbers(),
				functest.GenNumbers(),
			),
			func(args []cty.Value) bool {
				v1, err := Add(args[0], args[1])
				if err != nil {
					return false
				}

				v2, err := Subtract(v1, args[1])
				return err == nil && v2.RawEquals(args[0])
			},
		).Run,
	)
}

func TestSubtract(t *testing.T) {
	tests := []struct {
		A    cty.Value
		B    cty.Value
		Want cty.Value
	}{
		{
			cty.NumberIntVal(1),
			cty.NumberIntVal(2),
			cty.NumberIntVal(-1),
		},
		{
			cty.NumberIntVal(1),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Number),
		},
		{
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Number),
		},
		{
			cty.NumberIntVal(1),
			cty.DynamicVal,
			cty.UnknownVal(cty.Number),
		},
		{
			cty.DynamicVal,
			cty.DynamicVal,
			cty.UnknownVal(cty.Number),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Subtract(%#v,%#v)", test.A, test.B), func(t *testing.T) {
			got, err := Subtract(test.A, test.B)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}

	// Property-based tests against randomly-selected inputs
	t.Run(
		"succeeds for all numbers, including unknown numbers",
		functest.TestSuccessfulType(
			functest.GenFixedArgs(
				// We only allow one operand to be an infinity, because
				// it's not allowed to subtract an negative infinity from
				// an infinity.
				functest.GenNumbers().MaybeInfinity().MaybeAnnotated().MaybeDynamicVal(),
				functest.GenNumbers().MaybeAnnotated().MaybeDynamicVal(),
			),
			cty.Number,
			SubtractFunc.Call,
		).Run,
	)
	t.Run(
		"zero is the subtractive identity",
		functest.Test(
			functest.GenFixedArgs(
				functest.GenNumbers().MaybeInfinity(),
				functest.GenConstant(cty.Zero),
			),
			func(args []cty.Value) bool {
				v, err := SubtractFunc.Call(args)
				return err == nil && v.RawEquals(args[0])
			},
		).Run,
	)
	t.Run(
		"subtracting from zero is the same as negating",
		functest.Test(
			functest.GenFixedArgs(
				functest.GenConstant(cty.Zero),
				functest.GenNumbers().MaybeInfinity(),
			),
			func(args []cty.Value) bool {
				v1, err1 := SubtractFunc.Call(args)
				v2, err2 := Negate(args[1])
				return err1 == nil && err2 == nil && v1.RawEquals(v2)
			},
		).Run,
	)
	t.Run(
		"subtracting from zero is the same as negating",
		functest.Test(
			functest.GenFixedArgs(
				functest.GenConstant(cty.PositiveInfinity),
				functest.GenNumbers(),
			),
			func(args []cty.Value) bool {
				v, err := SubtractFunc.Call(args)
				return err == nil && v.RawEquals(cty.PositiveInfinity)
			},
		).Run,
	)
}

func TestMultiply(t *testing.T) {
	tests := []struct {
		A    cty.Value
		B    cty.Value
		Want cty.Value
	}{
		{
			cty.NumberIntVal(5),
			cty.NumberIntVal(2),
			cty.NumberIntVal(10),
		},
		{
			cty.NumberIntVal(1),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Number),
		},
		{
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Number),
		},
		{
			cty.NumberIntVal(1),
			cty.DynamicVal,
			cty.UnknownVal(cty.Number),
		},
		{
			cty.DynamicVal,
			cty.DynamicVal,
			cty.UnknownVal(cty.Number),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Multiply(%#v,%#v)", test.A, test.B), func(t *testing.T) {
			got, err := Multiply(test.A, test.B)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}

	// Property-based tests against randomly-selected inputs
	t.Run(
		"succeeds for all numbers, including unknown numbers",
		functest.TestSuccessfulType(
			functest.GenFixedArgs(
				// We only allow one operand to be an infinity, because
				// it's not allowed to add a negative infinity to a positive
				// infinity.
				functest.GenNumbers().MaybeInfinity().MaybeAnnotated().MaybeDynamicVal(),
				functest.GenNumbers().MaybeAnnotated().MaybeDynamicVal(),
			),
			cty.Number,
			MultiplyFunc.Call,
		).Run,
	)
	t.Run(
		"one is the multiplicative identity",
		functest.Test(
			functest.GenFixedArgs(
				functest.GenNumbers().MaybeInfinity(),
				functest.GenConstant(cty.NumberIntVal(1)),
			),
			func(args []cty.Value) bool {
				v, err := MultiplyFunc.Call(args)
				return err == nil && v.RawEquals(args[0])
			},
		).Run,
	)
	t.Run(
		"multiplication is commutative aside from infinities",
		functest.TestCommutative(functest.GenNumbers(), Multiply).Run,
	)
	t.Run(
		"multiplication is associative within reasonable precision bounds, aside from infinities",
		functest.TestAssociative(functest.GenNumbers(), Multiply).Run,
	)

}

func TestDivide(t *testing.T) {
	tests := []struct {
		A    cty.Value
		B    cty.Value
		Want cty.Value
	}{
		{
			cty.NumberIntVal(5),
			cty.NumberIntVal(2),
			cty.NumberFloatVal(2.5),
		},
		{
			cty.NumberIntVal(5),
			cty.NumberIntVal(0),
			cty.PositiveInfinity,
		},
		{
			cty.NumberIntVal(-5),
			cty.NumberIntVal(0),
			cty.NegativeInfinity,
		},
		{
			cty.NumberIntVal(1),
			cty.PositiveInfinity,
			cty.Zero,
		},
		{
			cty.NumberIntVal(1),
			cty.NegativeInfinity,
			cty.Zero,
		},
		{
			cty.NumberIntVal(1),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Number),
		},
		{
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Number),
		},
		{
			cty.NumberIntVal(1),
			cty.DynamicVal,
			cty.UnknownVal(cty.Number),
		},
		{
			cty.DynamicVal,
			cty.DynamicVal,
			cty.UnknownVal(cty.Number),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Divide(%#v,%#v)", test.A, test.B), func(t *testing.T) {
			got, err := Divide(test.A, test.B)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}

	// Property-based tests against randomly-selected inputs
	t.Run(
		"succeeds for all numbers, including unknown numbers, aside from division by zero",
		functest.TestSuccessfulType(
			functest.GenFixedArgs(
				functest.GenNumbers().MaybeInfinity().MaybeAnnotated().MaybeDynamicVal(),
				functest.GenNumbers().Where(func(v cty.Value) bool { return v != cty.Zero }).MaybeAnnotated().MaybeDynamicVal(),
			),
			cty.Number,
			DivideFunc.Call,
		).Run,
	)
}

func TestModulo(t *testing.T) {
	tests := []struct {
		A    cty.Value
		B    cty.Value
		Want cty.Value
	}{
		{
			cty.NumberIntVal(15),
			cty.NumberIntVal(10),
			cty.NumberIntVal(5),
		},
		{
			cty.NumberIntVal(0),
			cty.NumberIntVal(0),
			cty.NumberIntVal(0),
		},
		{
			cty.PositiveInfinity,
			cty.NumberIntVal(1),
			cty.PositiveInfinity,
		},
		{
			cty.NegativeInfinity,
			cty.NumberIntVal(1),
			cty.NegativeInfinity,
		},
		{
			cty.NumberIntVal(1),
			cty.PositiveInfinity,
			cty.PositiveInfinity,
		},
		{
			cty.NumberIntVal(1),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Number),
		},
		{
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Number),
		},
		{
			cty.NumberIntVal(1),
			cty.DynamicVal,
			cty.UnknownVal(cty.Number),
		},
		{
			cty.DynamicVal,
			cty.DynamicVal,
			cty.UnknownVal(cty.Number),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Modulo(%#v,%#v)", test.A, test.B), func(t *testing.T) {
			got, err := Modulo(test.A, test.B)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}

	// Property-based tests against randomly-selected inputs
	t.Run(
		"succeeds for all numbers, including unknown numbers, aside from modulo zero",
		functest.TestSuccessfulType(
			functest.GenFixedArgs(
				functest.GenNumbers().MaybeInfinity().MaybeAnnotated().MaybeDynamicVal(),
				functest.GenNumbers().Where(func(v cty.Value) bool { return v != cty.Zero }).MaybeAnnotated().MaybeDynamicVal(),
			),
			cty.Number,
			ModuloFunc.Call,
		).Run,
	)
}

func TestNegate(t *testing.T) {
	tests := []struct {
		Input cty.Value
		Want  cty.Value
	}{
		{
			cty.NumberIntVal(15),
			cty.NumberIntVal(-15),
		},
		{
			cty.NumberIntVal(15).Mark("blorp"),
			cty.NumberIntVal(-15).Mark("blorp"),
		},
		{
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Number),
		},
		{
			cty.DynamicVal,
			cty.UnknownVal(cty.Number),
		},
		{
			cty.UnknownVal(cty.Number).Mark("blorp"),
			cty.UnknownVal(cty.Number).Mark("blorp"),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Negate(%#v)", test.Input), func(t *testing.T) {
			got, err := Negate(test.Input)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}

	// Property-based tests against randomly-selected inputs
	t.Run(
		"succeeds for all numbers, including unknown numbers",
		functest.TestSuccessfulType(
			functest.GenFixedArgs(
				functest.GenNumbers().MaybeInfinity().MaybeAnnotated().MaybeDynamicVal(),
			),
			cty.Number,
			NegateFunc.Call,
		).Run,
	)
	t.Run(
		"negate is its own inverse, aside from DynamicVal",
		functest.TestInverse(
			functest.GenNumbers().MaybeInfinity().MaybeAnnotated(),
			Negate, Negate,
		).Run,
	)

}

func TestLessThan(t *testing.T) {
	tests := []struct {
		A    cty.Value
		B    cty.Value
		Want cty.Value
	}{
		{
			cty.NumberIntVal(1),
			cty.NumberIntVal(2),
			cty.True,
		},
		{
			cty.NumberIntVal(2),
			cty.NumberIntVal(1),
			cty.False,
		},
		{
			cty.NumberIntVal(2),
			cty.NumberIntVal(2),
			cty.False,
		},
		{
			cty.NumberIntVal(1),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Bool),
		},
		{
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Bool),
		},
		{
			cty.NumberIntVal(1),
			cty.DynamicVal,
			cty.UnknownVal(cty.Bool),
		},
		{
			cty.DynamicVal,
			cty.DynamicVal,
			cty.UnknownVal(cty.Bool),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("LessThan(%#v,%#v)", test.A, test.B), func(t *testing.T) {
			got, err := LessThan(test.A, test.B)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}

	// Property-based tests against randomly-selected inputs
	t.Run(
		"succeeds for all numbers, including unknown numbers",
		functest.TestSuccessfulType(
			functest.GenFixedArgs(
				functest.GenNumbers().MaybeInfinity().MaybeAnnotated().MaybeDynamicVal(),
				functest.GenNumbers().MaybeInfinity().MaybeAnnotated().MaybeDynamicVal(),
			),
			cty.Bool,
			LessThanFunc.Call,
		).Run,
	)
}

func TestLessThanOrEqualTo(t *testing.T) {
	tests := []struct {
		A    cty.Value
		B    cty.Value
		Want cty.Value
	}{
		{
			cty.NumberIntVal(1),
			cty.NumberIntVal(2),
			cty.True,
		},
		{
			cty.NumberIntVal(2),
			cty.NumberIntVal(1),
			cty.False,
		},
		{
			cty.NumberIntVal(2),
			cty.NumberIntVal(2),
			cty.True,
		},
		{
			cty.NumberIntVal(1),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Bool),
		},
		{
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Bool),
		},
		{
			cty.NumberIntVal(1),
			cty.DynamicVal,
			cty.UnknownVal(cty.Bool),
		},
		{
			cty.DynamicVal,
			cty.DynamicVal,
			cty.UnknownVal(cty.Bool),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("LessThanOrEqualTo(%#v,%#v)", test.A, test.B), func(t *testing.T) {
			got, err := LessThanOrEqualTo(test.A, test.B)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}

	// Property-based tests against randomly-selected inputs
	t.Run(
		"succeeds for all numbers, including unknown numbers",
		functest.TestSuccessfulType(
			functest.GenFixedArgs(
				functest.GenNumbers().MaybeInfinity().MaybeAnnotated().MaybeDynamicVal(),
				functest.GenNumbers().MaybeInfinity().MaybeAnnotated().MaybeDynamicVal(),
			),
			cty.Bool,
			LessThanOrEqualToFunc.Call,
		).Run,
	)
}

func TestGreaterThan(t *testing.T) {
	tests := []struct {
		A    cty.Value
		B    cty.Value
		Want cty.Value
	}{
		{
			cty.NumberIntVal(1),
			cty.NumberIntVal(2),
			cty.False,
		},
		{
			cty.NumberIntVal(2),
			cty.NumberIntVal(1),
			cty.True,
		},
		{
			cty.NumberIntVal(2),
			cty.NumberIntVal(2),
			cty.False,
		},
		{
			cty.NumberIntVal(1),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Bool),
		},
		{
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Bool),
		},
		{
			cty.NumberIntVal(1),
			cty.DynamicVal,
			cty.UnknownVal(cty.Bool),
		},
		{
			cty.DynamicVal,
			cty.DynamicVal,
			cty.UnknownVal(cty.Bool),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("GreaterThan(%#v,%#v)", test.A, test.B), func(t *testing.T) {
			got, err := GreaterThan(test.A, test.B)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}

	// Property-based tests against randomly-selected inputs
	t.Run(
		"succeeds for all numbers, including unknown numbers",
		functest.TestSuccessfulType(
			functest.GenFixedArgs(
				functest.GenNumbers().MaybeInfinity().MaybeAnnotated().MaybeDynamicVal(),
				functest.GenNumbers().MaybeInfinity().MaybeAnnotated().MaybeDynamicVal(),
			),
			cty.Bool,
			GreaterThanFunc.Call,
		).Run,
	)
}

func TestGreaterThanOrEqualTo(t *testing.T) {
	tests := []struct {
		A    cty.Value
		B    cty.Value
		Want cty.Value
	}{
		{
			cty.NumberIntVal(1),
			cty.NumberIntVal(2),
			cty.False,
		},
		{
			cty.NumberIntVal(2),
			cty.NumberIntVal(1),
			cty.True,
		},
		{
			cty.NumberIntVal(2),
			cty.NumberIntVal(2),
			cty.True,
		},
		{
			cty.NumberIntVal(1),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Bool),
		},
		{
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Bool),
		},
		{
			cty.NumberIntVal(1),
			cty.DynamicVal,
			cty.UnknownVal(cty.Bool),
		},
		{
			cty.DynamicVal,
			cty.DynamicVal,
			cty.UnknownVal(cty.Bool),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("GreaterThanOrEqualTo(%#v,%#v)", test.A, test.B), func(t *testing.T) {
			got, err := GreaterThanOrEqualTo(test.A, test.B)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}

	// Property-based tests against randomly-selected inputs
	t.Run(
		"succeeds for all numbers, including unknown numbers",
		functest.TestSuccessfulType(
			functest.GenFixedArgs(
				functest.GenNumbers().MaybeInfinity().MaybeAnnotated().MaybeDynamicVal(),
				functest.GenNumbers().MaybeInfinity().MaybeAnnotated().MaybeDynamicVal(),
			),
			cty.Bool,
			GreaterThanOrEqualToFunc.Call,
		).Run,
	)
}

func TestMin(t *testing.T) {
	tests := []struct {
		Inputs []cty.Value
		Want   cty.Value
	}{
		{
			[]cty.Value{cty.NumberIntVal(0)},
			cty.NumberIntVal(0),
		},
		{
			[]cty.Value{cty.NumberIntVal(-12)},
			cty.NumberIntVal(-12),
		},
		{
			[]cty.Value{cty.NumberIntVal(12)},
			cty.NumberIntVal(12),
		},
		{
			[]cty.Value{cty.NumberIntVal(-12), cty.NumberIntVal(0), cty.NumberIntVal(2)},
			cty.NumberIntVal(-12),
		},
		{
			[]cty.Value{cty.NegativeInfinity, cty.NumberIntVal(0)},
			cty.NegativeInfinity,
		},
		{
			[]cty.Value{cty.PositiveInfinity, cty.NumberIntVal(0)},
			cty.NumberIntVal(0),
		},
		{
			[]cty.Value{cty.NegativeInfinity},
			cty.NegativeInfinity,
		},
		{
			[]cty.Value{cty.PositiveInfinity, cty.UnknownVal(cty.Number)},
			cty.UnknownVal(cty.Number),
		},
		{
			[]cty.Value{cty.PositiveInfinity, cty.DynamicVal},
			cty.UnknownVal(cty.Number),
		},
		{
			[]cty.Value{cty.Zero.Mark(1), cty.NumberIntVal(1)},
			cty.Zero.Mark(1),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v", test.Inputs), func(t *testing.T) {
			got, err := Min(test.Inputs...)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}

	// Property-based tests against randomly-selected inputs
	t.Run(
		"succeeds as long as it has at least one number",
		functest.TestSuccessfulType(
			functest.GenFixedArgs(
				functest.GenNumbers().MaybeInfinity().MaybeAnnotated().MaybeDynamicVal(),
				functest.GenNumbers().MaybeInfinity().MaybeAnnotated().MaybeDynamicVal(),
			),
			cty.Bool,
			GreaterThanOrEqualToFunc.Call,
		).Run,
	)
}

func TestMax(t *testing.T) {
	tests := []struct {
		Inputs []cty.Value
		Want   cty.Value
	}{
		{
			[]cty.Value{cty.NumberIntVal(0)},
			cty.NumberIntVal(0),
		},
		{
			[]cty.Value{cty.NumberIntVal(-12)},
			cty.NumberIntVal(-12),
		},
		{
			[]cty.Value{cty.NumberIntVal(12)},
			cty.NumberIntVal(12),
		},
		{
			[]cty.Value{cty.NumberIntVal(-12), cty.NumberIntVal(0), cty.NumberIntVal(2)},
			cty.NumberIntVal(2),
		},
		{
			[]cty.Value{cty.NegativeInfinity, cty.NumberIntVal(0)},
			cty.NumberIntVal(0),
		},
		{
			[]cty.Value{cty.PositiveInfinity, cty.NumberIntVal(0)},
			cty.PositiveInfinity,
		},
		{
			[]cty.Value{cty.NegativeInfinity},
			cty.NegativeInfinity,
		},
		{
			[]cty.Value{cty.PositiveInfinity, cty.UnknownVal(cty.Number)},
			cty.UnknownVal(cty.Number),
		},
		{
			[]cty.Value{cty.PositiveInfinity, cty.DynamicVal},
			cty.UnknownVal(cty.Number),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v", test.Inputs), func(t *testing.T) {
			got, err := Max(test.Inputs...)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestInt(t *testing.T) {
	tests := []struct {
		Input cty.Value
		Want  cty.Value
	}{
		{
			cty.NumberIntVal(0),
			cty.NumberIntVal(0),
		},
		{
			cty.NumberIntVal(1),
			cty.NumberIntVal(1),
		},
		{
			cty.NumberIntVal(-1),
			cty.NumberIntVal(-1),
		},
		{
			cty.NumberFloatVal(1.3),
			cty.NumberIntVal(1),
		},
		{
			cty.NumberFloatVal(-1.7),
			cty.NumberIntVal(-1),
		},
		{
			cty.NumberFloatVal(-1.3),
			cty.NumberIntVal(-1),
		},
		{
			cty.NumberFloatVal(-1.7),
			cty.NumberIntVal(-1),
		},
		{
			cty.NumberVal(mustParseFloat("999999999999999999999999999999999999999999999999999999999999.7")),
			cty.NumberVal(mustParseFloat("999999999999999999999999999999999999999999999999999999999999")),
		},
		{
			cty.NumberVal(mustParseFloat("-999999999999999999999999999999999999999999999999999999999999.7")),
			cty.NumberVal(mustParseFloat("-999999999999999999999999999999999999999999999999999999999999")),
		},
	}

	for _, test := range tests {
		t.Run(test.Input.GoString(), func(t *testing.T) {
			got, err := Int(test.Input)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func mustParseFloat(s string) *big.Float {
	ret, _, err := big.ParseFloat(s, 10, 0, big.AwayFromZero)
	if err != nil {
		panic(err)
	}
	return ret
}

func TestCeil(t *testing.T) {
	tests := []struct {
		Num  cty.Value
		Want cty.Value
		Err  bool
	}{
		{
			cty.NumberFloatVal(-1.8),
			cty.NumberFloatVal(-1),
			false,
		},
		{
			cty.NumberFloatVal(1.2),
			cty.NumberFloatVal(2),
			false,
		},
		{
			cty.NumberFloatVal(math.Inf(1)),
			cty.NumberFloatVal(math.Inf(1)),
			false,
		},
		{
			cty.NumberFloatVal(math.Inf(-1)),
			cty.NumberFloatVal(math.Inf(-1)),
			false,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("ceil(%#v)", test.Num), func(t *testing.T) {
			got, err := Ceil(test.Num)

			if test.Err {
				if err == nil {
					t.Fatal("succeeded; want error")
				}
				return
			} else if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestFloor(t *testing.T) {
	tests := []struct {
		Num  cty.Value
		Want cty.Value
		Err  bool
	}{
		{
			cty.NumberFloatVal(-1.8),
			cty.NumberFloatVal(-2),
			false,
		},
		{
			cty.NumberFloatVal(1.2),
			cty.NumberFloatVal(1),
			false,
		},
		{
			cty.NumberFloatVal(math.Inf(1)),
			cty.NumberFloatVal(math.Inf(1)),
			false,
		},
		{
			cty.NumberFloatVal(math.Inf(-1)),
			cty.NumberFloatVal(math.Inf(-1)),
			false,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("floor(%#v)", test.Num), func(t *testing.T) {
			got, err := Floor(test.Num)

			if test.Err {
				if err == nil {
					t.Fatal("succeeded; want error")
				}
				return
			} else if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestLog(t *testing.T) {
	tests := []struct {
		Num  cty.Value
		Base cty.Value
		Want cty.Value
		Err  bool
	}{
		{
			cty.NumberFloatVal(1),
			cty.NumberFloatVal(10),
			cty.NumberFloatVal(0),
			false,
		},
		{
			cty.NumberFloatVal(10),
			cty.NumberFloatVal(10),
			cty.NumberFloatVal(1),
			false,
		},

		{
			cty.NumberFloatVal(0),
			cty.NumberFloatVal(10),
			cty.NegativeInfinity,
			false,
		},
		{
			cty.NumberFloatVal(10),
			cty.NumberFloatVal(0),
			cty.NumberFloatVal(-0),
			false,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("log(%#v, %#v)", test.Num, test.Base), func(t *testing.T) {
			got, err := Log(test.Num, test.Base)

			if test.Err {
				if err == nil {
					t.Fatal("succeeded; want error")
				}
				return
			} else if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestPow(t *testing.T) {
	tests := []struct {
		Num   cty.Value
		Power cty.Value
		Want  cty.Value
		Err   bool
	}{
		{
			cty.NumberFloatVal(1),
			cty.NumberFloatVal(0),
			cty.NumberFloatVal(1),
			false,
		},
		{
			cty.NumberFloatVal(1),
			cty.NumberFloatVal(1),
			cty.NumberFloatVal(1),
			false,
		},

		{
			cty.NumberFloatVal(2),
			cty.NumberFloatVal(0),
			cty.NumberFloatVal(1),
			false,
		},
		{
			cty.NumberFloatVal(2),
			cty.NumberFloatVal(1),
			cty.NumberFloatVal(2),
			false,
		},
		{
			cty.NumberFloatVal(3),
			cty.NumberFloatVal(2),
			cty.NumberFloatVal(9),
			false,
		},
		{
			cty.NumberFloatVal(-3),
			cty.NumberFloatVal(2),
			cty.NumberFloatVal(9),
			false,
		},
		{
			cty.NumberFloatVal(2),
			cty.NumberFloatVal(-2),
			cty.NumberFloatVal(0.25),
			false,
		},
		{
			cty.NumberFloatVal(0),
			cty.NumberFloatVal(2),
			cty.NumberFloatVal(0),
			false,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("pow(%#v, %#v)", test.Num, test.Power), func(t *testing.T) {
			got, err := Pow(test.Num, test.Power)

			if test.Err {
				if err == nil {
					t.Fatal("succeeded; want error")
				}
				return
			} else if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestSignum(t *testing.T) {
	tests := []struct {
		Num  cty.Value
		Want cty.Value
		Err  bool
	}{
		{
			cty.NumberFloatVal(0),
			cty.NumberFloatVal(0),
			false,
		},
		{
			cty.NumberFloatVal(12),
			cty.NumberFloatVal(1),
			false,
		},
		{
			cty.NumberFloatVal(-29),
			cty.NumberFloatVal(-1),
			false,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("signum(%#v)", test.Num), func(t *testing.T) {
			got, err := Signum(test.Num)

			if test.Err {
				if err == nil {
					t.Fatal("succeeded; want error")
				}
				return
			} else if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		Num  cty.Value
		Base cty.Value
		Want cty.Value
		Err  bool
	}{
		{
			cty.StringVal("128"),
			cty.NumberIntVal(10),
			cty.NumberIntVal(128),
			false,
		},
		{
			cty.StringVal("-128"),
			cty.NumberIntVal(10),
			cty.NumberIntVal(-128),
			false,
		},
		{
			cty.StringVal("00128"),
			cty.NumberIntVal(10),
			cty.NumberIntVal(128),
			false,
		},
		{
			cty.StringVal("-00128"),
			cty.NumberIntVal(10),
			cty.NumberIntVal(-128),
			false,
		},
		{
			cty.StringVal("FF00"),
			cty.NumberIntVal(16),
			cty.NumberIntVal(65280),
			false,
		},
		{
			cty.StringVal("ff00"),
			cty.NumberIntVal(16),
			cty.NumberIntVal(65280),
			false,
		},
		{
			cty.StringVal("-FF00"),
			cty.NumberIntVal(16),
			cty.NumberIntVal(-65280),
			false,
		},
		{
			cty.StringVal("00FF00"),
			cty.NumberIntVal(16),
			cty.NumberIntVal(65280),
			false,
		},
		{
			cty.StringVal("-00FF00"),
			cty.NumberIntVal(16),
			cty.NumberIntVal(-65280),
			false,
		},
		{
			cty.StringVal("1011111011101111"),
			cty.NumberIntVal(2),
			cty.NumberIntVal(48879),
			false,
		},
		{
			cty.StringVal("aA"),
			cty.NumberIntVal(62),
			cty.NumberIntVal(656),
			false,
		},
		{
			cty.StringVal("Aa"),
			cty.NumberIntVal(62),
			cty.NumberIntVal(2242),
			false,
		},
		{
			cty.StringVal("999999999999999999999999999999999999999999999999999999999999"),
			cty.NumberIntVal(10),
			cty.MustParseNumberVal("999999999999999999999999999999999999999999999999999999999999"),
			false,
		},
		{
			cty.StringVal("FF"),
			cty.NumberIntVal(10),
			cty.UnknownVal(cty.Number),
			true,
		},
		{
			cty.StringVal("00FF"),
			cty.NumberIntVal(10),
			cty.UnknownVal(cty.Number),
			true,
		},
		{
			cty.StringVal("-00FF"),
			cty.NumberIntVal(10),
			cty.UnknownVal(cty.Number),
			true,
		},
		{
			cty.NumberIntVal(2),
			cty.NumberIntVal(10),
			cty.UnknownVal(cty.Number),
			true,
		},
		{
			cty.StringVal("1"),
			cty.NumberIntVal(63),
			cty.UnknownVal(cty.Number),
			true,
		},
		{
			cty.StringVal("1"),
			cty.NumberIntVal(-1),
			cty.UnknownVal(cty.Number),
			true,
		},
		{
			cty.StringVal("1"),
			cty.NumberIntVal(1),
			cty.UnknownVal(cty.Number),
			true,
		},
		{
			cty.StringVal("1"),
			cty.NumberIntVal(0),
			cty.UnknownVal(cty.Number),
			true,
		},
		{
			cty.StringVal("1.2"),
			cty.NumberIntVal(10),
			cty.UnknownVal(cty.Number),
			true,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("parseint(%#v, %#v)", test.Num, test.Base), func(t *testing.T) {
			got, err := ParseInt(test.Num, test.Base)

			if test.Err {
				if err == nil {
					t.Fatal("succeeded; want error")
				}
				return
			} else if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}
