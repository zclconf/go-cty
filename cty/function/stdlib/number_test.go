package stdlib

import (
	"fmt"
	"math"
	"math/big"
	"testing"

	"github.com/zclconf/go-cty/cty"
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
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Number),
		},
		{
			cty.DynamicVal,
			cty.UnknownVal(cty.Number),
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
		{
			cty.MustParseNumberVal("99999999999999999999999999999999999999999999999999998.123"),
			cty.MustParseNumberVal("99999999999999999999999999999999999999999999999999999"),
			false,
		},
		{
			cty.MustParseNumberVal("-99999999999999999999999999999999999999999999999999998.123"),
			cty.MustParseNumberVal("-99999999999999999999999999999999999999999999999999998"),
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
		{
			cty.MustParseNumberVal("99999999999999999999999999999999999999999999999999999.123"),
			cty.MustParseNumberVal("99999999999999999999999999999999999999999999999999999"),
			false,
		},
		{
			cty.MustParseNumberVal("-99999999999999999999999999999999999999999999999999998.123"),
			cty.MustParseNumberVal("-99999999999999999999999999999999999999999999999999999"),
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
